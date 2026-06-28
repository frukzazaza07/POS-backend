package handler

import (
	"pos-backend/internal/service"
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
	search := c.Query("search")
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	items, total, err := h.svc.List(search, page, limit)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.Success(c, response.PaginatedData{
		Items: items,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

func (h *POSProductHandler) Get(c *fiber.Ctx) error {
	p, err := h.svc.GetByID(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "product not found")
	}
	return response.Success(c, p)
}

func (h *POSProductHandler) Create(c *fiber.Ctx) error {
	var req service.CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	p, err := h.svc.Create(req)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}
	return response.Created(c, p)
}

func (h *POSProductHandler) Update(c *fiber.Ctx) error {
	var req service.CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	p, err := h.svc.Update(c.Params("id"), req)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}
	return response.Success(c, p)
}

func (h *POSProductHandler) Delete(c *fiber.Ctx) error {
	if err := h.svc.Delete(c.Params("id")); err != nil {
		return response.Error(c, fiber.StatusNotFound, err.Error())
	}
	return response.Success(c, fiber.Map{"message": "deleted"})
}
