# Inventory System — Code Structure

Tech stack: **Go · Fiber v2 · GORM · PostgreSQL**  
Running at: `http://localhost:8080` (default)

---

## Directory Tree

```
investory-management-backend/
├── docs/                              # Auto-generated Swagger/OpenAPI docs
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
│
├── internal/
│   ├── database/
│   │   ├── migrate.go                 # GORM AutoMigrate for all models
│   │   ├── postgres.go                # PostgreSQL connection (DSN from env)
│   │   └── seed.go                    # Initial data seeder (admin user, sample items)
│   │
│   ├── handler/                       # HTTP handlers (thin — delegate to service)
│   │   ├── auth_handler.go            # POST /auth/register, POST /auth/login
│   │   ├── inventory_handler.go       # CRUD for inventory items + stock adjustment
│   │   ├── pos_handler.go             # POS-facing endpoints (API-key protected)
│   │   ├── product_handler.go         # Product CRUD + BOM management
│   │   ├── purchase_order_handler.go  # PO create / receive / cancel
│   │   ├── supplier_handler.go        # Supplier CRUD
│   │   └── webhook_handler.go         # Webhook subscription CRUD + test + delivery logs
│   │
│   ├── middleware/
│   │   ├── api_key_middleware.go      # Validates X-API-Key header (POS routes)
│   │   └── jwt_middleware.go          # Validates Bearer JWT (management routes)
│   │
│   ├── models/
│   │   ├── base_model.go              # UUID primary key, CreatedAt, UpdatedAt, soft-delete
│   │   ├── inventory_item_model.go    # InventoryItem — warehouse stock record
│   │   ├── pos_sale_log_model.go      # PosSaleLog — idempotency record per pos_order_id
│   │   ├── product_model.go           # Product + BOMItem association (many-to-many)
│   │   ├── purchase_order_model.go    # PurchaseOrder + PurchaseOrderItem
│   │   ├── stock_transaction_model.go # StockTransaction — immutable audit ledger
│   │   ├── supplier_model.go          # Supplier (contact info)
│   │   ├── user_model.go              # User (email, bcrypt password, role)
│   │   └── webhook_model.go           # WebhookSubscription + WebhookDelivery log
│   │
│   ├── repository/                    # GORM queries, no business logic
│   │   ├── inventory_repository.go
│   │   ├── pos_sale_log_repository.go
│   │   ├── product_repository.go
│   │   ├── purchase_order_repository.go
│   │   ├── stock_transaction_repository.go
│   │   ├── supplier_repository.go
│   │   ├── user_repository.go
│   │   └── webhook_repository.go
│   │
│   ├── router/
│   │   └── router.go                  # Route groups: /auth, /api/v1 (JWT), /api/v1/pos (API-key); barcode lookup in POS group
│   │
│   └── service/                       # Business logic layer
│       ├── auth_service.go            # bcrypt hash, JWT sign/verify
│       ├── inventory_service.go       # Stock level queries, is_low / is_out logic
│       ├── pos_service.go             # Idempotent sale processing, BOM deduction
│       ├── product_service.go         # Product + BOM CRUD, BOM validation
│       ├── purchase_order_service.go  # PO status machine, partial receive, stock IN
│       ├── stock_service.go           # StockTransaction recording (IN / OUT / ADJUSTMENT)
│       ├── supplier_service.go        # Supplier CRUD
│       └── webhook_service.go         # Fan-out delivery, HMAC signing, delivery log
│
├── pkg/
│   └── response/
│       └── response.go                # { status, message, data } envelope helpers
│
├── scripts/
│   └── test_endpoints.ps1             # PowerShell smoke-test script
│
├── .env                               # Local config (not committed)
├── .env.example                       # Config template
├── docker-compose.yml                 # App + PostgreSQL
├── Dockerfile                         # Multi-stage build (builder → alpine)
├── go.mod / go.sum
├── INTEGRATION.md                     # Integration guide for consumers (POS, frontend)
└── server.go                          # Entry point: load env, connect DB, start Fiber
```

---

## Models

### InventoryItem
Physical warehouse stock. Each item has a SKU, unit of measure, and quantity.

