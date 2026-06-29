package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// InventoryClient calls the Inventory system's POS API endpoints.
// Auth: X-API-Key header (server-to-server).
type InventoryClient struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func NewInventoryClient() *InventoryClient {
	return &InventoryClient{
		baseURL: os.Getenv("INVENTORY_BASE_URL"),
		apiKey:  os.Getenv("INVENTORY_API_KEY"),
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// --- Response wrappers ---

type inventoryResp struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// --- Stock levels ---

type StockLevel struct {
	InventoryItemID string  `json:"inventory_item_id"`
	SKU             string  `json:"sku"`
	Name            string  `json:"name"`
	Unit            string  `json:"unit"`
	QuantityInStock float64 `json:"quantity_in_stock"`
	MinQuantity     float64 `json:"min_quantity"`
	IsLow           bool    `json:"is_low"`
	IsOut           bool    `json:"is_out"`
}

func (c *InventoryClient) GetStockLevels() ([]StockLevel, error) {
	body, err := c.get("/api/v1/pos/stock/levels")
	if err != nil {
		return nil, err
	}
	var levels []StockLevel
	return levels, json.Unmarshal(body, &levels)
}

// --- Product availability ---

type ProductAvailability struct {
	PosProductID string               `json:"pos_product_id"`
	Name         string               `json:"name"`
	IsAvailable  bool                 `json:"is_available"`
	Details      []AvailabilityDetail `json:"details"`
}

type AvailabilityDetail struct {
	InventoryItemID string  `json:"inventory_item_id"`
	SKU             string  `json:"sku"`
	Name            string  `json:"name"`
	Required        float64 `json:"required"`
	Available       float64 `json:"available"`
	IsSufficient    bool    `json:"is_sufficient"`
}

func (c *InventoryClient) CheckAvailability(posProductID string, quantity int) (*ProductAvailability, error) {
	path := fmt.Sprintf("/api/v1/pos/products/%s/availability?quantity=%d", posProductID, quantity)
	body, err := c.get(path)
	if err != nil {
		return nil, err
	}
	var avail ProductAvailability
	return &avail, json.Unmarshal(body, &avail)
}

// --- Barcode lookup ---

type BarcodeProduct struct {
	ID           string          `json:"id"`
	PosProductID string          `json:"pos_product_id"`
	Name         string          `json:"name"`
	SKU          string          `json:"sku"`
	Barcode      string          `json:"barcode"`
	IsActive     bool            `json:"is_active"`
	BOM          []BarcodeItem   `json:"bom"`
	CreatedAt    string          `json:"created_at"`
	UpdatedAt    string          `json:"updated_at"`
}

type BarcodeItem struct {
	InventoryItemID string            `json:"inventory_item_id"`
	QuantityRequired float64          `json:"quantity_required"`
	InventoryItem   BarcodeItemDetail `json:"inventory_item"`
}

type BarcodeItemDetail struct {
	SKU             string  `json:"sku"`
	Name            string  `json:"name"`
	Unit            string  `json:"unit"`
	QuantityInStock float64 `json:"quantity_in_stock"`
}

func (c *InventoryClient) GetProductByBarcode(barcode string) (*BarcodeProduct, error) {
	body, err := c.get("/api/v1/pos/products/barcode/" + barcode)
	if err != nil {
		return nil, err
	}
	var product BarcodeProduct
	return &product, json.Unmarshal(body, &product)
}

// --- Stock deduction ---

type DeductRequest struct {
	PosOrderID string        `json:"pos_order_id"`
	Items      []DeductItem  `json:"items"`
}

type DeductItem struct {
	PosProductID string `json:"pos_product_id"`
	Quantity     int    `json:"quantity"`
}

type DeductResponse struct {
	PosOrderID string          `json:"pos_order_id"`
	Status     string          `json:"status"` // "processed" | "already_processed"
	Deductions []DeductDetail  `json:"deductions,omitempty"`
}

type DeductDetail struct {
	InventoryItemID   string  `json:"inventory_item_id"`
	SKU               string  `json:"sku"`
	Name              string  `json:"name"`
	QuantityDeducted  float64 `json:"quantity_deducted"`
	QuantityRemaining float64 `json:"quantity_remaining"`
}

func (c *InventoryClient) DeductStock(req DeductRequest) (*DeductResponse, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/v1/pos/stock/deduct", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("X-API-Key", c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("inventory unreachable: %w", err)
	}
	defer resp.Body.Close()

	var result inventoryResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.Status != "success" {
		return nil, fmt.Errorf("%s", result.Message)
	}

	var deduct DeductResponse
	return &deduct, json.Unmarshal(result.Data, &deduct)
}

// --- helpers ---

func (c *InventoryClient) get(path string) (json.RawMessage, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("inventory unreachable: %w", err)
	}
	defer resp.Body.Close()

	var result inventoryResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.Status != "success" {
		return nil, fmt.Errorf("inventory error: %s", result.Message)
	}
	return result.Data, nil
}
