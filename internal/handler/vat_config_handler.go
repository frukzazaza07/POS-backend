package handler

import (
	"errors"

	"pos-backend/internal/service"
	"pos-backend/pkg/i18n"
	"pos-backend/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type VatConfigHandler struct {
	svc *service.VatConfigService
}

func NewVatConfigHandler(svc *service.VatConfigService) *VatConfigHandler {
	return &VatConfigHandler{svc: svc}
}

func (h *VatConfigHandler) Get(c *fiber.Ctx) error {
	lang := i18n.Lang(c)
	config, err := h.svc.Get()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}
	return response.Success(c, i18n.T(lang, "config.vat_get"), config)
}

func (h *VatConfigHandler) Upsert(c *fiber.Ctx) error {
	lang := i18n.Lang(c)
	var req service.UpsertVatConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, i18n.T(lang, "err.invalid_body"))
	}

	config, err := h.svc.Upsert(req)
	if err != nil {
		var valErr *service.ValidationError
		if errors.As(err, &valErr) {
			return response.Error(c, fiber.StatusBadRequest, i18n.T(lang, err.Error()))
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}
	return response.Success(c, i18n.T(lang, "config.vat_updated"), config)
}
