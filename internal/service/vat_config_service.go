package service

import (
	"pos-backend/internal/models"
	"pos-backend/internal/repository"
)

type VatConfigService struct {
	repo *repository.VatConfigRepository
}

func NewVatConfigService(repo *repository.VatConfigRepository) *VatConfigService {
	return &VatConfigService{repo: repo}
}

type UpsertVatConfigRequest struct {
	Enabled          bool    `json:"enabled"`
	Rate             float64 `json:"rate"`
	PriceIncludesVat bool    `json:"price_includes_vat"`
}

func (s *VatConfigService) Get() (*models.VatConfig, error) {
	return s.repo.Get()
}

func (s *VatConfigService) Upsert(req UpsertVatConfigRequest) (*models.VatConfig, error) {
	if req.Rate < 0 || req.Rate > 100 {
		return nil, validationErr("vat rate must be between 0 and 100")
	}
	config := &models.VatConfig{
		Enabled:          req.Enabled,
		Rate:             req.Rate,
		PriceIncludesVat: req.PriceIncludesVat,
	}
	if err := s.repo.Upsert(config); err != nil {
		return nil, err
	}
	return s.repo.Get()
}
