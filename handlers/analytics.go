package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/azhai/gopaper/models"
	"github.com/azhai/gopaper/services"

	"github.com/labstack/echo/v5"
)

const (
	visitorCookieName   = "gp_visitor"
	visitorCookieMaxAge = 365 * 24 * 60 * 60
)

type AnalyticsHandler struct {
	analytics *services.Analytics
}

func NewAnalyticsHandler(analytics *services.Analytics) *AnalyticsHandler {
	return &AnalyticsHandler{analytics: analytics}
}

func (h *AnalyticsHandler) Stats(c *echo.Context) error {
	rangeKey := c.QueryParam("range")
	if rangeKey == "" {
		rangeKey = "1d"
	}
	report, ok := h.analytics.Report(time.Now(), rangeKey)
	if !ok {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    40000,
			Message: "不支持的时间范围",
		})
	}
	return c.JSON(http.StatusOK, models.SuccessResponse{Code: 0, Data: report})
}

// prepareView resolves the visitor before the response is written, so the cookie
// is part of the rendered response. Empty values mean the page is not tracked.
func (h *PageHandler) prepareView(c *echo.Context, article *models.Article) (visitor, kind string) {
	if h.analytics == nil || article == nil || article.Draft {
		return "", ""
	}
	kind = services.ContentKind(h.cache.GetDirType(article.DirPath), article.DirPath)
	if kind == "" {
		return "", ""
	}
	return visitorID(c), kind
}

func (h *PageHandler) recordView(article *models.Article, kind, visitor string) {
	if visitor == "" || kind == "" {
		return
	}
	title := article.Title
	if title == "" {
		title = article.Slug
	}
	if err := h.analytics.Record(articleHref(article.DirPath, article.Slug), title, kind, visitor); err != nil && h.logger != nil {
		h.logger.Warn("record page view failed", "slug", article.Slug, "error", err)
	}
}

// visitorID returns the anonymous browser identifier, creating it on first visit.
func visitorID(c *echo.Context) string {
	if cookie, err := c.Cookie(visitorCookieName); err == nil && validVisitorID(cookie.Value) {
		return cookie.Value
	}

	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return ""
	}
	id := hex.EncodeToString(raw)
	c.SetCookie(&http.Cookie{
		Name:     visitorCookieName,
		Value:    id,
		Path:     "/",
		MaxAge:   visitorCookieMaxAge,
		HttpOnly: true,
		Secure:   c.IsTLS() || strings.HasPrefix(strings.ToLower(c.Request().Header.Get("X-Forwarded-Proto")), "https"),
		SameSite: http.SameSiteLaxMode,
	})
	return id
}

func validVisitorID(value string) bool {
	if len(value) != 32 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
