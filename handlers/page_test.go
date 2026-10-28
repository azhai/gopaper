package handlers

import (
	"testing"

	"github.com/azhai/gopaper/models"
)

func TestLayoutToTemplate(t *testing.T) {
	tests := []struct{ in, want string }{
		{"home", "index"},
		{"list", "list"},
		{"article", "article"},
		{"", ""},
		{"unknown", ""},
	}
	for _, tt := range tests {
		if got := layoutToTemplate(tt.in); got != tt.want {
			t.Errorf("layoutToTemplate(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestIndexArticles(t *testing.T) {
	rootArticle := &models.Article{Slug: "about"}
	newsArticle := &models.Article{Slug: "restock"}
	tree := &models.SiteTree{
		Slug:     "",
		Articles: []*models.Article{rootArticle},
		Children: []*models.SiteTree{
			{Slug: "products", Articles: []*models.Article{{Slug: "usb-drive"}}},
			{Slug: "news", Articles: []*models.Article{newsArticle}},
		},
	}

	tests := []struct {
		name     string
		newsLink string
		want     []string
	}{
		{"news subtree", "/news", []string{"restock"}},
		{"empty link falls back to all", "", []string{"about", "usb-drive", "restock"}},
		{"missing dir falls back to all", "/updates", []string{"about", "usb-drive", "restock"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := indexArticles(tree, tt.newsLink)
			if len(got) != len(tt.want) {
				t.Fatalf("indexArticles() returned %d articles, want %d", len(got), len(tt.want))
			}
			for i, slug := range tt.want {
				if got[i].Slug != slug {
					t.Errorf("article %d = %q, want %q", i, got[i].Slug, slug)
				}
			}
		})
	}
}

func TestFilterProductsByCategory(t *testing.T) {
	products := []ProductView{
		{Title: "a", Category: "存储卡/microSD"},
		{Title: "b", Category: "存储卡/SD卡"},
		{Title: "c", Category: "U盘/USB-A"},
	}
	tests := []struct {
		name     string
		category string
		want     []string
	}{
		{"all", "", []string{"a", "b", "c"}},
		{"parent includes children", "存储卡", []string{"a", "b"}},
		{"child only", "存储卡/microSD", []string{"a"}},
		{"no match", "移动固态硬盘", []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterProductsByCategory(products, tt.category)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d products, want %d", len(got), len(tt.want))
			}
			for i, title := range tt.want {
				if got[i].Title != title {
					t.Errorf("product %d = %q, want %q", i, got[i].Title, title)
				}
			}
		})
	}
}

func TestProductPageHref(t *testing.T) {
	if got := productPageHref("/products", "", 1); got != "/products" {
		t.Errorf("page 1 = %q", got)
	}
	if got := productPageHref("/products", "", 2); got != "/products/page/2" {
		t.Errorf("page 2 = %q", got)
	}
	if got := productPageHref("/products", "cards/micro", 1); got != "/products?category=cards%2Fmicro" {
		t.Errorf("category page 1 = %q", got)
	}
	if got := productPageHref("/products", "cards/micro", 3); got != "/products/page/3?category=cards%2Fmicro" {
		t.Errorf("category page 3 = %q", got)
	}
}

func TestSimpleExcerpt(t *testing.T) {
	tests := []struct {
		name, in string
		n        int
		want     string
	}{
		{"short no truncate", "hello", 80, "hello"},
		{"long truncated", "this is a very long content that exceeds the limit", 10, "this is a ..."},
		{"newlines to spaces", "line1\nline2", 80, "line1 line2"},
		{"empty", "", 80, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := simpleExcerpt(tt.in, tt.n); got != tt.want {
				t.Errorf("simpleExcerpt() = %q, want %q", got, tt.want)
			}
		})
	}
}
