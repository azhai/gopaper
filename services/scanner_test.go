package services

import (
	"testing"

	"github.com/azhai/gopaper/models"
)

func TestCompareArticles(t *testing.T) {
	tests := []struct {
		name      string
		a, b      models.Article
		sortOrder string
		want      bool
	}{
		{"asc weight small first", models.Article{Weight: 1}, models.Article{Weight: 2}, "asc", true},
		{"asc weight large not first", models.Article{Weight: 2}, models.Article{Weight: 1}, "asc", false},
		{"desc weight large first", models.Article{Weight: 2}, models.Article{Weight: 1}, "desc", true},
		{"default asc when empty", models.Article{Weight: 1}, models.Article{Weight: 2}, "", true},
		{"same weight asc date old first", models.Article{Weight: 1, Date: "2025-01-01"}, models.Article{Weight: 1, Date: "2025-01-02"}, "asc", true},
		{"same weight desc date new first", models.Article{Weight: 1, Date: "2025-01-02"}, models.Article{Weight: 1, Date: "2025-01-01"}, "desc", true},
		{"same weight date asc title", models.Article{Weight: 1, Date: "x", Title: "a"}, models.Article{Weight: 1, Date: "x", Title: "b"}, "asc", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareArticles(&tt.a, &tt.b, tt.sortOrder)
			if got != tt.want {
				t.Errorf("compareArticles() = %v, want %v", got, tt.want)
			}
		})
	}
}
