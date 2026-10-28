package services

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	// AnalyticsFileName is the append-only event log kept inside the content directory.
	AnalyticsFileName = "_analytics.jsonl"

	KindProduct = "product"
	KindArticle = "article"
	KindNews    = "news"
)

const (
	analyticsRetention = 31 * 24 * time.Hour
	analyticsMaxEvents = 100000
)

// AnalyticsRange describes one selectable reporting window.
type AnalyticsRange struct {
	Key         string
	Label       string
	BucketLabel string
	Window      time.Duration
	Bucket      time.Duration
}

var analyticsRanges = map[string]AnalyticsRange{
	"5m": {Key: "5m", Label: "近5分钟", BucketLabel: "每分钟", Window: 5 * time.Minute, Bucket: time.Minute},
	"1h": {Key: "1h", Label: "近1小时", BucketLabel: "每5分钟", Window: time.Hour, Bucket: 5 * time.Minute},
	"1d": {Key: "1d", Label: "近24小时", BucketLabel: "每小时", Window: 24 * time.Hour, Bucket: time.Hour},
	"1w": {Key: "1w", Label: "近7天", BucketLabel: "每天", Window: 7 * 24 * time.Hour, Bucket: 24 * time.Hour},
	"1m": {Key: "1m", Label: "近30天", BucketLabel: "每天", Window: 30 * 24 * time.Hour, Bucket: 24 * time.Hour},
}

// AnalyticsRangeByKey returns the range for key, defaulting to the last 24 hours.
func AnalyticsRangeByKey(key string) (AnalyticsRange, bool) {
	if key == "" {
		key = "1d"
	}
	r, ok := analyticsRanges[key]
	return r, ok
}

// ContentKind maps content metadata to one of the tracked analytics kinds.
// Page-type content is skipped because it is not a product, article or news item.
func ContentKind(dirType, dirPath string) string {
	switch dirType {
	case KindProduct:
		return KindProduct
	case "page":
		return ""
	case KindNews:
		return KindNews
	}
	if dirPath == KindNews || strings.HasPrefix(dirPath, KindNews+"/") {
		return KindNews
	}
	return KindArticle
}

type analyticsEvent struct {
	At      time.Time `json:"at"`
	Path    string    `json:"path"`
	Title   string    `json:"title"`
	Kind    string    `json:"kind"`
	Visitor string    `json:"visitor"`
}

type AnalyticsKind struct {
	Kind     string `json:"kind"`
	Label    string `json:"label"`
	Views    int    `json:"views"`
	Users    int    `json:"users"`
	Contents int    `json:"contents"`
}

type AnalyticsPoint struct {
	Label string    `json:"label"`
	Hint  string    `json:"hint"`
	Start time.Time `json:"start"`
	Views int       `json:"views"`
	Users int       `json:"users"`
}

type AnalyticsContent struct {
	Path  string `json:"path"`
	Title string `json:"title"`
	Kind  string `json:"kind"`
	Views int    `json:"views"`
	Users int    `json:"users"`
}

type AnalyticsTotals struct {
	Views    int `json:"views"`
	Users    int `json:"users"`
	Contents int `json:"contents"`
}

type AnalyticsReport struct {
	Range       string             `json:"range"`
	RangeLabel  string             `json:"rangeLabel"`
	BucketLabel string             `json:"bucketLabel"`
	From        time.Time          `json:"from"`
	To          time.Time          `json:"to"`
	Totals      AnalyticsTotals    `json:"totals"`
	Kinds       []AnalyticsKind    `json:"kinds"`
	Series      []AnalyticsPoint   `json:"series"`
	Contents    []AnalyticsContent `json:"contents"`
}

// Analytics keeps page views in memory and appends them to a JSONL file.
type Analytics struct {
	mu        sync.RWMutex
	path      string
	logger    *slog.Logger
	events    []analyticsEvent
	file      *os.File
	lastPrune time.Time
}

// NewAnalytics loads the event log, drops expired entries and opens it for appending.
func NewAnalytics(contentDir string, logger *slog.Logger) *Analytics {
	a := &Analytics{
		path:      filepath.Join(contentDir, AnalyticsFileName),
		logger:    logger,
		lastPrune: time.Now(),
	}
	a.load(time.Now())
	if err := a.open(); err != nil && logger != nil {
		logger.Warn("open analytics log failed", "path", a.path, "error", err)
	}
	return a
}

// Record appends one page view. visitor must be an anonymous, stable browser identifier.
func (a *Analytics) Record(path, title, kind, visitor string) error {
	now := time.Now()
	a.mu.Lock()
	defer a.mu.Unlock()
	a.pruneLocked(now)
	return a.appendLocked(analyticsEvent{At: now, Path: path, Title: title, Kind: kind, Visitor: visitor})
}

