package handler

import "testing"

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
