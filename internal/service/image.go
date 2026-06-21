package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/azhai/gopaper/internal/model"

	"log/slog"
)

type ImageStore struct {
	uploadDir    string
	maxFileSize  int64
	allowedTypes map[string]bool
	logger       *slog.Logger
	cache        *CacheVault
}

func NewImageStore(uploadDir string, maxFileSize int64, cache *CacheVault, logger *slog.Logger) *ImageStore {
	return &ImageStore{
		uploadDir:   uploadDir,
		maxFileSize: maxFileSize,
		allowedTypes: map[string]bool{
			"image/jpeg":    true,
			"image/png":     true,
			"image/gif":     true,
			"image/webp":    true,
			"image/svg+xml": true,
		},
		logger: logger,
		cache:  cache,
	}
}

func (is *ImageStore) Upload(ctx context.Context, fileHeader *multipart.FileHeader) (*model.ImageInfo, error) {
	if err := is.validateFileType(fileHeader); err != nil {
		return nil, err
	}
	if err := is.validateFileSize(fileHeader); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(is.uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}

	fileName := fmt.Sprintf("%d-%s", time.Now().UnixMilli(), fileHeader.Filename)
	filePath := filepath.Join(is.uploadDir, fileName)

	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("open uploaded file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("create dest file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(filePath)
		return nil, fmt.Errorf("copy file: %w", err)
	}

	info := &model.ImageInfo{
		FileName:     fileName,
		OriginalName: fileHeader.Filename,
		FileSize:     fileHeader.Size,
		FileType:     fileHeader.Header.Get("Content-Type"),
		UploadPath:   "/uploads/" + fileName,
		UploadTime:   time.Now(),
	}

	is.logger.Info("image uploaded", "fileName", fileName)
	return info, nil
}

func (is *ImageStore) Delete(ctx context.Context, fileName string) error {
	filePath := filepath.Join(is.uploadDir, fileName)

	refs := is.findReferences(fileName)
	if len(refs) > 0 {
		return &ConflictError{Message: fmt.Sprintf("该图片仍被以下文章引用: %s", strings.Join(refs, ", "))}
	}

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove file: %w", err)
	}

	is.logger.Info("image deleted", "fileName", fileName)
	return nil
}

func (is *ImageStore) List(ctx context.Context, page, pageSize int) ([]model.ImageInfo, int, error) {
	entries, err := os.ReadDir(is.uploadDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, 0, nil
		}
		return nil, 0, fmt.Errorf("read upload dir: %w", err)
	}

	var images []model.ImageInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" && ext != ".svg" {
			continue
		}

		mimeType := "image/jpeg"
		switch ext {
		case ".png":
			mimeType = "image/png"
		case ".gif":
			mimeType = "image/gif"
		case ".webp":
			mimeType = "image/webp"
		case ".svg":
			mimeType = "image/svg+xml"
		}

		images = append(images, model.ImageInfo{
			FileName:     entry.Name(),
			OriginalName: strings.TrimPrefix(entry.Name(), fmt.Sprintf("%d-", info.ModTime().UnixMilli())),
			FileSize:     info.Size(),
			FileType:     mimeType,
			UploadPath:   "/uploads/" + entry.Name(),
			UploadTime:   info.ModTime(),
		})
	}

	sort.Slice(images, func(i, j int) bool {
		return images[i].UploadTime.After(images[j].UploadTime)
	})

	total := len(images)
	start := (page - 1) * pageSize
	if start >= total {
		return nil, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	return images[start:end], total, nil
}

func (is *ImageStore) validateFileType(header *multipart.FileHeader) error {
	contentType := header.Header.Get("Content-Type")
	if !is.allowedTypes[contentType] {
		return &FileTypeError{Message: "不支持的文件类型，仅允许jpg/png/gif/webp/svg"}
	}
	return nil
}

func (is *ImageStore) validateFileSize(header *multipart.FileHeader) error {
	if header.Size > is.maxFileSize {
		return &FileSizeError{Message: fmt.Sprintf("文件大小超过%dMB限制", is.maxFileSize/1024/1024)}
	}
	return nil
}

func (is *ImageStore) findReferences(fileName string) []string {
	articles := is.cache.GetAllArticles()
	var refs []string
	refStr := fileName
	for _, a := range articles {
		if strings.Contains(a.Content, refStr) {
			refs = append(refs, a.Slug)
		}
	}
	return refs
}

type FileTypeError struct {
	Message string
}

func (e *FileTypeError) Error() string {
	return e.Message
}

type FileSizeError struct {
	Message string
}

func (e *FileSizeError) Error() string {
	return e.Message
}
