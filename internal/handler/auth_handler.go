package handler

import (
	"pos-backend/internal/service"
	"pos-backend/pkg/i18n"
	"pos-backend/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	lang := i18n.Lang(c)
	var req service.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, i18n.T(lang, "err.invalid_body"))
	}
	if req.Email == "" || req.Password == "" {
		return response.Error(c, fiber.StatusBadRequest, i18n.T(lang, "err.email_password_required"))
	}

	token, user, err := h.authSvc.Login(req.Email, req.Password)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, i18n.T(lang, err.Error()))
	}

	return response.Success(c, i18n.T(lang, "auth.login_success"), fiber.Map{
		"token": token,
		"user":  user,
	})
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	lang := i18n.Lang(c)
	var req service.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, i18n.T(lang, "err.invalid_body"))
	}

	user, err := h.authSvc.Register(req)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, i18n.T(lang, err.Error()))
	}

	return response.Created(c, i18n.T(lang, "auth.register_success"), user)
}
