package models

// POSProduct is the local product catalog.
// PosProductID must match the pos_product_id registered in the Inventory system.
type POSProduct struct {
	BaseModel
	PosProductID string  `gorm:"uniqueIndex;not null" json:"pos_product_id"`
	Name         string  `gorm:"not null" json:"name"`
	Description  string  `json:"description"`
	Price        float64 `gorm:"not null" json:"price"`
	Category     string  `json:"category"`
	IsActive     bool    `gorm:"default:true" json:"is_active"`
}
