package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/azhai/gopaper/models"
	"github.com/azhai/gopaper/services"

	"github.com/labstack/echo/v5"
)

type PageHandler struct {
	cache     *services.CacheVault
	renderer  *services.Renderer
	templates atomic.Pointer[template.Template]
	siteURL   string
	dev       bool
	sse       *SSEServer
	themes    *services.ThemeForge
	analytics *services.Analytics
	logger    *slog.Logger
}

func NewPageHandler(cache *services.CacheVault, renderer *services.Renderer, siteURL string) *PageHandler {
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"safeHTML": func(s string) template.HTML { return template.HTML(s) },
	}).ParseGlob("themes/default/layouts/*.html"))

	ph := &PageHandler{cache: cache, renderer: renderer, siteURL: siteURL}
	ph.templates.Store(tmpl)
	return ph
}

// SetThemes enables per-request resolution of the scheduled site theme.
func (h *PageHandler) SetThemes(themes *services.ThemeForge, logger *slog.Logger) {
	h.themes = themes
	h.logger = logger
}

// SetAnalytics enables page-view tracking for content detail pages.
func (h *PageHandler) SetAnalytics(analytics *services.Analytics) {
	h.analytics = analytics
}

func (h *PageHandler) SetDevMode(sse *SSEServer) {
	h.dev = true
	h.sse = sse
}

func (h *PageHandler) reloadTemplates() error {
	tmpl, err := template.New("").Funcs(template.FuncMap{
		"safeHTML": func(s string) template.HTML { return template.HTML(s) },
	}).ParseGlob("themes/default/layouts/*.html")
	if err != nil {
		return err
	}
	h.templates.Store(tmpl)
	return nil
}

func (h *PageHandler) RSS(c *echo.Context) error {
	siteTree, ok := h.cache.GetSiteTree()
	if !ok || siteTree == nil {
		return c.String(http.StatusServiceUnavailable, "站点数据未加载")
	}
	return c.XMLBlob(http.StatusOK, GenerateRSS(h.siteURL, siteTree))
}

func (h *PageHandler) Sitemap(c *echo.Context) error {
	siteTree, ok := h.cache.GetSiteTree()
	if !ok || siteTree == nil {
		return c.String(http.StatusServiceUnavailable, "站点数据未加载")
	}
	return c.XMLBlob(http.StatusOK, GenerateSitemap(h.siteURL, siteTree))
}

func (h *PageHandler) Robots(c *echo.Context) error {
	return c.Blob(http.StatusOK, "text/plain; charset=utf-8", GenerateRobots(h.siteURL))
}

type NavItem struct {
	Label  string
	Href   string
	Active bool
}

type ArticleView struct {
	Title   string
	Author  string
	Date    string
	Tags    []string
	Href    string
	Summary string
}

type FeatureCardView struct {
	Title string
	Desc  string
	Icon  string
	Link  string
}

type ProductSpecView struct {
	Name      string
	Price     string
	SKU       string
	JDURL     string
	TaobaoURL string
}

type ImprovementView struct {
	Date    string
	Content string
}

type ProductView struct {
	Title        string
	Href         string
	Summary      string
	Category     string
	CategoryLeaf string
	Group        string
	Cover        string
	Gallery      []string
	PriceLabel   string
	Specs        []ProductSpecView
	Improvements []ImprovementView
}

type ProductGroup struct {
	Name     string
	Path     string
	Products []ProductView
}

// CategoryNode is one node of the two-level nested set built for product filters.
type CategoryNode struct {
	Name     string
	Path     string
	Depth    int
	Left     int
	Right    int
	Count    int
	Href     string
	Active   bool
	Children []CategoryNode
}

type PageData struct {
	Title     string
	SiteTitle string
	SiteDesc  string
	NavItems  []NavItem
	Year      int

	ThemeCSS template.CSS

	HeroTitle    string
	HeroSubtitle string
	HeroImage    string
	HeroCTAText  string
	HeroCTALink  string

	FeatureTitle string
	FeatureCards []FeatureCardView

	NewsTitle string
	NewsLink  string
	Articles  []ArticleView

	ContactEmail   string
	ContactPhone   string
	ContactAddress string
	FooterText     string
	ICP            string

	PageTitle string
	PageDesc  string

	Article    *models.Article
	Content    string
	DirType    string
	Pagination *Pagination

	Products          []ProductView
	ProductGroups     []ProductGroup
	ProductCategories []CategoryNode
	ActiveCategory    string
	ClearCategoryHref string
	Product           *ProductView
}

