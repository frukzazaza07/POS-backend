package service

import (
	"errors"
	"fmt"
	"pos-backend/internal/models"
	"pos-backend/internal/repository"
)

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
	Notes string             `json:"notes"`
	Items []OrderItemRequest `json:"items"`
}

// CreateOrder validates stock, creates the order in DB, then deducts from inventory (idempotent).
func (s *OrderService) CreateOrder(cashierID string, req CreateOrderRequest) (*models.Order, error) {
	if len(req.Items) == 0 {
		return nil, errors.New("order must have at least one item")
	}

	// 1. Resolve each item from local product catalog
	var orderItems []models.OrderItem
	var totalAmount float64

	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("quantity for %s must be positive", item.PosProductID)
		}
		product, err := s.productRepo.FindByPosProductID(item.PosProductID)
		if err != nil {
			return nil, fmt.Errorf("product not found: %s", item.PosProductID)
		}
		if !product.IsActive {
			return nil, fmt.Errorf("product is not active: %s", item.PosProductID)
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

	// 2. Create order in DB (PENDING)
	posOrderID := models.GeneratePosOrderID()
	order := &models.Order{
		PosOrderID:  posOrderID,
		CashierID:   cashierID,
		Status:      models.OrderStatusPending,
		TotalAmount: totalAmount,
		Notes:       req.Notes,
		Items:       orderItems,
	}
	if err := s.orderRepo.Create(order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// 3. Build deduct request for inventory
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
		// Stock insufficient or inventory unreachable — mark order as failed
		_ = s.orderRepo.UpdateStatus(order.ID, models.OrderStatusFailed, err.Error())
		order.Status = models.OrderStatusFailed
		order.FailReason = err.Error()
		return order, fmt.Errorf("stock deduction failed: %w", err)
	}

	// "already_processed" means idempotent replay — still a success
	_ = deductResp

	// 4. Mark completed
	if err := s.orderRepo.UpdateStatus(order.ID, models.OrderStatusCompleted, ""); err != nil {
		return nil, err
	}
	order.Status = models.OrderStatusCompleted

	// Reload with relations
	return s.orderRepo.FindByID(order.ID)
}

func (s *OrderService) ListOrders(page, limit int, cashierID string) ([]models.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.orderRepo.FindAll(page, limit, cashierID)
}

func (s *OrderService) GetOrder(id string) (*models.Order, error) {
	return s.orderRepo.FindByID(id)
}

func (s *OrderService) CancelOrder(id, requesterID, requesterRole string) error {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		return errors.New("order not found")
	}
	if order.Status != models.OrderStatusPending {
		return errors.New("only PENDING orders can be cancelled")
	}
	if requesterRole != "admin" && order.CashierID != requesterID {
		return errors.New("unauthorized to cancel this order")
	}
	return s.orderRepo.UpdateStatus(id, models.OrderStatusCancelled, "")
}
