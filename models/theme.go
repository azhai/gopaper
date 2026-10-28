package models

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

// Theme is a site-wide design preset covering layout, palette and typography.
// Themes are stored in content/_themes.toml and one of them becomes effective
// at its EnabledAt moment, which is how scheduled activation works.
type Theme struct {
	ID        string      `toml:"id" json:"id"`
	Name      string      `toml:"name" json:"name"`
	EnabledAt string      `toml:"enabled_at" json:"enabledAt"` // local "2006-01-02T15:04"
	Layout    ThemeLayout `toml:"layout" json:"layout"`
	Colors    ThemeColors `toml:"colors" json:"colors"`
	Fonts     ThemeFonts  `toml:"fonts" json:"fonts"`
}

type ThemeLayout struct {
	Container string `toml:"container" json:"container"` // narrow, default, wide, full
	Radius    string `toml:"radius" json:"radius"`       // sharp, soft, round, pill
	Header    string `toml:"header" json:"header"`       // gradient, solid, plain
	Hero      string `toml:"hero" json:"hero"`           // gradient, solid, plain
	Position  string `toml:"position" json:"position"`   // sticky, static
	Shadow    string `toml:"shadow" json:"shadow"`       // flat, soft, strong
}

type ThemeColors struct {
	Primary    string `toml:"primary" json:"primary"`
	Accent     string `toml:"accent" json:"accent"`
	Background string `toml:"background" json:"background"`
	Surface    string `toml:"surface" json:"surface"`
	Text       string `toml:"text" json:"text"`
	Muted      string `toml:"muted" json:"muted"`
	Border     string `toml:"border" json:"border"`
}

type ThemeFonts struct {
	Body    string `toml:"body" json:"body"`
	Heading string `toml:"heading" json:"heading"`
	Size    int    `toml:"size" json:"size"`
}

type ThemeFile struct {
	Themes []Theme `toml:"themes"`
}

// EnabledTheme describes which preset the site shows right now and which one
// takes over next, so the admin can tell whether a schedule is pending.
type EnabledTheme struct {
	CurrentID  string     `json:"currentId"`
	NextID     string     `json:"nextId"`
	NextAt     string     `json:"nextAt"`
	DefaultCSS string     `json:"defaultCss"`
	Themes     []ThemeCSS `json:"themes"`
	Presets    []Theme    `json:"presets"`
}

// ThemeCSS is a stored theme plus the CSS variable block it renders into.
type ThemeCSS struct {
	Theme
	CSS string `json:"css"`
}

var ThemeContainers = map[string]string{
	"narrow":  "880px",
	"default": "1120px",
	"wide":    "1320px",
	"full":    "1560px",
}

var ThemeRadii = map[string][2]string{
	"sharp": {"8px", "6px"},
	"soft":  {"16px", "12px"},
	"round": {"24px", "16px"},
	"pill":  {"32px", "20px"},
}

var ThemeFontStacks = map[string]string{
	"sans":  `"PingFang SC", "Microsoft YaHei", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Noto Sans SC", sans-serif`,
	"serif": `Georgia, "Songti SC", "STSong", "SimSun", "Noto Serif SC", serif`,
	"kai":   `"Kaiti SC", STKaiti, KaiTi, "Songti SC", Georgia, serif`,
	"mono":  `"SFMono-Regular", Consolas, "Courier New", "PingFang SC", monospace`,
}

var themeEnumDefaults = map[string]string{
	"container": "default",
	"radius":    "soft",
	"header":    "gradient",
	"hero":      "gradient",
	"position":  "sticky",
	"shadow":    "soft",
}

func DefaultThemeColors() ThemeColors {
	return ThemeColors{
		Primary:    "#6750a4",
		Accent:     "#625b71",
		Background: "#fffbfe",
		Surface:    "#ffffff",
		Text:       "#1c1b1f",
		Muted:      "#49454f",
		Border:     "#cac4d0",
	}
}

func DefaultTheme() Theme {
	return Theme{
		Name:   "Material 3",
		Layout: ThemeLayout{Container: "default", Radius: "soft", Header: "gradient", Hero: "gradient", Position: "sticky", Shadow: "soft"},
		Colors: DefaultThemeColors(),
		Fonts:  ThemeFonts{Body: "sans", Heading: "sans", Size: 15},
	}
}

// PresetThemes returns ready-to-use presets for the admin designer.
func PresetThemes() []Theme {
	macaron := DefaultTheme()
	macaron.ID, macaron.Name = "preset-macaron", "马卡龙"
	macaron.Colors = ThemeColors{Primary: "#a5536f", Accent: "#85405a", Background: "#fdf7f8", Surface: "#ffffff", Text: "#382e32", Muted: "#7a6c71", Border: "#f0e2e6"}
	macaron.Layout.Radius, macaron.Layout.Container = "pill", "wide"

	monet := DefaultTheme()
	monet.ID, monet.Name = "preset-monet", "莫奈"
	monet.Colors = ThemeColors{Primary: "#3f6f70", Accent: "#32595a", Background: "#f7faf8", Surface: "#fffdf9", Text: "#2a3736", Muted: "#667875", Border: "#dce7e3"}
	monet.Layout.Container = "narrow"
	monet.Fonts.Heading = "serif"

	red := DefaultTheme()
	red.ID, red.Name = "preset-red", "节庆红"
	red.Colors = ThemeColors{Primary: "#b23b30", Accent: "#922a22", Background: "#fdf9f8", Surface: "#ffffff", Text: "#332b2a", Muted: "#766a68", Border: "#ece1de"}
	red.Layout.Radius, red.Layout.Container = "round", "wide"
	red.Fonts.Heading = "serif"

	return []Theme{DefaultTheme(), macaron, monet, red}
}

