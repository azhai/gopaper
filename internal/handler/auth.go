package handler

import (
	"github.com/azhai/gopaper/internal/middleware"
	"github.com/azhai/gopaper/internal/model"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authGuard *middleware.AuthGuard
}

func NewAuthHandler(authGuard *middleware.AuthGuard) *AuthHandler {
	return &AuthHandler{authGuard: authGuard}
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req model.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, model.ErrorResponse{
			Code:    42201,
			Message: "请求格式错误",
		})
	}

	if req.Username == "" || req.Password == "" {
		return c.JSON(400, model.ErrorResponse{
			Code:    42201,
			Message: "用户名和密码不能为空",
		})
	}

	if h.authGuard.IsLocked(req.Username) {
		return c.JSON(403, model.ErrorResponse{
			Code:    40301,
			Message: "账户已锁定，请15分钟后重试",
		})
	}

	resp, err := h.authGuard.Login(req.Username, req.Password)
	if err != nil {
		return c.JSON(401, model.ErrorResponse{
			Code:    40101,
			Message: err.Error(),
		})
	}

	return c.JSON(200, model.SuccessResponse{
		Code:    0,
		Message: "登录成功",
		Data:    resp,
	})
}
