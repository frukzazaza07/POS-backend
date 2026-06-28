# Inventory Management — Integration Guide

ระบบนี้ทำหน้าที่เป็น **Inventory Backend** ที่รับข้อมูลจาก POS และระบบอื่น ๆ  
Base URL: `http://<host>:3000`

---

## สารบัญ

1. [Authentication](#1-authentication)
2. [POS System Integration](#2-pos-system-integration)
3. [Frontend / Management Integration](#3-frontend--management-integration)
4. [Webhook (รับ event จาก Inventory)](#4-webhook-รับ-event-จาก-inventory)
5. [Error Response Format](#5-error-response-format)
6. [Quick Reference — All Endpoints](#6-quick-reference--all-endpoints)

---

## 1. Authentication

ระบบมี **2 วิธี auth** แยกตาม use case:

### 1.1 JWT Bearer Token (สำหรับ Management / Frontend)

ใช้กับทุก endpoint ที่ขึ้นต้นด้วย `/api/v1/` (ยกเว้น POS endpoints)

**Login:**
```http
POST /auth/login
Content-Type: application/json

{
  "email": "admin@example.com",
  "password": "admin123"
}
```

**Response:**
```json
{
  "status": "success",
  "data": { "token": "eyJhbGciOiJIUzI1NiIs..." }
}
```

**ใช้งาน:** ใส่ header ทุก request ที่ต้องการ auth:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

Token หมดอายุใน **24 ชั่วโมง**

---

### 1.2 API Key (สำหรับ POS System — server-to-server)

ใช้กับ endpoint ที่ขึ้นต้นด้วย `/api/v1/pos/` เท่านั้น  
ตั้งค่า key ใน `.env` ด้วย `POS_API_KEY=<your-key>`

**ใช้งาน:**
```
X-API-Key: dev-pos-api-key-2026
```

หรือผ่าน query string (สำหรับ testing):
```
GET /api/v1/pos/stock/levels?api_key=dev-pos-api-key-2026
```

---

## 2. POS System Integration

### 2.1 Flow ภาพรวม

```
POS                          Inventory
 │                               │
 │── เปิดหน้าขาย ──────────────►│
 │◄── ดูสต็อกคงเหลือ ───────────│  GET /pos/stock/levels
 │                               │
 │── ลูกค้าสั่งออเดอร์ ─────────►│
 │◄── เช็กว่าขายได้ไหม ─────────│  GET /pos/products/:id/availability
 │                               │
 │── ขายเสร็จ ─────────────────►│
 │◄── ตัดสต็อก ─────────────────│  POST /pos/stock/deduct
 │                               │
 │◄═══════ Webhook ══════════════│  STOCK_UPDATED / STOCK_LOW / STOCK_OUT
```

---

### 2.2 ดูสต็อกคงเหลือ

ใช้เพื่อโหลดสต็อกทั้งหมดเมื่อ POS เปิดขึ้น หรือ sync เป็นระยะ

```http
GET /api/v1/pos/stock/levels
X-API-Key: <api-key>
```

**Response:**
```json
{
  "status": "success",
  "data": [
    {
      "inventory_item_id": "inv-coffee-beans-001",
      "sku": "RAW-COFFEE-BEANS",
      "name": "Coffee Beans (Arabica)",
      "unit": "g",
      "quantity_in_stock": 5000,
      "min_quantity": 500,
      "is_low": false,
      "is_out": false
    }
  ]
}
```

| Field | คำอธิบาย |
|---|---|
| `is_low` | `true` เมื่อ stock ≤ min_quantity |
| `is_out` | `true` เมื่อ stock = 0 |

---

### 2.3 เช็กว่าสินค้าขายได้ไหม (ก่อนรับออเดอร์)

```http
GET /api/v1/pos/products/{pos_product_id}/availability?quantity=2
X-API-Key: <api-key>
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "pos_product_id": "pos-latte",
    "name": "Cafe Latte",
    "is_available": true,
    "details": [
      {
        "inventory_item_id": "inv-coffee-beans-001",
        "sku": "RAW-COFFEE-BEANS",
        "name": "Coffee Beans",
        "required": 36,
        "available": 5000,
        "is_sufficient": true
      },
      {
        "inventory_item_id": "inv-milk-001",
        "sku": "RAW-MILK-FRESH",
        "name": "Fresh Milk",
        "required": 400,
        "available": 10000,
        "is_sufficient": true
      }
    ]
  }
}
```

> `is_available = false` เมื่อ inventory item ใด item หนึ่งใน BOM ไม่เพียงพอ

---

### 2.4 ตัดสต็อกหลังขาย (Idempotent)

**สำคัญ:** ต้องส่ง `pos_order_id` ที่ unique ต่อออเดอร์เสมอ  
ระบบจะ deduplicate — ถ้าส่งซ้ำ order เดิมจะ return `already_processed` ไม่ตัดสต็อกซ้ำ

```http
POST /api/v1/pos/stock/deduct
X-API-Key: <api-key>
Content-Type: application/json

{
  "pos_order_id": "ORDER-20260628-0001",
  "items": [
    { "pos_product_id": "pos-americano", "quantity": 1 },
    { "pos_product_id": "pos-latte",     "quantity": 2 }
  ]
}
```

**Response (สำเร็จ):**
```json
{
  "status": "success",
  "data": {
    "pos_order_id": "ORDER-20260628-0001",
    "status": "processed",
    "deductions": [
      {
        "inventory_item_id": "inv-coffee-beans-001",
        "sku": "RAW-COFFEE-BEANS",
        "name": "Coffee Beans",
        "quantity_deducted": 54,
        "quantity_remaining": 4946
      }
    ]
  }
}
```

**Response (ส่งซ้ำ):**
```json
{
  "status": "success",
  "data": {
    "pos_order_id": "ORDER-20260628-0001",
    "status": "already_processed"
  }
}
```

**Response (stock ไม่พอ — HTTP 400):**
```json
{
  "status": "error",
  "message": "insufficient stock for RAW-COFFEE-BEANS: have 10.00, need 18.00"
}
```

---

### 2.5 ผูก POS Product กับ Inventory

ก่อนที่ POS จะสามารถตัดสต็อกได้ ต้องมีการผูก `pos_product_id` กับ Inventory ผ่าน Management API:

**ขั้นตอน:**
1. สร้าง Product ใน Inventory โดยใส่ `pos_product_id` ให้ตรงกับ ID ใน POS
2. กำหนด BOM (สูตร) ว่า product นี้ใช้ inventory item อะไรบ้าง เท่าไหร่

```http
POST /api/v1/products
Authorization: Bearer <token>

{
  "pos_product_id": "pos-latte",
  "name": "Cafe Latte",
  "sku": "BEV-LATTE",
  "is_active": true
}
```

```http
PUT /api/v1/products/{id}/bom
Authorization: Bearer <token>

{
  "items": [
    { "inventory_item_id": "inv-coffee-beans-001", "quantity_required": 18 },
    { "inventory_item_id": "inv-milk-001",          "quantity_required": 200 },
    { "inventory_item_id": "inv-sugar-001",          "quantity_required": 5  },
    { "inventory_item_id": "inv-cup-hot-001",        "quantity_required": 1  }
  ]
}
```

> `quantity_required` = จำนวนที่ตัดต่อการขาย **1 หน่วย** ของสินค้านั้น

---

## 3. Frontend / Management Integration

ใช้ JWT token สำหรับทุก request ส่วนนี้

### 3.1 Suppliers

```
GET    /api/v1/suppliers?page=1&limit=20&search=coffee
POST   /api/v1/suppliers
GET    /api/v1/suppliers/{id}
PUT    /api/v1/suppliers/{id}
DELETE /api/v1/suppliers/{id}
```

**Create body:**
```json
{
  "name": "Coffee World Co.",
  "contact_name": "Somchai",
  "phone": "0812345678",
  "email": "order@coffeeworld.th",
  "address": "123 Silom Rd, Bangkok"
}
```

---

### 3.2 Inventory Items

```
GET    /api/v1/inventory/items?page=1&limit=20&search=beans
POST   /api/v1/inventory/items
GET    /api/v1/inventory/items/{id}
PUT    /api/v1/inventory/items/{id}
DELETE /api/v1/inventory/items/{id}
POST   /api/v1/inventory/items/{id}/adjust
GET    /api/v1/inventory/items/{id}/transactions
```

**Create body:**
```json
{
  "sku": "RAW-COFFEE-BEANS",
  "name": "Coffee Beans (Arabica)",
  "description": "Premium Arabica coffee beans",
  "unit": "g",
  "min_quantity": 500,
  "cost_per_unit": 0.5
}
```

**Adjust stock (manual):**
```json
{
  "quantity": 1000,
  "is_add": true,
  "note": "Manual restock from storage"
}
```
> `is_add: false` = ลด stock (เช่น สินค้าเสียหาย, ใช้เป็น sample)

**Transaction history response:**
```json
{
  "data": {
    "items": [
      {
        "transaction_type": "IN",
        "quantity": 2000,
        "quantity_before": 3000,
        "quantity_after": 5000,
        "reference_type": "PURCHASE_ORDER",
        "reference_id": "uuid-of-po",
        "note": "Full shipment",
        "created_at": "2026-06-28T00:02:43Z"
      }
    ],
    "total": 15, "page": 1, "limit": 20
  }
}
```

| `transaction_type` | ความหมาย |
|---|---|
| `IN` | รับของจาก Purchase Order |
| `OUT` | ตัดสต็อกจาก POS |
| `ADJUSTMENT_ADD` | เพิ่มแบบ manual |
| `ADJUSTMENT_REMOVE` | ลดแบบ manual |

---

### 3.3 Products & BOM

```
GET    /api/v1/products?page=1&limit=20&search=latte
POST   /api/v1/products
GET    /api/v1/products/{id}
PUT    /api/v1/products/{id}
DELETE /api/v1/products/{id}
GET    /api/v1/products/{id}/bom
PUT    /api/v1/products/{id}/bom   ← full replace
```

> **PUT /bom** เป็น full replace — ส่ง items ทั้งหมดที่ต้องการ ระบบจะลบของเก่าแล้วใส่ใหม่ทั้งหมด

---

### 3.4 Purchase Orders

```
GET    /api/v1/purchase-orders?status=DRAFT&page=1&limit=20
POST   /api/v1/purchase-orders
GET    /api/v1/purchase-orders/{id}
PUT    /api/v1/purchase-orders/{id}
POST   /api/v1/purchase-orders/{id}/receive
POST   /api/v1/purchase-orders/{id}/cancel
```

**สร้าง PO:**
```json
{
  "supplier_id": "uuid-of-supplier",
  "notes": "Monthly restock",
  "expected_at": "2026-07-05T09:00:00+07:00",
  "items": [
    { "inventory_item_id": "uuid", "quantity_ordered": 5000, "cost_per_unit": 0.45 }
  ]
}
```

**PO Status Flow:**
```
DRAFT → ORDERED → PARTIALLY_RECEIVED → RECEIVED
                                      (ถ้าทุก item ครบ)
      → CANCELLED  (ยกเลิกได้ก่อน RECEIVED)
```

**รับของ (อาจรับทีละส่วน):**
```json
{
  "note": "Received 1st batch",
  "items": [
    { "purchase_order_item_id": "uuid", "quantity_received": 2500 }
  ]
}
```
> รับไม่ครบทุก item → status = `PARTIALLY_RECEIVED`  
> รับครบทุก item → status = `RECEIVED` และ stock เพิ่มทันที

---

### 3.5 Webhooks (จัดการ subscription)

```
GET    /api/v1/webhooks
POST   /api/v1/webhooks
GET    /api/v1/webhooks/{id}
PUT    /api/v1/webhooks/{id}
DELETE /api/v1/webhooks/{id}
POST   /api/v1/webhooks/{id}/test
GET    /api/v1/webhooks/{id}/logs
```

**สร้าง Webhook:**
```json
{
  "name": "POS Stock Listener",
  "url": "https://your-pos-server.com/inventory-webhook",
  "secret": "your-hmac-secret-key",
  "events": ["STOCK_UPDATED", "STOCK_LOW", "STOCK_OUT"],
  "is_active": true
}
```

---

## 4. Webhook (รับ event จาก Inventory)

เมื่อ stock เปลี่ยนแปลง ระบบจะ POST ไปยัง URL ที่ลงทะเบียนไว้

### 4.1 Event Types

| Event | เกิดเมื่อ |
|---|---|
| `STOCK_UPDATED` | stock เปลี่ยนทุกกรณี (เพิ่ม/ลด) |
| `STOCK_LOW` | stock ≤ `min_quantity` หลัง deduct |
| `STOCK_OUT` | stock = 0 หลัง deduct |
| `TEST` | ส่งจาก `/webhooks/{id}/test` |

---

### 4.2 Payload Structure

```json
{
  "event": "STOCK_UPDATED",
  "timestamp": "2026-06-28T10:00:00Z",
  "data": {
    "inventory_item_id": "inv-coffee-beans-001",
    "sku": "RAW-COFFEE-BEANS",
    "name": "Coffee Beans (Arabica)",
    "quantity_in_stock": 4946,
    "unit": "g",
    "affected_pos_products": ["pos-americano", "pos-latte", "pos-cappuccino"]
  }
}
```

> `affected_pos_products` — รายการ `pos_product_id` ที่ใช้ inventory item นี้ใน BOM  
> ใช้เพื่อ **update UI ใน POS** ว่าสินค้าไหนอาจขายได้/ไม่ได้

---

### 4.3 Verify Signature (สำคัญ)

ระบบส่ง HMAC-SHA256 signature เพื่อให้ผู้รับยืนยันว่า request มาจาก Inventory จริง

**Headers ที่ส่งไป:**
```
X-Inventory-Event: STOCK_UPDATED
X-Inventory-Signature: sha256=abc123def456...
Content-Type: application/json
```

**วิธี verify (ตัวอย่าง Node.js):**
```js
const crypto = require('crypto');

function verifyInventorySignature(payload, signature, secret) {
  const expected = 'sha256=' + crypto
    .createHmac('sha256', secret)
    .update(payload)
    .digest('hex');
  return crypto.timingSafeEqual(
    Buffer.from(expected),
    Buffer.from(signature)
  );
}

// ใน webhook handler:
app.post('/inventory-webhook', (req, res) => {
  const rawBody  = req.rawBody; // ต้องเป็น raw buffer
  const sig      = req.headers['x-inventory-signature'];
  const event    = req.headers['x-inventory-event'];

  if (!verifyInventorySignature(rawBody, sig, process.env.WEBHOOK_SECRET)) {
    return res.status(401).send('Invalid signature');
  }

  const payload = JSON.parse(rawBody);
  console.log(`Event: ${event}`, payload.data);

  // อัพเดต POS stock cache
  updateStockCache(payload.data);
  res.status(200).send('OK');
});
```

**วิธี verify (Python):**
```python
import hmac, hashlib

def verify_signature(payload: bytes, signature: str, secret: str) -> bool:
    expected = 'sha256=' + hmac.new(
        secret.encode(), payload, hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(expected, signature)
```

---

### 4.4 Retry Policy

- ระบบพยายามส่ง **1 ครั้ง** (no automatic retry) — ดู delivery logs ที่ `GET /api/v1/webhooks/{id}/logs`
- Timeout: **10 วินาที** per request
- ถ้าต้องการ retry ให้ใช้ `POST /api/v1/webhooks/{id}/test` หรือ implement retry ที่ฝั่ง consumer

---

## 5. Error Response Format

ทุก error response มีรูปแบบเดียวกัน:

```json
{
  "status": "error",
  "message": "คำอธิบาย error"
}
```

| HTTP Status | ความหมาย |
|---|---|
| `400` | Request ไม่ถูกต้อง (validation, stock ไม่พอ) |
| `401` | ไม่มี token / API key ไม่ถูกต้อง |
| `404` | ไม่พบ resource |
| `500` | Server error |

---

## 6. Quick Reference — All Endpoints

### Auth
```
POST /auth/register
POST /auth/login
```

### Management API  (ต้องใช้ `Authorization: Bearer <token>`)
```
GET/POST        /api/v1/suppliers
GET/PUT/DELETE  /api/v1/suppliers/{id}

GET/POST        /api/v1/inventory/items
GET/PUT/DELETE  /api/v1/inventory/items/{id}
POST            /api/v1/inventory/items/{id}/adjust
GET             /api/v1/inventory/items/{id}/transactions

GET/POST        /api/v1/products
GET/PUT/DELETE  /api/v1/products/{id}
GET/PUT         /api/v1/products/{id}/bom

GET/POST        /api/v1/purchase-orders
GET/PUT         /api/v1/purchase-orders/{id}
POST            /api/v1/purchase-orders/{id}/receive
POST            /api/v1/purchase-orders/{id}/cancel

GET/POST        /api/v1/webhooks
GET/PUT/DELETE  /api/v1/webhooks/{id}
POST            /api/v1/webhooks/{id}/test
GET             /api/v1/webhooks/{id}/logs
```

### POS API  (ต้องใช้ `X-API-Key: <key>`)
```
POST /api/v1/pos/stock/deduct
GET  /api/v1/pos/stock/levels
GET  /api/v1/pos/products/{pos_product_id}/availability
```

### Docs
```
GET /docs            ← Scalar UI (interactive)
GET /swagger/doc.json ← OpenAPI JSON
```

---

## Environment Variables

```env
APP_PORT=3000
DB_HOST=localhost
DB_PORT=5433
DB_USER=postgres
DB_PASSWORD=<password>
DB_NAME=inventory_db
DB_SSLMODE=disable
JWT_SECRET=<strong-secret>
POS_API_KEY=<api-key-for-pos>
SEED=true   # ตั้งเป็น false ใน production
```
