package handler

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/azhai/gopaper/internal/model"
	"github.com/azhai/gopaper/web"
)

type Builder struct {
	pages     *PageHandler
	outputDir string
	logger    *slog.Logger
}

func NewBuilder(pages *PageHandler, outputDir string, logger *slog.Logger) *Builder {
	return &Builder{pages: pages, outputDir: outputDir, logger: logger}
}

func (b *Builder) Build(ctx context.Context) error {
	siteTree, ok := b.pages.cache.GetSiteTree()
	if !ok || siteTree == nil {
		return fmt.Errorf("站点数据未加载")
	}

	if err := os.MkdirAll(b.outputDir, 0755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	if err := b.buildIndex(ctx, siteTree); err != nil {
		return fmt.Errorf("build index: %w", err)
	}

	if err := b.buildTree(ctx, siteTree); err != nil {
		return err
	}

	for _, article := range siteTree.Articles {
		if err := b.writeArticle(ctx, article); err != nil {
			b.logger.Warn("build top article failed", "slug", article.Slug, "error", err)
		}
	}

	if err := b.copyStaticAssets(); err != nil {
		b.logger.Warn("copy static assets failed", "error", err)
	}

	if err := b.writeFeeds(siteTree); err != nil {
		b.logger.Warn("write feeds failed", "error", err)
	}

	if err := b.rewriteStaticRefs(); err != nil {
		b.logger.Warn("rewrite static refs failed", "error", err)
	}

	b.logger.Info("build complete", "output", b.outputDir)
	return nil
}

func (b *Builder) writeFeeds(siteTree *model.SiteTree) error {
	siteURL := b.pages.siteURL
	if err := b.writePage(filepath.Join(b.outputDir, "index.xml"), GenerateRSS(siteURL, siteTree)); err != nil {
		return err
	}
	if err := b.writePage(filepath.Join(b.outputDir, "sitemap.xml"), GenerateSitemap(siteURL, siteTree)); err != nil {
		return err
	}
	return b.writePage(filepath.Join(b.outputDir, "robots.txt"), GenerateRobots(siteURL))
}

func (b *Builder) buildIndex(ctx context.Context, siteTree *model.SiteTree) error {
	data := b.pages.buildIndexData(siteTree)

	buf, err := b.pages.renderToBytes("index", data)
	if err != nil {
		return err
	}
	return b.writePage(filepath.Join(b.outputDir, "index.html"), buf)
}

func (b *Builder) buildTree(ctx context.Context, node *model.SiteTree) error {
	for _, child := range node.Children {
		if err := b.buildDir(ctx, child); err != nil {
			b.logger.Warn("build dir failed", "dir", child.DirPath, "error", err)
		}
		if err := b.buildTree(ctx, child); err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) buildDir(ctx context.Context, node *model.SiteTree) error {
	if node.Meta != nil && node.Meta.DIR_TYPE == "page" && len(node.Articles) == 1 {
		article := node.Articles[0]
		outPath := filepath.Join(b.outputDir, filepath.FromSlash(node.DirPath), "index.html")
		return b.writeArticleAt(ctx, article, outPath)
	}

	for _, article := range node.Articles {
		if err := b.writeArticle(ctx, article); err != nil {
			b.logger.Warn("build article failed", "slug", article.Slug, "error", err)
		}
	}

	siteTree, _ := b.pages.cache.GetSiteTree()
	tmplName := "list"
	if node.Meta != nil && node.Meta.LAYOUT != "" {
		if name := layoutToTemplate(node.Meta.LAYOUT); name != "" {
			tmplName = name
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

	for page := 1; page <= totalPages; page++ {
		start := (page - 1) * pageSize
		end := start + pageSize
		if end > total {
			end = total
		}

		data := b.pages.buildBaseData(siteTree.Meta, node.DirPath)
		data.Title = node.Title
		data.PageTitle = node.Title
		if node.Meta != nil {
			data.PageDesc = node.Meta.SITE_DESC
			data.DirType = node.Meta.DIR_TYPE
		}
		data.Articles = toArticleViews(node.Articles[start:end])

		if tmplName == "index" && node.Meta != nil {
			for _, f := range node.Meta.FEATURES {
				data.FeatureCards = append(data.FeatureCards, FeatureCardView{
					Title: f.Title, Desc: f.Desc, Icon: f.Icon, Link: f.Link,
				})
			}
		}

		if totalPages > 1 && tmplName == "list" {
			base := "/" + node.DirPath
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

		buf, err := b.pages.renderToBytes(tmplName, data)
		if err != nil {
			return err
		}
		var outPath string
		if page == 1 {
			outPath = filepath.Join(b.outputDir, filepath.FromSlash(node.DirPath), "index.html")
		} else {
			outPath = filepath.Join(b.outputDir, filepath.FromSlash(node.DirPath), "page", strconv.Itoa(page), "index.html")
		}
		if err := b.writePage(outPath, buf); err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) writeArticle(ctx context.Context, article *model.Article) error {
	relPath := articleHref(article.DirPath, article.Slug)
	outPath := filepath.Join(b.outputDir, filepath.FromSlash(strings.TrimPrefix(relPath, "/")), "index.html")
	return b.writeArticleAt(ctx, article, outPath)
}

func (b *Builder) writeArticleAt(ctx context.Context, article *model.Article, outPath string) error {
	rendered, err := b.pages.renderer.RenderString(ctx, article.Content)
	if err != nil {
		rendered = article.Content
	}

	siteTree, _ := b.pages.cache.GetSiteTree()
	data := b.pages.buildBaseData(siteTree.Meta, article.DirPath)
	data.Title = article.Title
	data.Article = article
	data.Content = rendered
	data.DirType = b.pages.cache.GetDirType(article.DirPath)

	buf, err := b.pages.renderToBytes("article", data)
	if err != nil {
		return err
	}
	return b.writePage(outPath, buf)
}

func (b *Builder) copyStaticAssets() error {
	publicFS, err := fs.Sub(web.PublicFS, "public")
	if err != nil {
		return err
	}
	staticDir := filepath.Join(b.outputDir, "static")
	return fs.WalkDir(publicFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasPrefix(path, "admin/") {
			return nil
		}
		data, err := fs.ReadFile(publicFS, path)
		if err != nil {
			return err
		}

		outName := filepath.Base(path)
		if ext := filepath.Ext(path); ext == ".css" || ext == ".js" {
			sum := md5.Sum(data)
			stem := strings.TrimSuffix(outName, ext)
			outName = stem + "." + hex.EncodeToString(sum[:])[:8] + ext
		}

		outPath := filepath.Join(staticDir, filepath.FromSlash(filepath.Dir(path)), outName)
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}
		return os.WriteFile(outPath, data, 0644)
	})
}

func (b *Builder) rewriteStaticRefs() error {
	staticDir := filepath.Join(b.outputDir, "static")
	manifest := make(map[string]string)
	err := filepath.Walk(staticDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(staticDir, path)
		rel = filepath.ToSlash(rel)
		ext := filepath.Ext(rel)
		base := strings.TrimSuffix(rel, ext)
		if dot := strings.LastIndex(base, "."); dot > 0 {
			base = base[:dot]
		}
		manifest["/static/"+base+ext] = "/static/" + rel
		return nil
	})
	if err != nil {
		return err
	}

	return filepath.Walk(b.outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if filepath.Ext(path) != ".html" {
			return nil
		}
		if strings.HasPrefix(path, staticDir) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out := string(data)
		for orig, hashed := range manifest {
			out = strings.ReplaceAll(out, orig, hashed)
		}
		if out == string(data) {
			return nil
		}
		return os.WriteFile(path, []byte(out), 0644)
	})
}

func (b *Builder) writePage(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0644)
}
