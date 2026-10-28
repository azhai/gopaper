package handlers

import (
	"os"

	"github.com/azhai/gopaper/models"

	"github.com/labstack/echo/v5"
	"github.com/pelletier/go-toml/v2"
)

type LayoutHandler struct {
	configPath string
}

func NewLayoutHandler() *LayoutHandler {
	return &LayoutHandler{configPath: "themes/default/layouts.toml"}
}

func (h *LayoutHandler) GetLayouts(c *echo.Context) error {
	data, err := os.ReadFile(h.configPath)
	if err != nil {
		return c.JSON(200, models.SuccessResponse{
			Code: 0,
			Data: models.DefaultLayoutConfig(),
		})
	}
	var cfg models.LayoutConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return c.JSON(200, models.SuccessResponse{
			Code: 0,
			Data: models.DefaultLayoutConfig(),
		})
	}
	return c.JSON(200, models.SuccessResponse{Code: 0, Data: cfg})
}

func (h *LayoutHandler) SaveLayouts(c *echo.Context) error {
	var cfg models.LayoutConfig
	if err := c.Bind(&cfg); err != nil {
		return c.JSON(400, models.ErrorResponse{Code: 40000, Message: "参数错误"})
	}
	out, err := toml.Marshal(&cfg)
	if err != nil {
		return c.JSON(500, models.ErrorResponse{Code: 50000, Message: "序列化配置失败"})
	}
	if err := os.WriteFile(h.configPath, out, 0644); err != nil {
		return c.JSON(500, models.ErrorResponse{Code: 50000, Message: "写入配置失败"})
	}
	return c.JSON(200, models.SuccessResponse{Code: 0, Message: "布局配置已保存"})
}

func (h *LayoutHandler) GetRegions(c *echo.Context) error {
	tmplName := c.Param("name")
	data, err := os.ReadFile(h.configPath)
	if err != nil {
		return c.JSON(404, models.ErrorResponse{Code: 40400, Message: "布局配置不存在"})
	}
	var cfg models.LayoutConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return c.JSON(500, models.ErrorResponse{Code: 50000, Message: "解析配置失败"})
	}
	for _, t := range cfg.Templates {
		if t.Name == tmplName {
			return c.JSON(200, models.SuccessResponse{Code: 0, Data: t.Regions})
		}
	}
	return c.JSON(404, models.ErrorResponse{Code: 40400, Message: "模板不存在"})
}
