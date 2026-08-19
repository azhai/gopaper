package handler

import (
	"strconv"

	"github.com/azhai/gopaper/internal/model"
	"github.com/azhai/gopaper/internal/service"

	"github.com/labstack/echo/v4"
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

func (h *ArticleHandler) List(c echo.Context) error {
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

	var articles []*model.Article
	var total int
	var err error

	switch {
	case dir != "":
		articles, total, err = h.articleForge.ListByDir(c.Request().Context(), dir, page, pageSize)
	case typeFilter == "page" || typeFilter == "article":
		articles, total, err = h.articleForge.ListByType(c.Request().Context(), typeFilter, page, pageSize)
	default:
		articles, total, err = h.articleForge.ListAll(c.Request().Context(), page, pageSize)
	}

	if err != nil {
		return c.JSON(500, model.ErrorResponse{
			Code:    50000,
			Message: err.Error(),
		})
	}

	summaries := make([]*model.ArticleSummary, len(articles))
	for i, a := range articles {
		summaries[i] = model.ToSummary(a)
	}

	return c.JSON(200, model.PageResponse{
		Code:     0,
		Data:     summaries,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *ArticleHandler) Get(c echo.Context) error {
	slug := c.Param("slug")

	article, err := h.articleForge.GetBySlug(c.Request().Context(), slug)
	if err != nil {
		if _, ok := err.(*service.NotFoundError); ok {
			return c.JSON(404, model.ErrorResponse{
				Code:    40401,
				Message: "文章不存在",
			})
		}
		return c.JSON(500, model.ErrorResponse{
			Code:    50000,
			Message: err.Error(),
		})
	}

	return c.JSON(200, model.SuccessResponse{
		Code: 0,
		Data: article,
	})
}

func (h *ArticleHandler) Create(c echo.Context) error {
	var input model.ArticleInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(400, model.ErrorResponse{
			Code:    42201,
			Message: "请求格式错误",
		})
	}

	if err := h.articleForge.Create(c.Request().Context(), input); err != nil {
		if ve, ok := err.(*service.ValidationError); ok {
			return c.JSON(422, model.ErrorResponse{
				Code:    42201,
				Message: "MetaData校验失败",
				Data:    ve.Errors,
			})
		}
		if _, ok := err.(*service.ConflictError); ok {
			return c.JSON(409, model.ErrorResponse{
				Code:    40901,
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
		Message: "创建成功",
	})
}

func (h *ArticleHandler) Update(c echo.Context) error {
	slug := c.Param("slug")

	var input model.ArticleInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(400, model.ErrorResponse{
			Code:    42201,
			Message: "请求格式错误",
		})
	}

	if err := h.articleForge.Update(c.Request().Context(), slug, input); err != nil {
		if _, ok := err.(*service.NotFoundError); ok {
			return c.JSON(404, model.ErrorResponse{
				Code:    40401,
				Message: "文章不存在",
			})
		}
		return c.JSON(500, model.ErrorResponse{
			Code:    50000,
			Message: err.Error(),
		})
	}

	return c.JSON(200, model.SuccessResponse{
		Code:    0,
		Message: "更新成功",
	})
}

func (h *ArticleHandler) Delete(c echo.Context) error {
	slug := c.Param("slug")

	if err := h.articleForge.Delete(c.Request().Context(), slug); err != nil {
		if _, ok := err.(*service.NotFoundError); ok {
			return c.JSON(404, model.ErrorResponse{
				Code:    40401,
				Message: "文章不存在",
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

func (h *ArticleHandler) Preview(c echo.Context) error {
	var body struct {
		Content string `json:"content"`
	}
	if err := c.Bind(&body); err != nil {
		return c.JSON(400, model.ErrorResponse{
			Code:    42201,
			Message: "请求格式错误",
		})
	}

	html, err := h.renderer.RenderString(c.Request().Context(), body.Content)
	if err != nil {
		return c.JSON(500, model.ErrorResponse{
			Code:    50000,
			Message: "渲染失败",
		})
	}

	return c.JSON(200, model.SuccessResponse{
		Code: 0,
		Data: map[string]string{"html": html},
	})
}
