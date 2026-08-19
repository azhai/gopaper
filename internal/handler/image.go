package handler

import (
	"strconv"

	"github.com/azhai/gopaper/internal/model"
	"github.com/azhai/gopaper/internal/service"

	"github.com/labstack/echo/v4"
)

type ImageHandler struct {
	imageStore *service.ImageStore
}

func NewImageHandler(imageStore *service.ImageStore) *ImageHandler {
	return &ImageHandler{imageStore: imageStore}
}

func (h *ImageHandler) List(c echo.Context) error {
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
		return c.JSON(500, model.ErrorResponse{
			Code:    50000,
			Message: err.Error(),
		})
	}

	return c.JSON(200, model.PageResponse{
		Code:     0,
		Data:     images,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *ImageHandler) Upload(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(400, model.ErrorResponse{
			Code:    42201,
			Message: "未找到上传文件",
		})
	}

	info, err := h.imageStore.Upload(c.Request().Context(), file)
	if err != nil {
		if _, ok := err.(*service.FileTypeError); ok {
			return c.JSON(415, model.ErrorResponse{
				Code:    41501,
				Message: err.Error(),
			})
		}
		if _, ok := err.(*service.FileSizeError); ok {
			return c.JSON(413, model.ErrorResponse{
				Code:    41301,
				Message: err.Error(),
			})
		}
		return c.JSON(500, model.ErrorResponse{
			Code:    50000,
			Message: err.Error(),
		})
	}

	return c.JSON(201, model.SuccessResponse{
		Code:    0,
		Message: "上传成功",
		Data:    info,
	})
}

func (h *ImageHandler) Delete(c echo.Context) error {
	fileName := c.Param("fileName")

	if err := h.imageStore.Delete(c.Request().Context(), fileName); err != nil {
		if _, ok := err.(*service.ConflictError); ok {
			return c.JSON(409, model.ErrorResponse{
				Code:    40902,
				Message: err.Error(),
			})
		}
		return c.JSON(500, model.ErrorResponse{
			Code:    50000,
			Message: err.Error(),
		})
	}

	return c.JSON(200, model.SuccessResponse{
		Code:    0,
		Message: "删除成功",
	})
}
