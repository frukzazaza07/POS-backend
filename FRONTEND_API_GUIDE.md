# POS Backend — Frontend API Guide

Base URL: `http://localhost:4000`

All protected endpoints require:
```
Authorization: Bearer <token>
Content-Type: application/json
```

---

## Response Envelope

Every response uses this wrapper:

```json
{
  "status": "success" | "error",
  "message": "...",
  "data": { ... }
}
```

---

## Table of Contents

1. [TypeScript Types](#1-typescript-types)
2. [Axios Setup](#2-axios-setup)
3. [Authentication](#3-authentication)
4. [Products](#4-products)
5. [Orders & Payment Methods](#5-orders--payment-methods)
6. [Bank QR Config](#6-bank-qr-config)
7. [Pay Later Management](#7-pay-later-management)
8. [Stock](#8-stock)
9. [Role-Based Access](#9-role-based-access)
10. [Error Handling](#10-error-handling)
11. [Quick Reference](#11-quick-reference)

---

## 1. TypeScript Types

```ts
// src/types/api.ts

export type UserRole = 'admin' | 'cashier';
export type OrderStatus = 'PENDING' | 'COMPLETED' | 'CANCELLED' | 'FAILED';
export type PaymentMethod = 'CASH' | 'BANK_QRCODE' | 'PAY_LATER';

export interface User {
  id: string;
  name: string;
  email: string;
  role: UserRole;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface POSProduct {
  id: string;
  pos_product_id: string;
  name: string;
  description: string;
  price: number;
  category: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface OrderItem {
  id: string;
  order_id: string;
  pos_product_id: string;
  product_name: string;
  quantity: number;
  unit_price: number;
  subtotal: number;
}

export interface Order {
  id: string;
  pos_order_id: string;
  cashier_id: string;
  cashier?: User;
  status: OrderStatus;
  total_amount: number;
  payment_method: PaymentMethod;
  notes: string;
  fail_reason?: string;
  items?: OrderItem[];
  // Pay Later fields (present only when payment_method === 'PAY_LATER')
  customer_name?: string;
  customer_phone?: string;
  payment_due_date?: string;
  is_paid?: boolean;
  paid_at?: string | null;
  created_at: string;
  updated_at: string;
}

export interface BankQRConfig {
  id: string;
  bank_name: string;
  account_name: string;
  account_number: string;
  qr_image_url: string;
  is_active: boolean;
}

export interface StockItem {
  inventory_item_id: string;
  sku: string;
  name: string;
  unit: string;
  quantity_in_stock: number;
  min_quantity: number;
  is_low: boolean;
  is_out: boolean;
  synced_at: string;
}

export interface ProductAvailability {
  pos_product_id: string;
  name: string;
  is_available: boolean;
  details: Array<{
    inventory_item_id: string;
    sku: string;
    name: string;
    required: number;
    available: number;
    is_sufficient: boolean;
  }>;
}

export interface ApiResponse<T> {
  status: 'success' | 'error';
  message?: string;
  data?: T;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  limit: number;
}
```

---

## 2. Axios Setup

```ts
// src/lib/api.ts
import axios from 'axios';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? 'http://localhost:4000',
  headers: { 'Content-Type': 'application/json' },
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('pos_token');
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('pos_token');
      localStorage.removeItem('pos_user');
      window.location.href = '/login';
    }
    return Promise.reject(err);
  }
);

export default api;
```

`.env.local`:
```
VITE_API_URL=http://localhost:4000
```

---

## 3. Authentication

### Login
```
POST /auth/login
```
```json
{
  "email": "admin@pos.local",
  "password": "admin123"
}
```
**Response `data`:**
```json
{
  "token": "<jwt>",
  "user": { "id": "uuid", "name": "Admin", "email": "admin@pos.local", "role": "admin" }
}
```

```ts
// src/services/auth.ts
export async function login(email: string, password: string) {
  const { data } = await api.post<ApiResponse<{ token: string; user: User }>>('/auth/login', { email, password });
  localStorage.setItem('pos_token', data.data!.token);
  localStorage.setItem('pos_user', JSON.stringify(data.data!.user));
  return data.data!;
}

export function getCurrentUser(): User | null {
  const raw = localStorage.getItem('pos_user');
  return raw ? JSON.parse(raw) : null;
}

export function logout() {
  localStorage.removeItem('pos_token');
  localStorage.removeItem('pos_user');
  window.location.href = '/login';
}
```

### Register User *(admin only)*
```
POST /api/v1/users/register
```
```json
{
  "name": "Jane",
  "email": "jane@pos.local",
  "password": "secret123",
  "role": "cashier"
}
```

---

## 4. Products

| Method | Path | Role |
|---|---|---|
| `GET` | `/api/v1/products?page=1&limit=20` | any |
| `GET` | `/api/v1/products/:id` | any |
| `POST` | `/api/v1/products` | admin |
| `PUT` | `/api/v1/products/:id` | admin |
| `DELETE` | `/api/v1/products/:id` | admin |

**Create/Update body:**
```json
{
  "pos_product_id": "pos-latte",
  "name": "Cafe Latte",
  "description": "Espresso with steamed milk",
  "price": 65.00,
  "category": "drinks",
  "is_active": true
}
```
> `pos_product_id` must match the ID registered in the Inventory system.

```ts
// src/services/products.ts
export const getProducts = (params?: { page?: number; limit?: number }) =>
  api.get<ApiResponse<PaginatedResponse<POSProduct>>>('/api/v1/products', { params })
    .then(r => r.data.data!);

export const createProduct = (body: Partial<POSProduct>) =>
  api.post<ApiResponse<POSProduct>>('/api/v1/products', body).then(r => r.data.data!);

export const updateProduct = (id: string, body: Partial<POSProduct>) =>
  api.put<ApiResponse<POSProduct>>(`/api/v1/products/${id}`, body).then(r => r.data.data!);

export const deleteProduct = (id: string) =>
  api.delete(`/api/v1/products/${id}`);
```

---

## 5. Orders & Payment Methods

### Payment Method Options

| Value | When to use |
|---|---|
| `CASH` | Customer pays with cash at counter (default) |
| `BANK_QRCODE` | Customer scans store QR code and transfers |
| `PAY_LATER` | Customer pays later — saves name, phone, due date |

---

### Create Order
```
POST /api/v1/orders
```

**CASH (default — `payment_method` can be omitted):**
```json
{
  "payment_method": "CASH",
  "notes": "Table 3",
  "items": [
    { "pos_product_id": "pos-latte", "quantity": 2 }
  ]
}
```

**BANK_QRCODE:**
```json
{
  "payment_method": "BANK_QRCODE",
  "notes": "",
  "items": [
    { "pos_product_id": "pos-americano", "quantity": 1 }
  ]
}
```
After creating, call `GET /api/v1/config/bank-qr` to show the store's bank account / QR image to the customer.

**PAY_LATER:**
```json
{
  "payment_method": "PAY_LATER",
  "customer_name": "John Doe",
  "customer_phone": "0812345678",
  "payment_due_days": 7,
  "notes": "regular customer",
  "items": [
    { "pos_product_id": "pos-latte", "quantity": 3 }
  ]
}
```
- `customer_name` — **required**
- `customer_phone` — **required**
- `payment_due_days` — days until due (default **7** if omitted)

**Response `data` (HTTP 201):**
```json
{
  "id": "uuid",
  "pos_order_id": "ORDER-20260628-ABCD1234",
  "status": "COMPLETED",
  "total_amount": 195.00,
  "payment_method": "PAY_LATER",
  "customer_name": "John Doe",
  "customer_phone": "0812345678",
  "payment_due_date": "2026-07-05T10:00:00+07:00",
  "is_paid": false,
  "paid_at": null,
  "items": [ ... ]
}
```

```ts
// src/services/orders.ts
export type CreateOrderPayload = {
  payment_method?: PaymentMethod;
  notes?: string;
  items: Array<{ pos_product_id: string; quantity: number }>;
  // PAY_LATER only
  customer_name?: string;
  customer_phone?: string;
  payment_due_days?: number;
};

export const createOrder = (payload: CreateOrderPayload): Promise<Order> =>
  api.post<ApiResponse<Order>>('/api/v1/orders', payload).then(r => r.data.data!);
```

**Checkout component with payment method selection:**
```tsx
// src/components/Checkout.tsx
import { useState } from 'react';
import { createOrder } from '../services/orders';
import type { PaymentMethod } from '../types/api';

export default function Checkout({ cart }: { cart: Array<{ pos_product_id: string; price: number; quantity: number }> }) {
  const [method, setMethod] = useState<PaymentMethod>('CASH');
  const [customerName, setCustomerName] = useState('');
  const [customerPhone, setCustomerPhone] = useState('');
  const [dueDays, setDueDays] = useState(7);
  const [notes, setNotes] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const total = cart.reduce((s, i) => s + i.price * i.quantity, 0);

  const handleCheckout = async () => {
    setLoading(true); setError('');
    try {
      const order = await createOrder({
        payment_method: method,
        notes,
        items: cart.map(i => ({ pos_product_id: i.pos_product_id, quantity: i.quantity })),
        ...(method === 'PAY_LATER' && { customer_name: customerName, customer_phone: customerPhone, payment_due_days: dueDays }),
      });
      if (method === 'BANK_QRCODE') {
        // Redirect or show bank QR screen
        window.location.href = `/order/${order.id}/bank-qr`;
      } else {
        alert(`Order ${order.pos_order_id} done! ฿${order.total_amount}`);
      }
    } catch (err: any) {
      setError(err.response?.data?.message ?? 'Checkout failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <p>Total: ฿{total.toFixed(2)}</p>

      {/* Payment method selector */}
      <select value={method} onChange={e => setMethod(e.target.value as PaymentMethod)}>
        <option value="CASH">Cash</option>
        <option value="BANK_QRCODE">Bank QR Code</option>
        <option value="PAY_LATER">Pay Later</option>
      </select>

      {/* PAY_LATER extra fields */}
      {method === 'PAY_LATER' && (
        <div>
          <input placeholder="Customer name *" value={customerName} onChange={e => setCustomerName(e.target.value)} />
          <input placeholder="Phone *" value={customerPhone} onChange={e => setCustomerPhone(e.target.value)} />
          <input type="number" placeholder="Due days (default 7)" value={dueDays}
            onChange={e => setDueDays(Number(e.target.value))} min={1} />
        </div>
      )}

      <input placeholder="Notes" value={notes} onChange={e => setNotes(e.target.value)} />
      {error && <p style={{ color: 'red' }}>{error}</p>}
      <button onClick={handleCheckout} disabled={loading || cart.length === 0}>
        {loading ? 'Processing...' : 'Confirm Sale'}
      </button>
    </div>
  );
}
```

---

### List Orders
```
GET /api/v1/orders
```

**Query parameters:**

| Parameter | Type | Description |
|---|---|---|
| `page` | int | Page number (default 1) |
| `limit` | int | Items per page (default 20, max 100) |
| `payment_method` | string | Filter: `CASH`, `BANK_QRCODE`, `PAY_LATER` |
| `status` | string | Filter: `PENDING`, `COMPLETED`, `CANCELLED`, `FAILED` |
| `overdue` | bool | `true` = only overdue unpaid PAY_LATER orders |

> Cashiers see only their own orders. Admins see all.

```ts
export type OrderListParams = {
  page?: number;
  limit?: number;
  payment_method?: PaymentMethod;
  status?: OrderStatus;
  overdue?: boolean;
};

export const getOrders = (params?: OrderListParams): Promise<PaginatedResponse<Order>> =>
  api.get<ApiResponse<PaginatedResponse<Order>>>('/api/v1/orders', { params }).then(r => r.data.data!);
```

---

### Get Order
```
GET /api/v1/orders/:id
```

### Cancel Order
```
POST /api/v1/orders/:id/cancel
```
Only `PENDING` orders. Cashiers can cancel their own; admins can cancel any.

---

## 6. Bank QR Config

Used when `payment_method = "BANK_QRCODE"`. The admin sets the store's bank details once; the cashier screen shows the QR to the customer after order creation.

### Get Current Config
```
GET /api/v1/config/bank-qr
```
Returns `404` if not yet configured.

**Response `data`:**
```json
{
  "id": "uuid",
  "bank_name": "SCB",
  "account_name": "My Coffee Shop Co.",
  "account_number": "111-2-34567-8",
  "qr_image_url": "https://example.com/store-qr.png",
  "is_active": true
}
```

### Generate PromptPay QR Code Image
```
GET /api/v1/config/bank-qr/qrcode
GET /api/v1/config/bank-qr/qrcode?amount=185.00
```
Returns a **PNG image** (`image/png`) directly — use as `<img src="..." />`.

- No `?amount` → static QR (show once, any amount)
- `?amount=185.00` → dynamic QR with the exact amount embedded (Thai bank apps display it pre-filled)

Requires `promptpay_id` to be set in the config (see PUT below).

```tsx
// Show QR for a specific order amount
const token = localStorage.getItem('pos_token');
const qrSrc = `http://localhost:4000/api/v1/config/bank-qr/qrcode?amount=${order.total_amount}`;

// In JSX — add token via query param or proxy; or fetch as blob:
const [qrUrl, setQrUrl] = useState('');
useEffect(() => {
  fetch(qrSrc, { headers: { Authorization: `Bearer ${token}` } })
    .then(r => r.blob())
    .then(b => setQrUrl(URL.createObjectURL(b)));
}, [order.total_amount]);

return <img src={qrUrl} alt="PromptPay QR" style={{ width: 240 }} />;
```

---

### Set / Update Config *(admin only)*
```
PUT /api/v1/config/bank-qr
```
```json
{
  "bank_name": "SCB",
  "account_name": "My Coffee Shop Co.",
  "account_number": "111-2-34567-8",
  "promptpay_id": "0812345678",
  "qr_image_url": "https://example.com/store-qr.png"
}
```
- `bank_name`, `account_name`, `account_number` — **required**
- `promptpay_id` — phone number (10 digits, e.g. `0812345678`) or national/tax ID (13 digits); **required for QR generation**
- `qr_image_url` — optional static image URL (use QR generation endpoint instead)

```ts
// src/services/config.ts
import api from '../lib/api';
import type { ApiResponse, BankQRConfig } from '../types/api';

export const getBankQRConfig = (): Promise<BankQRConfig> =>
  api.get<ApiResponse<BankQRConfig>>('/api/v1/config/bank-qr').then(r => r.data.data!);

export const setBankQRConfig = (body: Omit<BankQRConfig, 'id' | 'is_active'>): Promise<BankQRConfig> =>
  api.put<ApiResponse<BankQRConfig>>('/api/v1/config/bank-qr', body).then(r => r.data.data!);
```

**Bank QR display screen:**
```tsx
// src/pages/BankQRScreen.tsx
import { useEffect, useState } from 'react';
import { getBankQRConfig } from '../services/config';
import type { BankQRConfig, Order } from '../types/api';

export default function BankQRScreen({ order }: { order: Order }) {
  const [config, setConfig] = useState<BankQRConfig | null>(null);

  useEffect(() => { getBankQRConfig().then(setConfig).catch(() => {}); }, []);

  if (!config) return <p>Bank QR not configured. Ask admin to set it up.</p>;

  return (
    <div>
      <h2>Scan to Pay</h2>
      {config.qr_image_url && <img src={config.qr_image_url} alt="Bank QR Code" style={{ width: 240 }} />}
      <p><strong>Bank:</strong> {config.bank_name}</p>
      <p><strong>Account:</strong> {config.account_name}</p>
      <p><strong>Number:</strong> {config.account_number}</p>
      <p><strong>Amount:</strong> ฿{order.total_amount.toFixed(2)}</p>
      <p style={{ color: '#888', fontSize: 12 }}>Order: {order.pos_order_id}</p>
    </div>
  );
}
```

**Admin config form:**
```tsx
// src/pages/admin/BankQRConfigPage.tsx
import { useEffect, useState } from 'react';
import { getBankQRConfig, setBankQRConfig } from '../services/config';

export default function BankQRConfigPage() {
  const [form, setForm] = useState({ bank_name: '', account_name: '', account_number: '', qr_image_url: '' });
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    getBankQRConfig().then(c => setForm({ bank_name: c.bank_name, account_name: c.account_name, account_number: c.account_number, qr_image_url: c.qr_image_url })).catch(() => {});
  }, []);

  const handleSave = async () => {
    await setBankQRConfig(form);
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

  return (
    <div>
      <h2>Bank QR Config</h2>
      <input placeholder="Bank name (e.g. SCB)" value={form.bank_name} onChange={e => setForm({ ...form, bank_name: e.target.value })} />
      <input placeholder="Account name" value={form.account_name} onChange={e => setForm({ ...form, account_name: e.target.value })} />
      <input placeholder="Account number" value={form.account_number} onChange={e => setForm({ ...form, account_number: e.target.value })} />
      <input placeholder="QR image URL (optional)" value={form.qr_image_url} onChange={e => setForm({ ...form, qr_image_url: e.target.value })} />
      <button onClick={handleSave}>Save</button>
      {saved && <span style={{ color: 'green' }}> Saved!</span>}
    </div>
  );
}
```

---

## 7. Pay Later Management

### How it works

1. Cashier creates order with `payment_method: "PAY_LATER"` + customer info
2. Stock is deducted immediately (goods leave)
3. Backend sends overdue alerts every **1 hour** while `is_paid = false` and `payment_due_date` has passed
4. When customer pays, admin calls **Mark as Paid**

### View Overdue PAY_LATER Orders *(admin)*
```
GET /api/v1/orders?payment_method=PAY_LATER&overdue=true
```

### View All PAY_LATER Orders
```
GET /api/v1/orders?payment_method=PAY_LATER
```

### Mark as Paid *(admin only)*
```
POST /api/v1/orders/:id/pay
```
Sets `is_paid = true` and records `paid_at`. Stops future alerts.

**Response `data`:**
```json
{
  "message": "marked as paid",
  "order": { "id": "...", "is_paid": true, "paid_at": "2026-07-03T09:00:00+07:00", ... }
}
```

```ts
// src/services/orders.ts (add to existing)
export const markOrderPaid = (id: string): Promise<{ message: string; order: Order }> =>
  api.post<ApiResponse<{ message: string; order: Order }>>(`/api/v1/orders/${id}/pay`).then(r => r.data.data!);
```

**Pay Later dashboard:**
```tsx
// src/pages/admin/PayLaterDashboard.tsx
import { useEffect, useState } from 'react';
import { getOrders, markOrderPaid } from '../services/orders';
import type { Order } from '../types/api';

export default function PayLaterDashboard() {
  const [overdueOrders, setOverdueOrders] = useState<Order[]>([]);

  const load = () =>
    getOrders({ payment_method: 'PAY_LATER', overdue: true, limit: 100 })
      .then(r => setOverdueOrders(r.items));

  useEffect(() => { load(); }, []);

  const handlePay = async (id: string) => {
    await markOrderPaid(id);
    load();
  };

  return (
    <div>
      <h2>Overdue Pay-Later Orders ({overdueOrders.length})</h2>
      {overdueOrders.length === 0 && <p>No overdue orders.</p>}
      {overdueOrders.map(o => (
        <div key={o.id} style={{ border: '1px solid red', padding: 12, marginBottom: 8 }}>
          <p><strong>{o.pos_order_id}</strong> — ฿{o.total_amount.toFixed(2)}</p>
          <p>Customer: {o.customer_name} ({o.customer_phone})</p>
          <p>Due: {o.payment_due_date ? new Date(o.payment_due_date).toLocaleDateString() : '-'}</p>
          <button onClick={() => handlePay(o.id)}>Mark as Paid</button>
        </div>
      ))}
    </div>
  );
}
```

### Overdue Alert Webhook

Set `ALERT_WEBHOOK_URL` in `.env` to receive a POST every hour for each overdue order:

```json
{
  "event": "PAY_LATER_OVERDUE",
  "order_id": "uuid",
  "pos_order_id": "ORDER-20260628-ABCD1234",
  "customer_name": "John Doe",
  "customer_phone": "0812345678",
  "total_amount": 195.00,
  "payment_due_date": "2026-07-05T10:00:00Z",
  "days_overdue": 3
}
```
Connect this to LINE Notify, Discord webhook, or any HTTP endpoint.

---

## 8. Stock

### Get Cached Stock (fast)
```
GET /api/v1/stock
```
Returns local cache — no round-trip to Inventory. Updated on startup, every 5 min, and via Inventory webhooks.

### Check Real-time Availability
```
GET /api/v1/stock/availability/:pos_product_id?quantity=2
```

### Force Re-sync *(admin only)*
```
POST /api/v1/stock/sync
```

```ts
// src/services/stock.ts
export const getStock = () =>
  api.get<ApiResponse<StockItem[]>>('/api/v1/stock').then(r => r.data.data!);

export const checkAvailability = (posProductId: string, quantity = 1) =>
  api.get<ApiResponse<ProductAvailability>>(`/api/v1/stock/availability/${posProductId}`, { params: { quantity } })
    .then(r => r.data.data!);

export const syncStock = () =>
  api.post<ApiResponse<{ message: string; count: number }>>('/api/v1/stock/sync').then(r => r.data.data!);
```

---

## 9. Role-Based Access

```ts
export const isAdmin = () => getCurrentUser()?.role === 'admin';
```

```tsx
{isAdmin() && <button onClick={() => syncStock()}>Sync Stock</button>}
{isAdmin() && <Link to="/admin/pay-later">Pay Later Dashboard</Link>}
{isAdmin() && <Link to="/admin/bank-qr">Bank QR Config</Link>}
```

---

## 10. Error Handling

```ts
// src/lib/errors.ts
import { AxiosError } from 'axios';

export function getErrorMessage(err: unknown, fallback = 'Something went wrong'): string {
  if (err instanceof AxiosError) return err.response?.data?.message ?? fallback;
  if (err instanceof Error) return err.message;
  return fallback;
}
```

| HTTP | Meaning |
|---|---|
| `400` | Validation error or business rule (see `message`) |
| `401` | Missing/invalid JWT — auto-redirected to `/login` |
| `403` | Admin-only action attempted by cashier |
| `404` | Resource not found |
| `502` | Inventory system unreachable |
| `500` | Internal server error |

---

## 11. Quick Reference

### All Endpoints

| Method | Path | Auth | Role |
|---|---|---|---|
| `POST` | `/auth/login` | — | public |
| `GET` | `/health` | — | public |
| `POST` | `/api/v1/users/register` | JWT | admin |
| `GET` | `/api/v1/products` | JWT | any |
| `GET` | `/api/v1/products/:id` | JWT | any |
| `POST` | `/api/v1/products` | JWT | admin |
| `PUT` | `/api/v1/products/:id` | JWT | admin |
| `DELETE` | `/api/v1/products/:id` | JWT | admin |
| `POST` | `/api/v1/orders` | JWT | any |
| `GET` | `/api/v1/orders` | JWT | any* |
| `GET` | `/api/v1/orders/:id` | JWT | any |
| `POST` | `/api/v1/orders/:id/cancel` | JWT | any** |
| `POST` | `/api/v1/orders/:id/pay` | JWT | admin |
| `GET` | `/api/v1/config/bank-qr/qrcode?amount=N` | JWT | any |
| `GET` | `/api/v1/config/bank-qr` | JWT | any |
| `PUT` | `/api/v1/config/bank-qr` | JWT | admin |
| `GET` | `/api/v1/stock` | JWT | any |
| `POST` | `/api/v1/stock/sync` | JWT | admin |
| `GET` | `/api/v1/stock/availability/:id?quantity=N` | JWT | any |

\* Cashiers see only their own orders.  
\*\* Cashiers can cancel only their own PENDING orders.

### Order `status` Lifecycle

```
PENDING → COMPLETED  (stock deducted, sale done)
        → FAILED     (inventory rejected or unreachable)
        → CANCELLED  (manually cancelled while PENDING)
```

### Payment Method Field Guide

| Field | CASH | BANK_QRCODE | PAY_LATER |
|---|---|---|---|
| `payment_method` | `"CASH"` | `"BANK_QRCODE"` | `"PAY_LATER"` |
| `customer_name` | — | — | required |
| `customer_phone` | — | — | required |
| `payment_due_days` | — | — | optional (default 7) |
| `is_paid` | — | — | `false` until `POST /:id/pay` |
| Show bank QR after order? | no | **yes** | no |

### `src/services/` file map

```
src/
├── lib/
│   ├── api.ts        # axios instance + interceptors
│   └── errors.ts     # getErrorMessage()
├── services/
│   ├── auth.ts       # login, logout, getCurrentUser
│   ├── products.ts   # CRUD
│   ├── orders.ts     # createOrder, getOrders, cancelOrder, markOrderPaid
│   ├── config.ts     # getBankQRConfig, setBankQRConfig
│   └── stock.ts      # getStock, checkAvailability, syncStock
└── types/
    └── api.ts        # all TypeScript types
```