type Pagination struct {
	Current  int
	Total    int
	HasPrev  bool
	HasNext  bool
	PrevHref string
	NextHref string
}

func (h *PageHandler) Index(c *echo.Context) error {
	siteTree, ok := h.cache.GetSiteTree()
	if !ok || siteTree == nil {
		return c.String(http.StatusInternalServerError, "站点数据未加载")
	}

	data := h.buildIndexData(siteTree)
	return h.renderTemplate(c, "index", data)
}

func (h *PageHandler) buildIndexData(siteTree *models.SiteTree) PageData {
	meta := siteTree.Meta
	allArticles := indexArticles(siteTree, meta.NEWS_LINK)
	data := h.buildBaseData(meta, "")
	data.Title = meta.SITE_TITLE

	if len(meta.FEATURES) > 0 {
		for _, f := range meta.FEATURES {
			data.FeatureCards = append(data.FeatureCards, FeatureCardView{
				Title: f.Title, Desc: f.Desc, Icon: f.Icon, Link: f.Link,
			})
		}
	}

	for _, a := range allArticles {
		if a.Position == "features" {
			desc := a.Summary
			if desc == "" {
				desc = simpleExcerpt(a.Content, 80)
			}
			data.FeatureCards = append(data.FeatureCards, FeatureCardView{
				Title: a.Title,
				Desc:  desc,
				Link:  articleHref(a.DirPath, a.Slug),
			})
		} else {
			data.Articles = append(data.Articles, ArticleView{
				Title:   a.Title,
				Author:  a.Author,
				Date:    a.Date,
				Tags:    a.Tags,
				Href:    articleHref(a.DirPath, a.Slug),
				Summary: a.Summary,
			})
		}
	}

	return data
}

func simpleExcerpt(content string, n int) string {
	s := strings.TrimSpace(content)
	s = strings.ReplaceAll(s, "\n", " ")
	runes := []rune(s)
	if len(runes) > n {
		return string(runes[:n]) + "..."
	}
	return s
}

