package handlers

import (
	"time"

	"github.com/azhai/gopaper/models"
	"github.com/azhai/gopaper/services"

	"github.com/labstack/echo/v5"
)

type ThemeHandler struct {
	themes *services.ThemeForge
}

func NewThemeHandler(themes *services.ThemeForge) *ThemeHandler {
	return &ThemeHandler{themes: themes}
}

// List returns every preset with its rendered CSS plus the effective schedule.
func (h *ThemeHandler) List(c *echo.Context) error {
	themes, err := h.themes.List()
	if err != nil {
		return themeError(c, err)
	}

	now := time.Now()
	effective, err := h.themes.Effective(now)
	if err != nil {
		return themeError(c, err)
	}
	next, err := h.themes.Next(now)
	if err != nil {
		return themeError(c, err)
	}

	data := models.EnabledTheme{
		Themes:     make([]models.ThemeCSS, 0, len(themes)),
		Presets:    models.PresetThemes(),
		DefaultCSS: models.DefaultTheme().CSS(),
	}
	for _, theme := range themes {
		data.Themes = append(data.Themes, models.ThemeCSS{Theme: theme, CSS: theme.CSS()})
	}
	if effective != nil {
		data.CurrentID = effective.ID
	}
	if next != nil {
		data.NextID = next.ID
		data.NextAt = next.EnabledAt
	}

	return c.JSON(200, models.SuccessResponse{Code: 0, Data: data})
}

func (h *ThemeHandler) Create(c *echo.Context) error {
	var input models.Theme
	if err := c.Bind(&input); err != nil {
		return c.JSON(400, models.ErrorResponse{Code: 40000, Message: "参数错误"})
	}

	theme, err := h.themes.Create(input)
	if err != nil {
		return themeError(c, err)
	}
	return c.JSON(201, models.SuccessResponse{
		Code:    0,
		Message: "主题已保存",
		Data:    models.ThemeCSS{Theme: theme, CSS: theme.CSS()},
	})
}

func (h *ThemeHandler) Update(c *echo.Context) error {
	var input models.Theme
	if err := c.Bind(&input); err != nil {
		return c.JSON(400, models.ErrorResponse{Code: 40000, Message: "参数错误"})
	}
	input.ID = c.Param("id")

	if err := h.themes.Update(input); err != nil {
		return themeError(c, err)
	}
	return c.JSON(200, models.SuccessResponse{Code: 0, Message: "主题已更新"})
}

func (h *ThemeHandler) Delete(c *echo.Context) error {
	if err := h.themes.Delete(c.Param("id")); err != nil {
		return themeError(c, err)
	}
	return c.JSON(200, models.SuccessResponse{Code: 0, Message: "主题已删除"})
}

type themeActivateInput struct {
	EnabledAt string `json:"enabledAt"`
}

// Activate starts a preset immediately or at the requested local time.
func (h *ThemeHandler) Activate(c *echo.Context) error {
	var input themeActivateInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(400, models.ErrorResponse{Code: 40000, Message: "参数错误"})
	}

	if err := h.themes.Activate(c.Param("id"), input.EnabledAt); err != nil {
		return themeError(c, err)
	}
	return c.JSON(200, models.SuccessResponse{Code: 0, Message: "启用时间已设置"})
}

// Preview renders the CSS a draft would produce, so the designer can show the
// exact styling the server is about to store.
func (h *ThemeHandler) Preview(c *echo.Context) error {
	var input models.Theme
	if err := c.Bind(&input); err != nil {
		return c.JSON(400, models.ErrorResponse{Code: 40000, Message: "参数错误"})
	}
	input.Normalize()

	return c.JSON(200, models.SuccessResponse{
		Code: 0,
		Data: map[string]string{"css": input.CSS()},
	})
}

func themeError(c *echo.Context, err error) error {
	switch e := err.(type) {
	case *services.ValidationError:
		return c.JSON(422, models.ErrorResponse{Code: 42201, Message: "参数校验失败", Data: e.Errors})
	case *services.ConflictError:
		return c.JSON(409, models.ErrorResponse{Code: 40901, Message: e.Error()})
	case *services.NotFoundError:
		return c.JSON(404, models.ErrorResponse{Code: 40401, Message: e.Error()})
	default:
		return c.JSON(500, models.ErrorResponse{Code: 50000, Message: err.Error()})
	}
}
