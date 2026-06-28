package models

type BankQRConfig struct {
	BaseModel
	BankName      string `gorm:"not null" json:"bank_name"`
	AccountName   string `gorm:"not null" json:"account_name"`
	AccountNumber string `gorm:"not null" json:"account_number"`
	QRImageURL    string `json:"qr_image_url"`
	PromptPayID   string `json:"promptpay_id"` // phone (0XXXXXXXXX) or national/tax ID (13 digits)
	IsActive      bool   `gorm:"default:true" json:"is_active"`
}