func (h *PageHandler) DirList(c *echo.Context) error {
	dir := c.Param("dir")

	article, ok := h.cache.GetArticle(dir)
	if ok {
		rendered, err := h.renderer.RenderString(c.Request().Context(), article.Content)
		if err != nil {
			rendered = article.Content
		}
		siteTree, _ := h.cache.GetSiteTree()
		data := h.buildBaseData(siteTree.Meta, article.DirPath)
		data.Title = article.Title
		data.Article = article
		data.Content = rendered
		visitor, kind := h.prepareView(c, article)
		if err := h.renderTemplate(c, "article", data); err != nil {
			return err
		}
		h.recordView(article, kind, visitor)
		return nil
	}

	siteTree, ok := h.cache.GetSiteTree()
	if !ok || siteTree == nil {
		return c.String(http.StatusInternalServerError, "站点数据未加载")
	}

	node := findNode(siteTree, dir)
	if node == nil {
		return c.String(http.StatusNotFound, "页面不存在")
	}

	if node.Meta != nil && node.Meta.DIR_TYPE == "product" {
		return h.renderProductList(c, siteTree, node, dir)
	}

	if node.Meta != nil && node.Meta.DIR_TYPE == "page" && len(node.Articles) == 1 {
		article := node.Articles[0]
		rendered, err := h.renderer.RenderString(c.Request().Context(), article.Content)
		if err != nil {
			rendered = article.Content
		}
		data := h.buildBaseData(siteTree.Meta, article.DirPath)
		data.Title = article.Title
		data.Article = article
		data.Content = rendered
		data.DirType = "page"
		return h.renderTemplate(c, "article", data)
	}

	page := pageNumber(c)

	pageSize := 10
	if node.Meta != nil && node.Meta.ARTICLES_PER_PAGE > 0 {
		pageSize = node.Meta.ARTICLES_PER_PAGE
	}
	total := len(node.Articles)
	totalPages := (total + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= total {
		return c.String(http.StatusNotFound, "页面不存在")
	}
	if end > total {
		end = total
	}

	data := h.buildBaseData(siteTree.Meta, dir)
	data.Title = node.Title
	data.PageTitle = node.Title
	tmplName := "list"
	if node.Meta != nil {
		data.PageDesc = node.Meta.SITE_DESC
		data.DirType = node.Meta.DIR_TYPE
		if name := layoutToTemplate(node.Meta.LAYOUT); name != "" {
			tmplName = name
		}
	}
	data.Articles = toArticleViews(node.Articles[start:end])

	if totalPages > 1 && tmplName == "list" {
		base := "/" + dir
		data.Pagination = &Pagination{
			Current: page,
			Total:   totalPages,
			HasPrev: page > 1,
			HasNext: page < totalPages,
		}
		if data.Pagination.HasPrev {
			if page-1 == 1 {
				data.Pagination.PrevHref = base
			} else {
				data.Pagination.PrevHref = fmt.Sprintf("%s/page/%d", base, page-1)
			}
		}
		if data.Pagination.HasNext {
			data.Pagination.NextHref = fmt.Sprintf("%s/page/%d", base, page+1)
		}
	}

	if tmplName == "index" && node.Meta != nil {
		for _, f := range node.Meta.FEATURES {
			data.FeatureCards = append(data.FeatureCards, FeatureCardView{
				Title: f.Title, Desc: f.Desc, Icon: f.Icon, Link: f.Link,
			})
		}
	}

	return h.renderTemplate(c, tmplName, data)
}

func (h *PageHandler) renderProductList(c *echo.Context, siteTree, node *models.SiteTree, dir string) error {
	all := buildProductViews(node.Articles)
	active := normalizeProductCategory(c.QueryParam("category"))
	products := filterProductsByCategory(all, active)

	page := pageNumber(c)
	pageSize := 9
	if node.Meta != nil && node.Meta.ARTICLES_PER_PAGE > 0 {
		pageSize = node.Meta.ARTICLES_PER_PAGE
	}
	total := len(products)
	totalPages := max(1, (total+pageSize-1)/pageSize)
	if page > totalPages {
		return c.String(http.StatusNotFound, "页面不存在")
	}
	start := min((page-1)*pageSize, total)
	end := min(start+pageSize, total)

	base := "/" + dir
	categories := buildCategoryNodes(all)
	prepareCategoryNodes(categories, base, active)

	data := h.buildBaseData(siteTree.Meta, dir)
	data.Title = node.Title
	data.PageTitle = node.Title
	if node.Meta != nil {
		data.PageDesc = node.Meta.SITE_DESC
	}
	data.DirType = "product"
	data.Products = products[start:end]
	data.ProductGroups = groupProducts(data.Products)
	data.ProductCategories = categories
	data.ActiveCategory = active
	data.ClearCategoryHref = base

	if totalPages > 1 {
		data.Pagination = &Pagination{
			Current: page,
			Total:   totalPages,
			HasPrev: page > 1,
			HasNext: page < totalPages,
		}
		if data.Pagination.HasPrev {
			data.Pagination.PrevHref = productPageHref(base, active, page-1)
		}
		if data.Pagination.HasNext {
			data.Pagination.NextHref = productPageHref(base, active, page+1)
		}
	}
	return h.renderTemplate(c, "list", data)
}

func pageNumber(c *echo.Context) int {
	raw := c.Param("num")
	if raw == "" {
		raw = c.QueryParam("page")
	}
	if n, err := strconv.Atoi(raw); err == nil && n > 0 {
		return n
	}
	return 1
}

func filterProductsByCategory(products []ProductView, category string) []ProductView {
	if category == "" {
		return products
	}
	filtered := make([]ProductView, 0, len(products))
	for _, product := range products {
		if product.Category == category || strings.HasPrefix(product.Category, category+"/") {
			filtered = append(filtered, product)
		}
	}
	return filtered
}

func productPageHref(base, category string, page int) string {
	href := base
	if page > 1 {
		href = fmt.Sprintf("%s/page/%d", base, page)
	}
	if category != "" {
		href += "?category=" + url.QueryEscape(category)
	}
	return href
}

func layoutToTemplate(layout string) string {
	switch layout {
	case "home":
		return "index"
	case "article":
		return "article"
	case "list":
		return "list"
	default:
		return ""
	}
}

func (h *PageHandler) ArticleDetail(c *echo.Context) error {
	slug := c.Param("slug")

	article, ok := h.cache.GetArticle(slug)
	if !ok {
		return c.String(http.StatusNotFound, "文章不存在")
	}

	return h.renderArticle(c, article)
}

func (h *PageHandler) ResolvePath(c *echo.Context) error {
	var segments []string
	for _, p := range []string{"d1", "d2", "d3", "d4", "d5"} {
		s := c.Param(p)
		if s == "" {
			break
		}
		segments = append(segments, s)
	}
	if len(segments) < 3 {
		return c.String(http.StatusNotFound, "页面不存在")
	}

	slug := segments[len(segments)-1]
	dirPath := strings.Join(segments[:len(segments)-1], "/")

	article, ok := h.cache.GetArticle(slug)
	if !ok || article.DirPath != dirPath {
		return c.String(http.StatusNotFound, "页面不存在")
	}

	return h.renderArticle(c, article)
}

func (h *PageHandler) renderArticle(c *echo.Context, article *models.Article) error {
	if article.Draft {
		return c.String(http.StatusNotFound, "文章不存在")
	}

	rendered, err := h.renderer.RenderString(c.Request().Context(), article.Content)
	if err != nil {
		rendered = article.Content
	}

	siteTree, _ := h.cache.GetSiteTree()
	data := h.buildBaseData(siteTree.Meta, article.DirPath)
	data.Title = article.Title
	data.Article = article
	data.Content = rendered
	data.DirType = h.cache.GetDirType(article.DirPath)
	if data.DirType == "product" {
		view := buildProductView(article)
		data.Product = &view
	}

	visitor, kind := h.prepareView(c, article)
	if err := h.renderTemplate(c, "article", data); err != nil {
		return err
	}
	h.recordView(article, kind, visitor)
	return nil
}

func (h *PageHandler) buildBaseData(meta *models.DirMeta, currentDir string) PageData {
	siteTree, _ := h.cache.GetSiteTree()

	data := PageData{
		Year:           time.Now().Year(),
		NavItems:       buildNavItems(siteTree, currentDir),
		HeroTitle:      meta.HERO_TITLE,
		HeroSubtitle:   meta.HERO_SUBTITLE,
		HeroImage:      meta.HERO_IMAGE,
		HeroCTAText:    meta.HERO_CTA_TEXT,
		HeroCTALink:    meta.HERO_CTA_LINK,
		FeatureTitle:   meta.FEATURE_TITLE,
		NewsTitle:      meta.NEWS_TITLE,
		NewsLink:       meta.NEWS_LINK,
		ContactEmail:   meta.CONTACT_EMAIL,
		ContactPhone:   meta.CONTACT_PHONE,
		ContactAddress: meta.CONTACT_ADDRESS,
		FooterText:     meta.FOOTER_TEXT,
		ICP:            meta.ICP,
	}

	if meta != nil {
		data.SiteTitle = meta.SITE_TITLE
		data.SiteDesc = meta.SITE_DESC
	}

	data.ThemeCSS = h.themeCSS()
	return data
}

func (h *PageHandler) themeCSS() template.CSS {
	if h.themes == nil {
		return ""
	}
	theme, err := h.themes.Effective(time.Now())
	if err != nil {
		if h.logger != nil {
			h.logger.Warn("resolve site theme failed", "error", err)
		}
		return ""
	}
	if theme == nil {
		return ""
	}
	return template.CSS(theme.CSS())
}

func (h *PageHandler) renderToBytes(name string, data PageData) ([]byte, error) {
	var buf []byte
	w := &writeBuffer{data: &buf}
	tmpl := h.templates.Load()
	if tmpl == nil {
		return nil, fmt.Errorf("templates not loaded")
	}
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		return nil, fmt.Errorf("render template %s: %w", name, err)
	}
	if h.dev {
		buf = injectLivereload(buf)
	}
	return buf, nil
}

