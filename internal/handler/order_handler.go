package handler

import (
	"pos-backend/internal/service"
	"pos-backend/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) Create(c *fiber.Ctx) error {
	cashierID, _ := c.Locals("userID").(string)

	var req service.CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid request body")
	}

	order, err := h.svc.CreateOrder(cashierID, req)
	if err != nil {
		// Order was created but stock deduction failed → return 400
		if order != nil {
			return response.Error(c, fiber.StatusBadRequest, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.Created(c, order)
}

func (h *OrderHandler) List(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	// Cashiers see only their own orders; admins see all
	role, _ := c.Locals("role").(string)
	cashierID := ""
	if role != "admin" {
		cashierID, _ = c.Locals("userID").(string)
	}

	orders, total, err := h.svc.ListOrders(page, limit, cashierID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.Success(c, response.PaginatedData{
		Items: orders,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

func (h *OrderHandler) Get(c *fiber.Ctx) error {
	order, err := h.svc.GetOrder(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "order not found")
	}
	return response.Success(c, order)
}

func (h *OrderHandler) Cancel(c *fiber.Ctx) error {
	requesterID, _ := c.Locals("userID").(string)
	requesterRole, _ := c.Locals("role").(string)

	if err := h.svc.CancelOrder(c.Params("id"), requesterID, requesterRole); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error())
	}
	return response.Success(c, fiber.Map{"message": "order cancelled"})
}
