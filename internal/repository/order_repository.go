package repository

import (
	"time"

	"pos-backend/internal/models"

	"gorm.io/gorm"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

type OrderFilter struct {
	CashierID     string
	PaymentMethod string
	Status        string
	OverdueOnly   bool
}

func (r *OrderRepository) Create(order *models.Order) error {
	return r.db.Create(order).Error
}

func (r *OrderRepository) FindAll(page, limit int, filter OrderFilter) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	q := r.db.Model(&models.Order{})
	if filter.CashierID != "" {
		q = q.Where("cashier_id = ?", filter.CashierID)
	}
	if filter.PaymentMethod != "" {
		q = q.Where("payment_method = ?", filter.PaymentMethod)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.OverdueOnly {
		q = q.Where("payment_method = ? AND is_paid = false AND payment_due_date < ? AND status = ?",
			models.PaymentPayLater, time.Now(), models.OrderStatusCompleted)
	}
	q.Count(&total)
	err := q.Preload("Items").Preload("Cashier").
		Order("created_at DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&orders).Error
	return orders, total, err
}

func (r *OrderRepository) FindByID(id string) (*models.Order, error) {
	var order models.Order
	if err := r.db.Preload("Items").Preload("Cashier").
		First(&order, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) FindByPosOrderID(posOrderID string) (*models.Order, error) {
	var order models.Order
	if err := r.db.First(&order, "pos_order_id = ?", posOrderID).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) UpdateStatus(id string, status models.OrderStatus, failReason string) error {
	updates := map[string]interface{}{"status": status}
	if failReason != "" {
		updates["fail_reason"] = failReason
	}
	return r.db.Model(&models.Order{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateCosts persists the order's total cost and per-item cost snapshots.
// Called after order creation once the authoritative cost is known (either the
// POSProduct fallback set at creation, or a more accurate figure the Inventory
// system returned in the stock-deduct response).
func (r *OrderRepository) UpdateCosts(orderID string, totalCost float64, items []models.OrderItem) error {
	if err := r.db.Model(&models.Order{}).Where("id = ?", orderID).
		Update("total_cost", totalCost).Error; err != nil {
		return err
	}
	for _, item := range items {
		if err := r.db.Model(&models.OrderItem{}).Where("id = ?", item.ID).
			Update("cost_price", item.CostPrice).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *OrderRepository) MarkAsPaid(id string) error {
	now := time.Now()
	return r.db.Model(&models.Order{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_paid": true,
		"paid_at": now,
	}).Error
}

func (r *OrderRepository) FindOverduePaylater() ([]models.Order, error) {
	var orders []models.Order
	err := r.db.Where(
		"payment_method = ? AND is_paid = false AND payment_due_date < ? AND status = ?",
		models.PaymentPayLater, time.Now(), models.OrderStatusCompleted,
	).Find(&orders).Error
	return orders, err
}