func injectLivereload(html []byte) []byte {
	script := []byte(`<script>new EventSource("/livereload").onmessage=function(){location.reload()};</script>`)
	return bytes.Replace(html, []byte("</body>"), append(script, []byte("</body>")...), 1)
}

func (h *PageHandler) renderTemplate(c *echo.Context, name string, data PageData) error {
	buf, err := h.renderToBytes(name, data)
	if err != nil {
		return err
	}
	return c.HTMLBlob(http.StatusOK, buf)
}

func buildNavItems(tree *models.SiteTree, currentDir string) []NavItem {
	items := []NavItem{{Label: "首页", Href: "/", Active: currentDir == ""}}
	if tree == nil {
		return items
	}
	for _, child := range tree.Children {
		if child.Meta != nil && child.Meta.NAV_HIDE {
			continue
		}
		items = append(items, NavItem{
			Label:  child.Title,
			Href:   "/" + child.Slug,
			Active: child.Slug == currentDir,
		})
	}
	return items
}

func buildProductViews(articles []*models.Article) []ProductView {
	views := make([]ProductView, 0, len(articles))
	for _, a := range articles {
		views = append(views, buildProductView(a))
	}
	return views
}

func buildGallery(a *models.Article) []string {
	images := make([]string, 0, len(a.Gallery)+1)
	seen := make(map[string]bool, len(a.Gallery)+1)
	add := func(src string) {
		if src == "" || seen[src] {
			return
		}
		seen[src] = true
		images = append(images, src)
	}
	add(a.Cover)
	for _, src := range a.Gallery {
		add(src)
	}
	return images
}

