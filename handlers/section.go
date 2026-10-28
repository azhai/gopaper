package handlers

import (
	"github.com/azhai/gopaper/models"
	"github.com/azhai/gopaper/services"

	"github.com/labstack/echo/v5"
)

type SectionHandler struct {
	sections *services.SectionForge
}

func NewSectionHandler(sections *services.SectionForge) *SectionHandler {
	return &SectionHandler{sections: sections}
}

func (h *SectionHandler) Create(c *echo.Context) error {
	var input models.SectionInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(400, models.ErrorResponse{Code: 40000, Message: "参数错误"})
	}

	name, err := h.sections.Create(c.Request().Context(), input)
	if err != nil {
		return sectionError(c, err)
	}
	return c.JSON(201, models.SuccessResponse{
		Code:    0,
		Message: "栏目已创建",
		Data:    map[string]string{"dirPath": name},
	})
}

func (h *SectionHandler) Update(c *echo.Context) error {
	var input models.SectionInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(400, models.ErrorResponse{Code: 40000, Message: "参数错误"})
	}

	if err := h.sections.Update(c.Request().Context(), input); err != nil {
		return sectionError(c, err)
	}
	return c.JSON(200, models.SuccessResponse{Code: 0, Message: "栏目已更新"})
}

func sectionError(c *echo.Context, err error) error {
	switch e := err.(type) {
	case *services.ValidationError:
		return c.JSON(422, models.ErrorResponse{Code: 42201, Message: "参数校验失败", Data: e.Errors})
	case *services.ConflictError:
		return c.JSON(409, models.ErrorResponse{Code: 40901, Message: e.Error()})
	case *services.NotFoundError:
		return c.JSON(404, models.ErrorResponse{Code: 40401, Message: e.Error()})
	default:
		return c.JSON(500, models.ErrorResponse{Code: 50000, Message: e.Error()})
	}
}
