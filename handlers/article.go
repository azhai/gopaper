package handlers

import (
	"strconv"

	"github.com/azhai/gopaper/models"
	"github.com/azhai/gopaper/services"

	"github.com/labstack/echo/v5"
)

type ArticleHandler struct {
	articleForge *services.ArticleForge
	renderer     *services.Renderer
}

func NewArticleHandler(articleForge *services.ArticleForge, renderer *services.Renderer) *ArticleHandler {
	return &ArticleHandler{
		articleForge: articleForge,
		renderer:     renderer,
	}
}

func (h *ArticleHandler) List(c *echo.Context) error {
	pageStr := c.QueryParam("page")
	if pageStr == "" {
		pageStr = "1"
	}
	page, _ := strconv.Atoi(pageStr)
	pageSizeStr := c.QueryParam("pageSize")
	if pageSizeStr == "" {
		pageSizeStr = "10"
	}
	pageSize, _ := strconv.Atoi(pageSizeStr)
	dir := c.QueryParam("dir")
	typeFilter := c.QueryParam("type")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var articles []*models.Article
	var total int
	var err error

	switch {
	case dir != "":
		articles, total, err = h.articleForge.ListByDir(c.Request().Context(), dir, page, pageSize)
	case typeFilter == "page" || typeFilter == "article" || typeFilter == "product":
		articles, total, err = h.articleForge.ListByType(c.Request().Context(), typeFilter, page, pageSize)
	default:
		articles, total, err = h.articleForge.ListAll(c.Request().Context(), page, pageSize)
	}

	if err != nil {
		return c.JSON(500, models.ErrorResponse{
			Code:    50000,
			Message: err.Error(),
		})
	}

	summaries := make([]*models.ArticleSummary, len(articles))
	for i, a := range articles {
		summaries[i] = models.ToSummary(a)
	}

	return c.JSON(200, models.PageResponse{
		Code:     0,
		Data:     summaries,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *ArticleHandler) Get(c *echo.Context) error {
	slug := c.Param("slug")

	article, err := h.articleForge.GetBySlug(c.Request().Context(), slug)
	if err != nil {
		if _, ok := err.(*services.NotFoundError); ok {
			return c.JSON(404, models.ErrorResponse{
				Code:    40401,
				Message: "文章不存在",
			})
		}
		return c.JSON(500, models.ErrorResponse{
			Code:    50000,
			Message: err.Error(),
		})
	}

	return c.JSON(200, models.SuccessResponse{
		Code: 0,
		Data: article,
	})
}

func (h *ArticleHandler) Create(c *echo.Context) error {
	var input models.ArticleInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(400, models.ErrorResponse{
			Code:    42201,
			Message: "请求格式错误",
		})
	}

	slug, err := h.articleForge.Create(c.Request().Context(), input)
	if err != nil {
		if ve, ok := err.(*services.ValidationError); ok {
			return c.JSON(422, models.ErrorResponse{
				Code:    42201,
				Message: "MetaData校验失败",
				Data:    ve.Errors,
			})
		}
		if _, ok := err.(*services.ConflictError); ok {
			return c.JSON(409, models.ErrorResponse{
				Code:    40901,
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
		Message: "创建成功",
		Data:    map[string]string{"slug": slug},
	})
}

func (h *ArticleHandler) Update(c *echo.Context) error {
	slug := c.Param("slug")

	var input models.ArticleInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(400, models.ErrorResponse{
			Code:    42201,
			Message: "请求格式错误",
		})
	}

	if err := h.articleForge.Update(c.Request().Context(), slug, input); err != nil {
		if _, ok := err.(*services.NotFoundError); ok {
			return c.JSON(404, models.ErrorResponse{
				Code:    40401,
				Message: "文章不存在",
			})
		}
		return c.JSON(500, models.ErrorResponse{
			Code:    50000,
			Message: err.Error(),
		})
	}

	return c.JSON(200, models.SuccessResponse{
		Code:    0,
		Message: "更新成功",
	})
}

func (h *ArticleHandler) Delete(c *echo.Context) error {
	slug := c.Param("slug")

	if err := h.articleForge.Delete(c.Request().Context(), slug); err != nil {
		if _, ok := err.(*services.NotFoundError); ok {
			return c.JSON(404, models.ErrorResponse{
				Code:    40401,
				Message: "文章不存在",
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

func (h *ArticleHandler) Preview(c *echo.Context) error {
	var body struct {
		Content string `json:"content"`
	}
	if err := c.Bind(&body); err != nil {
		return c.JSON(400, models.ErrorResponse{
			Code:    42201,
			Message: "请求格式错误",
		})
	}

	html, err := h.renderer.RenderString(c.Request().Context(), body.Content)
	if err != nil {
		return c.JSON(500, models.ErrorResponse{
			Code:    50000,
			Message: "渲染失败",
		})
	}

	return c.JSON(200, models.SuccessResponse{
		Code: 0,
		Data: map[string]string{"html": html},
	})
}
