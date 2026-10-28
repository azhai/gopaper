package models

import (
	"strings"
	"testing"
)

func TestThemeCSSTokens(t *testing.T) {
	theme := DefaultTheme()
	theme.Normalize()
	css := theme.CSS()

	for _, want := range []string{
		":root{", "--c-bg:#fffbfe", "--c-primary:#6750a4", "--c-text-rgb:28,27,31",
		"--container:1120px", "--radius:16px", "--radius-sm:12px",
		"--font-size:15px", "--header-pos:sticky", "--shadow-a:1", "linear-gradient(",
	} {
		if !strings.Contains(css, want) {
			t.Fatalf("CSS() missing %q in %s", want, css)
		}
	}
	if strings.Contains(css, "<") {
		t.Fatalf("CSS() must not be able to break out of a style tag: %s", css)
	}
}

func TestPresetThemes(t *testing.T) {
	presets := PresetThemes()
	if len(presets) != 4 {
		t.Fatalf("PresetThemes() returned %d presets, want 4", len(presets))
	}
	want := []string{"Material 3", "马卡龙", "莫奈", "节庆红"}
	for i, name := range want {
		if presets[i].Name != name {
			t.Errorf("preset %d = %q, want %q", i, presets[i].Name, name)
		}
	}
	red := presets[3]
	if red.Colors.Primary != "#b23b30" || red.Colors.Accent != "#922a22" {
		t.Errorf("red preset colors = %+v", red.Colors)
	}
	for _, p := range presets {
		p.Normalize()
		if css := p.CSS(); !strings.Contains(css, "--c-primary:"+p.Colors.Primary) {
			t.Errorf("preset %q CSS missing primary %s", p.Name, p.Colors.Primary)
		}
	}
}

func TestThemeCSSReflectsTokens(t *testing.T) {
	theme := DefaultTheme()
	theme.Colors.Primary = "#1c3f5e"
	theme.Layout.Container = "wide"
	theme.Layout.Radius = "sharp"
	theme.Layout.Position = "static"
	theme.Layout.Shadow = "flat"
	theme.Layout.Header = "solid"
	theme.Fonts.Heading = "serif"
	theme.Fonts.Size = 18
	theme.Normalize()

	css := theme.CSS()
	for _, want := range []string{
		"--c-primary:#1c3f5e", "--container:1320px", "--radius:8px",
		"--header-pos:static", "--shadow-a:0", "--font-size:18px",
		"--header-bg:#1c3f5e",
	} {
		if !strings.Contains(css, want) {
			t.Fatalf("CSS() missing %q in %s", want, css)
		}
	}
	if !strings.Contains(css, "Georgia") {
		t.Fatalf("heading font not applied: %s", css)
	}
	// dark primary means light text on the header band
	if !strings.Contains(css, "--header-text:#fffdf7") {
		t.Fatalf("expected light header text: %s", css)
	}
}

func TestThemeNormalizeFallsBackToDefaults(t *testing.T) {
	defaults := DefaultTheme()
	theme := Theme{
		Layout: ThemeLayout{Container: "huge", Radius: "square", Header: "sidebar", Hero: "gradient", Position: "floating", Shadow: "mega"},
		Colors: ThemeColors{Primary: "red", Accent: "#zzzzzz", Text: "#0f0"},
		Fonts:  ThemeFonts{Body: "comic", Heading: "kai-ish", Size: 30},
	}
	theme.Normalize()

	if theme.Layout.Container != defaults.Layout.Container || theme.Layout.Radius != defaults.Layout.Radius {
		t.Fatalf("layout enums not reset: %+v", theme.Layout)
	}
	if theme.Layout.Header != defaults.Layout.Header || theme.Layout.Position != defaults.Layout.Position {
		t.Fatalf("layout enums not reset: %+v", theme.Layout)
	}
	if theme.Colors.Primary != defaults.Colors.Primary || theme.Colors.Accent != defaults.Colors.Accent {
		t.Fatalf("invalid hex not reset: %+v", theme.Colors)
	}
	if theme.Colors.Text != "#00ff00" {
		t.Fatalf("3-digit hex should expand, got %q", theme.Colors.Text)
	}
	if theme.Fonts.Body != "sans" || theme.Fonts.Heading != "sans" || theme.Fonts.Size != 15 {
		t.Fatalf("font tokens not reset: %+v", theme.Fonts)
	}
	if css := theme.CSS(); css == "" {
		t.Fatal("empty CSS after normalize")
	}
}
