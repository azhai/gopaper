package handlers

import (
	"strconv"

	"github.com/azhai/gopaper/models"
	"github.com/azhai/gopaper/services"

	"github.com/labstack/echo/v5"
)

type ImageHandler struct {
	imageStore *services.ImageStore
}

func NewImageHandler(imageStore *services.ImageStore) *ImageHandler {
	return &ImageHandler{imageStore: imageStore}
}

func (h *ImageHandler) List(c *echo.Context) error {
	pageStr := c.QueryParam("page")
	if pageStr == "" {
		pageStr = "1"
	}
	page, _ := strconv.Atoi(pageStr)
	pageSizeStr := c.QueryParam("pageSize")
	if pageSizeStr == "" {
		pageSizeStr = "20"
	}
	pageSize, _ := strconv.Atoi(pageSizeStr)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	images, total, err := h.imageStore.List(c.Request().Context(), page, pageSize)
	if err != nil {
		return c.JSON(500, models.ErrorResponse{
			Code:    50000,
			Message: err.Error(),
		})
	}

	return c.JSON(200, models.PageResponse{
		Code:     0,
		Data:     images,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *ImageHandler) Upload(c *echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(400, models.ErrorResponse{
			Code:    42201,
			Message: "未找到上传文件",
		})
	}

	info, err := h.imageStore.Upload(c.Request().Context(), file)
	if err != nil {
		if _, ok := err.(*services.FileTypeError); ok {
			return c.JSON(415, models.ErrorResponse{
				Code:    41501,
				Message: err.Error(),
			})
		}
		if _, ok := err.(*services.FileSizeError); ok {
			return c.JSON(413, models.ErrorResponse{
				Code:    41301,
				Message: err.Error(),
			})
		}
		return c.JSON(500, models.ErrorResponse{
			Code:    50000,
			Message: err.Error(),
		})
	}

	return c.JSON(201, models.SuccessResponse{
		Code:    0,
		Message: "上传成功",
		Data:    info,
	})
}

func (h *ImageHandler) Delete(c *echo.Context) error {
	fileName := c.Param("fileName")

	if err := h.imageStore.Delete(c.Request().Context(), fileName); err != nil {
		if _, ok := err.(*services.ConflictError); ok {
			return c.JSON(409, models.ErrorResponse{
				Code:    40902,
				Message: err.Error(),
			})
		}
		return c.JSON(500, models.ErrorResponse{
			Code:    50000,
			Message: err.Error(),
		})
	}

	return c.JSON(200, models.SuccessResponse{
		Code:    0,
		Message: "删除成功",
	})
}
