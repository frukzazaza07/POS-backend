package handler

import (
	"errors"

	"pos-backend/internal/repository"
	"pos-backend/internal/service"
	"pos-backend/pkg/i18n"
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
	lang := i18n.Lang(c)
	cashierID, _ := c.Locals("userID").(string)

	var req service.CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, i18n.T(lang, "err.invalid_body"))
	}

	order, err := h.svc.CreateOrder(cashierID, req)
	if err != nil {
		var valErr *service.ValidationError
		if errors.As(err, &valErr) {
			return response.Error(c, fiber.StatusBadRequest, i18n.T(lang, err.Error()))
		}
		if order != nil {
			// stock deduction failed — order saved as FAILED in DB
			return response.Error(c, fiber.StatusBadGateway, i18n.T(lang, "err.stock_deduction_failed"))
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.Created(c, i18n.T(lang, "order.created"), order)
}

func (h *OrderHandler) List(c *fiber.Ctx) error {
	lang := i18n.Lang(c)
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

	return response.Success(c, i18n.T(lang, "order.list"), response.PaginatedData{
		Items: orders,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

func (h *OrderHandler) Get(c *fiber.Ctx) error {
	lang := i18n.Lang(c)
	order, err := h.svc.GetOrder(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, i18n.T(lang, "err.order_not_found"))
	}
	return response.Success(c, i18n.T(lang, "order.get"), order)
}

func (h *OrderHandler) Cancel(c *fiber.Ctx) error {
	lang := i18n.Lang(c)
	requesterID, _ := c.Locals("userID").(string)
	requesterRole, _ := c.Locals("role").(string)

	if err := h.svc.CancelOrder(c.Params("id"), requesterID, requesterRole); err != nil {
		var valErr *service.ValidationError
		if errors.As(err, &valErr) {
			return response.Error(c, fiber.StatusBadRequest, i18n.T(lang, err.Error()))
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}
	return response.Success(c, i18n.T(lang, "order.cancelled"), nil)
}

func (h *OrderHandler) MarkPaid(c *fiber.Ctx) error {
	lang := i18n.Lang(c)
	if err := h.svc.MarkAsPaid(c.Params("id")); err != nil {
		var valErr *service.ValidationError
		if errors.As(err, &valErr) {
			return response.Error(c, fiber.StatusBadRequest, i18n.T(lang, err.Error()))
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}
	order, _ := h.svc.GetOrder(c.Params("id"))
	return response.Success(c, i18n.T(lang, "order.paid"), order)
}
