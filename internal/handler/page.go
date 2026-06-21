package handler

import (
	"fmt"
	"html/template"
	"time"

	"github.com/azhai/gopaper/internal/model"
	"github.com/azhai/gopaper/internal/service"

	"github.com/gofiber/fiber/v3"
)

type PageHandler struct {
	cache     *service.CacheVault
	renderer  *service.Renderer
	templates *template.Template
}

func NewPageHandler(cache *service.CacheVault, renderer *service.Renderer) *PageHandler {
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"safeHTML": func(s string) template.HTML { return template.HTML(s) },
	}).ParseGlob("web/templates/*.html"))

	return &PageHandler{cache: cache, renderer: renderer, templates: tmpl}
}

type NavItem struct {
	Label  string
	Href   string
	Active bool
}

type ArticleView struct {
	Title  string
	Author string
	Date   string
	Tags   []string
	Href   string
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

	// Hero
	HeroTitle    string
	HeroSubtitle string
	HeroImage    string
	HeroCTAText  string
	HeroCTALink  string

	// Features
	FeatureTitle string
	FeatureCards []FeatureCardView

	// News
	NewsTitle string
	NewsLink  string
	Articles  []ArticleView

	// Contact
	ContactEmail   string
	ContactPhone   string
	ContactAddress string
	FooterText     string
	ICP            string

	// List page
	PageTitle string
	PageDesc  string

	// Article page
	Article *model.Article
	Content string
	DirType string // page, news, docs
}

func (h *PageHandler) Index(c fiber.Ctx) error {
	siteTree, ok := h.cache.GetSiteTree()
	if !ok || siteTree == nil {
		return c.Status(500).SendString("站点数据未加载")
	}

	var allArticles []*model.Article
	collectArticles(siteTree, &allArticles)

	meta := siteTree.Meta
	data := h.buildBaseData(meta, "")
	data.Title = meta.SITE_TITLE
	data.Articles = toArticleViews(allArticles)

	if len(meta.FEATURES) > 0 {
		for _, f := range meta.FEATURES {
			data.FeatureCards = append(data.FeatureCards, FeatureCardView{
				Title: f.Title, Desc: f.Desc, Icon: f.Icon, Link: f.Link,
			})
		}
	}

	return h.renderTemplate(c, "index", data)
}

func (h *PageHandler) DirList(c fiber.Ctx) error {
	dir := c.Params("dir")

	article, ok := h.cache.GetArticle(dir)
	if ok {
		rendered, err := h.renderer.RenderString(c.Context(), article.Content)
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
		return c.Status(500).SendString("站点数据未加载")
	}

	node := findNode(siteTree, dir)
	if node == nil {
		return c.Status(404).SendString("页面不存在")
	}

	// page 类型目录：只有一篇文章时直接显示内容，无需列表
	if node.Meta != nil && node.Meta.DIR_TYPE == "page" && len(node.Articles) == 1 {
		article := node.Articles[0]
		rendered, err := h.renderer.RenderString(c.Context(), article.Content)
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

	data := h.buildBaseData(siteTree.Meta, dir)
	data.Title = node.Title
	data.PageTitle = node.Title
	if node.Meta != nil {
		data.PageDesc = node.Meta.SITE_DESC
		data.DirType = node.Meta.DIR_TYPE
	}
	data.Articles = toArticleViews(node.Articles)

	return h.renderTemplate(c, "list", data)
}

func (h *PageHandler) ArticleDetail(c fiber.Ctx) error {
	slug := c.Params("slug")

	article, ok := h.cache.GetArticle(slug)
	if !ok {
		return c.Status(404).SendString("文章不存在")
	}

	rendered, err := h.renderer.RenderString(c.Context(), article.Content)
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

func (h *PageHandler) renderTemplate(c fiber.Ctx, name string, data PageData) error {
	var buf []byte
	w := &writeBuffer{data: &buf}
	if err := h.templates.ExecuteTemplate(w, name, data); err != nil {
		return fmt.Errorf("render template %s: %w", name, err)
	}
	c.Type("html")
	return c.Send(buf)
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
			Title:  a.Title,
			Author: a.Author,
			Date:   a.Date,
			Tags:   a.Tags,
			Href:   articleHref(a.DirPath, a.Slug),
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
