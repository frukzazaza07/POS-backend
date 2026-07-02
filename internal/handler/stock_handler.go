package handler

import (
	"strings"

	"pos-backend/internal/service"
	"pos-backend/pkg/i18n"
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

func (h *StockHandler) GetCachedStock(c *fiber.Ctx) error {
	lang := i18n.Lang(c)
	items, err := h.syncSvc.GetCachedStock()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error())
	}
	return response.Success(c, i18n.T(lang, "stock.list"), items)
}

func (h *StockHandler) SyncStock(c *fiber.Ctx) error {
	lang := i18n.Lang(c)
	if err := h.syncSvc.SyncFromInventory(); err != nil {
		return response.Error(c, fiber.StatusBadGateway, i18n.T(lang, "err.inventory_sync_failed")+": "+err.Error())
	}

	items, _ := h.syncSvc.GetCachedStock()
	return response.Success(c, i18n.T(lang, "stock.synced"), fiber.Map{
		"count": len(items),
	})
}

func (h *StockHandler) GetProductByBarcode(c *fiber.Ctx) error {
	lang := i18n.Lang(c)
	barcode := c.Params("barcode")
	product, err := h.invClient.GetProductByBarcode(barcode)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return response.Error(c, fiber.StatusNotFound, i18n.T(lang, "err.product_not_found"))
		}
		return response.Error(c, fiber.StatusBadGateway, i18n.T(lang, "err.inventory_lookup_failed")+": "+err.Error())
	}
	return response.Success(c, i18n.T(lang, "stock.barcode"), product)
}

func (h *StockHandler) CheckAvailability(c *fiber.Ctx) error {
	lang := i18n.Lang(c)
	posProductID := c.Params("pos_product_id")
	quantity := c.QueryInt("quantity", 1)

	avail, err := h.invClient.CheckAvailability(posProductID, quantity)
	if err != nil {
		return response.Error(c, fiber.StatusBadGateway, i18n.T(lang, "err.inventory_check_failed")+": "+err.Error())
	}
	return response.Success(c, i18n.T(lang, "stock.availability"), avail)
}