func (a *Analytics) recordAt(at time.Time, path, title, kind, visitor string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.appendLocked(analyticsEvent{At: at, Path: path, Title: title, Kind: kind, Visitor: visitor})
}

// Close flushes and closes the event log.
func (a *Analytics) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.file == nil {
		return nil
	}
	err := a.file.Close()
	a.file = nil
	return err
}

func (a *Analytics) load(now time.Time) {
	f, err := os.Open(a.path)
	if err != nil {
		if !os.IsNotExist(err) && a.logger != nil {
			a.logger.Warn("read analytics log failed", "path", a.path, "error", err)
		}
		return
	}
	defer f.Close()

	cutoff := now.Add(-analyticsRetention)
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	dropped := false
	for sc.Scan() {
		var ev analyticsEvent
		if err := json.Unmarshal(sc.Bytes(), &ev); err != nil {
			dropped = true
			continue
		}
		if ev.Path == "" || ev.Visitor == "" || !validKind(ev.Kind) || ev.At.Before(cutoff) {
			dropped = true
			continue
		}
		a.events = append(a.events, ev)
	}
	if err := sc.Err(); err != nil && a.logger != nil {
		a.logger.Warn("scan analytics log failed", "path", a.path, "error", err)
	}
	if len(a.events) > analyticsMaxEvents {
		a.events = a.events[len(a.events)-analyticsMaxEvents:]
		dropped = true
	}
	if dropped {
		a.compactLocked()
	}
}

func (a *Analytics) open() error {
	f, err := os.OpenFile(a.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	a.file = f
	return nil
}

// compactLocked rewrites the log with only the events currently held in memory.
// ponytail: compaction runs at startup and on an in-memory overflow; a long-running
// process keeps expired lines on disk until the next restart.
func (a *Analytics) compactLocked() {
	tmp := a.path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		if a.logger != nil {
			a.logger.Warn("compact analytics log failed", "path", a.path, "error", err)
		}
		return
	}
	w := bufio.NewWriter(f)
	for _, ev := range a.events {
		line, err := json.Marshal(ev)
		if err != nil {
			continue
		}
		if _, err := w.Write(append(line, '\n')); err != nil {
			break
		}
	}
	if err := w.Flush(); err != nil {
		f.Close()
		os.Remove(tmp)
		return
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return
	}
	if err := os.Rename(tmp, a.path); err != nil {
		os.Remove(tmp)
		if a.logger != nil {
			a.logger.Warn("replace analytics log failed", "path", a.path, "error", err)
		}
	}
}

func (a *Analytics) pruneLocked(now time.Time) {
	if now.Sub(a.lastPrune) < time.Minute {
		return
	}
	a.lastPrune = now
	cutoff := now.Add(-analyticsRetention)
	i := 0
	for i < len(a.events) && a.events[i].At.Before(cutoff) {
		i++
	}
	if i == 0 {
		return
	}
	a.events = append([]analyticsEvent(nil), a.events[i:]...)
}

func (a *Analytics) appendLocked(ev analyticsEvent) error {
	ev.Path = strings.TrimSpace(ev.Path)
	ev.Title = truncateRunes(strings.TrimSpace(ev.Title), 200)
	if ev.At.IsZero() {
		ev.At = time.Now()
	}
	if ev.Path == "" || ev.Visitor == "" || !validKind(ev.Kind) {
		return errors.New("invalid analytics event")
	}
	if len(a.events) >= analyticsMaxEvents {
		cut := analyticsMaxEvents / 5
		a.events = append([]analyticsEvent(nil), a.events[cut:]...)
		if a.file != nil {
			_ = a.file.Close()
			a.file = nil
			a.compactLocked()
			if err := a.open(); err != nil && a.logger != nil {
				a.logger.Warn("reopen analytics log failed", "path", a.path, "error", err)
			}
		}
	}
	a.events = append(a.events, ev)
	if a.file == nil {
		return nil
	}
	line, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("encode analytics event: %w", err)
	}
	if _, err := a.file.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("write analytics event: %w", err)
	}
	return nil
}