// Normalize fills empty tokens with the built-in defaults so stored themes are
// always complete and the CSS renderer needs no nil handling.
func (t *Theme) Normalize() {
	if t.Colors.Primary = cleanHex(t.Colors.Primary); t.Colors.Primary == "" {
		t.Colors.Primary = DefaultThemeColors().Primary
	}
	if t.Colors.Accent = cleanHex(t.Colors.Accent); t.Colors.Accent == "" {
		t.Colors.Accent = DefaultThemeColors().Accent
	}
	if t.Colors.Background = cleanHex(t.Colors.Background); t.Colors.Background == "" {
		t.Colors.Background = DefaultThemeColors().Background
	}
	if t.Colors.Surface = cleanHex(t.Colors.Surface); t.Colors.Surface == "" {
		t.Colors.Surface = DefaultThemeColors().Surface
	}
	if t.Colors.Text = cleanHex(t.Colors.Text); t.Colors.Text == "" {
		t.Colors.Text = DefaultThemeColors().Text
	}
	if t.Colors.Muted = cleanHex(t.Colors.Muted); t.Colors.Muted == "" {
		t.Colors.Muted = DefaultThemeColors().Muted
	}
	if t.Colors.Border = cleanHex(t.Colors.Border); t.Colors.Border == "" {
		t.Colors.Border = DefaultThemeColors().Border
	}

	t.Layout.Container = themeEnum("container", t.Layout.Container)
	t.Layout.Radius = themeEnum("radius", t.Layout.Radius)
	t.Layout.Header = themeEnum("header", t.Layout.Header)
	t.Layout.Hero = themeEnum("hero", t.Layout.Hero)
	t.Layout.Position = themeEnum("position", t.Layout.Position)
	t.Layout.Shadow = themeEnum("shadow", t.Layout.Shadow)

	t.Fonts.Body = themeFont(t.Fonts.Body)
	t.Fonts.Heading = themeFont(t.Fonts.Heading)
	if t.Fonts.Size < 14 || t.Fonts.Size > 20 {
		t.Fonts.Size = 15
	}
}

func themeEnum(kind, value string) string {
	def := themeEnumDefaults[kind]
	switch kind {
	case "container":
		if _, ok := ThemeContainers[value]; ok {
			return value
		}
	case "radius":
		if _, ok := ThemeRadii[value]; ok {
			return value
		}
	case "header", "hero":
		if value == "gradient" || value == "solid" || value == "plain" {
			return value
		}
	case "position":
		if value == "sticky" || value == "static" {
			return value
		}
	case "shadow":
		if value == "flat" || value == "soft" || value == "strong" {
			return value
		}
	}
	return def
}

func themeFont(value string) string {
	if _, ok := ThemeFontStacks[value]; ok {
		return value
	}
	return "sans"
}

// CSS renders the theme as a CSS custom property block. Values come from the
// validated token sets above, so the output contains no free-form text.
func (t Theme) CSS() string {
	c := t.Colors
	primarySoft := lighten(c.Primary, 0.3)
	primaryLight := lighten(c.Primary, 0.55)
	accentSoft := lighten(c.Accent, 0.3)
	headerTint := mix(primarySoft, c.Surface, 0.2)
	headerBase := bandBase(t.Layout.Header, headerTint, c)
	heroBase := bandBase(t.Layout.Hero, mix(headerTint, c.Surface, 0.5), c)
	footerBg := darken(c.Text, 0.25)

	radius := ThemeRadii[t.Layout.Radius]
	vars := []string{
		"--c-bg:" + c.Background,
		"--c-surface:" + c.Surface,
		"--c-surface-soft:" + mix(c.Primary, c.Background, 0.9),
		"--c-text:" + c.Text,
		"--c-text-rgb:" + rgbTriplet(c.Text),
		"--c-muted:" + c.Muted,
		"--c-faint:" + mix(c.Muted, c.Background, 0.25),
		"--c-meta:" + mix(c.Primary, c.Muted, 0.4),
		"--c-primary:" + c.Primary,
		"--c-primary-rgb:" + rgbTriplet(c.Primary),
		"--c-primary-ink:" + darken(c.Primary, 0.2),
		"--c-primary-soft:" + primarySoft,
		"--c-primary-light:" + primaryLight,
		"--c-primary-on-dark:" + lighten(c.Primary, 0.55),
		"--c-tint:" + mix(c.Primary, c.Surface, 0.55),
		"--c-tint-soft:" + mix(c.Primary, c.Surface, 0.75),
		"--c-accent:" + c.Accent,
		"--c-accent-soft:" + accentSoft,
		"--c-border:" + c.Border,
		"--c-border-strong:" + mix(c.Primary, c.Surface, 0.45),
		"--c-footer-bg:" + footerBg,
		"--c-footer-rgb:" + rgbTriplet(footerBg),
		"--c-footer-text:" + mix(lighten(c.Primary, 0.55), c.Muted, 0.3),
		"--header-bg:" + headerSurface(t.Layout.Header, headerTint, primaryLight, c),
		"--header-text:" + readable(headerBase, c.Text),
		"--header-border:" + headerBorder(t.Layout.Header, c),
		"--header-pos:" + position(t.Layout.Position),
		"--hero-bg:" + heroSurface(t.Layout.Hero, headerTint, c),
		"--hero-text:" + readable(heroBase, c.Text),
		"--shadow-rgb:" + rgbTriplet(primarySoft),
		"--shadow-a:" + shadowAlpha(t.Layout.Shadow),
		"--container:" + ThemeContainers[t.Layout.Container],
		"--radius:" + radius[0],
		"--radius-sm:" + radius[1],
		"--font-body:" + ThemeFontStacks[t.Fonts.Body],
		"--font-heading:" + ThemeFontStacks[t.Fonts.Heading],
		"--font-size:" + strconv.Itoa(t.Fonts.Size) + "px",
	}

	return ":root{" + strings.Join(vars, ";") + ";}"
}

