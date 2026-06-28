package service

import (
	"log"
	"pos-backend/internal/models"
	"pos-backend/internal/repository"
	"time"
)

type StockSyncService struct {
	stockRepo *repository.StockCacheRepository
	invClient *InventoryClient
}

func NewStockSyncService(repo *repository.StockCacheRepository, client *InventoryClient) *StockSyncService {
	return &StockSyncService{stockRepo: repo, invClient: client}
}

// SyncFromInventory fetches all stock levels from the Inventory system and upserts into local cache.
func (s *StockSyncService) SyncFromInventory() error {
	levels, err := s.invClient.GetStockLevels()
	if err != nil {
		return err
	}

	now := time.Now()
	for _, l := range levels {
		item := models.StockCache{
			InventoryItemID: l.InventoryItemID,
			SKU:             l.SKU,
			Name:            l.Name,
			Unit:            l.Unit,
			QuantityInStock: l.QuantityInStock,
			MinQuantity:     l.MinQuantity,
			IsLow:           l.IsLow,
			IsOut:           l.IsOut,
			SyncedAt:        now,
		}
		if err := s.stockRepo.Upsert(item); err != nil {
			log.Printf("stock sync: upsert %s failed: %v", l.InventoryItemID, err)
		}
	}

	log.Printf("stock sync: synced %d items from inventory", len(levels))
	return nil
}

// UpdateFromWebhook updates a single item in the cache — called by the webhook handler.
func (s *StockSyncService) UpdateFromWebhook(l StockLevel) error {
	item := models.StockCache{
		InventoryItemID: l.InventoryItemID,
		SKU:             l.SKU,
		Name:            l.Name,
		Unit:            l.Unit,
		QuantityInStock: l.QuantityInStock,
		MinQuantity:     l.MinQuantity,
		IsLow:           l.IsLow,
		IsOut:           l.IsOut,
	}
	return s.stockRepo.Upsert(item)
}

func (s *StockSyncService) GetCachedStock() ([]models.StockCache, error) {
	return s.stockRepo.FindAll()
}

// StartPeriodicSync runs a background sync every interval.
func (s *StockSyncService) StartPeriodicSync(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			if err := s.SyncFromInventory(); err != nil {
				log.Printf("periodic stock sync failed: %v", err)
			}
		}
	}()
}
