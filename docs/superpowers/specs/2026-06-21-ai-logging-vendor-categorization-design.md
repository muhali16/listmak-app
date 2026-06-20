# AI Request Logging & Vendor Categorization Design

**Date:** 2026-06-21
**Status:** Approved

## Overview

Two related features:
1. Log every AI call (input, output, latency, status) to PostgreSQL for internal admin visibility
2. Add "Kelompokkan per lokasi" button on OrderListView that triggers AI vendor scan and switches to vendor-grouped display

## Data Model

### New table: `ai_logs` (PostgreSQL, not SQLite)

```go
type AILog struct {
    ID         uint      `gorm:"primaryKey"`
    OrderID    *uint     `gorm:"index"`           // nullable FK → orders.id
    Input      string    `gorm:"type:text"`        // orderDetail sent to AI
    Output     string    `gorm:"type:text"`        // vendor name returned
    Model      string                              // e.g. "accounts/fireworks/..."
    Provider   string                              // "fireworks" | "noop"
    LatencyMs  int64                               // call duration in ms
    Status     string                              // "success" | "failed"
    ErrorMsg   string                              // populated when status="failed"
    CreatedAt  time.Time
}
```

`OrderID` is nullable because `ExtractVendor` may be called from public share form (no order ID yet at call time).

### Modified: `users` table

Add field:
```go
Role string `gorm:"default:'user'"` // "user" | "admin"
```

## Role & Access Control

**Bootstrap via env:**
```
ADMIN_EMAILS=muhammadali55214@gmail.com,other@example.com
```

On Google OAuth callback (`auth_controller.go`):
- Parse `ADMIN_EMAILS`, split by `,`, trim whitespace
- If user email in list → set `role = "admin"` on upsert
- Email in env always wins (overrides manual demotion on login)

**Manual promotion by admin:**
```
PATCH /api/v1/admin/users/:id/role
Body: { "role": "admin" | "user" }
```
Persisted in PostgreSQL. Survives across logins (unless email is also in env list).

**JWT:** encode `role` field in claims so middleware can check without DB hit.

**`AdminOnly` middleware:** checks `role == "admin"` from JWT claim. Returns 403 if not admin.

## AIService Changes

### Signature change

```go
type AIService interface {
    ExtractVendor(orderDetail string, orderID *uint) (string, error)
}
```

AIService receives `AILogRepository` at construction time:

```go
func NewFireworksAIService(apiKey, model string, logRepo AILogRepository) AIService
```

`noopAIService` also accepts logRepo but does not log (no real call = nothing to record).

### AILogRepository interface

```go
type AILogRepository interface {
    Create(log *models.AILog) error
}
```

### Logging flow inside `ExtractVendor`

```
start := time.Now()
// build & send HTTP request
if err or non-200:
    logRepo.Create({ Status: "failed", ErrorMsg: ..., LatencyMs: elapsed, Output: "" })
    return "", err
// parse response
logRepo.Create({ Status: "success", Output: vendor, LatencyMs: elapsed })
return vendor, nil
```

## New Endpoint: Scan Vendors

```
POST /api/v1/listmaks/:id/scan-vendors
Auth: required (JWT cookie)
```

**Behavior:**
1. Fetch all orders for listmak `:id`
2. Filter: only orders where `vendor_name == ""` (skip already-categorized)
3. For each filtered order: call `AIService.ExtractVendor(order.order_detail, &order.ID)`
4. If result non-empty: update `orders.vendor_name` in DB
5. Each call logs to `ai_logs` automatically (via AIService)
6. Return all orders for the listmak (refreshed)

**Response:** same shape as `GET /api/v1/listmaks/:id/orders`

No parallelism — sequential calls to avoid rate-limiting Fireworks API.

## Admin Endpoints

All behind `AuthMiddleware` + `AdminOnly`:

```
GET  /api/v1/admin/ai-logs           → paginated list, newest first, limit 50
PATCH /api/v1/admin/users/:id/role   → { "role": "admin" | "user" }
```

`GET /api/v1/admin/ai-logs` response fields: `id, order_id, input, output, model, provider, latency_ms, status, error_msg, created_at`

## Frontend: OrderListView Changes

### Vendor scan button

Toggle button near summary card or above search:
- Label: "Kelompokkan per lokasi"
- Click triggers:
  1. `POST /api/v1/listmaks/:id/scan-vendors` (with loading state)
  2. On success: update `this.orders` with returned data
  3. Switch `groupMode` to `'vendor'`

### Computed: vendor grouping

New computed `vendorGroups` (used when `groupMode === 'vendor'`):
- Group orders by `vendor_name`
- Orders with empty `vendor_name` → group key `"Belum ada lokasi"`
- Same card structure as existing name-grouped view
- No paid toggle on vendor groups (paid is per-person, not per-vendor)

### Toggle behavior

`groupMode` data property: `'name'` (default) | `'vendor'`

Button shows active state when `groupMode === 'vendor'`. Click again → back to `'name'`, no re-scan.

## Frontend: Admin Dashboard

Route: `/admin/ai-logs`
- Guard: redirect to `/today` if `user.role !== 'admin'`
- Hidden from `Sidebar.vue` and `BottomNav.vue` (admin accesses direct URL)
- `meta: { requiresAuth: true }`

Table columns: Waktu · Input (truncated 60 chars) · Output · Provider · Latency · Status (badge hijau/merah)

Pagination: load-more button, 50 records per page.

## Wiring (container.go)

```
aiLogRepo (PostgreSQL DB)
  → NewFireworksAIService(apiKey, model, aiLogRepo)
  → OrderController / ShareController
```

Migration: `AutoMigrate(&models.AILog{})` added to `configs/migration.go`.
`Role` field migration: GORM AutoMigrate adds column to existing `users` table safely.