func headerSurface(style, softTint, lighterTint string, c ThemeColors) string {
	switch style {
	case "solid":
		return c.Primary
	case "plain":
		return c.Surface
	default:
		return fmt.Sprintf("linear-gradient(90deg,%s 0%%,%s 100%%)", softTint, lighterTint)
	}
}

func heroSurface(style, softTint string, c ThemeColors) string {
	switch style {
	case "solid":
		return c.Primary
	case "plain":
		return c.Background
	default:
		mid := mix(softTint, c.Surface, 0.55)
		return fmt.Sprintf("linear-gradient(135deg,%s 0%%,%s 55%%,%s 100%%)", softTint, mid, c.Surface)
	}
}

// bandBase is the dominant colour of a header/hero band, used for contrast.
func bandBase(style, gradientMid string, c ThemeColors) string {
	switch style {
	case "solid":
		return c.Primary
	case "plain":
		return c.Surface
	default:
		return gradientMid
	}
}

func headerBorder(style string, c ThemeColors) string {
	if style == "plain" {
		return "1px solid " + c.Border
	}
	return "0 solid transparent"
}

func position(pos string) string {
	if pos == "static" {
		return "static"
	}
	return "sticky"
}

func shadowAlpha(strength string) string {
	switch strength {
	case "flat":
		return "0"
	case "strong":
		return "1.8"
	default:
		return "1"
	}
}

// ValidThemeHex returns the canonical #rrggbb form of value, or "" when the
// value is not a CSS colour hex.
func ValidThemeHex(value string) string { return cleanHex(value) }

func cleanHex(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if !strings.HasPrefix(value, "#") {
		return ""
	}
	digits := value[1:]
	switch len(digits) {
	case 3:
		for _, r := range digits {
			if !isHexDigit(byte(r)) {
				return ""
			}
		}
		var expanded strings.Builder
		expanded.WriteByte('#')
		for _, r := range digits {
			expanded.WriteRune(r)
			expanded.WriteRune(r)
		}
		return expanded.String()
	case 6:
		for _, r := range digits {
			if !isHexDigit(byte(r)) {
				return ""
			}
		}
		return value
	default:
		return ""
	}
}

func isHexDigit(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f')
}

func parseHex(value string) color.RGBA {
	rgb, _ := strconv.ParseUint(value[1:], 16, 32)
	return color.RGBA{R: uint8(rgb >> 16), G: uint8(rgb >> 8), B: uint8(rgb), A: 255}
}

func toHex(c color.RGBA) string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

// mix blends a toward b; weight is the share of b.
func mix(a, b string, weight float64) string {
	x, y := parseHex(a), parseHex(b)
	blended := color.RGBA{
		R: uint8(float64(x.R) + (float64(y.R)-float64(x.R))*weight),
		G: uint8(float64(x.G) + (float64(y.G)-float64(x.G))*weight),
		B: uint8(float64(x.B) + (float64(y.B)-float64(x.B))*weight),
		A: 255,
	}
	return toHex(blended)
}

func lighten(value string, weight float64) string { return mix(value, "#ffffff", weight) }
func darken(value string, weight float64) string  { return mix(value, "#000000", weight) }

func rgbTriplet(value string) string {
	c := parseHex(value)
	return strconv.Itoa(int(c.R)) + "," + strconv.Itoa(int(c.G)) + "," + strconv.Itoa(int(c.B))
}

// readable picks the text colour that keeps contrast on the given background.
func readable(background, fallback string) string {
	c := parseHex(background)
	luminance := (0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B)) / 255
	if luminance < 0.55 {
		return "#fffdf7"
	}
	return fallback
}
