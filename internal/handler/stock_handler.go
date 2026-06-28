package handler

import (
	"pos-backend/internal/service"
	"pos-backend/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type StockHandler struct {
	syncSvc   *service.StockSyncService
	invClient *service.InventoryClient
}

func NewStockHandler(syncSvc *service.StockSyncService, invClient *service.InventoryClient) *StockHandler {
	return &StockHandler{syncSvc: syncSvc, invClient: invClient}
}

// GetCachedStock returns the local stock cache — fast, no inventory round-trip.
func (h *StockHandler) GetCachedStock(c *fiber.Ctx) error {
	items, err := h.syncSvc.GetCachedStock()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}
	return response.Success(c, items)
}

// SyncStock forces a full re-sync from the Inventory system (admin only).
func (h *StockHandler) SyncStock(c *fiber.Ctx) error {
	if err := h.syncSvc.SyncFromInventory(); err != nil {
		return response.Error(c, fiber.StatusBadGateway, "inventory sync failed: "+err.Error())
	}

	items, _ := h.syncSvc.GetCachedStock()
	return response.Success(c, fiber.Map{
		"message": "sync complete",
		"count":   len(items),
	})
}

// CheckAvailability proxies to the Inventory system in real time.
func (h *StockHandler) CheckAvailability(c *fiber.Ctx) error {
	posProductID := c.Params("pos_product_id")
	quantity := c.QueryInt("quantity", 1)

	avail, err := h.invClient.CheckAvailability(posProductID, quantity)
	if err != nil {
		return response.Error(c, fiber.StatusBadGateway, "inventory check failed: "+err.Error())
	}
	return response.Success(c, avail)
}
