package models

// VatConfig is a singleton (one row) config for how VAT is applied to orders.
// PriceIncludesVat controls how Rate is applied to an order's TotalAmount:
//   - true  (typical Thai retail): product prices already include VAT.
//     TotalAmount stays as-is; VatAmount is the tax portion backed out of it.
//   - false: product prices are pre-tax. VAT is added on top, increasing
//     TotalAmount by VatAmount.
type VatConfig struct {
	BaseModel
	Enabled          bool    `gorm:"default:false" json:"enabled"`
	Rate             float64 `gorm:"default:7" json:"rate"` // percent, e.g. 7 = 7%
	PriceIncludesVat bool    `gorm:"default:true" json:"price_includes_vat"`
}
