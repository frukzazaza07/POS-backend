package repository

import (
	"pos-backend/internal/models"

	"gorm.io/gorm"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(order *models.Order) error {
	return r.db.Create(order).Error
}

func (r *OrderRepository) FindAll(page, limit int, cashierID string) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	q := r.db.Model(&models.Order{})
	if cashierID != "" {
		q = q.Where("cashier_id = ?", cashierID)
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
