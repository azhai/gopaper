package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/azhai/gopaper/internal/common"
	"github.com/azhai/gopaper/internal/model"

	"log/slog"
)

type ArticleForge struct {
	contentDir string
	cache      *CacheVault
	scanner    *Scanner
	logger     *slog.Logger
}

func NewArticleForge(contentDir string, cache *CacheVault, scanner *Scanner, logger *slog.Logger) *ArticleForge {
	return &ArticleForge{
		contentDir: contentDir,
		cache:      cache,
		scanner:    scanner,
		logger:     logger,
	}
}

func (af *ArticleForge) Create(ctx context.Context, input model.ArticleInput) error {
	errors := common.ValidateArticleInput(input.DirPath, input.Title, input.Slug, input.Author, input.Date, input.Tags)
	if len(errors) > 0 {
		return &ValidationError{Errors: errors}
	}

	slug := input.Slug
	if slug == "" {
		slug = common.GenerateSlug(input.Title)
	}

	if af.checkSlugConflict(input.DirPath, slug) {
		return &ConflictError{Message: "该Slug已存在，请更换"}
	}

	dirPath := filepath.Join(af.contentDir, input.DirPath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	fileName := slug + ".md"
	filePath := filepath.Join(dirPath, fileName)

	content := af.buildFileContent(input, slug)

	if err := af.writeAtomic(filePath, []byte(content)); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	af.logger.Info("article created", "slug", slug, "dir", input.DirPath)

	return af.cache.Refresh(ctx, af.scanner)
}

func (af *ArticleForge) Update(ctx context.Context, slug string, input model.ArticleInput) error {
	article, ok := af.cache.GetArticle(slug)
	if !ok {
		return &NotFoundError{Message: "文章不存在"}
	}

	if input.Title != "" {
		article.Title = input.Title
	}
	if input.Author != "" {
		article.Author = input.Author
	}
	if input.Date != "" {
		article.Date = input.Date
	}
	if input.Tags != nil {
		article.Tags = input.Tags
	}
	if input.Comments != nil {
		article.Comments = *input.Comments
	}
	if input.Weight != 0 {
		article.Weight = input.Weight
	}
	article.Position = input.Position
	if input.Content != "" {
		article.Content = input.Content
	}

	content := af.buildFileContent(model.ArticleInput{
		DirPath:  article.DirPath,
		Title:    article.Title,
		Slug:     article.Slug,
		Author:   article.Author,
		Date:     article.Date,
		Tags:     article.Tags,
		Comments: &article.Comments,
		Weight:   article.Weight,
		Position: article.Position,
		Content:  article.Content,
	}, article.Slug)

	if err := af.writeAtomic(article.FilePath, []byte(content)); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	af.logger.Info("article updated", "slug", slug)

	return af.cache.Refresh(ctx, af.scanner)
}

func (af *ArticleForge) Delete(ctx context.Context, slug string) error {
	article, ok := af.cache.GetArticle(slug)
	if !ok {
		return &NotFoundError{Message: "文章不存在"}
	}

	if err := os.Remove(article.FilePath); err != nil {
		return fmt.Errorf("remove file: %w", err)
	}

	af.logger.Info("article deleted", "slug", slug)

	return af.cache.Refresh(ctx, af.scanner)
}

func (af *ArticleForge) GetBySlug(ctx context.Context, slug string) (*model.Article, error) {
	article, ok := af.cache.GetArticle(slug)
	if !ok {
		return nil, &NotFoundError{Message: "文章不存在"}
	}
	return article, nil
}

func (af *ArticleForge) ListByDir(ctx context.Context, dirPath string, page, pageSize int) ([]*model.Article, int, error) {
	articles, ok := af.cache.GetArticleList(dirPath)
	if !ok {
		return nil, 0, nil
	}

	total := len(articles)
	start := (page - 1) * pageSize
	if start >= total {
		return nil, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	return articles[start:end], total, nil
}

func (af *ArticleForge) ListAll(ctx context.Context, page, pageSize int) ([]*model.Article, int, error) {
	articles := af.cache.GetAllArticles()
	total := len(articles)
	start := (page - 1) * pageSize
	if start >= total {
		return nil, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return articles[start:end], total, nil
}

// ListByType returns articles whose directory DIR_TYPE matches the given filter.
// typeFilter "page" returns only page-type articles; "article" returns non-page articles.
func (af *ArticleForge) ListByType(ctx context.Context, typeFilter string, page, pageSize int) ([]*model.Article, int, error) {
	all := af.cache.GetAllArticles()
	filtered := make([]*model.Article, 0, len(all))
	for _, a := range all {
		dirType := af.cache.GetDirType(a.DirPath)
		match := false
		if typeFilter == "page" {
			match = dirType == "page"
		} else { // "article"
			match = dirType != "page"
		}
		if match {
			filtered = append(filtered, a)
		}
	}

	total := len(filtered)
	start := (page - 1) * pageSize
	if start >= total {
		return nil, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return filtered[start:end], total, nil
}

func (af *ArticleForge) checkSlugConflict(dirPath, slug string) bool {
	articles, ok := af.cache.GetArticleList(dirPath)
	if !ok {
		return false
	}
	for _, a := range articles {
		if a.Slug == slug {
			return true
		}
	}
	return false
}

func (af *ArticleForge) writeAtomic(filePath string, content []byte) error {
	tmpPath := filePath + ".tmp"
	if err := os.WriteFile(tmpPath, content, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, filePath)
}

func (af *ArticleForge) buildFileContent(input model.ArticleInput, slug string) string {
	var sb strings.Builder
	sb.WriteString("+++\n")
	sb.WriteString(fmt.Sprintf("title = %q\n", input.Title))
	sb.WriteString(fmt.Sprintf("slug = %q\n", slug))
	if input.Author != "" {
		sb.WriteString(fmt.Sprintf("author = %q\n", input.Author))
	}
	if input.Date != "" {
		sb.WriteString(fmt.Sprintf("date = %q\n", input.Date))
	} else {
		sb.WriteString(fmt.Sprintf("date = %q\n", time.Now().Format("2006-01-02")))
	}
	if len(input.Tags) > 0 {
		tags := make([]string, len(input.Tags))
		for i, t := range input.Tags {
			tags[i] = fmt.Sprintf("%q", t)
		}
		sb.WriteString(fmt.Sprintf("tags = [%s]\n", strings.Join(tags, ", ")))
	}
	comments := true
	if input.Comments != nil {
		comments = *input.Comments
	}
	sb.WriteString(fmt.Sprintf("comments = %v\n", comments))
	if input.Weight != 0 {
		sb.WriteString(fmt.Sprintf("weight = %d\n", input.Weight))
	}
	if input.Position != "" {
		sb.WriteString(fmt.Sprintf("position = %q\n", input.Position))
	}
	sb.WriteString("+++\n\n")
	sb.WriteString(input.Content)
	return sb.String()
}

type ValidationError struct {
	Errors map[string]string
}

func (e *ValidationError) Error() string {
	return "校验失败"
}

type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

type ConflictError struct {
	Message string
}

func (e *ConflictError) Error() string {
	return e.Message
}
