package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/azhai/gopaper/common"
	"github.com/azhai/gopaper/models"

	"log/slog"

	"github.com/pelletier/go-toml/v2"
)

// SectionForge manages navigation sections (top-level content directories).
type SectionForge struct {
	contentDir string
	cache      *CacheVault
	scanner    *Scanner
	logger     *slog.Logger
}

func NewSectionForge(contentDir string, cache *CacheVault, scanner *Scanner, logger *slog.Logger) *SectionForge {
	return &SectionForge{contentDir: contentDir, cache: cache, scanner: scanner, logger: logger}
}

var sectionReserved = map[string]bool{
	"api": true, "admin": true, "static": true, "uploads": true,
	"page": true, "livereload": true, "preview": true,
}

var sectionTypes = map[string]bool{"page": true, "news": true, "docs": true, "product": true}

func (sf *SectionForge) Create(ctx context.Context, input models.SectionInput) (string, error) {
	name := common.GenerateSlug(input.Name)
	if !common.ValidateSlug(name) {
		return "", &ValidationError{Errors: map[string]string{"name": "网址标识仅支持小写字母、数字和连字符"}}
	}
	if sectionReserved[name] {
		return "", &ValidationError{Errors: map[string]string{"name": "该网址标识为系统保留，请更换"}}
	}
	if strings.TrimSpace(input.Title) == "" {
		return "", &ValidationError{Errors: map[string]string{"title": "栏目名称不能为空"}}
	}
	if len(input.Title) > 60 {
		return "", &ValidationError{Errors: map[string]string{"title": "栏目名称不超过60个字符"}}
	}
	dirType := input.Type
	if !sectionTypes[dirType] {
		dirType = "page"
	}

	dir := filepath.Join(sf.contentDir, name)
	if _, err := os.Stat(dir); err == nil {
		return "", &ConflictError{Message: "该网址标识已被占用"}
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create section dir: %w", err)
	}

	meta := models.DirMeta{
		SITE_TITLE: input.Title,
		SITE_DESC:  input.Title,
		DIR_TYPE:   dirType,
		NAV_ORDER:  input.NavOrder,
		LAYOUT:     input.Layout,
		SORT_ORDER: input.SortOrder,
	}
	if err := sf.writeMeta(dir, meta); err != nil {
		return "", err
	}

	sf.logger.Info("section created", "dir", name, "type", dirType)
	if err := sf.cache.Refresh(ctx, sf.scanner); err != nil {
		return name, err
	}
	return name, nil
}

func (sf *SectionForge) Update(ctx context.Context, input models.SectionInput) error {
	if !sf.sectionExists(input.DirPath) {
		return &NotFoundError{Message: "栏目不存在"}
	}
	if strings.TrimSpace(input.Title) == "" {
		return &ValidationError{Errors: map[string]string{"title": "栏目名称不能为空"}}
	}

	dir := filepath.Join(sf.contentDir, input.DirPath)
	meta := sf.readMeta(dir)
	meta.SITE_TITLE = input.Title
	meta.NAV_ORDER = input.NavOrder
	meta.NAV_HIDE = input.NavHide
	if sectionTypes[input.Type] {
		meta.DIR_TYPE = input.Type
	}
	meta.LAYOUT = input.Layout
	meta.SORT_ORDER = input.SortOrder
	if err := sf.writeMeta(dir, meta); err != nil {
		return err
	}

	sf.logger.Info("section updated", "dir", input.DirPath)
	return sf.cache.Refresh(ctx, sf.scanner)
}

func (sf *SectionForge) sectionExists(dirPath string) bool {
	if dirPath == "" || strings.Contains(dirPath, "..") || strings.ContainsAny(dirPath, `/\`) {
		return false
	}
	for _, d := range sf.cache.GetDirs() {
		if d.DirPath == dirPath {
			return true
		}
	}
	return false
}

func (sf *SectionForge) readMeta(dir string) models.DirMeta {
	data, err := os.ReadFile(filepath.Join(dir, "_meta.toml"))
	if err != nil {
		return models.DefaultDirMeta()
	}
	var meta models.DirMeta
	if err := toml.Unmarshal(data, &meta); err != nil {
		return models.DefaultDirMeta()
	}
	if meta.ARTICLES_PER_PAGE == 0 {
		meta.ARTICLES_PER_PAGE = 10
	}
	return meta
}

func (sf *SectionForge) writeMeta(dir string, meta models.DirMeta) error {
	out, err := toml.Marshal(&meta)
	if err != nil {
		return fmt.Errorf("encode section meta: %w", err)
	}
	path := filepath.Join(dir, "_meta.toml")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, out, 0644); err != nil {
		return fmt.Errorf("write section meta: %w", err)
	}
	return os.Rename(tmp, path)
}
