package handler

import "pos-backend/internal/models"

// Cost and profit figures are admin-only. Cashiers get the same product/order
// payloads with these fields zeroed out before serialization.

func redactProductCost(p *models.POSProduct, role string) {
	if p == nil || role == "admin" {
		return
	}
	p.CostPrice = 0
}

func redactProductCosts(items []models.POSProduct, role string) {
	if role == "admin" {
		return
	}
	for i := range items {
		items[i].CostPrice = 0
	}
}

func redactOrderCost(o *models.Order, role string) {
	if o == nil || role == "admin" {
		return
	}
	o.TotalCost = 0
	o.Profit = 0
	for i := range o.Items {
		o.Items[i].CostPrice = 0
	}
}

func redactOrderCosts(orders []models.Order, role string) {
	if role == "admin" {
		return
	}
	for i := range orders {
		redactOrderCost(&orders[i], role)
	}
}
