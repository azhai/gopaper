package services

import (
	"strings"
	"testing"

	"github.com/azhai/gopaper/models"
)

func TestSlugCandidate(t *testing.T) {
	cases := []struct {
		name  string
		title string
		want  string
		fall  bool
	}{
		{name: "english title", title: "Hello World", want: "hello-world"},
		{name: "punctuation", title: "Go 1.27 / Release!", want: "go-1-27-release"},
		{name: "chinese falls back", title: "公司新闻", fall: true},
		{name: "empty falls back", title: "!!!", fall: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := slugCandidate(tc.title)
			if tc.fall {
				if !strings.HasPrefix(got, "p-") {
					t.Fatalf("slugCandidate(%q) = %q, want p- prefix", tc.title, got)
				}
				return
			}
			if got != tc.want {
				t.Fatalf("slugCandidate(%q) = %q, want %q", tc.title, got, tc.want)
			}
		})
	}
}

func TestValidateProductFields(t *testing.T) {
	validSpec := models.ProductSpec{Name: "标准版", SKU: "CMS-STD"}
	cases := []struct {
		name      string
		input     models.ArticleInput
		isProduct bool
		wantErr   string
	}{
		{
			name: "valid product",
			input: models.ArticleInput{
				Category: "软件产品/企业建站",
				Specs:    []models.ProductSpec{validSpec},
				Improvements: []models.ProductImprovement{
					{Date: "2025-06-18", Content: "新增草稿自动保存", Visible: true},
				},
			},
			isProduct: true,
		},
		{
			name:      "missing sku",
			input:     models.ArticleInput{Specs: []models.ProductSpec{{Name: "标准版"}}},
			isProduct: true,
			wantErr:   "specs.0.sku",
		},
		{
			name: "duplicate sku",
			input: models.ArticleInput{Specs: []models.ProductSpec{
				validSpec,
				{Name: "专业版", SKU: "cms-std"},
			}},
			isProduct: true,
			wantErr:   "specs.1.sku",
		},
		{
			name:      "product needs a spec",
			input:     models.ArticleInput{},
			isProduct: true,
			wantErr:   "specs",
		},
		{
			name:      "category max two levels",
			input:     models.ArticleInput{Category: "软件产品/建站/模板", Specs: []models.ProductSpec{validSpec}},
			isProduct: true,
			wantErr:   "category",
		},
		{
			name: "invalid improvement date and content",
			input: models.ArticleInput{
				Specs:        []models.ProductSpec{validSpec},
				Improvements: []models.ProductImprovement{{Date: "2025/06/18"}},
			},
			isProduct: true,
			wantErr:   "improvements.0.date",
		},
		{
			name:      "non-product may omit specs",
			input:     models.ArticleInput{},
			isProduct: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			errors := validateProductFields(tc.input, tc.isProduct)
			if tc.wantErr == "" {
				if len(errors) != 0 {
					t.Fatalf("unexpected errors: %v", errors)
				}
				return
			}
			if _, ok := errors[tc.wantErr]; !ok {
				t.Fatalf("missing error %q in %v", tc.wantErr, errors)
			}
		})
	}
}
