package handler

import (
	"os"
	"time"

	"github.com/azhai/gopaper/internal/model"
	"github.com/azhai/gopaper/internal/service"

	"github.com/labstack/echo/v4"
	"github.com/pelletier/go-toml/v2"
)

type CacheHandler struct {
	cache      *service.CacheVault
	scanner    *service.Scanner
	contentDir string
}

func NewCacheHandler(cache *service.CacheVault, scanner *service.Scanner, contentDir string) *CacheHandler {
	return &CacheHandler{cache: cache, scanner: scanner, contentDir: contentDir}
}

func (h *CacheHandler) Refresh(c echo.Context) error {
	start := time.Now()

	if err := h.cache.Refresh(c.Request().Context(), h.scanner); err != nil {
		return c.JSON(500, model.ErrorResponse{
			Code:    50000,
			Message: "缓存刷新失败",
		})
	}

	articles := h.cache.GetAllArticles()
	duration := time.Since(start)

	return c.JSON(200, model.SuccessResponse{
		Code:    0,
		Message: "缓存刷新完成",
		Data: map[string]any{
			"articleCount": len(articles),
			"duration":     duration.String(),
		},
	})
}

func (h *CacheHandler) ListDirs(c echo.Context) error {
	dirs := h.cache.GetDirs()
	return c.JSON(200, model.SuccessResponse{
		Code: 0,
		Data: dirs,
	})
}

func (h *CacheHandler) GetSettings(c echo.Context) error {
	metaPath := h.contentDir + "/_meta.toml"
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return c.JSON(500, model.ErrorResponse{Code: 50000, Message: "读取配置失败"})
	}
	var meta model.DirMeta
	if err := toml.Unmarshal(data, &meta); err != nil {
		return c.JSON(500, model.ErrorResponse{Code: 50000, Message: "解析配置失败"})
	}
	return c.JSON(200, model.SuccessResponse{Code: 0, Data: meta})
}

func (h *CacheHandler) UpdateSettings(c echo.Context) error {
	var meta model.DirMeta
	if err := c.Bind(&meta); err != nil {
		return c.JSON(400, model.ErrorResponse{Code: 40000, Message: "参数错误"})
	}
	out, err := toml.Marshal(&meta)
	if err != nil {
		return c.JSON(500, model.ErrorResponse{Code: 50000, Message: "序列化配置失败"})
	}
	metaPath := h.contentDir + "/_meta.toml"
	if err := os.WriteFile(metaPath, out, 0644); err != nil {
		return c.JSON(500, model.ErrorResponse{Code: 50000, Message: "写入配置失败"})
	}
	if err := h.cache.Refresh(c.Request().Context(), h.scanner); err != nil {
		return c.JSON(500, model.ErrorResponse{Code: 50000, Message: "缓存刷新失败"})
	}
	return c.JSON(200, model.SuccessResponse{Code: 0, Message: "设置已保存并刷新缓存"})
}
