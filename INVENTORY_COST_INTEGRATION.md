# Inventory System — Cost & Profit Integration Guide

Written for whoever implements this on the **Inventory system** side. POS
already has a working, self-contained cost/profit feature (admin-set
`cost_price` per product, snapshotted onto each order). This document
describes the *optional* upgrade: letting Inventory compute the real
recipe cost so POS no longer needs a manually maintained number.

Nothing here is required for POS's cost/profit feature to work today — it's
already live using the manual fallback. This is a follow-up for accuracy.

---

## Why

POS tracks `price` (what the customer pays) and needs `cost` (what the
ingredients cost) to report profit. Right now `cost_price` is a flat number
an admin types into the POS product form. It drifts from reality the moment
ingredient prices change, and it ignores the actual recipe (BOM) — a Latte
and an Espresso can't have accurate independent costs without knowing how
much coffee, milk, and cups each one actually consumes.

Inventory already owns the BOM (`quantity_required` per ingredient per
product — see `barcode` lookup) and the stock item records. Adding a `cost`
field to inventory items and computing recipe cost server-side, at the same
place stock is deducted, keeps POS simple and keeps cost logic in one place
(no duplicated recipe math in two systems that can drift apart).

---

## What to add

### 1. `unit_cost` on inventory items

Add a cost-per-unit field to whatever your inventory item record is (the
thing currently exposed via `GET /api/v1/pos/stock/levels` and the `bom[].
inventory_item` object in the barcode endpoint). Same unit as
`quantity_in_stock` (e.g. cost per gram, per ml, per piece) — not a
per-package cost.

```
inventory_item.unit_cost: number   // e.g. 0.85 (THB per gram of coffee beans)
```

Whether this is a manually-entered purchase cost, a weighted average, or
FIFO/LIFO doesn't matter to POS — POS only ever consumes the current value.

### 2. Include recipe cost in the stock-deduct response

This is the field POS actually reads. `POST /api/v1/pos/stock/deduct`
already receives the exact order items (`pos_product_id` + `quantity`) and
already resolves each product's BOM to deduct stock. At that same point,
compute cost — you have the ingredient quantities and their `unit_cost`
right there.

**Current response** (unchanged, still required):
```json
{
  "status": "success",
  "data": {
    "pos_order_id": "ORDER-20260628-0001",
    "status": "processed",
    "deductions": [
      { "inventory_item_id": "inv-coffee-beans-001", "sku": "RAW-COFFEE-BEANS",
        "name": "Coffee Beans", "quantity_deducted": 54, "quantity_remaining": 4946 }
    ]
  }
}
```

**Add a `cost_breakdown` array**, one entry per distinct `pos_product_id` in
the request — not per ingredient, POS wants it pre-aggregated per product:

```json
{
  "status": "success",
  "data": {
    "pos_order_id": "ORDER-20260628-0001",
    "status": "processed",
    "deductions": [ /* unchanged */ ],
    "cost_breakdown": [
      {
        "pos_product_id": "pos-latte",
        "unit_cost": 22.50,
        "total_cost": 45.00
      },
      {
        "pos_product_id": "pos-espresso",
        "unit_cost": 15.00,
        "total_cost": 15.00
      }
    ]
  }
}
```

- `unit_cost` = cost of one unit of that product (Σ `bom[].quantity_required
  × inventory_item.unit_cost` for that product's recipe)
- `total_cost` = `unit_cost × quantity` for that line of the order

This field is **additive and optional from POS's perspective** — POS's Go
client (`internal/service/inventory_client.go`) already has a
`CostBreakdown []CostBreakdownItem` field on the deduct response struct that
safely stays empty until you start sending it. No POS deploy is needed on
your side to ship this; POS will pick it up automatically the next order
after you start returning it.

### 3. (Optional) Expose `unit_cost` on read-only endpoints too

Not required for the core feature, but useful if you ever want an admin
screen in Inventory itself to show recipe cost before a sale happens:

- `GET /api/v1/pos/stock/levels` → add `unit_cost` per item
- `GET /api/v1/pos/products/{pos_product_id}/availability` → add `unit_cost`
  to each `details[]` entry
- `GET /api/v1/pos/products/barcode/{barcode}` → add `unit_cost` to
  `bom[].inventory_item`

POS doesn't currently read `unit_cost` from any of these three — only from
`cost_breakdown` in the deduct response — so these are purely optional
transparency, not a dependency.

---

## What NOT to change

- **Don't remove or rename** anything in the existing `deductions` array —
  POS still uses it for the standard idempotency/audit trail.
- **Don't require `cost_breakdown`** to be present for the deduct call to
  succeed — if a product has no BOM cost data yet, omit it from the array
  (or send `unit_cost: 0`) rather than failing the whole deduction. Stock
  deduction must never fail because of a costing gap.
- **Don't change the `already_processed` idempotency behavior** — cost data
  is a snapshot, it doesn't need to be recomputed or re-sent on a duplicate
  `pos_order_id` request.

---

## Rollout

1. Add `unit_cost` to your inventory item storage (nullable/defaults to 0 is
   fine — POS treats missing cost as 0, same as an admin who hasn't set
   `cost_price` yet).
2. Backfill or manually enter `unit_cost` for your raw materials.
3. Ship `cost_breakdown` in the deduct response.
4. Watch POS reports (`GET /api/v1/reports/products/top`,
   `GET /api/v1/reports/summary` — admin JWT required) — `total_cost` and
   `profit` per order will start reflecting real recipe cost the moment an
   order is placed after your change ships. No coordination needed beyond
   that; POS overrides its manual `cost_price` fallback automatically
   whenever `cost_breakdown` is non-empty.
5. Once `unit_cost` coverage is good, you can stop telling admins to
   maintain `cost_price` manually in POS — it becomes a fallback for
   products Inventory hasn't costed yet, not the primary source.

---

## Reference: what POS already has (no action needed here)

For context — this is already live in POS, described here so you know what
you're feeding data into:

- `POSProduct.cost_price` — admin-editable manual cost, used as the
  fallback until `cost_breakdown` arrives.
- `OrderItem.cost_price` — snapshotted per unit at order time (from the
  fallback, then overwritten if `cost_breakdown` is present in the deduct
  response).
- `Order.total_cost` — sum of item costs; `Order.profit` — derived as
  `total_amount - total_cost`, computed at read time, never stored.
- All of the above are **admin-only** — stripped from every response
  (`products`, `orders`) for the `cashier` role. Reports (`/api/v1/reports/*`)
  are admin-only routes already, so cost/profit fields there need no
  additional gating.
