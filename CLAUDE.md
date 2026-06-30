# POS Backend — Project Rules for Claude

## Project Overview

Go (Fiber v2 + GORM) POS backend that integrates with a separate Inventory system.

| Service | Port | Purpose |
|---|---|---|
| POS Backend | 4000 | This project |
| Inventory System | 8080 | External — must be running for stock sync/deduct |
| PostgreSQL | 5432 | Via Docker (`docker compose up -d postgres`) |

---

## How to Run

```powershell
docker compose up -d postgres   # start DB
go run .                        # start POS backend
```

Default admin: `admin@pos.local` / `admin123`

---

## Architecture Rules

### Layer boundaries — never cross them
```
handler → service → repository → model
```
- Handlers only parse HTTP, call one service method, return response
- Services hold all business logic — no GORM calls directly
- Repositories only do GORM queries — no business logic
- Models are pure structs — no methods that hit the DB

### File naming convention
```
internal/models/       *_model.go
internal/repository/   *_repository.go
internal/service/      *_service.go
internal/handler/      *_handler.go
```

### Adding a new feature checklist
1. Model (`internal/models/`)
2. Add model to `database.Migrate()` in `internal/database/migrate.go`
3. Repository (`internal/repository/`)
4. Service (`internal/service/`)
5. Handler (`internal/handler/`)
6. Route in `internal/router/router.go`
7. Wire repo/service/handler in `server.go`
8. Update `FRONTEND_API_GUIDE.md`

---

## API Rules

### Auth
- Public routes: `POST /auth/login`, `POST /webhook/inventory`, `GET /health`
- All `/api/v1/*` routes require `Authorization: Bearer <JWT>`
- Admin-only routes use `middleware.AdminOnly()`

### Response envelope — always use `pkg/response`
```go
response.Success(c, data)       // 200
response.Created(c, data)       // 201
response.Error(c, status, msg)  // 4xx/5xx
```
Never call `c.JSON()` directly in handlers.

### Pagination — use `response.PaginatedData`
```go
response.Success(c, response.PaginatedData{
    Items: items, Total: total, Page: page, Limit: limit,
})
```

---

## Payment Methods

Orders support three payment methods:

| Value | Required extra fields |
|---|---|
| `CASH` | none (default) |
| `BANK_QRCODE` | none — frontend fetches `GET /api/v1/config/bank-qr/qrcode` separately |
| `PAY_LATER` | `customer_name`, `customer_phone` (required); `payment_due_days` (default 7) |

### Pay Later alert loop
- Runs every **1 hour** via `PayLaterAlertService.StartAlertLoop`
- Fires for orders where: `payment_method=PAY_LATER AND is_paid=false AND payment_due_date < now() AND status=COMPLETED`
- Logs to stdout; optionally POSTs to `ALERT_WEBHOOK_URL` env var
- Stops alerting once admin calls `POST /api/v1/orders/:id/pay`

### PromptPay QR code
- `GET /api/v1/config/bank-qr/qrcode` → `image/png`
- `GET /api/v1/config/bank-qr/qrcode?amount=185.00` → dynamic QR with amount
- Requires `promptpay_id` set in bank QR config (phone 10-digit or national ID 13-digit)
- Built by `pkg/promptpay` — EMVCo standard, CRC-16/CCITT-FALSE

---

## Inventory Integration Rules

- POS calls Inventory with `X-API-Key` header (never JWT)
- Stock deduction is **idempotent** — always pass unique `pos_order_id`
- If inventory is unreachable, order status → `FAILED` (never silently succeed)
- Stock cache syncs: startup, every 5 min, and on webhook push
- Webhook endpoint (`POST /webhook/inventory`) is public but HMAC-verified

### Three Inventory calls in order flow
```
1. CheckAvailability  →  GET /api/v1/pos/products/:id/availability
2. Create order in DB (PENDING)
3. DeductStock        →  POST /api/v1/pos/stock/deduct
4. Update order → COMPLETED (or FAILED)
```

---

## Environment Variables

```env
APP_PORT=4000
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=pos_db
DB_SSLMODE=disable
JWT_SECRET=your-super-secret-jwt-key-change-in-production
INVENTORY_BASE_URL=http://localhost:8080
INVENTORY_API_KEY=dev-pos-api-key-2026
INVENTORY_WEBHOOK_SECRET=your-webhook-hmac-secret
ALERT_WEBHOOK_URL=           # optional: POST overdue PAY_LATER alerts here
SEED=true                    # set false in production
```

---

## Testing Rules

### Before committing — always run
```powershell
go build ./...    # must compile clean
go test ./...     # must pass
```

### Manual integration test order
1. `POST /auth/login` → get token
2. `POST /api/v1/stock/sync` → verify sync count
3. `GET /api/v1/stock` → verify cache populated
4. `GET /api/v1/stock/availability/:id?quantity=1` → verify available
5. `POST /api/v1/orders` with `CASH` → verify COMPLETED
6. `POST /api/v1/orders` with `BANK_QRCODE` → verify COMPLETED
7. `POST /api/v1/orders` with `PAY_LATER` → verify customer fields + due date saved
8. `POST /api/v1/orders/:id/pay` → verify is_paid=true
9. `GET /api/v1/config/bank-qr/qrcode` → verify image/png response
10. `GET /api/v1/orders?overdue=true` → verify filter works

### Validation that must always reject
| Case | Expected |
|---|---|
| `PAY_LATER` without `customer_name` | 400 |
| `PAY_LATER` without `customer_phone` | 400 |
| Invalid `payment_method` value | 400 |
| `POST /orders/:id/pay` on non-PAY_LATER order | 400 |
| `POST /orders/:id/pay` on already-paid order | 400 |
| `POST /orders/:id/cancel` on COMPLETED order | 400 |
| QR code endpoint without `promptpay_id` set | 422 |

---

## Do Not

- **Do not edit files outside `c:\projects\POS-backend\`** — you are the POS-backend developer only; the Inventory system (`investory-management-backend`) is a separate project owned by a separate team
- Do not add business logic in handlers
- Do not call GORM directly in services
- Do not skip `middleware.AdminOnly()` on write config/user endpoints
- Do not change the `pos_order_id` format — Inventory uses it for idempotency
- Do not remove the `already_processed` check in `DeductStock`
- Do not add `fmt.Println` — use `log.Printf`
- Do not commit `.env` — it is in `.gitignore`

---

## Key Files

| File | Purpose |
|---|---|
| `server.go` | Entry point — wire everything |
| `internal/router/router.go` | All routes |
| `internal/service/inventory_client.go` | All calls to Inventory system |
| `internal/service/paylater_alert_service.go` | Overdue alert loop |
| `pkg/promptpay/promptpay.go` | EMVCo PromptPay QR payload builder |
| `FRONTEND_API_GUIDE.md` | API reference for frontend team |
| `INVENTORY_INTEGRATION.md` | How POS integrates with Inventory |
| `INTEGRATION.md` | Inventory system's full API reference |
| `INVESTORY_SYSTEM_STRUCTURE.md` | Inventory system code structure |
