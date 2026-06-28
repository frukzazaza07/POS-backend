package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusCompleted OrderStatus = "COMPLETED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
	OrderStatusFailed    OrderStatus = "FAILED"
)

type Order struct {
	BaseModel
	PosOrderID  string      `gorm:"uniqueIndex;not null" json:"pos_order_id"`
	CashierID   string      `gorm:"not null" json:"cashier_id"`
	Cashier     *User       `gorm:"foreignKey:CashierID" json:"cashier,omitempty"`
	Status      OrderStatus `gorm:"type:varchar(20);default:PENDING;not null" json:"status"`
	TotalAmount float64     `gorm:"not null" json:"total_amount"`
	Notes       string      `json:"notes"`
	FailReason  string      `json:"fail_reason,omitempty"`
	Items       []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

type OrderItem struct {
	BaseModel
	OrderID      string  `gorm:"not null;index" json:"order_id"`
	PosProductID string  `gorm:"not null" json:"pos_product_id"`
	ProductName  string  `gorm:"not null" json:"product_name"`
	Quantity     int     `gorm:"not null" json:"quantity"`
	UnitPrice    float64 `gorm:"not null" json:"unit_price"`
	Subtotal     float64 `gorm:"not null" json:"subtotal"`
}

func GeneratePosOrderID() string {
	suffix := strings.ToUpper(uuid.New().String()[:8])
	return fmt.Sprintf("ORDER-%s-%s", time.Now().Format("20060102"), suffix)
}
