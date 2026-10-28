package services

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/azhai/gopaper/models"
)

func newTestThemeForge(t *testing.T) *ThemeForge {
	t.Helper()
	return NewThemeForge(t.TempDir(), slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func at(value string) time.Time {
	at, err := time.ParseInLocation("2006-01-02T15:04", value, time.Local)
	if err != nil {
		panic(err)
	}
	return at
}

func TestPickThemeByTime(t *testing.T) {
	themes := []models.Theme{
		{ID: "a", EnabledAt: "2026-01-01T00:00"},
		{ID: "b", EnabledAt: "2026-03-01T09:30"},
		{ID: "c", EnabledAt: "2026-06-01T09:30"},
		{ID: "bad", EnabledAt: "not-a-time"},
	}

	cases := []struct {
		name     string
		now      time.Time
		current  string
		upcoming string
	}{
		{name: "before all", now: at("2025-12-31T23:00"), current: "", upcoming: "a"},
		{name: "first wins", now: at("2026-02-01T00:00"), current: "a", upcoming: "b"},
		{name: "latest past wins", now: at("2026-04-15T12:00"), current: "b", upcoming: "c"},
		{name: "after all", now: at("2026-12-31T00:00"), current: "c", upcoming: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := pickThemeByTime(themes, tc.now, false); stringOf(got) != tc.current {
				t.Fatalf("current = %q, want %q", stringOf(got), tc.current)
			}
			if got := pickThemeByTime(themes, tc.now, true); stringOf(got) != tc.upcoming {
				t.Fatalf("upcoming = %q, want %q", stringOf(got), tc.upcoming)
			}
		})
	}
}

func stringOf(theme *models.Theme) string {
	if theme == nil {
		return ""
	}
	return theme.ID
}

func TestImmediateActivationWinsSameMinute(t *testing.T) {
	forge := newTestThemeForge(t)

	first, err := forge.Create(models.Theme{Name: "甲", Colors: models.ThemeColors{Primary: "#111111"}})
	if err != nil {
		t.Fatalf("Create 甲: %v", err)
	}
	second, err := forge.Create(models.Theme{Name: "乙", Colors: models.ThemeColors{Primary: "#222222"}})
	if err != nil {
		t.Fatalf("Create 乙: %v", err)
	}

	if err := forge.Activate(first.ID, ""); err != nil {
		t.Fatalf("Activate 甲: %v", err)
	}
	effective, err := forge.Effective(time.Now())
	if err != nil || effective == nil || effective.ID != first.ID {
		t.Fatalf("Effective after activating 甲 = %+v, %v", effective, err)
	}

	if err := forge.Activate(second.ID, ""); err != nil {
		t.Fatalf("Activate 乙: %v", err)
	}
	effective, err = forge.Effective(time.Now())
	if err != nil || effective == nil || effective.ID != second.ID {
		t.Fatalf("Effective after activating 乙 in the same minute = %+v, %v", effective, err)
	}

	themes, err := forge.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, theme := range themes {
		if len(theme.EnabledAt) != len("2026-01-02T15:04") {
			t.Fatalf("stored 启用时间 %q is not minute precision", theme.EnabledAt)
		}
	}
}

func TestThemeForgeLifecycle(t *testing.T) {
	forge := newTestThemeForge(t)

	created, err := forge.Create(models.Theme{
		Name:      "开年红",
		EnabledAt: "2026-03-01T09:00",
		Colors:    models.ThemeColors{Primary: "#c0392b"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == "" {
		t.Fatal("Create did not assign an id")
	}

	if effective, err := forge.Effective(at("2026-02-28T10:00")); err != nil || effective != nil {
		t.Fatalf("Effective before schedule = %v, %v; want none", stringOf(effective), err)
	}
	effective, err := forge.Effective(at("2026-03-01T09:01"))
	if err != nil {
		t.Fatalf("Effective: %v", err)
	}
	if effective == nil || effective.Colors.Primary != "#c0392b" {
		t.Fatalf("Effective = %+v, want the scheduled theme", effective)
	}

	if err := forge.Activate(created.ID, "2026-02-01T08:00"); err != nil {
		t.Fatalf("Activate: %v", err)
	}
	themes, err := forge.List()
	if err != nil || len(themes) != 1 || themes[0].EnabledAt != "2026-02-01T08:00" {
		t.Fatalf("List after Activate = %+v, %v", themes, err)
	}

	if err := forge.Update(models.Theme{ID: created.ID, Name: "开年红", Colors: models.ThemeColors{Primary: "#0f62fe"}}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	themes, _ = forge.List()
	if themes[0].Colors.Primary != "#0f62fe" || themes[0].EnabledAt != "2026-02-01T08:00" {
		t.Fatalf("Update lost fields: %+v", themes[0])
	}

	if err := forge.Delete(created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if themes, err := forge.List(); err != nil || len(themes) != 0 {
		t.Fatalf("List after Delete = %+v, %v", themes, err)
	}
	if err := forge.Delete(created.ID); err == nil {
		t.Fatal("Delete of unknown id should fail")
	}
}

func TestThemeForgeValidation(t *testing.T) {
	forge := newTestThemeForge(t)

	cases := []struct {
		name  string
		input models.Theme
	}{
		{name: "missing name", input: models.Theme{}},
		{name: "bad color", input: models.Theme{Name: "坏颜色", Colors: models.ThemeColors{Primary: "not-a-hex"}}},
		{name: "bad time", input: models.Theme{Name: "坏时间", EnabledAt: "2026-13-45T99:99"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := forge.Create(tc.input); err == nil {
				t.Fatal("Create should reject invalid input")
			}
		})
	}

	if _, err := forge.Create(models.Theme{Name: "同一名称"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := forge.Create(models.Theme{Name: "同一名称"}); err == nil {
		t.Fatal("duplicate name should be rejected")
	}
}
