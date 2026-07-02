package handler

import (
	"pos-backend/internal/service"
	"pos-backend/pkg/i18n"
	"pos-backend/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type POSProductHandler struct {
	svc *service.POSProductService
}

func NewPOSProductHandler(svc *service.POSProductService) *POSProductHandler {
	return &POSProductHandler{svc: svc}
}

func (h *POSProductHandler) List(c *fiber.Ctx) error {
	lang := i18n.Lang(c)
	search := c.Query("search")
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	items, total, err := h.svc.List(search, page, limit)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.Success(c, i18n.T(lang, "product.list"), response.PaginatedData{
		Items: items,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

func (h *POSProductHandler) Get(c *fiber.Ctx) error {
	lang := i18n.Lang(c)
	p, err := h.svc.GetByID(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, i18n.T(lang, "err.product_not_found"))
	}
	return response.Success(c, i18n.T(lang, "product.get"), p)
}

func (h *POSProductHandler) Create(c *fiber.Ctx) error {
	lang := i18n.Lang(c)
	var req service.CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, i18n.T(lang, "err.invalid_body"))
	}

	p, err := h.svc.Create(req)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, i18n.T(lang, err.Error()))
	}
	return response.Created(c, i18n.T(lang, "product.create"), p)
}

func (h *POSProductHandler) Update(c *fiber.Ctx) error {
	lang := i18n.Lang(c)
	var req service.CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, i18n.T(lang, "err.invalid_body"))
	}

	p, err := h.svc.Update(c.Params("id"), req)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, i18n.T(lang, err.Error()))
	}
	return response.Success(c, i18n.T(lang, "product.update"), p)
}

func (h *POSProductHandler) Delete(c *fiber.Ctx) error {
	lang := i18n.Lang(c)
	if err := h.svc.Delete(c.Params("id")); err != nil {
		return response.Error(c, fiber.StatusNotFound, i18n.T(lang, err.Error()))
	}
	return response.Success(c, i18n.T(lang, "product.delete"), nil)
}
