package services

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/azhai/gopaper/common"
	"github.com/azhai/gopaper/models"

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

func (af *ArticleForge) Create(ctx context.Context, input models.ArticleInput) (string, error) {
	input = normalizeProductInput(input)
	errors := common.ValidateArticleInput(input.DirPath, input.Title, input.Slug, input.Author, input.Date, input.Tags)
	mergeErrors(errors, validateProductFields(input, af.cache.GetDirType(input.DirPath) == "product"))
	if len(errors) > 0 {
		return "", &ValidationError{Errors: errors}
	}

	slug, err := af.pickSlug(input.DirPath, input.Title, input.Slug)
	if err != nil {
		return "", err
	}

	dirPath := filepath.Join(af.contentDir, input.DirPath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("create dir: %w", err)
	}

	fileName := slug + ".md"
	filePath := filepath.Join(dirPath, fileName)

	content := af.buildFileContent(input, slug)

	if err := af.writeAtomic(filePath, []byte(content)); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	af.logger.Info("article created", "slug", slug, "dir", input.DirPath)

	if err := af.cache.Refresh(ctx, af.scanner); err != nil {
		return slug, err
	}
	return slug, nil
}

func (af *ArticleForge) Update(ctx context.Context, slug string, input models.ArticleInput) error {
	article, ok := af.cache.GetArticle(slug)
	if !ok {
		return &NotFoundError{Message: "文章不存在"}
	}

	input = normalizeProductInput(input)
	errors := common.ValidateArticleInput(article.DirPath, input.Title, "", input.Author, input.Date, input.Tags)
	delete(errors, "dirPath")
	mergeErrors(errors, validateProductFields(input, af.cache.GetDirType(article.DirPath) == "product"))
	if len(errors) > 0 {
		return &ValidationError{Errors: errors}
	}

	article.Title = input.Title
	article.Author = input.Author
	article.Date = input.Date
	article.Tags = input.Tags
	article.Comments = true
	if input.Comments != nil {
		article.Comments = *input.Comments
	}
	article.Weight = input.Weight
	article.Position = input.Position
	if input.Draft != nil {
		article.Draft = *input.Draft
	}
	article.Summary = input.Summary
	article.Category = input.Category
	article.Cover = input.Cover
	if len(input.Gallery) > 0 {
		article.Gallery = input.Gallery
	}
	article.Specs = input.Specs
	article.Improvements = input.Improvements
	article.Content = input.Content

	content := af.buildFileContent(models.ArticleInput{
		DirPath:      article.DirPath,
		Title:        article.Title,
		Slug:         article.Slug,
		Author:       article.Author,
		Date:         article.Date,
		Tags:         article.Tags,
		Comments:     &article.Comments,
		Weight:       article.Weight,
		Position:     article.Position,
		Draft:        &article.Draft,
		Summary:      article.Summary,
		Category:     article.Category,
		Cover:        article.Cover,
		Gallery:      article.Gallery,
		Specs:        article.Specs,
		Improvements: article.Improvements,
		Content:      article.Content,
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

func (af *ArticleForge) GetBySlug(ctx context.Context, slug string) (*models.Article, error) {
	article, ok := af.cache.GetArticle(slug)
	if !ok {
		return nil, &NotFoundError{Message: "文章不存在"}
	}
	return article, nil
}

func (af *ArticleForge) ListByDir(ctx context.Context, dirPath string, page, pageSize int) ([]*models.Article, int, error) {
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

func (af *ArticleForge) ListAll(ctx context.Context, page, pageSize int) ([]*models.Article, int, error) {
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
// typeFilter "page" returns page-type articles; "product" returns product pages;
// "article" returns everything else.
func (af *ArticleForge) ListByType(ctx context.Context, typeFilter string, page, pageSize int) ([]*models.Article, int, error) {
	all := af.cache.GetAllArticles()
	filtered := make([]*models.Article, 0, len(all))
	for _, a := range all {
		dirType := af.cache.GetDirType(a.DirPath)
		match := false
		switch typeFilter {
		case "page":
			match = dirType == "page"
		case "product":
			match = dirType == "product"
		default:
			match = dirType != "page" && dirType != "product"
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

func (af *ArticleForge) pickSlug(dirPath, title, requested string) (string, error) {
	generated := requested == ""
	slug := requested
	if generated {
		slug = slugCandidate(title)
	}
	if !generated && af.checkSlugConflict(dirPath, slug) {
		return "", &ConflictError{Message: "该Slug已存在，请更换"}
	}
	if !generated {
		return slug, nil
	}
	if !af.checkSlugConflict(dirPath, slug) {
		return slug, nil
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d", slug, i)
		if !af.checkSlugConflict(dirPath, candidate) {
			return candidate, nil
		}
	}
}

// slugCandidate turns a title into a URL identifier. Chinese titles produce no
// ascii characters, so fall back to a timestamp to keep the file addressable.
func slugCandidate(title string) string {
	if slug := common.GenerateSlug(title); slug != "" {
		return slug
	}
	return "p-" + time.Now().Format("20060102-150405")
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

func (af *ArticleForge) buildFileContent(input models.ArticleInput, slug string) string {
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
	if input.Draft != nil && *input.Draft {
		sb.WriteString(fmt.Sprintf("draft = %v\n", *input.Draft))
	}
	if input.Summary != "" {
		sb.WriteString(fmt.Sprintf("summary = %q\n", input.Summary))
	}
	if input.Category != "" {
		sb.WriteString(fmt.Sprintf("category = %q\n", input.Category))
	}
	if input.Cover != "" {
		sb.WriteString(fmt.Sprintf("cover = %q\n", input.Cover))
	}
	if len(input.Gallery) > 0 {
		gallery := make([]string, len(input.Gallery))
		for i, g := range input.Gallery {
			gallery[i] = fmt.Sprintf("%q", g)
		}
		sb.WriteString(fmt.Sprintf("gallery = [%s]\n", strings.Join(gallery, ", ")))
	}
	for _, spec := range input.Specs {
		sb.WriteString("\n[[specs]]\n")
		sb.WriteString(fmt.Sprintf("name = %q\n", spec.Name))
		if spec.Price != "" {
			sb.WriteString(fmt.Sprintf("price = %q\n", spec.Price))
		}
		if spec.SKU != "" {
			sb.WriteString(fmt.Sprintf("sku = %q\n", spec.SKU))
		}
		if spec.JDURL != "" {
			sb.WriteString(fmt.Sprintf("jd_url = %q\n", spec.JDURL))
		}
		if spec.TaobaoURL != "" {
			sb.WriteString(fmt.Sprintf("taobao_url = %q\n", spec.TaobaoURL))
		}
	}
	for _, improvement := range input.Improvements {
		sb.WriteString("\n[[improvements]]\n")
		sb.WriteString(fmt.Sprintf("date = %q\n", improvement.Date))
		sb.WriteString(fmt.Sprintf("content = %q\n", improvement.Content))
		sb.WriteString(fmt.Sprintf("visible = %v\n", improvement.Visible))
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

const (
	maxCategoryDepth = 2
	maxCategoryLen   = 40
	maxImprovements  = 50
	maxImprovement   = 200
)

func normalizeProductInput(input models.ArticleInput) models.ArticleInput {
	input.Category = normalizeCategoryPath(input.Category)
	input.Cover = strings.TrimSpace(input.Cover)
	if len(input.Gallery) > 0 {
		gallery := make([]string, 0, len(input.Gallery))
		for _, g := range input.Gallery {
			if g = strings.TrimSpace(g); g != "" {
				gallery = append(gallery, g)
			}
		}
		input.Gallery = gallery
	}
	specs := make([]models.ProductSpec, 0, len(input.Specs))
	for _, spec := range input.Specs {
		spec.Name = strings.TrimSpace(spec.Name)
		spec.Price = strings.TrimSpace(spec.Price)
		spec.SKU = strings.TrimSpace(spec.SKU)
		spec.JDURL = strings.TrimSpace(spec.JDURL)
		spec.TaobaoURL = strings.TrimSpace(spec.TaobaoURL)
		specs = append(specs, spec)
	}
	input.Specs = specs
	improvements := make([]models.ProductImprovement, 0, len(input.Improvements))
	for _, improvement := range input.Improvements {
		improvement.Date = strings.TrimSpace(improvement.Date)
		improvement.Content = strings.TrimSpace(improvement.Content)
		improvements = append(improvements, improvement)
	}
	input.Improvements = improvements
	return input
}

func normalizeCategoryPath(raw string) string {
	parts := make([]string, 0, maxCategoryDepth)
	for _, part := range strings.Split(raw, "/") {
		if part = strings.TrimSpace(part); part != "" {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, "/")
}

func validateProductFields(input models.ArticleInput, isProduct bool) map[string]string {
	errors := make(map[string]string)
	if input.Category != "" {
		parts := strings.Split(input.Category, "/")
		total := 0
		if len(parts) > maxCategoryDepth {
			errors["category"] = "分类最多两级，用 / 分隔"
		}
		for _, part := range parts {
			n := utf8.RuneCountInString(part)
			if n == 0 {
				errors["category"] = "分类层级不能为空"
				break
			}
			if n > 20 {
				errors["category"] = "每级分类不超过20个字符"
				break
			}
			total += n
		}
		if total > maxCategoryLen {
			errors["category"] = "分类总长度不超过40个字符"
		}
	}
	if !validCoverURL(input.Cover) {
		errors["cover"] = "封面地址需为站内路径或以 http(s):// 开头"
	}
	if len(input.Gallery) > 20 {
		errors["gallery"] = "产品图不超过20张"
	}
	for _, g := range input.Gallery {
		if !validCoverURL(g) {
			errors["gallery"] = "图片地址需为站内路径或以 http(s):// 开头"
			break
		}
	}
	if isProduct && len(input.Specs) == 0 {
		errors["specs"] = "产品至少需要一个规格，每个规格都要有 SKU"
	}
	if len(input.Specs) > 30 {
		errors["specs"] = "规格数量不超过30个"
	}
	skuSeen := make(map[string]int, len(input.Specs))
	for i, spec := range input.Specs {
		prefix := fmt.Sprintf("specs.%d", i)
		if spec.Name == "" {
			errors[prefix+".name"] = "规格名称不能为空"
		} else if utf8.RuneCountInString(spec.Name) > 60 {
			errors[prefix+".name"] = "规格名称不超过60个字符"
		}
		if utf8.RuneCountInString(spec.Price) > 30 {
			errors[prefix+".price"] = "价格文字不超过30个字符"
		}
		if spec.SKU == "" {
			errors[prefix+".sku"] = "每个规格都要填写 SKU"
		} else if utf8.RuneCountInString(spec.SKU) > 60 {
			errors[prefix+".sku"] = "SKU不超过60个字符"
		} else if prev, ok := skuSeen[strings.ToLower(spec.SKU)]; ok {
			errors[prefix+".sku"] = fmt.Sprintf("SKU 与规格 %d 重复", prev+1)
		} else {
			skuSeen[strings.ToLower(spec.SKU)] = i
		}
		if !validExternalURL(spec.JDURL) {
			errors[prefix+".jdUrl"] = "京东链接需以 http(s):// 开头"
		}
		if !validExternalURL(spec.TaobaoURL) {
			errors[prefix+".taobaoUrl"] = "淘宝链接需以 http(s):// 开头"
		}
	}
	if len(input.Improvements) > maxImprovements {
		errors["improvements"] = fmt.Sprintf("改进记录不超过%d条", maxImprovements)
	}
	for i, improvement := range input.Improvements {
		prefix := fmt.Sprintf("improvements.%d", i)
		if !validDate(improvement.Date) {
			errors[prefix+".date"] = "改进日期需为 YYYY-MM-DD"
		}
		if improvement.Content == "" {
			errors[prefix+".content"] = "改进内容不能为空"
		} else if utf8.RuneCountInString(improvement.Content) > maxImprovement {
			errors[prefix+".content"] = fmt.Sprintf("改进内容不超过%d个字符", maxImprovement)
		}
	}
	return errors
}

func validDate(raw string) bool {
	parsed, err := time.Parse("2006-01-02", raw)
	return err == nil && parsed.Format("2006-01-02") == raw
}

func validCoverURL(raw string) bool {
	if raw == "" {
		return true
	}
	if strings.HasPrefix(raw, "/") {
		return !strings.HasPrefix(raw, "//")
	}
	return validExternalURL(raw)
}

func validExternalURL(raw string) bool {
	if raw == "" {
		return true
	}
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func mergeErrors(dst, src map[string]string) {
	for key, value := range src {
		dst[key] = value
	}
}
