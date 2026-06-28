package repository

import (
	"pos-backend/internal/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StockCacheRepository struct {
	db *gorm.DB
}

func NewStockCacheRepository(db *gorm.DB) *StockCacheRepository {
	return &StockCacheRepository{db: db}
}

func (r *StockCacheRepository) Upsert(item models.StockCache) error {
	item.SyncedAt = time.Now()
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "inventory_item_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"sku", "name", "unit", "quantity_in_stock", "min_quantity", "is_low", "is_out", "synced_at"}),
	}).Create(&item).Error
}

func (r *StockCacheRepository) FindAll() ([]models.StockCache, error) {
	var items []models.StockCache
	err := r.db.Order("name ASC").Find(&items).Error
	return items, err
}

func (r *StockCacheRepository) FindByInventoryItemID(id string) (*models.StockCache, error) {
	var item models.StockCache
	if err := r.db.First(&item, "inventory_item_id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}