// Report aggregates views and unique visitors for the selected range.
func (a *Analytics) Report(now time.Time, rangeKey string) (AnalyticsReport, bool) {
	r, ok := AnalyticsRangeByKey(rangeKey)
	if !ok {
		return AnalyticsReport{}, false
	}

	starts := bucketStarts(now, r)
	from := starts[0]
	to := starts[len(starts)-1].Add(r.Bucket)

	report := AnalyticsReport{
		Range:       r.Key,
		RangeLabel:  r.Label,
		BucketLabel: r.BucketLabel,
		From:        from,
		To:          to,
		Kinds:       make([]AnalyticsKind, 0, 3),
		Series:      make([]AnalyticsPoint, len(starts)),
		Contents:    make([]AnalyticsContent, 0),
	}

	seriesUsers := make([]map[string]struct{}, len(starts))
	for i := range seriesUsers {
		seriesUsers[i] = map[string]struct{}{}
	}
	totalUsers := map[string]struct{}{}
	totalViews := 0
	kindViews := map[string]int{}
	kindUsers := map[string]map[string]struct{}{}
	kindPaths := map[string]map[string]struct{}{}
	contents := map[string]*contentAggregate{}

	a.mu.RLock()
	for _, ev := range a.events {
		if ev.At.Before(from) || !ev.At.Before(to) {
			continue
		}
		idx := bucketIndex(starts, r.Bucket, ev.At)
		if idx < 0 {
			continue
		}
		report.Series[idx].Views++
		seriesUsers[idx][ev.Visitor] = struct{}{}
		totalUsers[ev.Visitor] = struct{}{}
		totalViews++
		kindViews[ev.Kind]++
		if kindUsers[ev.Kind] == nil {
			kindUsers[ev.Kind] = map[string]struct{}{}
			kindPaths[ev.Kind] = map[string]struct{}{}
		}
		kindUsers[ev.Kind][ev.Visitor] = struct{}{}
		kindPaths[ev.Kind][ev.Path] = struct{}{}
		agg := contents[ev.Path]
		if agg == nil {
			agg = &contentAggregate{path: ev.Path, kind: ev.Kind, users: map[string]struct{}{}}
			contents[ev.Path] = agg
		}
		if ev.Title != "" {
			agg.title = ev.Title
		}
		agg.views++
		agg.users[ev.Visitor] = struct{}{}
	}
	a.mu.RUnlock()

	for i := range starts {
		label, hint := bucketLabel(starts[i], r.Bucket)
		report.Series[i].Label = label
		report.Series[i].Hint = hint
		report.Series[i].Start = starts[i]
		report.Series[i].Users = len(seriesUsers[i])
	}

	for _, kind := range []struct{ key, label string }{
		{KindProduct, "产品"},
		{KindArticle, "文章"},
		{KindNews, "新闻"},
	} {
		report.Kinds = append(report.Kinds, AnalyticsKind{
			Kind:     kind.key,
			Label:    kind.label,
			Views:    kindViews[kind.key],
			Users:    len(kindUsers[kind.key]),
			Contents: len(kindPaths[kind.key]),
		})
	}

	for _, agg := range contents {
		title := agg.title
		if title == "" {
			title = agg.path
		}
		report.Contents = append(report.Contents, AnalyticsContent{
			Path:  agg.path,
			Title: title,
			Kind:  agg.kind,
			Views: agg.views,
			Users: len(agg.users),
		})
	}
	sort.Slice(report.Contents, func(i, j int) bool {
		if report.Contents[i].Views != report.Contents[j].Views {
			return report.Contents[i].Views > report.Contents[j].Views
		}
		if report.Contents[i].Users != report.Contents[j].Users {
			return report.Contents[i].Users > report.Contents[j].Users
		}
		return report.Contents[i].Path < report.Contents[j].Path
	})

	report.Totals = AnalyticsTotals{
		Views:    totalViews,
		Users:    len(totalUsers),
		Contents: len(contents),
	}
	return report, true
}

type contentAggregate struct {
	path  string
	title string
	kind  string
	views int
	users map[string]struct{}
}

func bucketStarts(now time.Time, r AnalyticsRange) []time.Time {
	n := int(r.Window / r.Bucket)
	if n < 1 {
		n = 1
	}
	starts := make([]time.Time, n)
	var last time.Time
	if r.Bucket >= 24*time.Hour {
		last = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	} else {
		last = now.Truncate(r.Bucket)
	}
	for i := n - 1; i >= 0; i-- {
		starts[i] = last.Add(-time.Duration(n-1-i) * r.Bucket)
	}
	return starts
}

func bucketIndex(starts []time.Time, bucket time.Duration, at time.Time) int {
	i := sort.Search(len(starts), func(i int) bool { return starts[i].After(at) })
	i--
	if i < 0 || !at.Before(starts[i].Add(bucket)) {
		return -1
	}
	return i
}

func bucketLabel(start time.Time, bucket time.Duration) (label, hint string) {
	if bucket >= 24*time.Hour {
		return start.Format("01-02"), fmt.Sprintf("%d月%d日 %s", start.Month(), start.Day(), weekdayCN[start.Weekday()])
	}
	return start.Format("15:04"), start.Format("15:04") + "–" + start.Add(bucket).Format("15:04")
}

var weekdayCN = [...]string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}

func validKind(kind string) bool {
	return kind == KindProduct || kind == KindArticle || kind == KindNews
}

func truncateRunes(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}
