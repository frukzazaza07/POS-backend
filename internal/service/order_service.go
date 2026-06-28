package service

import (
	"fmt"
	"pos-backend/internal/models"
	"pos-backend/internal/repository"
	"time"
)

// ValidationError marks client-caused errors (bad input, business rule violations).
// The handler maps these to HTTP 400 instead of 500.
type ValidationError struct{ msg string }

func (e *ValidationError) Error() string { return e.msg }

func validationErr(msg string) error          { return &ValidationError{msg: msg} }
func validationErrf(f string, a ...any) error { return &ValidationError{msg: fmt.Sprintf(f, a...)} }

type OrderService struct {
	orderRepo   *repository.OrderRepository
	productRepo *repository.POSProductRepository
	invClient   *InventoryClient
}

func NewOrderService(
	orderRepo *repository.OrderRepository,
	productRepo *repository.POSProductRepository,
	invClient *InventoryClient,
) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		invClient:   invClient,
	}
}

type OrderItemRequest struct {
	PosProductID string `json:"pos_product_id"`
	Quantity     int    `json:"quantity"`
}

type CreateOrderRequest struct {
	Notes          string             `json:"notes"`
	Items          []OrderItemRequest `json:"items"`
	PaymentMethod  string             `json:"payment_method"`
	CustomerName   string             `json:"customer_name"`
	CustomerPhone  string             `json:"customer_phone"`
	PaymentDueDays int                `json:"payment_due_days"`
}

func (s *OrderService) CreateOrder(cashierID string, req CreateOrderRequest) (*models.Order, error) {
	if len(req.Items) == 0 {
		return nil, validationErr("order must have at least one item")
	}

	paymentMethod := models.PaymentMethod(req.PaymentMethod)
	if paymentMethod == "" {
		paymentMethod = models.PaymentCash
	}
	switch paymentMethod {
	case models.PaymentCash, models.PaymentBankQR, models.PaymentPayLater:
	default:
		return nil, validationErrf("invalid payment_method: %s (allowed: CASH, BANK_QRCODE, PAY_LATER)", req.PaymentMethod)
	}
	if paymentMethod == models.PaymentPayLater {
		if req.CustomerName == "" || req.CustomerPhone == "" {
			return nil, validationErr("customer_name and customer_phone are required for PAY_LATER")
		}
	}

	var orderItems []models.OrderItem
	var totalAmount float64

	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return nil, validationErrf("quantity for %s must be positive", item.PosProductID)
		}
		product, err := s.productRepo.FindByPosProductID(item.PosProductID)
		if err != nil {
			return nil, validationErrf("product not found: %s", item.PosProductID)
		}
		if !product.IsActive {
			return nil, validationErrf("product is not active: %s", item.PosProductID)
		}
		subtotal := product.Price * float64(item.Quantity)
		totalAmount += subtotal
		orderItems = append(orderItems, models.OrderItem{
			PosProductID: item.PosProductID,
			ProductName:  product.Name,
			Quantity:     item.Quantity,
			UnitPrice:    product.Price,
			Subtotal:     subtotal,
		})
	}

	posOrderID := models.GeneratePosOrderID()
	order := &models.Order{
		PosOrderID:    posOrderID,
		CashierID:     cashierID,
		Status:        models.OrderStatusPending,
		TotalAmount:   totalAmount,
		Notes:         req.Notes,
		Items:         orderItems,
		PaymentMethod: paymentMethod,
	}

	if paymentMethod == models.PaymentPayLater {
		order.CustomerName = req.CustomerName
		order.CustomerPhone = req.CustomerPhone
		dueDays := req.PaymentDueDays
		if dueDays <= 0 {
			dueDays = 7
		}
		due := time.Now().AddDate(0, 0, dueDays)
		order.PaymentDueDate = &due
	}

	if err := s.orderRepo.Create(order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	var deductItems []DeductItem
	for _, item := range req.Items {
		deductItems = append(deductItems, DeductItem{
			PosProductID: item.PosProductID,
			Quantity:     item.Quantity,
		})
	}

	deductResp, err := s.invClient.DeductStock(DeductRequest{
		PosOrderID: posOrderID,
		Items:      deductItems,
	})
	if err != nil {
		_ = s.orderRepo.UpdateStatus(order.ID, models.OrderStatusFailed, err.Error())
		order.Status = models.OrderStatusFailed
		order.FailReason = err.Error()
		return order, fmt.Errorf("stock deduction failed: %w", err)
	}
	_ = deductResp

	if err := s.orderRepo.UpdateStatus(order.ID, models.OrderStatusCompleted, ""); err != nil {
		return nil, err
	}
	order.Status = models.OrderStatusCompleted

	return s.orderRepo.FindByID(order.ID)
}

func (s *OrderService) ListOrders(page, limit int, filter repository.OrderFilter) ([]models.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.orderRepo.FindAll(page, limit, filter)
}

func (s *OrderService) GetOrder(id string) (*models.Order, error) {
	return s.orderRepo.FindByID(id)
}

func (s *OrderService) CancelOrder(id, requesterID, requesterRole string) error {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		return validationErr("order not found")
	}
	if order.Status != models.OrderStatusPending {
		return validationErr("only PENDING orders can be cancelled")
	}
	if requesterRole != "admin" && order.CashierID != requesterID {
		return validationErr("unauthorized to cancel this order")
	}
	return s.orderRepo.UpdateStatus(id, models.OrderStatusCancelled, "")
}

func (s *OrderService) MarkAsPaid(id string) error {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		return validationErr("order not found")
	}
	if order.PaymentMethod != models.PaymentPayLater {
		return validationErr("only PAY_LATER orders can be marked as paid")
	}
	if order.IsPaid {
		return validationErr("order is already paid")
	}
	if order.Status != models.OrderStatusCompleted {
		return validationErr("only COMPLETED orders can be marked as paid")
	}
	return s.orderRepo.MarkAsPaid(id)
}
