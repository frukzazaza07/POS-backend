package handler

import (
	"errors"
	"pos-backend/internal/repository"
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
		var valErr *service.ValidationError
		if order != nil || errors.As(err, &valErr) {
			return response.Error(c, fiber.StatusBadRequest, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.Created(c, order)
}

func (h *OrderHandler) List(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	role, _ := c.Locals("role").(string)
	filter := repository.OrderFilter{
		PaymentMethod: c.Query("payment_method"),
		Status:        c.Query("status"),
		OverdueOnly:   c.QueryBool("overdue"),
	}
	if role != "admin" {
		filter.CashierID, _ = c.Locals("userID").(string)
	}

	orders, total, err := h.svc.ListOrders(page, limit, filter)
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
		var valErr *service.ValidationError
		if errors.As(err, &valErr) {
			return response.Error(c, fiber.StatusBadRequest, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}
	return response.Success(c, fiber.Map{"message": "order cancelled"})
}

func (h *OrderHandler) MarkPaid(c *fiber.Ctx) error {
	if err := h.svc.MarkAsPaid(c.Params("id")); err != nil {
		var valErr *service.ValidationError
		if errors.As(err, &valErr) {
			return response.Error(c, fiber.StatusBadRequest, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}
	order, _ := h.svc.GetOrder(c.Params("id"))
	return response.Success(c, fiber.Map{"message": "marked as paid", "order": order})
}
