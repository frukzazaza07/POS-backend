package repository

import (
	"pos-backend/internal/models"

	"gorm.io/gorm"
)

type POSProductRepository struct {
	db *gorm.DB
}

func NewPOSProductRepository(db *gorm.DB) *POSProductRepository {
	return &POSProductRepository{db: db}
}

func (r *POSProductRepository) Create(p *models.POSProduct) error {
	return r.db.Create(p).Error
}

func (r *POSProductRepository) FindAll(search string, page, limit int) ([]models.POSProduct, int64, error) {
	var items []models.POSProduct
	var total int64

	q := r.db.Model(&models.POSProduct{})
	if search != "" {
		q = q.Where("name ILIKE ? OR pos_product_id ILIKE ? OR category ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	q.Count(&total)
	err := q.Offset((page - 1) * limit).Limit(limit).Find(&items).Error
	return items, total, err
}

func (r *POSProductRepository) FindByID(id string) (*models.POSProduct, error) {
	var p models.POSProduct
	if err := r.db.First(&p, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *POSProductRepository) FindByPosProductID(posProductID string) (*models.POSProduct, error) {
	var p models.POSProduct
	if err := r.db.First(&p, "pos_product_id = ?", posProductID).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *POSProductRepository) Update(p *models.POSProduct) error {
	return r.db.Save(p).Error
}

func (r *POSProductRepository) Delete(id string) error {
	return r.db.Delete(&models.POSProduct{}, "id = ?", id).Error
}
