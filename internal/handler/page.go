package handler

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/azhai/gopaper/internal/model"
	"github.com/azhai/gopaper/internal/service"

	"github.com/labstack/echo/v4"
)

type PageHandler struct {
	cache     *service.CacheVault
	renderer  *service.Renderer
	templates atomic.Pointer[template.Template]
	siteURL   string
	dev       bool
	sse       *SSEServer
}

func NewPageHandler(cache *service.CacheVault, renderer *service.Renderer, siteURL string) *PageHandler {
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"safeHTML": func(s string) template.HTML { return template.HTML(s) },
	}).ParseGlob("themes/default/layouts/*.html"))

	ph := &PageHandler{cache: cache, renderer: renderer, siteURL: siteURL}
	ph.templates.Store(tmpl)
	return ph
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

func (h *PageHandler) RSS(c echo.Context) error {
	siteTree, ok := h.cache.GetSiteTree()
	if !ok || siteTree == nil {
		return c.String(http.StatusServiceUnavailable, "站点数据未加载")
	}
	return c.XMLBlob(http.StatusOK, GenerateRSS(h.siteURL, siteTree))
}

func (h *PageHandler) Sitemap(c echo.Context) error {
	siteTree, ok := h.cache.GetSiteTree()
	if !ok || siteTree == nil {
		return c.String(http.StatusServiceUnavailable, "站点数据未加载")
	}
	return c.XMLBlob(http.StatusOK, GenerateSitemap(h.siteURL, siteTree))
}

func (h *PageHandler) Robots(c echo.Context) error {
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

type PageData struct {
	Title     string
	SiteTitle string
	SiteDesc  string
	NavItems  []NavItem
	Year      int

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

	Article *model.Article
	Content string
	DirType string
	Pagination *Pagination
}

type Pagination struct {
	Current  int
	Total    int
	HasPrev  bool
	HasNext  bool
	PrevHref string
	NextHref string
}

func (h *PageHandler) Index(c echo.Context) error {
	siteTree, ok := h.cache.GetSiteTree()
	if !ok || siteTree == nil {
		return c.String(http.StatusInternalServerError, "站点数据未加载")
	}

	data := h.buildIndexData(siteTree)
	return h.renderTemplate(c, "index", data)
}

func (h *PageHandler) buildIndexData(siteTree *model.SiteTree) PageData {
	var allArticles []*model.Article
	collectArticles(siteTree, &allArticles)

	meta := siteTree.Meta
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
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}

func (h *PageHandler) DirList(c echo.Context) error {
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
		return h.renderTemplate(c, "article", data)
	}

	siteTree, ok := h.cache.GetSiteTree()
	if !ok || siteTree == nil {
		return c.String(http.StatusInternalServerError, "站点数据未加载")
	}

	node := findNode(siteTree, dir)
	if node == nil {
		return c.String(http.StatusNotFound, "页面不存在")
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

	page := 1
	if num := c.Param("num"); num != "" {
		if n, err := strconv.Atoi(num); err == nil && n > 0 {
			page = n
		}
	} else if q := c.QueryParam("page"); q != "" {
		if n, err := strconv.Atoi(q); err == nil && n > 0 {
			page = n
		}
	}

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

func (h *PageHandler) ArticleDetail(c echo.Context) error {
	slug := c.Param("slug")

	article, ok := h.cache.GetArticle(slug)
	if !ok {
		return c.String(http.StatusNotFound, "文章不存在")
	}

	return h.renderArticle(c, article)
}

func (h *PageHandler) ResolvePath(c echo.Context) error {
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

func (h *PageHandler) renderArticle(c echo.Context, article *model.Article) error {
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

	return h.renderTemplate(c, "article", data)
}

func (h *PageHandler) buildBaseData(meta *model.DirMeta, currentDir string) PageData {
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

	return data
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

func (h *PageHandler) renderTemplate(c echo.Context, name string, data PageData) error {
	buf, err := h.renderToBytes(name, data)
	if err != nil {
		return err
	}
	return c.HTMLBlob(http.StatusOK, buf)
}

func buildNavItems(tree *model.SiteTree, currentDir string) []NavItem {
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

func toArticleViews(articles []*model.Article) []ArticleView {
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

func collectArticles(node *model.SiteTree, articles *[]*model.Article) {
	*articles = append(*articles, node.Articles...)
	for _, child := range node.Children {
		collectArticles(child, articles)
	}
}

func findNode(tree *model.SiteTree, slug string) *model.SiteTree {
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
