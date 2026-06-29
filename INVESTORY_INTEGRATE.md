# Inventory System Integration

This document describes how the POS Backend integrates with the Inventory system.

---

## Overview

The POS Backend communicates with the Inventory system in two directions:

| Direction | Mechanism | Purpose |
|---|---|---|
| POS → Inventory | REST API (`X-API-Key`) | Check stock, deduct stock after order |
| Inventory → POS | Webhook (`HMAC-SHA256`) | Real-time stock level updates |

---

## Environment Variables

```env
INVENTORY_BASE_URL=http://localhost:8080
INVENTORY_API_KEY=dev-pos-api-key-2026
INVENTORY_WEBHOOK_SECRET=your-webhook-hmac-secret
```

| Variable | Description |
|---|---|
| `INVENTORY_BASE_URL` | Base URL of the Inventory system |
| `INVENTORY_API_KEY` | API key sent in `X-API-Key` header on every outbound request |
| `INVENTORY_WEBHOOK_SECRET` | Shared secret used to verify inbound webhook HMAC signatures |

---

## Outbound: POS → Inventory API

All requests set `X-API-Key: <INVENTORY_API_KEY>` in the header.

### 1. Get Stock Levels

Fetches all inventory item stock levels. Used on startup and periodic sync.

```
GET {INVENTORY_BASE_URL}/api/v1/pos/stock/levels
```

**Response `data` array:**
```json
[
  {
    "inventory_item_id": "uuid",
    "sku": "ITEM-001",
    "name": "Coffee Beans",
    "unit": "kg",
    "quantity_in_stock": 12.5,
    "min_quantity": 5.0,
    "is_low": false,
    "is_out": false
  }
]
```

### 2. Check Product Availability

Checks whether a POS product can be made given current stock. Called in real time before order creation.

```
GET {INVENTORY_BASE_URL}/api/v1/pos/products/{pos_product_id}/availability?quantity={qty}
```

**Response `data`:**
```json
{
  "pos_product_id": "uuid",
  "name": "Latte",
  "is_available": true,
  "details": [
    {
      "inventory_item_id": "uuid",
      "sku": "MILK-001",
      "name": "Fresh Milk",
      "required": 0.2,
      "available": 5.0,
      "is_sufficient": true
    }
  ]
}
```

### 3. Lookup Product by Barcode

Called when a cashier scans a product barcode. Returns the product details including its BOM with current inventory quantities, so the POS can verify availability before adding to the order.

```
GET {INVENTORY_BASE_URL}/api/v1/pos/products/barcode/{barcode}
X-API-Key: <api-key>
```

**Response `data`:**
```json
{
  "id": "uuid",
  "pos_product_id": "pos-latte",
  "name": "Cafe Latte",
  "sku": "BEV-LATTE",
  "barcode": "1234567890128",
  "is_active": true,
  "bom": [
    {
      "inventory_item_id": "uuid",
      "quantity_required": 18,
      "inventory_item": {
        "sku": "RAW-COFFEE-BEANS",
        "name": "Coffee Beans",
        "unit": "g",
        "quantity_in_stock": 4946
      }
    }
  ],
  "created_at": "2026-06-28T00:00:00Z",
  "updated_at": "2026-06-29T00:00:00Z"
}
```

Returns `404` if no product with that barcode exists.

---

### 4. Deduct Stock

Called after an order is confirmed. Idempotent — the Inventory system uses `pos_order_id` to prevent double-deduction.

```
POST {INVENTORY_BASE_URL}/api/v1/pos/stock/deduct
Content-Type: application/json
```

**Request body:**
```json
{
  "pos_order_id": "uuid",
  "items": [
    { "pos_product_id": "uuid", "quantity": 2 }
  ]
}
```

**Response `data`:**
```json
{
  "pos_order_id": "uuid",
  "status": "processed",
  "deductions": [
    {
      "inventory_item_id": "uuid",
      "sku": "MILK-001",
      "name": "Fresh Milk",
      "quantity_deducted": 0.4,
      "quantity_remaining": 4.6
    }
  ]
}
```

`status` is `"processed"` on first call or `"already_processed"` if the order was already deducted.

---

## Inbound: Inventory → POS Webhook

The Inventory system pushes stock change events to:

```
POST {POS_BASE_URL}/webhook/inventory
```

This endpoint is **public** (no JWT), but verifies an HMAC-SHA256 signature.

### Request Headers

| Header | Value |
|---|---|
| `X-Inventory-Event` | Event type string (e.g. `STOCK_UPDATED`) |
| `X-Inventory-Signature` | `sha256=<hex-hmac>` of the raw request body |

### Signature Verification

The POS backend computes:

```
HMAC-SHA256(rawBody, INVENTORY_WEBHOOK_SECRET)
```

and compares it to the `X-Inventory-Signature` header value (prefix `sha256=`). Requests with an invalid or missing signature (when `INVENTORY_WEBHOOK_SECRET` is set) are rejected with `401`.

### Supported Events

| Event | Action |
|---|---|
| `STOCK_UPDATED` | Updates stock cache for the item |
| `STOCK_LOW` | Updates cache, sets `is_low = true` |
| `STOCK_OUT` | Updates cache, sets `is_out = true` |
| `TEST` | Acknowledged and logged, no action |
| *(any other)* | Logged and ignored |

### Webhook Payload

```json
{
  "event": "STOCK_UPDATED",
  "timestamp": "2026-06-28T10:00:00Z",
  "data": {
    "inventory_item_id": "uuid",
    "sku": "MILK-001",
    "name": "Fresh Milk",
    "quantity_in_stock": 4.6,
    "unit": "L",
    "affected_pos_products": ["uuid-latte", "uuid-cappuccino"]
  }
}
```

---

## Stock Cache

The POS Backend maintains a local `stock_caches` table to avoid hitting the Inventory system on every request.

| Sync Trigger | Frequency |
|---|---|
| App startup | Once |
| Periodic background sync | Every 5 minutes |
| Webhook push from Inventory | Real-time, per event |
| Manual admin trigger (`POST /api/v1/stock/sync`) | On demand |

### POS Stock & Product API Endpoints

All require `Authorization: Bearer <JWT>`.

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/stock` | Read local stock cache |
| `POST` | `/api/v1/stock/sync` | Force full re-sync from Inventory *(admin only)* |
| `GET` | `/api/v1/stock/availability/:pos_product_id?quantity=N` | Real-time availability check via Inventory |
| `GET` | `/api/v1/products/barcode/:barcode` | Look up product by barcode (proxies to Inventory) |

---

## Order Flow

```
Cashier creates order
        │
        ▼
CheckAvailability (Inventory API) ──► insufficient → reject
        │ sufficient
        ▼
  Save order (status: pending)
        │
        ▼
  DeductStock (Inventory API) ──────► error → order status: failed
        │ success
        ▼
  Order status: completed
```

---

## Response Envelope

All Inventory API responses follow this envelope:

```json
{
  "status": "success" | "error",
  "message": "...",
  "data": { ... }
}
```

The POS client returns an error for any response where `status != "success"`.
