package repository

import (
	"errors"

	"pos-backend/internal/models"

	"gorm.io/gorm"
)

type VatConfigRepository struct {
	db *gorm.DB
}

func NewVatConfigRepository(db *gorm.DB) *VatConfigRepository {
	return &VatConfigRepository{db: db}
}

// Get returns the singleton VAT config, or a disabled default if none has
// been set yet — order creation must never fail just because VAT hasn't been
// configured.
func (r *VatConfigRepository) Get() (*models.VatConfig, error) {
	var config models.VatConfig
	err := r.db.First(&config).Error
	if err == nil {
		return &config, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &models.VatConfig{Enabled: false, Rate: 7, PriceIncludesVat: true}, nil
	}
	return nil, err
}

func (r *VatConfigRepository) Upsert(config *models.VatConfig) error {
	var existing models.VatConfig
	if err := r.db.First(&existing).Error; err != nil {
		return r.db.Create(config).Error
	}
	config.ID = existing.ID
	return r.db.Save(config).Error
}
