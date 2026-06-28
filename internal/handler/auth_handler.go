package handler

import (
	"pos-backend/internal/service"
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
	var req service.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}
	if req.Email == "" || req.Password == "" {
		return response.Error(c, fiber.StatusBadRequest, "email and password are required")
	}

	token, user, err := h.authSvc.Login(req.Email, req.Password)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, err.Error())
	}

	return response.Success(c, fiber.Map{
		"token": token,
		"user":  user,
	})
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req service.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	user, err := h.authSvc.Register(req)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}

	return response.Created(c, user)
}
