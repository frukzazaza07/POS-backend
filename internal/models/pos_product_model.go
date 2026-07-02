package models

// POSProduct is the local product catalog.
// PosProductID must match the pos_product_id registered in the Inventory system.
type POSProduct struct {
	BaseModel
	PosProductID string  `gorm:"uniqueIndex;not null" json:"pos_product_id"`
	Name         string  `gorm:"not null" json:"name"`
	Description  string  `json:"description"`
	Price        float64 `gorm:"not null" json:"price"`
	// CostPrice is the per-unit cost used to compute profit. Admin-only — stripped
	// from responses for non-admin roles. Manually set unless synced from the
	// Inventory system's recipe cost (see INVENTORY_COST_INTEGRATION.md).
	CostPrice float64 `gorm:"default:0" json:"cost_price,omitempty"`
	Category  string  `json:"category"`
	IsActive  bool    `gorm:"default:true" json:"is_active"`
}
