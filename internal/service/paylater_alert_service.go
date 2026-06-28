package service

import (
	"bytes"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"os"
	"pos-backend/internal/repository"
	"time"
)

type PayLaterAlertService struct {
	orderRepo *repository.OrderRepository
	alertURL  string
}

func NewPayLaterAlertService(orderRepo *repository.OrderRepository) *PayLaterAlertService {
	return &PayLaterAlertService{
		orderRepo: orderRepo,
		alertURL:  os.Getenv("ALERT_WEBHOOK_URL"),
	}
}

type overdueAlert struct {
	Event          string    `json:"event"`
	OrderID        string    `json:"order_id"`
	PosOrderID     string    `json:"pos_order_id"`
	CustomerName   string    `json:"customer_name"`
	CustomerPhone  string    `json:"customer_phone"`
	TotalAmount    float64   `json:"total_amount"`
	PaymentDueDate time.Time `json:"payment_due_date"`
	DaysOverdue    int       `json:"days_overdue"`
}

// CheckAndAlert finds all overdue unpaid PAY_LATER orders and logs/POSTs an alert for each.
// Called by StartAlertLoop — runs until the order is marked paid.
func (s *PayLaterAlertService) CheckAndAlert() {
	orders, err := s.orderRepo.FindOverduePaylater()
	if err != nil {
		log.Printf("paylater-alert: query failed: %v", err)
		return
	}
	if len(orders) == 0 {
		return
	}
	log.Printf("paylater-alert: %d overdue PAY_LATER order(s) found", len(orders))

	for _, o := range orders {
		daysOverdue := 0
		var dueDate time.Time
		if o.PaymentDueDate != nil {
			dueDate = *o.PaymentDueDate
			daysOverdue = int(math.Ceil(time.Since(dueDate).Hours() / 24))
		}

		alert := overdueAlert{
			Event:          "PAY_LATER_OVERDUE",
			OrderID:        o.ID,
			PosOrderID:     o.PosOrderID,
			CustomerName:   o.CustomerName,
			CustomerPhone:  o.CustomerPhone,
			TotalAmount:    o.TotalAmount,
			PaymentDueDate: dueDate,
			DaysOverdue:    daysOverdue,
		}

		log.Printf("paylater-alert: [OVERDUE] order=%s customer=%s phone=%s amount=%.2f days_overdue=%d",
			o.PosOrderID, o.CustomerName, o.CustomerPhone, o.TotalAmount, daysOverdue)

		if s.alertURL != "" {
			s.postAlert(alert)
		}
	}
}

func (s *PayLaterAlertService) postAlert(alert overdueAlert) {
	payload, err := json.Marshal(alert)
	if err != nil {
		log.Printf("paylater-alert: marshal error: %v", err)
		return
	}
	resp, err := http.Post(s.alertURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		log.Printf("paylater-alert: webhook POST failed: %v", err)
		return
	}
	defer resp.Body.Close()
	log.Printf("paylater-alert: webhook sent for %s (HTTP %d)", alert.PosOrderID, resp.StatusCode)
}

// StartAlertLoop runs CheckAndAlert every interval until stopped.
func (s *PayLaterAlertService) StartAlertLoop(interval time.Duration) {
	go func() {
		// Run once immediately on startup to catch any pre-existing overdue orders
		s.CheckAndAlert()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			s.CheckAndAlert()
		}
	}()
}
