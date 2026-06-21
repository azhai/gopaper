package handler

import (
	"strconv"

	"github.com/azhai/gopaper/internal/service"

	"github.com/azhai/gopaper/internal/model"

	"github.com/gofiber/fiber/v3"
)

type ImageHandler struct {
	imageStore *service.ImageStore
}

func NewImageHandler(imageStore *service.ImageStore) *ImageHandler {
	return &ImageHandler{imageStore: imageStore}
}

func (h *ImageHandler) List(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	images, total, err := h.imageStore.List(c.Context(), page, pageSize)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Code:    50000,
			Message: err.Error(),
		})
	}

	return c.JSON(model.PageResponse{
		Code:     0,
		Data:     images,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *ImageHandler) Upload(c fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Code:    42201,
			Message: "未找到上传文件",
		})
	}

	info, err := h.imageStore.Upload(c.Context(), file)
	if err != nil {
		if _, ok := err.(*service.FileTypeError); ok {
			return c.Status(415).JSON(model.ErrorResponse{
				Code:    41501,
				Message: err.Error(),
			})
		}
		if _, ok := err.(*service.FileSizeError); ok {
			return c.Status(413).JSON(model.ErrorResponse{
				Code:    41301,
				Message: err.Error(),
			})
		}
		return c.Status(500).JSON(model.ErrorResponse{
			Code:    50000,
			Message: err.Error(),
		})
	}

	return c.Status(201).JSON(model.SuccessResponse{
		Code:    0,
		Message: "上传成功",
		Data:    info,
	})
}

func (h *ImageHandler) Delete(c fiber.Ctx) error {
	fileName := c.Params("fileName")

	if err := h.imageStore.Delete(c.Context(), fileName); err != nil {
		if _, ok := err.(*service.ConflictError); ok {
			return c.Status(409).JSON(model.ErrorResponse{
				Code:    40902,
				Message: err.Error(),
			})
		}
		return c.Status(500).JSON(model.ErrorResponse{
			Code:    50000,
			Message: err.Error(),
		})
	}

	return c.JSON(model.SuccessResponse{
		Code:    0,
		Message: "删除成功",
	})
}
