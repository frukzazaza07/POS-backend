package handler

import (
	"time"

	"pos-backend/internal/service"
	"pos-backend/pkg/i18n"
	"pos-backend/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type ReportHandler struct {
	svc *service.ReportService
}

func NewReportHandler(svc *service.ReportService) *ReportHandler {
	return &ReportHandler{svc: svc}
}

// parseDateRange reads ?from=YYYY-MM-DD&to=YYYY-MM-DD, defaulting to the last 30 days.
func parseDateRange(c *fiber.Ctx) (from, to time.Time) {
	now := time.Now()
	to = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	from = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -29)

	if f := c.Query("from"); f != "" {
		if t, err := time.ParseInLocation("2006-01-02", f, now.Location()); err == nil {
			from = t
		}
	}
	if t := c.Query("to"); t != "" {
		if parsed, err := time.ParseInLocation("2006-01-02", t, now.Location()); err == nil {
			to = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 0, now.Location())
		}
	}
	return
}

func (h *ReportHandler) Summary(c *fiber.Ctx) error {
	lang := i18n.Lang(c)
	from, to := parseDateRange(c)
	data, err := h.svc.GetSummary(from, to)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}
	return response.Success(c, i18n.T(lang, "report.summary"), data)
}

func (h *ReportHandler) DailyRevenue(c *fiber.Ctx) error {
	lang := i18n.Lang(c)
	from, to := parseDateRange(c)
	data, err := h.svc.GetDailyRevenue(from, to)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}
	return response.Success(c, i18n.T(lang, "report.daily_revenue"), data)
}

func (h *ReportHandler) TopProducts(c *fiber.Ctx) error {
	lang := i18n.Lang(c)
	from, to := parseDateRange(c)
	limit := c.QueryInt("limit", 10)
	data, err := h.svc.GetTopProducts(from, to, limit)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}
	return response.Success(c, i18n.T(lang, "report.top_products"), data)
}

func (h *ReportHandler) CategoryRevenue(c *fiber.Ctx) error {
	lang := i18n.Lang(c)
	from, to := parseDateRange(c)
	data, err := h.svc.GetCategoryRevenue(from, to)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}
	return response.Success(c, i18n.T(lang, "report.category_revenue"), data)
}

func (h *ReportHandler) CashierSales(c *fiber.Ctx) error {
	lang := i18n.Lang(c)
	from, to := parseDateRange(c)
	data, err := h.svc.GetCashierSales(from, to)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}
	return response.Success(c, i18n.T(lang, "report.cashier_sales"), data)
}

func (h *ReportHandler) OverduePayLater(c *fiber.Ctx) error {
	lang := i18n.Lang(c)
	data, err := h.svc.GetOverduePayLater()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}
	return response.Success(c, i18n.T(lang, "report.overdue_paylater"), data)
}
