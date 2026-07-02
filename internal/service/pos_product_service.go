package service

import (
	"errors"
	"pos-backend/internal/models"
	"pos-backend/internal/repository"
)

type POSProductService struct {
	repo *repository.POSProductRepository
}

func NewPOSProductService(repo *repository.POSProductRepository) *POSProductService {
	return &POSProductService{repo: repo}
}

type CreateProductRequest struct {
	PosProductID string   `json:"pos_product_id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Price        *float64 `json:"price"`
	CostPrice    *float64 `json:"cost_price"`
	Category     string   `json:"category"`
	IsActive     *bool    `json:"is_active"`
}

func (s *POSProductService) Create(req CreateProductRequest) (*models.POSProduct, error) {
	if req.PosProductID == "" || req.Name == "" {
		return nil, errors.New("pos_product_id and name are required")
	}
	if req.Price == nil || *req.Price < 0 {
		return nil, errors.New("price must be non-negative")
	}
	if req.CostPrice != nil && *req.CostPrice < 0 {
		return nil, errors.New("cost_price must be non-negative")
	}
	if existing, _ := s.repo.FindByPosProductID(req.PosProductID); existing != nil {
		return nil, errors.New("pos_product_id already exists")
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	costPrice := 0.0
	if req.CostPrice != nil {
		costPrice = *req.CostPrice
	}

	p := &models.POSProduct{
		PosProductID: req.PosProductID,
		Name:         req.Name,
		Description:  req.Description,
		Price:        *req.Price,
		CostPrice:    costPrice,
		Category:     req.Category,
		IsActive:     isActive,
	}
	if err := s.repo.Create(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *POSProductService) List(search string, page, limit int) ([]models.POSProduct, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.FindAll(search, page, limit)
}

func (s *POSProductService) GetByID(id string) (*models.POSProduct, error) {
	return s.repo.FindByID(id)
}

func (s *POSProductService) Update(id string, req CreateProductRequest) (*models.POSProduct, error) {
	p, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("product not found")
	}

	if req.Name != "" {
		p.Name = req.Name
	}
	if req.Description != "" {
		p.Description = req.Description
	}
	if req.Price != nil && *req.Price >= 0 {
		p.Price = *req.Price
	}
	if req.CostPrice != nil && *req.CostPrice >= 0 {
		p.CostPrice = *req.CostPrice
	}
	if req.Category != "" {
		p.Category = req.Category
	}
	if req.IsActive != nil {
		p.IsActive = *req.IsActive
	}

	if err := s.repo.Update(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *POSProductService) Delete(id string) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return errors.New("product not found")
	}
	return s.repo.Delete(id)
}