func buildProductView(a *models.Article) ProductView {
	view := ProductView{
		Title:   a.Title,
		Href:    articleHref(a.DirPath, a.Slug),
		Summary: a.Summary,
		Cover:   a.Cover,
		Gallery: buildGallery(a),
		Specs:   make([]ProductSpecView, 0, len(a.Specs)),
	}
	if view.Summary == "" {
		view.Summary = simpleExcerpt(a.Content, 100)
	}
	view.Category = normalizeProductCategory(a.Category)
	if view.Category == "" {
		view.Category = "未分类"
	}
	view.CategoryLeaf = leafCategory(view.Category)
	view.Group = topCategory(view.Category)
	for _, improvement := range a.Improvements {
		content := strings.TrimSpace(improvement.Content)
		if !improvement.Visible || content == "" {
			continue
		}
		view.Improvements = append(view.Improvements, ImprovementView{
			Date:    improvement.Date,
			Content: content,
		})
	}
	sort.SliceStable(view.Improvements, func(i, j int) bool {
		return view.Improvements[i].Date > view.Improvements[j].Date
	})
	prices := make([]string, 0, len(a.Specs))
	for _, spec := range a.Specs {
		view.Specs = append(view.Specs, ProductSpecView{
			Name:      spec.Name,
			Price:     spec.Price,
			SKU:       spec.SKU,
			JDURL:     spec.JDURL,
			TaobaoURL: spec.TaobaoURL,
		})
		if spec.Price != "" && !containsString(prices, spec.Price) {
			prices = append(prices, spec.Price)
		}
	}
	switch len(prices) {
	case 0:
		view.PriceLabel = "询价"
	case 1:
		view.PriceLabel = prices[0]
	default:
		view.PriceLabel = prices[0] + " 起"
	}
	return view
}

func groupProducts(products []ProductView) []ProductGroup {
	groups := make([]ProductGroup, 0)
	index := make(map[string]int)
	for _, product := range products {
		i, ok := index[product.Group]
		if !ok {
			i = len(groups)
			index[product.Group] = i
			groups = append(groups, ProductGroup{Name: product.Group, Path: product.Group})
		}
		groups[i].Products = append(groups[i].Products, product)
	}
	sort.SliceStable(groups, func(i, j int) bool {
		li, ri := categoryRank(groups[i].Path, groups[i].Name), categoryRank(groups[j].Path, groups[j].Name)
		if li != ri {
			return li < ri
		}
		return groups[i].Name < groups[j].Name
	})
	return groups
}

const maxProductCategoryDepth = 2

type categoryTreeNode struct {
	name     string
	path     string
	depth    int
	count    int
	left     int
	right    int
	children []*categoryTreeNode
}

// buildCategoryNodes derives a nested set from product categories. Paths are
// stored as materialized paths ("parent/child"); the depth cap keeps the tree
// at two levels while the nested set lets a parent filter include children.
func buildCategoryNodes(products []ProductView) []CategoryNode {
	root := &categoryTreeNode{}
	index := make(map[string]*categoryTreeNode)
	for _, product := range products {
		segments := categorySegments(product.Category)
		parent := root
		for depth, name := range segments {
			if depth >= maxProductCategoryDepth {
				break
			}
			path := strings.Join(segments[:depth+1], "/")
			node, ok := index[path]
			if !ok {
				node = &categoryTreeNode{name: name, path: path, depth: depth + 1}
				index[path] = node
				parent.children = append(parent.children, node)
			}
			node.count++
			parent = node
		}
	}
	sortCategoryTree(root)

	counter := 1
	var assign func(*categoryTreeNode)
	assign = func(node *categoryTreeNode) {
		node.left = counter
		counter++
		for _, child := range node.children {
			assign(child)
		}
		node.right = counter
		counter++
	}
	for _, child := range root.children {
		assign(child)
	}

	nodes := make([]CategoryNode, 0, len(root.children))
	for _, child := range root.children {
		nodes = append(nodes, toCategoryNode(child))
	}
	return nodes
}

