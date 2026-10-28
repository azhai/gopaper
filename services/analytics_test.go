package services

import (
	"testing"
	"time"
)

func TestContentKind(t *testing.T) {
	tests := []struct {
		name    string
		dirType string
		dirPath string
		want    string
	}{
		{name: "product", dirType: "product", dirPath: "products", want: KindProduct},
		{name: "news section", dirType: "news", dirPath: "news", want: KindNews},
		{name: "news by path", dirType: "docs", dirPath: "news/guides", want: KindNews},
		{name: "docs as article", dirType: "docs", dirPath: "docs", want: KindArticle},
		{name: "root article", dirType: "", dirPath: "", want: KindArticle},
		{name: "page skipped", dirType: "page", dirPath: "about", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContentKind(tt.dirType, tt.dirPath); got != tt.want {
				t.Fatalf("ContentKind(%q, %q) = %q, want %q", tt.dirType, tt.dirPath, got, tt.want)
			}
		})
	}
}

func TestAnalyticsReport(t *testing.T) {
	a := NewAnalytics(t.TempDir(), nil)
	defer a.Close()

	now := time.Now()
	record := func(at time.Time, path, title, kind, visitor string) {
		t.Helper()
		if err := a.recordAt(at, path, title, kind, visitor); err != nil {
			t.Fatalf("record event: %v", err)
		}
	}
	record(now.Add(-2*time.Minute), "/products/old", "旧标题", KindProduct, "v1")
	record(now.Add(-90*time.Second), "/products/old", "新标题", KindProduct, "v1")
	record(now.Add(-30*time.Minute), "/news/launch", "上新", KindNews, "v2")
	record(now.Add(-2*time.Hour), "/docs/guide", "指南", KindArticle, "v1")
	record(now.Add(-2*24*time.Hour), "/news/launch", "上新", KindNews, "v3")
	record(now.Add(-40*24*time.Hour), "/products/old", "旧标题", KindProduct, "v9")

	tests := []struct {
		rangeKey string
		points   int
		views    int
		users    int
		contents int
		kinds    map[string][2]int // views, users
	}{
		{
			rangeKey: "5m", points: 5, views: 2, users: 1, contents: 1,
			kinds: map[string][2]int{KindProduct: {2, 1}, KindArticle: {0, 0}, KindNews: {0, 0}},
		},
		{
			rangeKey: "1h", points: 12, views: 3, users: 2, contents: 2,
			kinds: map[string][2]int{KindProduct: {2, 1}, KindArticle: {0, 0}, KindNews: {1, 1}},
		},
		{
			rangeKey: "1d", points: 24, views: 4, users: 2, contents: 3,
			kinds: map[string][2]int{KindProduct: {2, 1}, KindArticle: {1, 1}, KindNews: {1, 1}},
		},
		{
			rangeKey: "1w", points: 7, views: 5, users: 3, contents: 3,
			kinds: map[string][2]int{KindProduct: {2, 1}, KindArticle: {1, 1}, KindNews: {2, 2}},
		},
		{
			rangeKey: "1m", points: 30, views: 5, users: 3, contents: 3,
			kinds: map[string][2]int{KindProduct: {2, 1}, KindArticle: {1, 1}, KindNews: {2, 2}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.rangeKey, func(t *testing.T) {
			report, ok := a.Report(now, tt.rangeKey)
			if !ok {
				t.Fatalf("range %q not accepted", tt.rangeKey)
			}
			if got := report.Totals.Views; got != tt.views {
				t.Errorf("views = %d, want %d", got, tt.views)
			}
			if got := report.Totals.Users; got != tt.users {
				t.Errorf("users = %d, want %d", got, tt.users)
			}
			if got := report.Totals.Contents; got != tt.contents {
				t.Errorf("contents = %d, want %d", got, tt.contents)
			}
			if len(report.Series) != tt.points {
				t.Fatalf("series points = %d, want %d", len(report.Series), tt.points)
			}
			seriesViews := 0
			for _, point := range report.Series {
				seriesViews += point.Views
			}
			if seriesViews != tt.views {
				t.Errorf("series views = %d, want %d", seriesViews, tt.views)
			}
			if len(report.Kinds) != 3 {
				t.Fatalf("kinds = %d, want 3", len(report.Kinds))
			}
			for _, kind := range report.Kinds {
				want := tt.kinds[kind.Kind]
				if kind.Views != want[0] || kind.Users != want[1] {
					t.Errorf("%s = %d views / %d users, want %d / %d", kind.Kind, kind.Views, kind.Users, want[0], want[1])
				}
			}
		})
	}

	report, _ := a.Report(now, "1w")
	if len(report.Contents) == 0 {
		t.Fatal("contents list is empty")
	}
	if report.Contents[0].Path != "/news/launch" || report.Contents[0].Views != 2 {
		t.Errorf("top content = %+v, want /news/launch with 2 views", report.Contents[0])
	}
	for _, item := range report.Contents {
		if item.Path == "/products/old" && item.Title != "新标题" {
			t.Errorf("product title = %q, want latest title 新标题", item.Title)
		}
	}
}

func TestAnalyticsPersistence(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()

	a := NewAnalytics(dir, nil)
	if err := a.recordAt(now.Add(-time.Minute), "/docs/guide", "指南", KindArticle, "v1"); err != nil {
		t.Fatalf("record: %v", err)
	}
	if err := a.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	reopened := NewAnalytics(dir, nil)
	defer reopened.Close()
	report, ok := reopened.Report(now, "1h")
	if !ok {
		t.Fatal("report rejected")
	}
	if report.Totals.Views != 1 || report.Totals.Users != 1 {
		t.Fatalf("reloaded totals = %+v, want 1 view / 1 user", report.Totals)
	}
	if got := report.Kinds[1]; got.Kind != KindArticle || got.Views != 1 {
		t.Fatalf("article kind = %+v, want 1 view", got)
	}
}

func TestAnalyticsRejectsUnknownRange(t *testing.T) {
	a := NewAnalytics(t.TempDir(), nil)
	defer a.Close()
	if _, ok := a.Report(time.Now(), "2y"); ok {
		t.Fatal("unknown range accepted")
	}
	if r, ok := AnalyticsRangeByKey(""); !ok || r.Key != "1d" {
		t.Fatalf("default range = %+v, %v", r, ok)
	}
}
