package handler

import (
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/azhai/gopaper/internal/model"
)

func GenerateRSS(siteURL string, siteTree *model.SiteTree) []byte {
	base := strings.TrimRight(siteURL, "/")
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	sb.WriteString(`<rss version="2.0">` + "\n<channel>\n")

	meta := siteTree.Meta
	title := "Untitled"
	desc := ""
	if meta != nil {
		if meta.SITE_TITLE != "" {
			title = meta.SITE_TITLE
		}
		desc = meta.SITE_DESC
	}
	sb.WriteString(fmt.Sprintf("  <title>%s</title>\n", html.EscapeString(title)))
	sb.WriteString(fmt.Sprintf("  <link>%s</link>\n", html.EscapeString(base+"/")))
	if desc != "" {
		sb.WriteString(fmt.Sprintf("  <description>%s</description>\n", html.EscapeString(desc)))
	}

	var allArticles []*model.Article
	collectArticles(siteTree, &allArticles)

	for _, a := range allArticles {
		href := base + articleHref(a.DirPath, a.Slug)
		sb.WriteString("  <item>\n")
		sb.WriteString(fmt.Sprintf("    <title>%s</title>\n", html.EscapeString(a.Title)))
		sb.WriteString(fmt.Sprintf("    <link>%s</link>\n", html.EscapeString(href)))
		sb.WriteString(fmt.Sprintf("    <guid>%s</guid>\n", html.EscapeString(href)))
		if a.Date != "" {
			sb.WriteString(fmt.Sprintf("    <pubDate>%s</pubDate>\n", html.EscapeString(formatRSSDate(a.Date))))
		}
		sb.WriteString("  </item>\n")
	}

	sb.WriteString("</channel>\n</rss>\n")
	return []byte(sb.String())
}

func GenerateSitemap(siteURL string, siteTree *model.SiteTree) []byte {
	base := strings.TrimRight(siteURL, "/")
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	sb.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")

	sb.WriteString(fmt.Sprintf("  <url><loc>%s</loc></url>\n", html.EscapeString(base+"/")))

	for _, child := range siteTree.Children {
		loc := base + "/" + child.Slug + "/"
		sb.WriteString(fmt.Sprintf("  <url><loc>%s</loc></url>\n", html.EscapeString(loc)))
	}

	var allArticles []*model.Article
	collectArticles(siteTree, &allArticles)
	for _, a := range allArticles {
		loc := base + articleHref(a.DirPath, a.Slug) + "/"
		sb.WriteString(fmt.Sprintf("  <url><loc>%s</loc></url>\n", html.EscapeString(loc)))
	}

	sb.WriteString("</urlset>\n")
	return []byte(sb.String())
}

func GenerateRobots(siteURL string) []byte {
	base := strings.TrimRight(siteURL, "/")
	return []byte(fmt.Sprintf("User-agent: *\nAllow: /\nSitemap: %s/sitemap.xml\n", base))
}

func formatRSSDate(dateStr string) string {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format(time.RFC1123Z)
}