func toCategoryNode(node *categoryTreeNode) CategoryNode {
	out := CategoryNode{
		Name:  node.name,
		Path:  node.path,
		Depth: node.depth,
		Left:  node.left,
		Right: node.right,
		Count: node.count,
	}
	for _, child := range node.children {
		out.Children = append(out.Children, toCategoryNode(child))
	}
	return out
}

// categoryRanks pins the merchandising order; Chinese names otherwise sort by
// UTF-8 byte order, which looks random to customers.
var categoryRanks = map[string]int{
	"存储卡":           1,
	"存储卡/SD卡":       1,
	"存储卡/microSD":   2,
	"存储卡/CFexpress": 3,
	"U盘":            2,
	"U盘/USB-A":      1,
	"U盘/双接口":        2,
	"移动固态硬盘":        3,
	"移动固态硬盘/便携SSD":  1,
	"未分类":           99,
}

func categoryRank(path, name string) int {
	if rank, ok := categoryRanks[path]; ok {
		return rank
	}
	if rank, ok := categoryRanks[name]; ok {
		return rank
	}
	return 50
}

func prepareCategoryNodes(nodes []CategoryNode, base, active string) {
	for i := range nodes {
		nodes[i].Href = productPageHref(base, nodes[i].Path, 1)
		nodes[i].Active = nodes[i].Path == active
		prepareCategoryNodes(nodes[i].Children, base, active)
	}
}

func sortCategoryTree(node *categoryTreeNode) {
	sort.SliceStable(node.children, func(i, j int) bool {
		left, right := node.children[i], node.children[j]
		lr, rr := categoryRank(left.path, left.name), categoryRank(right.path, right.name)
		if lr != rr {
			return lr < rr
		}
		return left.name < right.name
	})
	for _, child := range node.children {
		sortCategoryTree(child)
	}
}

func normalizeProductCategory(raw string) string {
	segments := categorySegments(raw)
	return strings.Join(segments, "/")
}

func categorySegments(raw string) []string {
	segments := make([]string, 0, maxProductCategoryDepth)
	for _, part := range strings.Split(raw, "/") {
		if part = strings.TrimSpace(part); part != "" {
			segments = append(segments, part)
		}
	}
	return segments
}

func topCategory(path string) string {
	if segments := categorySegments(path); len(segments) > 0 {
		return segments[0]
	}
	return "未分类"
}

func leafCategory(path string) string {
	segments := categorySegments(path)
	if len(segments) == 0 {
		return "未分类"
	}
	return segments[len(segments)-1]
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func toArticleViews(articles []*models.Article) []ArticleView {
	views := make([]ArticleView, 0, len(articles))
	for _, a := range articles {
		views = append(views, ArticleView{
			Title:   a.Title,
			Author:  a.Author,
			Date:    a.Date,
			Tags:    a.Tags,
			Href:    articleHref(a.DirPath, a.Slug),
			Summary: a.Summary,
		})
	}
	return views
}

func indexArticles(siteTree *models.SiteTree, newsLink string) []*models.Article {
	if slug := strings.Trim(newsLink, "/"); slug != "" {
		if node := findNode(siteTree, slug); node != nil && node != siteTree {
			var articles []*models.Article
			collectArticles(node, &articles)
			return articles
		}
	}
	var articles []*models.Article
	collectArticles(siteTree, &articles)
	return articles
}

func collectArticles(node *models.SiteTree, articles *[]*models.Article) {
	*articles = append(*articles, node.Articles...)
	for _, child := range node.Children {
		collectArticles(child, articles)
	}
}

func findNode(tree *models.SiteTree, slug string) *models.SiteTree {
	if tree.Slug == slug {
		return tree
	}
	for _, child := range tree.Children {
		if found := findNode(child, slug); found != nil {
			return found
		}
	}
	return nil
}

func articleHref(dirPath, slug string) string {
	if dirPath == "" {
		return "/" + slug
	}
	return "/" + dirPath + "/" + slug
}

type writeBuffer struct {
	data *[]byte
}

func (w *writeBuffer) Write(p []byte) (int, error) {
	*w.data = append(*w.data, p...)
	return len(p), nil
}
