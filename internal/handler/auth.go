package handler

import (
	"github.com/azhai/gopaper/internal/middleware"
	"github.com/azhai/gopaper/internal/model"

	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	authGuard *middleware.AuthGuard
}

func NewAuthHandler(authGuard *middleware.AuthGuard) *AuthHandler {
	return &AuthHandler{authGuard: authGuard}
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Code:    42201,
			Message: "请求格式错误",
		})
	}

	if req.Username == "" || req.Password == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Code:    42201,
			Message: "用户名和密码不能为空",
		})
	}

	if h.authGuard.IsLocked(req.Username) {
		return c.Status(403).JSON(model.ErrorResponse{
			Code:    40301,
			Message: "账户已锁定，请15分钟后重试",
		})
	}

	resp, err := h.authGuard.Login(req.Username, req.Password)
	if err != nil {
		return c.Status(401).JSON(model.ErrorResponse{
			Code:    40101,
			Message: err.Error(),
		})
	}

	return c.JSON(model.SuccessResponse{
		Code:    0,
		Message: "登录成功",
		Data:    resp,
	})
}
