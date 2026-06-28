package repository

import (
	"pos-backend/internal/models"

	"gorm.io/gorm"
)

type BankQRConfigRepository struct {
	db *gorm.DB
}

func NewBankQRConfigRepository(db *gorm.DB) *BankQRConfigRepository {
	return &BankQRConfigRepository{db: db}
}

func (r *BankQRConfigRepository) Get() (*models.BankQRConfig, error) {
	var config models.BankQRConfig
	if err := r.db.Where("is_active = true").First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *BankQRConfigRepository) Upsert(config *models.BankQRConfig) error {
	var existing models.BankQRConfig
	if err := r.db.First(&existing).Error; err != nil {
		return r.db.Create(config).Error
	}
	config.ID = existing.ID
	return r.db.Save(config).Error
}
