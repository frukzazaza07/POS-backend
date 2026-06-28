package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"pos-backend/internal/service"
	"pos-backend/pkg/response"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type WebhookHandler struct {
	syncSvc *service.StockSyncService
}

func NewWebhookHandler(syncSvc *service.StockSyncService) *WebhookHandler {
	return &WebhookHandler{syncSvc: syncSvc}
}

// InventoryWebhookPayload mirrors what the Inventory system sends.
type InventoryWebhookPayload struct {
	Event     string          `json:"event"`
	Timestamp string          `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

type webhookStockData struct {
	InventoryItemID     string   `json:"inventory_item_id"`
	SKU                 string   `json:"sku"`
	Name                string   `json:"name"`
	QuantityInStock     float64  `json:"quantity_in_stock"`
	Unit                string   `json:"unit"`
	AffectedPosProducts []string `json:"affected_pos_products"`
}

// InventoryEvent receives webhook events from the Inventory system.
// Verifies HMAC-SHA256 signature before processing.
func (h *WebhookHandler) InventoryEvent(c *fiber.Ctx) error {
	rawBody := c.Body()
	sig := c.Get("X-Inventory-Signature")
	event := c.Get("X-Inventory-Event")

	secret := os.Getenv("INVENTORY_WEBHOOK_SECRET")
	if secret != "" && sig != "" {
		if !verifySignature(rawBody, sig, secret) {
			return response.Error(c, fiber.StatusUnauthorized, "invalid webhook signature")
		}
	}

	var payload InventoryWebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid payload")
	}

	log.Printf("webhook: received event=%s", event)

	switch payload.Event {
	case "STOCK_UPDATED", "STOCK_LOW", "STOCK_OUT":
		var data webhookStockData
		if err := json.Unmarshal(payload.Data, &data); err != nil {
			return response.Error(c, fiber.StatusBadRequest, "invalid stock data")
		}

		sl := service.StockLevel{
			InventoryItemID: data.InventoryItemID,
			SKU:             data.SKU,
			Name:            data.Name,
			Unit:            data.Unit,
			QuantityInStock: data.QuantityInStock,
			IsLow:           payload.Event == "STOCK_LOW",
			IsOut:           payload.Event == "STOCK_OUT",
		}

		if err := h.syncSvc.UpdateFromWebhook(sl); err != nil {
			log.Printf("webhook: cache update failed: %v", err)
		} else {
			log.Printf("webhook: updated cache for %s (qty=%.2f)", data.SKU, data.QuantityInStock)
		}

	case "TEST":
		log.Printf("webhook: test event received from inventory")

	default:
		log.Printf("webhook: unknown event %s — ignored", payload.Event)
	}

	return c.SendStatus(fiber.StatusOK)
}

func verifySignature(payload []byte, signature, secret string) bool {
	expected := fmt.Sprintf("sha256=%s", computeHMAC(payload, secret))
	return hmac.Equal([]byte(expected), []byte(strings.TrimSpace(signature)))
}

func computeHMAC(data []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}
