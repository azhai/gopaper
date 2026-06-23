package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/azhai/gopaper/internal/common"
	"github.com/azhai/gopaper/internal/model"

	"log/slog"

	"github.com/pelletier/go-toml/v2"
)

type Scanner struct {
	contentDir string
	logger     *slog.Logger
}

func NewScanner(contentDir string, logger *slog.Logger) *Scanner {
	return &Scanner{
		contentDir: contentDir,
		logger:     logger,
	}
}

func (s *Scanner) ScanAll(ctx context.Context) (*model.SiteTree, []model.Article, map[string]model.DirMeta, error) {
	rootMeta := s.readDirMeta(s.contentDir)
	dirMetas := map[string]model.DirMeta{"": rootMeta}

	var articles []model.Article
	siteTree := &model.SiteTree{
		Title:   rootMeta.SITE_TITLE,
		Slug:    "",
		DirPath: "",
		Meta:    &rootMeta,
	}

	err := s.scanDir(ctx, s.contentDir, "", rootMeta, siteTree, &articles, dirMetas)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("scan dir: %w", err)
	}

	sortArticles(articles)

	return siteTree, articles, dirMetas, nil
}

func (s *Scanner) scanDir(ctx context.Context, absDir, relDir string, parentMeta model.DirMeta, tree *model.SiteTree, articles *[]model.Article, dirMetas map[string]model.DirMeta) error {
	entries, err := os.ReadDir(absDir)
	if err != nil {
		s.logger.Warn("read dir failed", "dir", absDir, "error", err)
		return nil
	}

	dirMeta := s.readDirMeta(absDir)
	mergedMeta := model.MergeDirMeta(parentMeta, dirMeta)
	dirMetas[relDir] = mergedMeta

	if tree.Title == "" || tree.Title == filepath.Base(relDir) {
		if mergedMeta.SITE_TITLE != "" {
			tree.Title = mergedMeta.SITE_TITLE
		} else if tree.Title == "" {
			tree.Title = filepath.Base(relDir)
			if relDir == "" {
				tree.Title = mergedMeta.SITE_TITLE
			}
		}
	}
	tree.Slug = common.GenerateSlug(filepath.Base(relDir))
	tree.DirPath = relDir
	tree.Meta = &mergedMeta

	for _, entry := range entries {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		name := entry.Name()
		if strings.HasPrefix(name, ".") || name == "_meta.toml" {
			continue
		}

		fullPath := filepath.Join(absDir, name)

		if entry.IsDir() {
			childRelDir := filepath.Join(relDir, name)
			childTree := &model.SiteTree{
				Title: name,
			}
			err := s.scanDir(ctx, fullPath, childRelDir, mergedMeta, childTree, articles, dirMetas)
			if err != nil {
				return err
			}
			if !mergedMeta.NAV_HIDE {
				tree.Children = append(tree.Children, childTree)
			}
			continue
		}

		if !strings.HasSuffix(name, ".md") {
			continue
		}

		article, err := s.parseFile(fullPath, relDir)
		if err != nil {
			s.logger.Warn("parse file failed",
				"file", fullPath,
				"error", err,
			)
			continue
		}
		*articles = append(*articles, *article)
		tree.Articles = append(tree.Articles, article)
	}

	sortSiteTreeChildren(tree.Children)
	return nil
}

func (s *Scanner) parseFile(filePath, dirPath string) (*model.Article, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	metaData, content := s.parseMetaData(data)

	slug := metaData.Slug
	if slug == "" {
		slug = common.GenerateSlug(filepath.Base(filePath))
	}

	comments := true
	if metaData.Comments != nil {
		comments = *metaData.Comments
	}

	article := &model.Article{
		Slug:     slug,
		Title:    metaData.Title,
		Author:   metaData.Author,
		Date:     metaData.Date,
		Tags:     metaData.Tags,
		Comments: comments,
		Weight:   metaData.Weight,
		Position: metaData.Position,
		Content:  string(content),
		DirPath:  dirPath,
		FilePath: filePath,
	}

	if article.Title == "" {
		article.Title = slug
	}

	return article, nil
}

func (s *Scanner) parseMetaData(data []byte) (model.MetaData, []byte) {
	var meta model.MetaData
	content := data

	str := string(data)
	if !strings.HasPrefix(str, "+++\n") && !strings.HasPrefix(str, "+++\r\n") {
		return meta, content
	}

	endIdx := strings.Index(str[4:], "+++")
	if endIdx < 0 {
		return meta, content
	}

	tomlContent := str[4 : endIdx+4]
	content = []byte(str[endIdx+7:])

	if err := toml.Unmarshal([]byte(tomlContent), &meta); err != nil {
		s.logger.Warn("parse toml meta failed", "error", err)
		return meta, content
	}

	return meta, content
}

func (s *Scanner) readDirMeta(dirPath string) model.DirMeta {
	metaPath := filepath.Join(dirPath, "_meta.toml")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return model.DefaultDirMeta()
	}

	var meta model.DirMeta
	if err := toml.Unmarshal(data, &meta); err != nil {
		s.logger.Warn("parse _meta.toml failed", "path", metaPath, "error", err)
		return model.DefaultDirMeta()
	}

	if meta.ARTICLES_PER_PAGE == 0 {
		meta.ARTICLES_PER_PAGE = 10
	}

	return meta
}

func sortArticles(articles []model.Article) {
	sort.SliceStable(articles, func(i, j int) bool {
		if articles[i].Weight != articles[j].Weight {
			return articles[i].Weight > articles[j].Weight
		}
		if articles[i].Date != articles[j].Date {
			return articles[i].Date > articles[j].Date
		}
		return articles[i].Title < articles[j].Title
	})
}

func sortSiteTreeChildren(children []*model.SiteTree) {
	sort.SliceStable(children, func(i, j int) bool {
		mi := children[i].Meta
		mj := children[j].Meta
		oi, oj := 0, 0
		if mi != nil {
			oi = mi.NAV_ORDER
		}
		if mj != nil {
			oj = mj.NAV_ORDER
		}
		if oi != oj {
			return oi < oj
		}
		return children[i].Title < children[j].Title
	})
}
