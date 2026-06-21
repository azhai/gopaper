package handler

import (
	"strconv"

	"github.com/azhai/gopaper/internal/model"
	"github.com/azhai/gopaper/internal/service"

	"github.com/gofiber/fiber/v3"
)

type ArticleHandler struct {
	articleForge *service.ArticleForge
	renderer     *service.Renderer
}

func NewArticleHandler(articleForge *service.ArticleForge, renderer *service.Renderer) *ArticleHandler {
	return &ArticleHandler{
		articleForge: articleForge,
		renderer:     renderer,
	}
}

func (h *ArticleHandler) List(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize", "10"))
	dir := c.Query("dir")
	typeFilter := c.Query("type") // "page" | "article"

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var articles []*model.Article
	var total int
	var err error

	switch {
	case dir != "":
		articles, total, err = h.articleForge.ListByDir(c.Context(), dir, page, pageSize)
	case typeFilter == "page" || typeFilter == "article":
		articles, total, err = h.articleForge.ListByType(c.Context(), typeFilter, page, pageSize)
	default:
		articles, total, err = h.articleForge.ListAll(c.Context(), page, pageSize)
	}

	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Code:    50000,
			Message: err.Error(),
		})
	}

	summaries := make([]*model.ArticleSummary, len(articles))
	for i, a := range articles {
		summaries[i] = model.ToSummary(a)
	}

	return c.JSON(model.PageResponse{
		Code:     0,
		Data:     summaries,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *ArticleHandler) Get(c fiber.Ctx) error {
	slug := c.Params("slug")

	article, err := h.articleForge.GetBySlug(c.Context(), slug)
	if err != nil {
		if _, ok := err.(*service.NotFoundError); ok {
			return c.Status(404).JSON(model.ErrorResponse{
				Code:    40401,
				Message: "文章不存在",
			})
		}
		return c.Status(500).JSON(model.ErrorResponse{
			Code:    50000,
			Message: err.Error(),
		})
	}

	return c.JSON(model.SuccessResponse{
		Code: 0,
		Data: article,
	})
}

func (h *ArticleHandler) Create(c fiber.Ctx) error {
	var input model.ArticleInput
	if err := c.Bind().Body(&input); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Code:    42201,
			Message: "请求格式错误",
		})
	}

	if err := h.articleForge.Create(c.Context(), input); err != nil {
		if ve, ok := err.(*service.ValidationError); ok {
			return c.Status(422).JSON(model.ErrorResponse{
				Code:    42201,
				Message: "MetaData校验失败",
				Data:    ve.Errors,
			})
		}
		if _, ok := err.(*service.ConflictError); ok {
			return c.Status(409).JSON(model.ErrorResponse{
				Code:    40901,
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
		Message: "创建成功",
	})
}

func (h *ArticleHandler) Update(c fiber.Ctx) error {
	slug := c.Params("slug")

	var input model.ArticleInput
	if err := c.Bind().Body(&input); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Code:    42201,
			Message: "请求格式错误",
		})
	}

	if err := h.articleForge.Update(c.Context(), slug, input); err != nil {
		if _, ok := err.(*service.NotFoundError); ok {
			return c.Status(404).JSON(model.ErrorResponse{
				Code:    40401,
				Message: "文章不存在",
			})
		}
		return c.Status(500).JSON(model.ErrorResponse{
			Code:    50000,
			Message: err.Error(),
		})
	}

	return c.JSON(model.SuccessResponse{
		Code:    0,
		Message: "更新成功",
	})
}

func (h *ArticleHandler) Delete(c fiber.Ctx) error {
	slug := c.Params("slug")

	if err := h.articleForge.Delete(c.Context(), slug); err != nil {
		if _, ok := err.(*service.NotFoundError); ok {
			return c.Status(404).JSON(model.ErrorResponse{
				Code:    40401,
				Message: "文章不存在",
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

func (h *ArticleHandler) Preview(c fiber.Ctx) error {
	var body struct {
		Content string `json:"content"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Code:    42201,
			Message: "请求格式错误",
		})
	}

	html, err := h.renderer.RenderString(c.Context(), body.Content)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Code:    50000,
			Message: "渲染失败",
		})
	}

	return c.JSON(model.SuccessResponse{
		Code: 0,
		Data: map[string]string{"html": html},
	})
}
