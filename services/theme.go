package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/azhai/gopaper/models"

	"github.com/pelletier/go-toml/v2"
)

// ThemeForge owns the site theme presets in content/_themes.toml and resolves
// which preset is effective at a given moment, enabling scheduled activation.
type ThemeForge struct {
	path   string
	logger *slog.Logger

	mu       sync.Mutex
	cached   []models.Theme
	loadedOK bool
	loadedAt time.Time
	loadedSz int64
}

const themeTimeLayout = "2006-01-02T15:04"

func NewThemeForge(contentDir string, logger *slog.Logger) *ThemeForge {
	return &ThemeForge{path: filepath.Join(contentDir, "_themes.toml"), logger: logger}
}

// List returns every preset, re-reading the file when it changed on disk.
func (tf *ThemeForge) List() ([]models.Theme, error) {
	tf.mu.Lock()
	defer tf.mu.Unlock()
	if err := tf.reloadLocked(); err != nil {
		return nil, err
	}
	return append([]models.Theme(nil), tf.cached...), nil
}

// Effective returns the preset with the latest activation time not after now.
func (tf *ThemeForge) Effective(now time.Time) (*models.Theme, error) {
	themes, err := tf.List()
	if err != nil {
		return nil, err
	}
	return pickThemeByTime(themes, now, false), nil
}

// Next returns the earliest preset scheduled after now, if any.
func (tf *ThemeForge) Next(now time.Time) (*models.Theme, error) {
	themes, err := tf.List()
	if err != nil {
		return nil, err
	}
	return pickThemeByTime(themes, now, true), nil
}

func pickThemeByTime(themes []models.Theme, now time.Time, upcoming bool) *models.Theme {
	var found *models.Theme
	var foundAt time.Time
	for i := range themes {
		at, err := parseThemeTime(themes[i].EnabledAt)
		if err != nil {
			continue
		}
		match := !at.After(now)
		if upcoming {
			match = at.After(now)
		}
		if !match {
			continue
		}
		// Minute precision means two activations can tie; Activate appends the
		// theme it enables, so the later entry is the one the operator intended.
		if found == nil || (upcoming && at.Before(foundAt)) || (!upcoming && !at.Before(foundAt)) {
			theme := themes[i]
			found, foundAt = &theme, at
		}
	}
	return found
}

func (tf *ThemeForge) Create(input models.Theme) (models.Theme, error) {
	themes, err := tf.List()
	if err != nil {
		return models.Theme{}, err
	}
	if len(themes) >= 20 {
		return models.Theme{}, &ValidationError{Errors: map[string]string{"themes": "最多保存20套主题"}}
	}

	input.ID = ""
	if err := tf.validate(&input, themes); err != nil {
		return models.Theme{}, err
	}
	input.ID = newThemeID()
	if input.EnabledAt == "" {
		input.EnabledAt = time.Now().Format(themeTimeLayout)
	}

	themes = append(themes, input)
	if err := tf.save(themes); err != nil {
		return models.Theme{}, err
	}
	tf.logger.Info("theme created", "id", input.ID, "name", input.Name, "enabled_at", input.EnabledAt)
	return input, nil
}

func (tf *ThemeForge) Update(input models.Theme) error {
	themes, err := tf.List()
	if err != nil {
		return err
	}
	i := indexOfTheme(themes, input.ID)
	if i < 0 {
		return &NotFoundError{Message: "主题不存在"}
	}
	existing := themes[i]
	if err := tf.validate(&input, themes); err != nil {
		return err
	}
	if input.EnabledAt == "" {
		input.EnabledAt = existing.EnabledAt
	}
	input.ID = existing.ID
	themes[i] = input

	if err := tf.save(themes); err != nil {
		return err
	}
	tf.logger.Info("theme updated", "id", input.ID, "enabled_at", input.EnabledAt)
	return nil
}

func (tf *ThemeForge) Delete(id string) error {
	themes, err := tf.List()
	if err != nil {
		return err
	}
	i := indexOfTheme(themes, id)
	if i < 0 {
		return &NotFoundError{Message: "主题不存在"}
	}
	kept := append([]models.Theme{}, themes[:i]...)
	kept = append(kept, themes[i+1:]...)

	if err := tf.save(kept); err != nil {
		return err
	}
	tf.logger.Info("theme deleted", "id", id)
	return nil
}

// Activate schedules a preset to start now (empty when) or at the given time.
func (tf *ThemeForge) Activate(id, when string) error {
	themes, err := tf.List()
	if err != nil {
		return err
	}
	i := indexOfTheme(themes, id)
	if i < 0 {
		return &NotFoundError{Message: "主题不存在"}
	}
	at := strings.TrimSpace(when)
	if at == "" {
		at = time.Now().Format(themeTimeLayout)
	} else if _, err := parseThemeTime(at); err != nil {
		return &ValidationError{Errors: map[string]string{"enabledAt": "启用时间格式不正确"}}
	}
	themes[i].EnabledAt = at
	themes = moveThemeLast(themes, i)

	if err := tf.save(themes); err != nil {
		return err
	}
	tf.logger.Info("theme activated", "id", id, "enabled_at", at)
	return nil
}

