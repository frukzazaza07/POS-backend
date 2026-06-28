# POS Backend — React Frontend Integration Guide

Base URL: `http://localhost:4000`  
All protected endpoints require `Authorization: Bearer <token>` in the request header.

---

## Table of Contents

1. [TypeScript Types](#1-typescript-types)
2. [Axios Setup](#2-axios-setup)
3. [Authentication](#3-authentication)
4. [Products](#4-products)
5. [Orders](#5-orders)
6. [Stock](#6-stock)
7. [Error Handling](#7-error-handling)
8. [Role-Based Access](#8-role-based-access)
9. [Quick Reference](#9-quick-reference)

---

## 1. TypeScript Types

Create `src/types/api.ts` and paste everything below. These mirror the Go models exactly.

```ts
// src/types/api.ts

export type UserRole = 'admin' | 'cashier';
export type OrderStatus = 'PENDING' | 'COMPLETED' | 'CANCELLED' | 'FAILED';

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
  pos_product_id: string;   // matches Inventory system's pos_product_id
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
  notes: string;
  fail_reason?: string;
  items?: OrderItem[];
  created_at: string;
  updated_at: string;
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

export interface AvailabilityDetail {
  inventory_item_id: string;
  sku: string;
  name: string;
  required: number;
  available: number;
  is_sufficient: boolean;
}

export interface ProductAvailability {
  pos_product_id: string;
  name: string;
  is_available: boolean;
  details: AvailabilityDetail[];
}

// Generic API response wrappers
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

Install axios: `npm install axios`

Create `src/lib/api.ts`:

```ts
// src/lib/api.ts
import axios from 'axios';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? 'http://localhost:4000',
  headers: { 'Content-Type': 'application/json' },
});

// Attach JWT token to every request automatically
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('pos_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Redirect to login on 401
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

Add to `.env.local`:
```
VITE_API_URL=http://localhost:4000
```

---

## 3. Authentication

### 3.1 Login

**Endpoint:** `POST /auth/login`  
**Auth:** None (public)

**Request body:**
```json
{
  "email": "admin@pos.local",
  "password": "admin123"
}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": "uuid",
      "name": "Admin",
      "email": "admin@pos.local",
      "role": "admin",
      "is_active": true,
      "created_at": "2026-06-28T00:00:00Z",
      "updated_at": "2026-06-28T00:00:00Z"
    }
  }
}
```

**React code:**
```ts
// src/services/auth.ts
import api from '../lib/api';
import type { User, ApiResponse } from '../types/api';

interface LoginResponse {
  token: string;
  user: User;
}

export async function login(email: string, password: string): Promise<LoginResponse> {
  const { data } = await api.post<ApiResponse<LoginResponse>>('/auth/login', {
    email,
    password,
  });
  const result = data.data!;

  // Persist token and user
  localStorage.setItem('pos_token', result.token);
  localStorage.setItem('pos_user', JSON.stringify(result.user));

  return result;
}

export function logout() {
  localStorage.removeItem('pos_token');
  localStorage.removeItem('pos_user');
  window.location.href = '/login';
}

export function getCurrentUser(): User | null {
  const raw = localStorage.getItem('pos_user');
  return raw ? JSON.parse(raw) : null;
}

export function isAuthenticated(): boolean {
  return !!localStorage.getItem('pos_token');
}
```

**Login form example:**
```tsx
// src/pages/Login.tsx
import { useState } from 'react';
import { login } from '../services/auth';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    try {
      await login(email, password);
      window.location.href = '/';
    } catch (err: any) {
      setError(err.response?.data?.message ?? 'Login failed');
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <input value={email} onChange={(e) => setEmail(e.target.value)} placeholder="Email" />
      <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder="Password" />
      {error && <p style={{ color: 'red' }}>{error}</p>}
      <button type="submit">Login</button>
    </form>
  );
}
```

---

### 3.2 Register New User (admin only)

**Endpoint:** `POST /api/v1/users/register`  
**Auth:** JWT — admin only

**Request body:**
```json
{
  "name": "Jane Doe",
  "email": "jane@pos.local",
  "password": "securepass",
  "role": "cashier"
}
```

`role` can be `"admin"` or `"cashier"` (defaults to `"cashier"` if omitted).

**React code:**
```ts
export async function registerUser(payload: {
  name: string;
  email: string;
  password: string;
  role?: 'admin' | 'cashier';
}): Promise<User> {
  const { data } = await api.post<ApiResponse<User>>('/api/v1/users/register', payload);
  return data.data!;
}
```

---

## 4. Products

### 4.1 List Products

**Endpoint:** `GET /api/v1/products`  
**Auth:** JWT (any role)  
**Query params:** `search`, `page` (default 1), `limit` (default 20)

**Response:**
```json
{
  "status": "success",
  "data": {
    "items": [
      {
        "id": "uuid",
        "pos_product_id": "pos-latte",
        "name": "Cafe Latte",
        "description": "Espresso with steamed milk",
        "price": 65.00,
        "category": "beverages",
        "is_active": true,
        "created_at": "2026-06-28T00:00:00Z",
        "updated_at": "2026-06-28T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "limit": 20
  }
}
```

**React code:**
```ts
// src/services/products.ts
import api from '../lib/api';
import type { ApiResponse, PaginatedResponse, POSProduct } from '../types/api';

export async function getProducts(params?: {
  search?: string;
  page?: number;
  limit?: number;
}): Promise<PaginatedResponse<POSProduct>> {
  const { data } = await api.get<ApiResponse<PaginatedResponse<POSProduct>>>(
    '/api/v1/products',
    { params }
  );
  return data.data!;
}
```

**Hook example:**
```tsx
// src/hooks/useProducts.ts
import { useEffect, useState } from 'react';
import { getProducts } from '../services/products';
import type { POSProduct } from '../types/api';

export function useProducts(search = '', page = 1) {
  const [products, setProducts] = useState<POSProduct[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    setLoading(true);
    getProducts({ search, page, limit: 20 })
      .then((res) => {
        setProducts(res.items);
        setTotal(res.total);
      })
      .catch((err) => setError(err.response?.data?.message ?? 'Failed to load'))
      .finally(() => setLoading(false));
  }, [search, page]);

  return { products, total, loading, error };
}
```

---

### 4.2 Get Single Product

**Endpoint:** `GET /api/v1/products/:id`  
**Auth:** JWT

```ts
export async function getProduct(id: string): Promise<POSProduct> {
  const { data } = await api.get<ApiResponse<POSProduct>>(`/api/v1/products/${id}`);
  return data.data!;
}
```

---

### 4.3 Create Product (admin only)

**Endpoint:** `POST /api/v1/products`  
**Auth:** JWT — admin only

**Request body:**
```json
{
  "pos_product_id": "pos-latte",
  "name": "Cafe Latte",
  "description": "Espresso with steamed milk",
  "price": 65.00,
  "category": "beverages",
  "is_active": true
}
```

> `pos_product_id` **must match** the `pos_product_id` you registered in the Inventory system for stock deduction to work.

```ts
export async function createProduct(payload: {
  pos_product_id: string;
  name: string;
  description?: string;
  price: number;
  category?: string;
  is_active?: boolean;
}): Promise<POSProduct> {
  const { data } = await api.post<ApiResponse<POSProduct>>('/api/v1/products', payload);
  return data.data!;
}
```

---

### 4.4 Update Product (admin only)

**Endpoint:** `PUT /api/v1/products/:id`  
**Auth:** JWT — admin only

```ts
export async function updateProduct(
  id: string,
  payload: Partial<{
    name: string;
    description: string;
    price: number;
    category: string;
    is_active: boolean;
  }>
): Promise<POSProduct> {
  const { data } = await api.put<ApiResponse<POSProduct>>(
    `/api/v1/products/${id}`,
    payload
  );
  return data.data!;
}
```

---

### 4.5 Delete Product (admin only)

**Endpoint:** `DELETE /api/v1/products/:id`  
**Auth:** JWT — admin only

```ts
export async function deleteProduct(id: string): Promise<void> {
  await api.delete(`/api/v1/products/${id}`);
}
```

---

## 5. Orders

### 5.1 Create Order (process a sale)

**Endpoint:** `POST /api/v1/orders`  
**Auth:** JWT (any role)

**Request body:**
```json
{
  "notes": "Table 5, extra sugar",
  "items": [
    { "pos_product_id": "pos-latte",     "quantity": 2 },
    { "pos_product_id": "pos-americano", "quantity": 1 }
  ]
}
```

**Response — success (`HTTP 201`):**
```json
{
  "status": "success",
  "data": {
    "id": "uuid",
    "pos_order_id": "ORDER-20260628-A1B2C3D4",
    "cashier_id": "uuid",
    "status": "COMPLETED",
    "total_amount": 195.00,
    "notes": "Table 5, extra sugar",
    "items": [
      {
        "id": "uuid",
        "order_id": "uuid",
        "pos_product_id": "pos-latte",
        "product_name": "Cafe Latte",
        "quantity": 2,
        "unit_price": 65.00,
        "subtotal": 130.00
      },
      {
        "id": "uuid",
        "order_id": "uuid",
        "pos_product_id": "pos-americano",
        "product_name": "Americano",
        "quantity": 1,
        "unit_price": 65.00,
        "subtotal": 65.00
      }
    ],
    "created_at": "2026-06-28T10:00:00Z",
    "updated_at": "2026-06-28T10:00:00Z"
  }
}
```

**Response — stock insufficient (`HTTP 400`):**
```json
{
  "status": "error",
  "message": "stock deduction failed: insufficient stock for RAW-COFFEE: have 10.00, need 18.00"
}
```

**Order status values:**

| Status | Meaning |
|---|---|
| `PENDING` | Created, awaiting inventory deduction |
| `COMPLETED` | Stock deducted successfully — sale done |
| `FAILED` | Inventory deduction failed (see `fail_reason`) |
| `CANCELLED` | Manually cancelled before processing |

**React code:**
```ts
// src/services/orders.ts
import api from '../lib/api';
import type { ApiResponse, Order } from '../types/api';

export interface CreateOrderPayload {
  notes?: string;
  items: Array<{
    pos_product_id: string;
    quantity: number;
  }>;
}

export async function createOrder(payload: CreateOrderPayload): Promise<Order> {
  const { data } = await api.post<ApiResponse<Order>>('/api/v1/orders', payload);
  return data.data!;
}
```

**Checkout component example:**
```tsx
// src/components/Checkout.tsx
import { useState } from 'react';
import { createOrder } from '../services/orders';
import type { POSProduct } from '../types/api';

interface CartItem {
  product: POSProduct;
  quantity: number;
}

export default function Checkout({ cart }: { cart: CartItem[] }) {
  const [notes, setNotes] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const total = cart.reduce((sum, i) => sum + i.product.price * i.quantity, 0);

  const handleCheckout = async () => {
    setLoading(true);
    setError('');
    try {
      const order = await createOrder({
        notes,
        items: cart.map((i) => ({
          pos_product_id: i.product.pos_product_id,
          quantity: i.quantity,
        })),
      });
      alert(`Order ${order.pos_order_id} completed! Total: ฿${order.total_amount}`);
    } catch (err: any) {
      setError(err.response?.data?.message ?? 'Checkout failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <p>Total: ฿{total.toFixed(2)}</p>
      <input value={notes} onChange={(e) => setNotes(e.target.value)} placeholder="Notes" />
      {error && <p style={{ color: 'red' }}>{error}</p>}
      <button onClick={handleCheckout} disabled={loading || cart.length === 0}>
        {loading ? 'Processing...' : 'Confirm Sale'}
      </button>
    </div>
  );
}
```

---

### 5.2 List Orders

**Endpoint:** `GET /api/v1/orders`  
**Auth:** JWT  
**Query params:** `page` (default 1), `limit` (default 20)

> Cashiers see only their own orders. Admins see all orders.

**Response:**
```json
{
  "status": "success",
  "data": {
    "items": [ /* Order[] */ ],
    "total": 42,
    "page": 1,
    "limit": 20
  }
}
```

```ts
export async function getOrders(params?: {
  page?: number;
  limit?: number;
}): Promise<PaginatedResponse<Order>> {
  const { data } = await api.get<ApiResponse<PaginatedResponse<Order>>>(
    '/api/v1/orders',
    { params }
  );
  return data.data!;
}
```

---

### 5.3 Get Single Order

**Endpoint:** `GET /api/v1/orders/:id`  
**Auth:** JWT

```ts
export async function getOrder(id: string): Promise<Order> {
  const { data } = await api.get<ApiResponse<Order>>(`/api/v1/orders/${id}`);
  return data.data!;
}
```

---

### 5.4 Cancel Order

**Endpoint:** `POST /api/v1/orders/:id/cancel`  
**Auth:** JWT  
**Rules:** Only `PENDING` orders can be cancelled. Cashiers can only cancel their own orders; admins can cancel any.

**Response:**
```json
{
  "status": "success",
  "data": { "message": "order cancelled" }
}
```

```ts
export async function cancelOrder(id: string): Promise<void> {
  await api.post(`/api/v1/orders/${id}/cancel`);
}
```

---

## 6. Stock

### 6.1 Get Cached Stock Levels

**Endpoint:** `GET /api/v1/stock`  
**Auth:** JWT  

Returns the local cache — fast, no network call to Inventory. Updated on app startup, every 5 minutes, and via webhook events.

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
      "quantity_in_stock": 4946.0,
      "min_quantity": 500.0,
      "is_low": false,
      "is_out": false,
      "synced_at": "2026-06-28T10:00:00Z"
    }
  ]
}
```

```ts
// src/services/stock.ts
import api from '../lib/api';
import type { ApiResponse, StockItem, ProductAvailability } from '../types/api';

export async function getStock(): Promise<StockItem[]> {
  const { data } = await api.get<ApiResponse<StockItem[]>>('/api/v1/stock');
  return data.data!;
}
```

**Stock indicator component:**
```tsx
// src/components/StockBadge.tsx
import type { StockItem } from '../types/api';

export function StockBadge({ item }: { item: StockItem }) {
  if (item.is_out) return <span style={{ color: 'red' }}>Out of Stock</span>;
  if (item.is_low) return <span style={{ color: 'orange' }}>Low Stock</span>;
  return <span style={{ color: 'green' }}>In Stock</span>;
}
```

---

### 6.2 Check Real-Time Availability

**Endpoint:** `GET /api/v1/stock/availability/:pos_product_id`  
**Auth:** JWT  
**Query params:** `quantity` (default 1)

Use this before showing a product as available in the order screen.

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
        "available": 4946,
        "is_sufficient": true
      }
    ]
  }
}
```

```ts
export async function checkAvailability(
  posProductId: string,
  quantity = 1
): Promise<ProductAvailability> {
  const { data } = await api.get<ApiResponse<ProductAvailability>>(
    `/api/v1/stock/availability/${posProductId}`,
    { params: { quantity } }
  );
  return data.data!;
}
```

---

### 6.3 Force Stock Sync (admin only)

**Endpoint:** `POST /api/v1/stock/sync`  
**Auth:** JWT — admin only

Triggers an immediate full sync from the Inventory system.

**Response:**
```json
{
  "status": "success",
  "data": { "message": "sync complete", "count": 12 }
}
```

```ts
export async function syncStock(): Promise<{ message: string; count: number }> {
  const { data } = await api.post<ApiResponse<{ message: string; count: number }>>(
    '/api/v1/stock/sync'
  );
  return data.data!;
}
```

---

## 7. Error Handling

Every error response has the same shape:

```json
{
  "status": "error",
  "message": "description of what went wrong"
}
```

| HTTP Status | Meaning |
|---|---|
| `400` | Bad request — validation failed or stock insufficient |
| `401` | Missing or invalid JWT token (redirected to `/login` automatically) |
| `403` | Admin-only endpoint — cashier attempted access |
| `404` | Resource not found |
| `502` | Inventory system unreachable |
| `500` | Server error |

**Centralised error extraction helper:**

```ts
// src/lib/errors.ts
import { AxiosError } from 'axios';

export function getErrorMessage(err: unknown, fallback = 'Something went wrong'): string {
  if (err instanceof AxiosError) {
    return err.response?.data?.message ?? fallback;
  }
  if (err instanceof Error) return err.message;
  return fallback;
}
```

**Usage:**
```tsx
import { getErrorMessage } from '../lib/errors';

try {
  await createOrder(payload);
} catch (err) {
  setError(getErrorMessage(err, 'Failed to process order'));
}
```

---

## 8. Role-Based Access

```ts
// src/lib/auth.ts
import { getCurrentUser } from '../services/auth';

export function isAdmin(): boolean {
  return getCurrentUser()?.role === 'admin';
}
```

**Conditional UI rendering:**
```tsx
import { isAdmin } from '../lib/auth';

// Only show admin actions
{isAdmin() && (
  <button onClick={() => syncStock()}>Sync Stock</button>
)}

{isAdmin() && (
  <button onClick={() => deleteProduct(id)}>Delete Product</button>
)}
```

**Protected route example (React Router v6):**
```tsx
// src/components/AdminRoute.tsx
import { Navigate, Outlet } from 'react-router-dom';
import { getCurrentUser } from '../services/auth';

export function AdminRoute() {
  const user = getCurrentUser();
  if (!user) return <Navigate to="/login" />;
  if (user.role !== 'admin') return <Navigate to="/" />;
  return <Outlet />;
}

// In router setup:
// <Route element={<AdminRoute />}>
//   <Route path="/admin/products" element={<ProductManagement />} />
//   <Route path="/admin/users" element={<UserManagement />} />
// </Route>
```

---

## 9. Quick Reference

### All Endpoints

| Method | Path | Auth | Role |
|---|---|---|---|
| `POST` | `/auth/login` | — | public |
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
| `GET` | `/api/v1/stock` | JWT | any |
| `POST` | `/api/v1/stock/sync` | JWT | admin |
| `GET` | `/api/v1/stock/availability/:pos_product_id?quantity=N` | JWT | any |
| `GET` | `/health` | — | public |

\* Cashiers see only their own orders; admins see all.  
\*\* Cashiers can cancel only their own PENDING orders; admins can cancel any.

### `src/services/` file map

```
src/
├── lib/
│   ├── api.ts          # axios instance + interceptors
│   ├── auth.ts         # isAdmin() helper
│   └── errors.ts       # getErrorMessage()
├── services/
│   ├── auth.ts         # login, logout, getCurrentUser
│   ├── products.ts     # getProducts, getProduct, createProduct, updateProduct, deleteProduct
│   ├── orders.ts       # createOrder, getOrders, getOrder, cancelOrder
│   └── stock.ts        # getStock, checkAvailability, syncStock
├── hooks/
│   └── useProducts.ts  # example data-fetching hook
└── types/
    └── api.ts          # all TypeScript types
```

### Token Lifecycle

- Token is valid for **24 hours**
- Stored in `localStorage` under key `pos_token`
- Automatically attached to every request via the axios interceptor
- On `401` response, token is cleared and user is redirected to `/login`
- No refresh token — user must log in again after expiry