| Field | Type | Notes |
|---|---|---|
| `id` | UUID | Primary key |
| `sku` | string | Unique, e.g. `RAW-COFFEE-BEANS` |
| `name` | string | Display name |
| `unit` | string | `g`, `ml`, `pcs`, etc. |
| `quantity_in_stock` | float64 | Current stock level |
| `min_quantity` | float64 | Threshold for `is_low` alert |
| `cost_per_unit` | float64 | For purchase order valuation |

Computed flags (not stored):
- `is_low` = `quantity_in_stock <= min_quantity`
- `is_out` = `quantity_in_stock == 0`

---

### Product + BOMItem
A Product is a menu item sold via POS. Its Bill of Materials (BOM) defines which inventory items are consumed per unit sold.

```
Product "Cafe Latte"
  └── BOMItem: Coffee Beans   18 g
  └── BOMItem: Fresh Milk    200 ml
  └── BOMItem: Sugar           5 g
  └── BOMItem: Hot Cup         1 pcs
```

| Field (Product) | Notes |
|---|---|
| `pos_product_id` | Must match POS system's product ID for stock deduction |
| `sku` | Internal SKU |
| `is_active` | Inactive products are excluded from availability checks |

---

### StockTransaction
Immutable audit ledger. Every stock change writes one row.

| `transaction_type` | When created |
|---|---|
| `IN` | Purchase order received |
| `OUT` | POS sale deduction |
| `ADJUSTMENT_ADD` | Manual addition |
| `ADJUSTMENT_REMOVE` | Manual reduction (damage, loss) |

---

### PosSaleLog
Idempotency guard for `POST /api/v1/pos/stock/deduct`. One row per `pos_order_id`. If a row already exists, the endpoint returns `already_processed` without re-deducting.

---

### PurchaseOrder + PurchaseOrderItem
Tracks ordered goods from a supplier.

**Status machine:**
```
DRAFT → ORDERED → PARTIALLY_RECEIVED → RECEIVED
      ↘ CANCELLED  (before RECEIVED)
```
Stock (`IN` transaction) is added when `POST /purchase-orders/:id/receive` is called. Partial receives are supported — status becomes `PARTIALLY_RECEIVED` until all items are fully received.

---

### WebhookSubscription + WebhookDelivery
- `WebhookSubscription` — URL, HMAC secret, subscribed event list, active flag
- `WebhookDelivery` — per-delivery log (HTTP status, response body, duration)

---

## Auth Layers

| Route group | Auth method | Header |
|---|---|---|
| `/api/v1/pos/*` | API Key | `X-API-Key: <key>` |
| `/api/v1/*` (management) | JWT Bearer | `Authorization: Bearer <token>` |
| `/auth/*` | None (public) | — |

---

## Webhook Fan-out Flow

```
Stock deducted (POS sale)
        │
        ▼
  webhook_service.Dispatch(event, data)
        │
        ├── Find all active subscriptions for this event
        │
        └── For each subscription:
              POST payload to URL
              Sign with HMAC-SHA256
              Log result in WebhookDelivery
              Set headers: X-Inventory-Event, X-Inventory-Signature
```

---

## Environment Variables

```env
APP_PORT=8080
APP_ENV=development

DB_HOST=localhost
DB_PORT=5433
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=inventory_db
DB_SSLMODE=disable

JWT_SECRET=your-jwt-secret
POS_API_KEY=dev-pos-api-key-2026

SEED=true
```

---

## API Docs

Interactive docs are served at:
```
GET http://localhost:8080/docs          ← Scalar UI
GET http://localhost:8080/swagger/doc.json  ← OpenAPI JSON
```

Generate/refresh after code changes:
```bash
swag init
```

---

## How POS Backend Connects to This System

The POS Backend ([pos-backend](../)) calls four endpoints on this system:

| POS action | Inventory endpoint |
|---|---|
| Startup stock sync | `GET /api/v1/pos/stock/levels` |
| Barcode scan lookup | `GET /api/v1/pos/products/barcode/{barcode}` |
| Availability check before order | `GET /api/v1/pos/products/:id/availability` |
| Stock deduction after order | `POST /api/v1/pos/stock/deduct` |

The Inventory system calls back to POS via webhook when stock changes.

See [INVENTORY_INTEGRATION.md](./INVENTORY_INTEGRATION.md) for the POS-side implementation details.  
See [INTEGRATION.md](./INTEGRATION.md) for the Inventory system's full API reference.