func (tf *ThemeForge) validate(theme *models.Theme, all []models.Theme) error {
	theme.Name = strings.TrimSpace(theme.Name)
	if theme.Name == "" {
		return &ValidationError{Errors: map[string]string{"name": "主题名称不能为空"}}
	}
	if len([]rune(theme.Name)) > 40 {
		return &ValidationError{Errors: map[string]string{"name": "主题名称不超过40个字符"}}
	}

	colorFields := []struct{ key, value string }{
		{"primary", theme.Colors.Primary},
		{"accent", theme.Colors.Accent},
		{"background", theme.Colors.Background},
		{"surface", theme.Colors.Surface},
		{"text", theme.Colors.Text},
		{"muted", theme.Colors.Muted},
		{"border", theme.Colors.Border},
	}
	for _, field := range colorFields {
		if field.value != "" && models.ValidThemeHex(field.value) == "" {
			return &ValidationError{Errors: map[string]string{field.key: "颜色需写成 #rgb 或 #rrggbb"}}
		}
	}

	if theme.EnabledAt != "" {
		if _, err := parseThemeTime(theme.EnabledAt); err != nil {
			return &ValidationError{Errors: map[string]string{"enabledAt": "启用时间格式不正确"}}
		}
	}
	for _, other := range all {
		if other.ID != theme.ID && strings.EqualFold(strings.TrimSpace(other.Name), theme.Name) {
			return &ConflictError{Message: "已有同名主题，请换一个名称"}
		}
	}

	theme.Normalize()
	return nil
}

func (tf *ThemeForge) save(themes []models.Theme) error {
	for i := range themes {
		themes[i].Normalize()
	}
	out, err := toml.Marshal(&models.ThemeFile{Themes: themes})
	if err != nil {
		return fmt.Errorf("encode themes: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(tf.path), 0755); err != nil {
		return fmt.Errorf("create content dir: %w", err)
	}
	tmp := tf.path + ".tmp"
	if err := os.WriteFile(tmp, out, 0644); err != nil {
		return fmt.Errorf("write themes: %w", err)
	}
	if err := os.Rename(tmp, tf.path); err != nil {
		return fmt.Errorf("replace themes: %w", err)
	}

	tf.mu.Lock()
	defer tf.mu.Unlock()
	tf.cached = themes
	tf.memoStatLocked()
	return nil
}

func (tf *ThemeForge) reloadLocked() error {
	info, err := os.Stat(tf.path)
	if errors.Is(err, os.ErrNotExist) {
		tf.cached, tf.loadedOK = nil, false
		return nil
	}
	if err != nil {
		return fmt.Errorf("stat themes: %w", err)
	}
	if tf.loadedOK && info.ModTime().Equal(tf.loadedAt) && info.Size() == tf.loadedSz {
		return nil
	}

	data, err := os.ReadFile(tf.path)
	if err != nil {
		return fmt.Errorf("read themes: %w", err)
	}
	var file models.ThemeFile
	if err := toml.Unmarshal(data, &file); err != nil {
		tf.logger.Warn("parse _themes.toml failed", "path", tf.path, "error", err)
		tf.cached = nil
		tf.loadedOK = false
		return nil
	}
	for i := range file.Themes {
		file.Themes[i].Normalize()
	}
	tf.cached = file.Themes
	tf.memoStatLocked()
	return nil
}

func (tf *ThemeForge) memoStatLocked() {
	info, err := os.Stat(tf.path)
	if err != nil {
		tf.loadedOK = false
		return
	}
	tf.loadedOK, tf.loadedAt, tf.loadedSz = true, info.ModTime(), info.Size()
}

func parseThemeTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	for _, layout := range []string{themeTimeLayout, "2006-01-02T15:04:05", time.RFC3339, "2006-01-02"} {
		if at, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return at, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid theme time %q", value)
}

// moveThemeLast puts the activated theme at the end of the file so it wins any
// same-minute tie against a theme enabled earlier.
func moveThemeLast(themes []models.Theme, i int) []models.Theme {
	moved := make([]models.Theme, 0, len(themes))
	moved = append(moved, themes[:i]...)
	moved = append(moved, themes[i+1:]...)
	return append(moved, themes[i])
}

func indexOfTheme(themes []models.Theme, id string) int {
	if id == "" {
		return -1
	}
	for i := range themes {
		if themes[i].ID == id {
			return i
		}
	}
	return -1
}

func newThemeID() string {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return "t" + time.Now().Format("20060102150405")
	}
	return "t-" + hex.EncodeToString(buf)
}
