package models

import "time"

// StockCache is a local mirror of the Inventory system's stock levels.
// Updated on startup and kept fresh via webhooks from the Inventory system.
type StockCache struct {
	InventoryItemID string    `gorm:"primaryKey" json:"inventory_item_id"`
	SKU             string    `gorm:"not null" json:"sku"`
	Name            string    `gorm:"not null" json:"name"`
	Unit            string    `json:"unit"`
	QuantityInStock float64   `gorm:"not null" json:"quantity_in_stock"`
	MinQuantity     float64   `json:"min_quantity"`
	IsLow           bool      `json:"is_low"`
	IsOut           bool      `json:"is_out"`
	SyncedAt        time.Time `json:"synced_at"`
}
