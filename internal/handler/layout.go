package handler

import (
	"os"

	"github.com/azhai/gopaper/internal/model"
	"github.com/gofiber/fiber/v3"
	"github.com/pelletier/go-toml/v2"
)

type LayoutHandler struct {
	configPath string
}

func NewLayoutHandler() *LayoutHandler {
	return &LayoutHandler{configPath: "templates/layouts.toml"}
}

// GetLayouts returns all templates and their regions.
func (h *LayoutHandler) GetLayouts(c fiber.Ctx) error {
	data, err := os.ReadFile(h.configPath)
	if err != nil {
		return c.JSON(model.SuccessResponse{
			Code: 0,
			Data: model.DefaultLayoutConfig(),
		})
	}
	var cfg model.LayoutConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return c.JSON(model.SuccessResponse{
			Code: 0,
			Data: model.DefaultLayoutConfig(),
		})
	}
	return c.JSON(model.SuccessResponse{Code: 0, Data: cfg})
}

// SaveLayouts saves the entire layout config.
func (h *LayoutHandler) SaveLayouts(c fiber.Ctx) error {
	var cfg model.LayoutConfig
	if err := c.Bind().JSON(&cfg); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{Code: 40000, Message: "参数错误"})
	}
	out, err := toml.Marshal(&cfg)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{Code: 50000, Message: "序列化配置失败"})
	}
	if err := os.WriteFile(h.configPath, out, 0644); err != nil {
		return c.Status(500).JSON(model.ErrorResponse{Code: 50000, Message: "写入配置失败"})
	}
	return c.JSON(model.SuccessResponse{Code: 0, Message: "布局配置已保存"})
}

// GetRegions returns the regions for a specific template.
func (h *LayoutHandler) GetRegions(c fiber.Ctx) error {
	tmplName := c.Params("name")
	data, err := os.ReadFile(h.configPath)
	if err != nil {
		return c.Status(404).JSON(model.ErrorResponse{Code: 40400, Message: "布局配置不存在"})
	}
	var cfg model.LayoutConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return c.Status(500).JSON(model.ErrorResponse{Code: 50000, Message: "解析配置失败"})
	}
	for _, t := range cfg.Templates {
		if t.Name == tmplName {
			return c.JSON(model.SuccessResponse{Code: 0, Data: t.Regions})
		}
	}
	return c.Status(404).JSON(model.ErrorResponse{Code: 40400, Message: "模板不存在"})
}
