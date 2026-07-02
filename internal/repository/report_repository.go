package repository

import (
	"time"

	"pos-backend/internal/models"

	"gorm.io/gorm"
)

type ReportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

type StatusCount struct {
	Status      string  `json:"status"`
	Count       int64   `json:"count"`
	TotalAmount float64 `json:"total_amount"`
}

type PaymentMethodCount struct {
	PaymentMethod string  `json:"payment_method"`
	Count         int64   `json:"count"`
	TotalAmount   float64 `json:"total_amount"`
}

type DailyRevenue struct {
	Date       string  `json:"date"`
	Revenue    float64 `json:"revenue"`
	OrderCount int64   `json:"order_count"`
}

type TopProduct struct {
	PosProductID string  `json:"pos_product_id"`
	ProductName  string  `json:"product_name"`
	TotalQty     int64   `json:"total_qty"`
	TotalRevenue float64 `json:"total_revenue"`
	TotalCost    float64 `json:"total_cost"`
	Profit       float64 `json:"profit"`
}

type CategoryRevenue struct {
	Category   string  `json:"category"`
	Revenue    float64 `json:"revenue"`
	OrderCount int64   `json:"order_count"`
}

type CashierSales struct {
	CashierID   string  `json:"cashier_id"`
	CashierName string  `json:"cashier_name"`
	OrderCount  int64   `json:"order_count"`
	Revenue     float64 `json:"revenue"`
}

func (r *ReportRepository) GetCompletedRevenueSummary(from, to time.Time) (revenue, cost float64, count int64, err error) {
	type result struct {
		TotalRevenue float64
		TotalCost    float64
		OrderCount   int64
	}
	var res result
	err = r.db.Model(&models.Order{}).
		Select("COALESCE(SUM(total_amount), 0) as total_revenue, COALESCE(SUM(total_cost), 0) as total_cost, COUNT(*) as order_count").
		Where("status = ? AND created_at BETWEEN ? AND ?", models.OrderStatusCompleted, from, to).
		Scan(&res).Error
	return res.TotalRevenue, res.TotalCost, res.OrderCount, err
}

func (r *ReportRepository) GetStatusCounts(from, to time.Time) ([]StatusCount, error) {
	var results []StatusCount
	err := r.db.Model(&models.Order{}).
		Select("status, COUNT(*) as count, COALESCE(SUM(total_amount), 0) as total_amount").
		Where("created_at BETWEEN ? AND ?", from, to).
		Group("status").
		Scan(&results).Error
	return results, err
}

func (r *ReportRepository) GetPaymentMethodCounts(from, to time.Time) ([]PaymentMethodCount, error) {
	var results []PaymentMethodCount
	err := r.db.Model(&models.Order{}).
		Select("payment_method, COUNT(*) as count, COALESCE(SUM(CASE WHEN status = 'COMPLETED' THEN total_amount ELSE 0 END), 0) as total_amount").
		Where("created_at BETWEEN ? AND ?", from, to).
		Group("payment_method").
		Scan(&results).Error
	return results, err
}

func (r *ReportRepository) GetDailyRevenue(from, to time.Time) ([]DailyRevenue, error) {
	var results []DailyRevenue
	err := r.db.Raw(`
		SELECT DATE(created_at) AS date,
		       COALESCE(SUM(total_amount), 0) AS revenue,
		       COUNT(*) AS order_count
		FROM orders
		WHERE status = ? AND created_at BETWEEN ? AND ? AND deleted_at IS NULL
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`, models.OrderStatusCompleted, from, to).Scan(&results).Error
	return results, err
}

func (r *ReportRepository) GetTopProducts(from, to time.Time, limit int) ([]TopProduct, error) {
	var results []TopProduct
	err := r.db.Raw(`
		SELECT oi.pos_product_id,
		       oi.product_name,
		       SUM(oi.quantity)                          AS total_qty,
		       SUM(oi.subtotal)                           AS total_revenue,
		       SUM(oi.cost_price * oi.quantity)           AS total_cost,
		       SUM(oi.subtotal - oi.cost_price * oi.quantity) AS profit
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE o.status = ? AND o.created_at BETWEEN ? AND ?
		  AND o.deleted_at IS NULL AND oi.deleted_at IS NULL
		GROUP BY oi.pos_product_id, oi.product_name
		ORDER BY total_qty DESC
		LIMIT ?
	`, models.OrderStatusCompleted, from, to, limit).Scan(&results).Error
	return results, err
}

func (r *ReportRepository) GetCategoryRevenue(from, to time.Time) ([]CategoryRevenue, error) {
	var results []CategoryRevenue
	err := r.db.Raw(`
		SELECT COALESCE(NULLIF(p.category, ''), 'Uncategorized') AS category,
		       SUM(oi.subtotal)       AS revenue,
		       COUNT(DISTINCT o.id)   AS order_count
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		LEFT JOIN pos_products p ON p.pos_product_id = oi.pos_product_id AND p.deleted_at IS NULL
		WHERE o.status = ? AND o.created_at BETWEEN ? AND ?
		  AND o.deleted_at IS NULL AND oi.deleted_at IS NULL
		GROUP BY COALESCE(NULLIF(p.category, ''), 'Uncategorized')
		ORDER BY revenue DESC
	`, models.OrderStatusCompleted, from, to).Scan(&results).Error
	return results, err
}

func (r *ReportRepository) GetCashierSales(from, to time.Time) ([]CashierSales, error) {
	var results []CashierSales
	err := r.db.Raw(`
		SELECT o.cashier_id,
		       u.name                          AS cashier_name,
		       COUNT(*)                         AS order_count,
		       COALESCE(SUM(o.total_amount), 0) AS revenue
		FROM orders o
		JOIN users u ON u.id = o.cashier_id AND u.deleted_at IS NULL
		WHERE o.status = ? AND o.created_at BETWEEN ? AND ? AND o.deleted_at IS NULL
		GROUP BY o.cashier_id, u.name
		ORDER BY revenue DESC
	`, models.OrderStatusCompleted, from, to).Scan(&results).Error
	return results, err
}

func (r *ReportRepository) GetOverduePayLater() ([]models.Order, error) {
	var orders []models.Order
	err := r.db.Preload("Cashier").Preload("Items").
		Where("payment_method = ? AND is_paid = false AND payment_due_date < NOW() AND status = ?",
			models.PaymentPayLater, models.OrderStatusCompleted).
		Order("payment_due_date ASC").
		Find(&orders).Error
	return orders, err
}
