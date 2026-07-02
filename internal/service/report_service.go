package service

import (
	"time"

	"pos-backend/internal/models"
	"pos-backend/internal/repository"
)

type ReportService struct {
	repo *repository.ReportRepository
}

func NewReportService(repo *repository.ReportRepository) *ReportService {
	return &ReportService{repo: repo}
}

type SummaryReport struct {
	From          string                        `json:"from"`
	To            string                        `json:"to"`
	TotalRevenue  float64                       `json:"total_revenue"`
	OrderCount    int64                         `json:"order_count"`
	AvgOrderValue float64                       `json:"avg_order_value"`
	ByStatus      []repository.StatusCount      `json:"by_status"`
	ByPayment     []repository.PaymentMethodCount `json:"by_payment_method"`
}

func (s *ReportService) GetSummary(from, to time.Time) (*SummaryReport, error) {
	revenue, count, err := s.repo.GetCompletedRevenueSummary(from, to)
	if err != nil {
		return nil, err
	}
	byStatus, err := s.repo.GetStatusCounts(from, to)
	if err != nil {
		return nil, err
	}
	byPayment, err := s.repo.GetPaymentMethodCounts(from, to)
	if err != nil {
		return nil, err
	}
	avg := 0.0
	if count > 0 {
		avg = revenue / float64(count)
	}
	return &SummaryReport{
		From:          from.Format("2006-01-02"),
		To:            to.Format("2006-01-02"),
		TotalRevenue:  revenue,
		OrderCount:    count,
		AvgOrderValue: avg,
		ByStatus:      byStatus,
		ByPayment:     byPayment,
	}, nil
}

func (s *ReportService) GetDailyRevenue(from, to time.Time) ([]repository.DailyRevenue, error) {
	return s.repo.GetDailyRevenue(from, to)
}

func (s *ReportService) GetTopProducts(from, to time.Time, limit int) ([]repository.TopProduct, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	return s.repo.GetTopProducts(from, to, limit)
}

func (s *ReportService) GetCategoryRevenue(from, to time.Time) ([]repository.CategoryRevenue, error) {
	return s.repo.GetCategoryRevenue(from, to)
}

func (s *ReportService) GetCashierSales(from, to time.Time) ([]repository.CashierSales, error) {
	return s.repo.GetCashierSales(from, to)
}

func (s *ReportService) GetOverduePayLater() ([]models.Order, error) {
	return s.repo.GetOverduePayLater()
}
