package service

import (
	"testing"

	"github.com/azhai/gopaper/internal/model"
)

func TestCompareArticles(t *testing.T) {
	tests := []struct {
		name      string
		a, b      model.Article
		sortOrder string
		want      bool
	}{
		{"asc weight small first", model.Article{Weight: 1}, model.Article{Weight: 2}, "asc", true},
		{"asc weight large not first", model.Article{Weight: 2}, model.Article{Weight: 1}, "asc", false},
		{"desc weight large first", model.Article{Weight: 2}, model.Article{Weight: 1}, "desc", true},
		{"default asc when empty", model.Article{Weight: 1}, model.Article{Weight: 2}, "", true},
		{"same weight asc date old first", model.Article{Weight: 1, Date: "2025-01-01"}, model.Article{Weight: 1, Date: "2025-01-02"}, "asc", true},
		{"same weight desc date new first", model.Article{Weight: 1, Date: "2025-01-02"}, model.Article{Weight: 1, Date: "2025-01-01"}, "desc", true},
		{"same weight date asc title", model.Article{Weight: 1, Date: "x", Title: "a"}, model.Article{Weight: 1, Date: "x", Title: "b"}, "asc", true},
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
