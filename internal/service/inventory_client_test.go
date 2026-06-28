package service_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"pos-backend/internal/service"
)

func mockServer(t *testing.T, apiKey string, mux *http.ServeMux) *service.InventoryClient {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != apiKey {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "unauthorized"})
			return
		}
		mux.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)

	os.Setenv("INVENTORY_BASE_URL", srv.URL)
	os.Setenv("INVENTORY_API_KEY", apiKey)
	return service.NewInventoryClient()
}

func jsonResp(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func TestGetStockLevels(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/pos/stock/levels", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, map[string]interface{}{
			"status": "success",
			"data": []map[string]interface{}{
				{
					"inventory_item_id": "inv-coffee-001",
					"sku":               "RAW-COFFEE",
					"name":              "Coffee Beans",
					"unit":              "g",
					"quantity_in_stock": 5000.0,
					"min_quantity":      500.0,
					"is_low":            false,
					"is_out":            false,
				},
			},
		})
	})

	client := mockServer(t, "test-api-key", mux)
	levels, err := client.GetStockLevels()
	if err != nil {
		t.Fatalf("GetStockLevels: %v", err)
	}
	if len(levels) != 1 {
		t.Fatalf("want 1 level, got %d", len(levels))
	}
	if levels[0].InventoryItemID != "inv-coffee-001" {
		t.Errorf("want inv-coffee-001, got %s", levels[0].InventoryItemID)
	}
	if levels[0].QuantityInStock != 5000 {
		t.Errorf("want qty=5000, got %.2f", levels[0].QuantityInStock)
	}
}

func TestCheckAvailability(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/pos/products/pos-latte/availability", func(w http.ResponseWriter, r *http.Request) {
		qty := r.URL.Query().Get("quantity")
		if qty != "2" {
			http.Error(w, "wrong quantity", http.StatusBadRequest)
			return
		}
		jsonResp(w, map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"pos_product_id": "pos-latte",
				"name":           "Cafe Latte",
				"is_available":   true,
				"details":        []interface{}{},
			},
		})
	})

	client := mockServer(t, "test-api-key", mux)
	avail, err := client.CheckAvailability("pos-latte", 2)
	if err != nil {
		t.Fatalf("CheckAvailability: %v", err)
	}
	if !avail.IsAvailable {
		t.Error("want is_available=true")
	}
	if avail.PosProductID != "pos-latte" {
		t.Errorf("want pos-latte, got %s", avail.PosProductID)
	}
}

func TestDeductStock_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/pos/stock/deduct", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req service.DeductRequest
		json.NewDecoder(r.Body).Decode(&req)
		jsonResp(w, map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"pos_order_id": req.PosOrderID,
				"status":       "processed",
				"deductions":   []interface{}{},
			},
		})
	})

	client := mockServer(t, "test-api-key", mux)
	resp, err := client.DeductStock(service.DeductRequest{
		PosOrderID: "ORDER-20260628-ABCD1234",
		Items:      []service.DeductItem{{PosProductID: "pos-latte", Quantity: 2}},
	})
	if err != nil {
		t.Fatalf("DeductStock: %v", err)
	}
	if resp.Status != "processed" {
		t.Errorf("want status=processed, got %s", resp.Status)
	}
	if resp.PosOrderID != "ORDER-20260628-ABCD1234" {
		t.Errorf("want ORDER-20260628-ABCD1234, got %s", resp.PosOrderID)
	}
}

func TestDeductStock_AlreadyProcessed(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/pos/stock/deduct", func(w http.ResponseWriter, r *http.Request) {
		jsonResp(w, map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"pos_order_id": "ORDER-DUPE",
				"status":       "already_processed",
			},
		})
	})

	client := mockServer(t, "test-api-key", mux)
	resp, err := client.DeductStock(service.DeductRequest{
		PosOrderID: "ORDER-DUPE",
		Items:      []service.DeductItem{{PosProductID: "pos-latte", Quantity: 1}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "already_processed" {
		t.Errorf("want already_processed, got %s", resp.Status)
	}
}

func TestDeductStock_InsufficientStock(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/pos/stock/deduct", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		jsonResp(w, map[string]string{
			"status":  "error",
			"message": "insufficient stock for RAW-COFFEE: have 10.00, need 18.00",
		})
	})

	client := mockServer(t, "test-api-key", mux)
	_, err := client.DeductStock(service.DeductRequest{
		PosOrderID: "ORDER-LOW",
		Items:      []service.DeductItem{{PosProductID: "pos-latte", Quantity: 10}},
	})
	if err == nil {
		t.Fatal("expected error for insufficient stock")
	}
}

func TestGetStockLevels_WrongAPIKey(t *testing.T) {
	mux := http.NewServeMux()
	client := mockServer(t, "correct-key", mux)

	os.Setenv("INVENTORY_API_KEY", "wrong-key")
	client = service.NewInventoryClient()

	_, err := client.GetStockLevels()
	if err == nil {
		t.Fatal("expected error for wrong API key")
	}
}
