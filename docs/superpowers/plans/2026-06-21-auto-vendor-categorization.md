# Auto Vendor Categorization + Food Suggestions — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.
>
> **CONFIRM GATE:** Each task ends with ⏸️. Do NOT start the next task until the user confirms. Update checkboxes as you complete each step — this is the session history for continuity across Claude Code sessions.

**Goal:** Add async vendor categorization via Fireworks.ai LLM + food autocomplete suggestions on the shared order form, with vendor grouping in the OB order view.

**Architecture:** `orders.vendor_name` field populated async after each order save via Fireworks.ai goroutine. Public endpoint serves food suggestions from historical orders via Postgres ILIKE. Frontend shows autocomplete + location nudge on shared form, and groups orders by vendor in OrderListView with inline edit chip.

**Tech Stack:** Go 1.25 (Gin + GORM + Postgres), Vue 3 Options API, Fireworks.ai REST (OpenAI-compatible), pg_trgm extension

## Global Constraints

- Module: `github.com/muhali16/listmak-service`
- All API responses via `utils.SendResponse()` or matching `c.JSON()` pattern — no exceptions
- Frontend uses Vue **Options API** — match existing file style exactly (no Composition API / `<script setup>`)
- AI goroutine is fire-and-forget — order save must never fail due to AI failure
- Fireworks.ai HTTP timeout: 10 seconds
- No breaking changes to existing endpoint contracts or interfaces

## Execution History

Check boxes as tasks complete. A new session can resume from the first unchecked task.

- [ ] Task 1: Schema + Model — `vendor_name` column, pg_trgm
- [ ] Task 2: AI Service — `internal/services/ai_service.go`
- [ ] Task 3: Order Repository — `UpdateVendorName`, `GetFoodSuggestions`
- [ ] Task 4: Order Service — AI injection, goroutine, new service methods
- [ ] Task 5: Food Suggestions endpoint — `GET /share-links/:shareId/food-suggestions`
- [ ] Task 6: Vendor Override endpoint — `PATCH /orders/:id/vendor`
- [ ] Task 7: Wire DI + env vars — `container.go`, `.env`
- [ ] Task 8: Frontend SharedListMakView — autocomplete + nudge
- [ ] Task 9: Frontend OrderListView — vendor grouping + inline edit

---

### Task 1: Schema + Model

**Files:**
- Modify: `backend/internal/models/order_model.go`
- Modify: `backend/internal/configs/migration.go`

**Interfaces:**
- Produces: `models.Order.VendorName string` — used by all subsequent tasks

- [ ] **Step 1: Add `VendorName` to Order struct**

In `backend/internal/models/order_model.go`, add `VendorName` after `IsPaid`/`PaidAt` and before `AddedVia`:

```go
type Order struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	ListmakID   uint           `gorm:"not null;index" json:"listmak_id"`
	Name        string         `gorm:"type:varchar(100);not null;index" json:"name"`
	OrderDetail string         `gorm:"type:text;not null" json:"order_detail"`
	Price       float64        `gorm:"type:decimal(12,2);default:0" json:"price"`
	Qty         int            `gorm:"default:1" json:"qty"`
	TotalPrice  float64        `gorm:"type:decimal(12,2);generated:always as (price * qty) stored;<-:false" json:"total_price"`
	IsPaid      bool           `gorm:"default:false;index" json:"is_paid"`
	PaidAt      *time.Time     `json:"paid_at"`
	VendorName  string         `gorm:"type:varchar(100)" json:"vendor_name"`
	AddedVia    string         `gorm:"type:varchar(20);default:'parse'" json:"added_via"`
	AddedAt     time.Time      `gorm:"default:CURRENT_TIMESTAMP;<-:create" json:"added_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
```

- [ ] **Step 2: Enable pg_trgm + GIN index in migration**

In `backend/internal/configs/migration.go`, add two `Exec` calls before `AutoMigrate`:

```go
func AutoMigrate(db *gorm.DB) {
	if db.Migrator().HasTable("users") {
		db.Exec(`ALTER TABLE users DROP CONSTRAINT IF EXISTS idx_users_google_id`)
		db.Exec(`ALTER TABLE users DROP CONSTRAINT IF EXISTS idx_users_email`)
	}

	// Enable trigram extension and GIN index for food suggestions search
	db.Exec(`CREATE EXTENSION IF NOT EXISTS pg_trgm`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_orders_detail_trgm ON orders USING GIN (order_detail gin_trgm_ops) WHERE deleted_at IS NULL`)

	if err := db.AutoMigrate(models.ModelRegistry()...); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}
	log.Println("AutoMigrate completed")
}
```

- [ ] **Step 3: Start server and verify migration**

```bash
cd backend && go run ./cmd/api
```

Expected output includes: `AutoMigrate completed`

In a separate terminal, verify column exists:

```bash
PGPASSWORD=akimusta psql -h localhost -U listmak_user -d listmak_app -c "\d orders" | grep vendor_name
```

Expected: `vendor_name | character varying(100)`

- [ ] **Step 4: Commit**

```bash
git add backend/internal/models/order_model.go backend/internal/configs/migration.go
git commit -m "feat: add vendor_name to orders model, enable pg_trgm extension"
```

⏸️ **STOP — Konfirmasi ke user sebelum Task 2**

---

### Task 2: AI Service (Fireworks.ai)

**Files:**
- Create: `backend/internal/services/ai_service.go`

**Interfaces:**
- Produces:
  - `services.AIService` interface: `ExtractVendor(orderDetail string) (string, error)`
  - `services.NewFireworksAIService(apiKey, model string) AIService`
  - `services.NewNoopAIService() AIService` — used when env key is empty

- [ ] **Step 1: Create `ai_service.go`**

Create `backend/internal/services/ai_service.go` with full content:

```go
package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type AIService interface {
	ExtractVendor(orderDetail string) (string, error)
}

type fireworksAIService struct {
	apiKey string
	model  string
	client *http.Client
}

func NewFireworksAIService(apiKey, model string) AIService {
	return &fireworksAIService{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

type noopAIService struct{}

func NewNoopAIService() AIService { return &noopAIService{} }

func (n *noopAIService) ExtractVendor(_ string) (string, error) { return "", nil }

func (s *fireworksAIService) ExtractVendor(orderDetail string) (string, error) {
	prompt := fmt.Sprintf(
		"Dari pesanan makanan berikut, ekstrak nama resto atau kategori tempatnya. "+
			"Jika ada nama tempat atau nama orang (contoh: 'pak donan', 'warung bu sari') gunakan itu. "+
			"Jika tidak ada nama tempat eksplisit, infer dari jenis makanannya "+
			"(contoh: 'soto daging' → 'Soto', 'dimsum' → 'Dimsum', 'ayam madura' → 'Ayam Madura'). "+
			"Jawab HANYA nama tempatnya saja, tanpa penjelasan, tanpa tanda kutip, singkat. "+
			"Pesanan: %s", orderDetail,
	)

	reqBody, _ := json.Marshal(map[string]interface{}{
		"model": s.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":  30,
		"temperature": 0,
	})

	req, err := http.NewRequest("POST", "https://api.fireworks.ai/inference/v1/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("fireworks API %d: %s", resp.StatusCode, string(b))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 {
		return "", nil
	}

	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd backend && go build ./...
```

Expected: no errors (container.go not yet changed, so `NewOrderService` call there will still compile with old signature — we haven't changed the service yet).

- [ ] **Step 3: Commit**

```bash
git add backend/internal/services/ai_service.go
git commit -m "feat: add Fireworks.ai service for async vendor name extraction"
```

⏸️ **STOP — Konfirmasi ke user sebelum Task 3**

---

### Task 3: Order Repository — UpdateVendorName + GetFoodSuggestions

**Files:**
- Modify: `backend/internal/repository/order_repository.go`

**Interfaces:**
- Produces:
  - `OrderRepository.UpdateVendorName(id uint, vendor string) error`
  - `OrderRepository.GetFoodSuggestions(listmakID uint, query string) ([]string, error)`

- [ ] **Step 1: Add two methods to `OrderRepository` interface**

In `backend/internal/repository/order_repository.go`, update the interface:

```go
type OrderRepository interface {
	GetOrdersByListmakId(listmakId uint, isPaid *bool, search string) ([]models.Order, error)
	GetOrderById(id uint) (models.Order, error)
	CreateOrder(order models.Order) (models.Order, error)
	CreateOrders(orders []models.Order) ([]models.Order, error)
	UpdateOrder(order models.Order) (models.Order, error)
	UpdateOrdersPaidByName(listmakId uint, name string, isPaid bool) (int64, error)
	DeleteOrder(id uint) error
	UpdateVendorName(id uint, vendor string) error
	GetFoodSuggestions(listmakID uint, query string) ([]string, error)
}
```

- [ ] **Step 2: Implement `UpdateVendorName`**

Add at the end of `backend/internal/repository/order_repository.go`:

```go
func (r *orderRepository) UpdateVendorName(id uint, vendor string) error {
	return r.db.Model(&models.Order{}).Where("id = ?", id).Update("vendor_name", vendor).Error
}
```

- [ ] **Step 3: Implement `GetFoodSuggestions`**

Add after `UpdateVendorName`:

```go
func (r *orderRepository) GetFoodSuggestions(listmakID uint, query string) ([]string, error) {
	var results []string

	// Empty query: return popular items from this listmak by frequency
	if query == "" {
		err := r.db.Model(&models.Order{}).
			Select("order_detail").
			Where("listmak_id = ? AND deleted_at IS NULL", listmakID).
			Group("order_detail").
			Order("COUNT(*) DESC").
			Limit(8).
			Pluck("order_detail", &results).Error
		return results, err
	}

	likePattern := "%" + query + "%"

	// Priority 1: same listmak, up to 5 results
	var sameListmak []string
	if err := r.db.Model(&models.Order{}).
		Where("listmak_id = ? AND order_detail ILIKE ? AND deleted_at IS NULL", listmakID, likePattern).
		Distinct("order_detail").
		Order("order_detail").
		Limit(5).
		Pluck("order_detail", &sameListmak).Error; err != nil {
		return nil, err
	}
	results = append(results, sameListmak...)

	// Priority 2: global, pad up to 8 total
	remaining := 8 - len(results)
	if remaining > 0 {
		q := r.db.Model(&models.Order{}).
			Where("order_detail ILIKE ? AND deleted_at IS NULL", likePattern).
			Distinct("order_detail").
			Order("order_detail").
			Limit(remaining)
		if len(sameListmak) > 0 {
			q = q.Where("order_detail NOT IN ?", sameListmak)
		}
		var global []string
		if err := q.Pluck("order_detail", &global).Error; err != nil {
			return nil, err
		}
		results = append(results, global...)
	}

	return results, nil
}
```

- [ ] **Step 4: Verify compilation**

```bash
cd backend && go build ./...
```

Expected: no errors.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/repository/order_repository.go
git commit -m "feat: add UpdateVendorName and GetFoodSuggestions to order repository"
```

⏸️ **STOP — Konfirmasi ke user sebelum Task 4**

---

### Task 4: Order Service — AI Injection + All Service Updates

**Files:**
- Modify: `backend/internal/services/order_service.go`

**Interfaces:**
- Consumes:
  - `services.AIService.ExtractVendor(string) (string, error)` (Task 2)
  - `repository.OrderRepository.UpdateVendorName(uint, string) error` (Task 3)
  - `repository.OrderRepository.GetFoodSuggestions(uint, string) ([]string, error)` (Task 3)
- Produces:
  - `services.NewOrderService(orderRepo, listmakRepo, ai AIService) OrderService` — updated signature
  - `OrderService.GetFoodSuggestions(listmakID uint, query string) ([]string, error)`
  - `OrderService.UpdateVendorName(id uint, vendorName string) error`

- [ ] **Step 1: Update `OrderService` interface**

In `backend/internal/services/order_service.go`, add two methods to the interface:

```go
type OrderService interface {
	GetOrdersByListmakId(listmakId uint, isPaid *bool, search string) ([]models.Order, error)
	CreateOrder(order models.Order) (models.Order, error)
	CreateOrdersBulk(listmakId uint, orders []models.Order) (int, []models.Order, error)
	UpdateOrder(order models.Order) (models.Order, error)
	UpdateOrderPaidStatus(id uint, isPaid bool) (models.Order, error)
	UpdateOrdersPaidByName(listmakId uint, name string, isPaid bool) (int64, error)
	DeleteOrder(id uint) error
	GetFoodSuggestions(listmakID uint, query string) ([]string, error)
	UpdateVendorName(id uint, vendorName string) error
}
```

- [ ] **Step 2: Update `orderService` struct + constructor**

```go
type orderService struct {
	orderRepo   repository.OrderRepository
	listmakRepo repository.ListmakRepository
	ai          AIService
}

func NewOrderService(orderRepo repository.OrderRepository, listmakRepo repository.ListmakRepository, ai AIService) OrderService {
	return &orderService{
		orderRepo:   orderRepo,
		listmakRepo: listmakRepo,
		ai:          ai,
	}
}
```

- [ ] **Step 3: Add private `fillVendorAsync` helper**

Add before `GetOrdersByListmakId`:

```go
func (s *orderService) fillVendorAsync(id uint, orderDetail string) {
	vendor, err := s.ai.ExtractVendor(orderDetail)
	if err != nil || vendor == "" {
		return
	}
	s.orderRepo.UpdateVendorName(id, vendor)
}
```

- [ ] **Step 4: Update `CreateOrder` to spawn goroutine**

```go
func (s *orderService) CreateOrder(order models.Order) (models.Order, error) {
	order.TotalPrice = order.Price * float64(order.Qty)

	newOrder, err := s.orderRepo.CreateOrder(order)
	if err != nil {
		return models.Order{}, err
	}
	s.updateListmakTotals(order.ListmakID)

	go s.fillVendorAsync(newOrder.ID, newOrder.OrderDetail)
	return newOrder, nil
}
```

- [ ] **Step 5: Update `CreateOrdersBulk` to spawn goroutine per order**

```go
func (s *orderService) CreateOrdersBulk(listmakId uint, orders []models.Order) (int, []models.Order, error) {
	for i := range orders {
		orders[i].ListmakID = listmakId
		orders[i].TotalPrice = orders[i].Price * float64(orders[i].Qty)
		if orders[i].AddedVia == "" {
			orders[i].AddedVia = "parse"
		}
	}

	createdOrders, err := s.orderRepo.CreateOrders(orders)
	if err != nil {
		return 0, nil, err
	}

	s.updateListmakTotals(listmakId)

	for _, o := range createdOrders {
		go s.fillVendorAsync(o.ID, o.OrderDetail)
	}
	return len(createdOrders), createdOrders, nil
}
```

- [ ] **Step 6: Add `GetFoodSuggestions` and `UpdateVendorName` service methods**

```go
func (s *orderService) GetFoodSuggestions(listmakID uint, query string) ([]string, error) {
	return s.orderRepo.GetFoodSuggestions(listmakID, query)
}

func (s *orderService) UpdateVendorName(id uint, vendorName string) error {
	return s.orderRepo.UpdateVendorName(id, vendorName)
}
```

- [ ] **Step 7: Verify compilation**

```bash
cd backend && go build ./...
```

Expected: compilation error in `container.go` — `NewOrderService` called with 2 args, now needs 3. This is expected and fixed in Task 7.

- [ ] **Step 8: Commit**

```bash
git add backend/internal/services/order_service.go
git commit -m "feat: inject AIService into order service, add async vendor extraction goroutine"
```

⏸️ **STOP — Konfirmasi ke user sebelum Task 5**

---

### Task 5: Food Suggestions Endpoint

**Files:**
- Modify: `backend/internal/handlers/controllers/share_controller.go`
- Modify: `backend/internal/routes/share_routes.go`

**Interfaces:**
- Consumes: `orderService.GetFoodSuggestions(uint, string) ([]string, error)` (Task 4)
- Produces: `GET /api/v1/share-links/:shareId/food-suggestions?q=<query>` → `{ success: true, data: ["string", ...] }`

- [ ] **Step 1: Add `GetFoodSuggestions` to `ShareController` interface**

In `backend/internal/handlers/controllers/share_controller.go`, add to the interface:

```go
type ShareController interface {
	CreateShareLink(c *gin.Context)
	GetShareLink(c *gin.Context)
	DeleteShareLink(c *gin.Context)
	SubmitOrderViaShare(c *gin.Context)
	GetOrdersViaShare(c *gin.Context)
	CreateViewShare(c *gin.Context)
	GetViewShare(c *gin.Context)
	GetActiveSharesForListmak(c *gin.Context)
	GetFoodSuggestions(c *gin.Context)
}
```

- [ ] **Step 2: Add handler implementation**

Add after `GetOrdersViaShare` method in the same file. Ensure `"strings"` is in the import block (add if not present):

```go
func (sc *shareController) GetFoodSuggestions(c *gin.Context) {
	shareId := c.Param("shareId")

	share, err := sc.shareService.GetShareLink(shareId)
	if err != nil {
		utils.SendResponse(c, http.StatusNotFound, false, "Share link tidak ditemukan", nil)
		return
	}

	query := strings.TrimSpace(c.Query("q"))
	suggestions, err := sc.orderService.GetFoodSuggestions(share.ListmakID, query)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Gagal mengambil saran", nil)
		return
	}

	utils.SendResponse(c, http.StatusOK, true, "OK", suggestions)
}
```

- [ ] **Step 3: Register route**

In `backend/internal/routes/share_routes.go`, add inside the public `shareLinks` block:

```go
shareLinks.GET("/:shareId", sc.GetShareLink)
shareLinks.GET("/:shareId/orders", sc.GetOrdersViaShare)
shareLinks.POST("/:shareId/orders", sc.SubmitOrderViaShare)
shareLinks.GET("/:shareId/food-suggestions", sc.GetFoodSuggestions)
```

- [ ] **Step 4: Verify compilation (container error expected)**

```bash
cd backend && go build ./...
```

Expected: only `container.go` error about `NewOrderService` argument count.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/handlers/controllers/share_controller.go backend/internal/routes/share_routes.go
git commit -m "feat: add GET /share-links/:shareId/food-suggestions endpoint"
```

⏸️ **STOP — Konfirmasi ke user sebelum Task 6**

---

### Task 6: Vendor Override Endpoint

**Files:**
- Modify: `backend/internal/handlers/controllers/order_controller.go`
- Modify: `backend/internal/routes/listmak_routes.go`

**Interfaces:**
- Consumes: `orderService.UpdateVendorName(uint, string) error` (Task 4)
- Produces: `PATCH /api/v1/orders/:id/vendor` → `{ success: true, data: { id, vendor_name } }`

- [ ] **Step 1: Add `UpdateOrderVendor` to `OrderController` interface**

In `backend/internal/handlers/controllers/order_controller.go`:

```go
type OrderController interface {
	GetOrders(c *gin.Context)
	CreateOrder(c *gin.Context)
	CreateOrdersBulk(c *gin.Context)
	UpdateOrder(c *gin.Context)
	UpdateOrderPaid(c *gin.Context)
	UpdateOrdersPaidByName(c *gin.Context)
	DeleteOrder(c *gin.Context)
	UpdateOrderVendor(c *gin.Context)
}
```

- [ ] **Step 2: Add handler implementation**

Add at the end of `order_controller.go`. Ensure `"strings"` is in imports (add if not present):

```go
func (oc *orderController) UpdateOrderVendor(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid order ID", nil)
		return
	}

	var payload struct {
		VendorName string `json:"vendor_name"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid payload", nil)
		return
	}

	vendorName := strings.TrimSpace(payload.VendorName)
	if err := oc.orderService.UpdateVendorName(uint(id), vendorName); err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Gagal update vendor", nil)
		return
	}

	utils.SendResponse(c, http.StatusOK, true, "Vendor berhasil diupdate", gin.H{
		"id":          id,
		"vendor_name": vendorName,
	})
}
```

- [ ] **Step 3: Register route**

In `backend/internal/routes/listmak_routes.go`, add inside `orderGroup`:

```go
orderGroup := r.Group("/orders")
orderGroup.Use(middlewares.AuthMiddleware())
{
    orderGroup.PUT("/:id", oc.UpdateOrder)
    orderGroup.PATCH("/:id/paid", oc.UpdateOrderPaid)
    orderGroup.PATCH("/:id/vendor", oc.UpdateOrderVendor)
    orderGroup.DELETE("/:id", oc.DeleteOrder)
}
```

- [ ] **Step 4: Verify compilation (container error still expected)**

```bash
cd backend && go build ./...
```

Expected: only `container.go` error.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/handlers/controllers/order_controller.go backend/internal/routes/listmak_routes.go
git commit -m "feat: add PATCH /orders/:id/vendor endpoint for manual vendor override"
```

⏸️ **STOP — Konfirmasi ke user sebelum Task 7**

---

### Task 7: Wire DI + Env Vars

**Files:**
- Modify: `backend/internal/handlers/controllers/container.go`
- Modify: `backend/.env`
- Modify: `backend/.env.example`

**Interfaces:**
- Consumes: `services.NewFireworksAIService(apiKey, model string)`, `services.NewNoopAIService()` (Task 2), updated `services.NewOrderService(repo, repo, ai)` (Task 4)
- Produces: fully wired backend, server starts clean with no compilation errors

- [ ] **Step 1: Update `container.go`**

Replace full content of `backend/internal/handlers/controllers/container.go`:

```go
package controllers

import (
	"os"

	"github.com/muhali16/listmak-service/internal/repository"
	"github.com/muhali16/listmak-service/internal/services"
	"gorm.io/gorm"
)

type Container struct {
	UserController    UserController
	AuthController    AuthController
	ListmakController ListmakController
	OrderController   OrderController
	ShareController   ShareController
}

func InitContainer(db *gorm.DB) *Container {
	userRepo := repository.NewUserRepository(db)
	listmakRepo := repository.NewListmakRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	shareRepo := repository.NewShareLinkRepository(db)
	viewShareRepo := repository.NewViewShareRepository(db)

	var aiService services.AIService
	apiKey := os.Getenv("FIREWORKS_API_KEY")
	model := os.Getenv("FIREWORKS_MODEL")
	if apiKey != "" && model != "" {
		aiService = services.NewFireworksAIService(apiKey, model)
	} else {
		aiService = services.NewNoopAIService()
	}

	userService := services.NewUserService(userRepo)
	listmakService := services.NewListmakService(listmakRepo)
	orderService := services.NewOrderService(orderRepo, listmakRepo, aiService)
	shareService := services.NewShareService(shareRepo, viewShareRepo, listmakRepo)

	userController := NewUserController(userService)
	authController := NewAuthController(userService)
	listmakController := NewListmakController(listmakService)
	orderController := NewOrderController(orderService)
	shareController := NewShareController(shareService, orderService)

	return &Container{
		UserController:    userController,
		AuthController:    authController,
		ListmakController: listmakController,
		OrderController:   orderController,
		ShareController:   shareController,
	}
}
```

- [ ] **Step 2: Add env vars to `.env`**

In `backend/.env`, add (use your real Fireworks.ai key):

```
FIREWORKS_API_KEY=your_fireworks_api_key_here
FIREWORKS_MODEL=accounts/fireworks/models/llama-v3p1-8b-instruct
```

- [ ] **Step 3: Add env vars to `.env.example`**

In `backend/.env.example`, add:

```
FIREWORKS_API_KEY=
FIREWORKS_MODEL=accounts/fireworks/models/llama-v3p1-8b-instruct
```

- [ ] **Step 4: Full build**

```bash
cd backend && go build ./...
```

Expected: **no errors**.

- [ ] **Step 5: Start server + test food suggestions**

```bash
cd backend && air
```

Get a valid shareId from DB:

```bash
PGPASSWORD=akimusta psql -h localhost -U listmak_user -d listmak_app -c "SELECT share_id FROM share_links LIMIT 1;"
```

Test suggestions endpoint:

```bash
curl "http://localhost:9001/api/v1/share-links/<shareId>/food-suggestions?q=nasi"
```

Expected: `{"success":true,"data":["Nasi goreng","Nasi goreng pak dodi pedas",...]}`

Test empty query:

```bash
curl "http://localhost:9001/api/v1/share-links/<shareId>/food-suggestions"
```

Expected: popular items array.

- [ ] **Step 6: Test async vendor extraction**

Submit an order via share link:

```bash
curl -X POST "http://localhost:9001/api/v1/share-links/<shareId>/orders" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test AI","order_detail":"Bakso bakar pak donan"}'
```

Wait 5 seconds, then check DB:

```bash
PGPASSWORD=akimusta psql -h localhost -U listmak_user -d listmak_app \
  -c "SELECT id, name, order_detail, vendor_name FROM orders ORDER BY id DESC LIMIT 3;"
```

Expected: `vendor_name` shows `Pak Donan` or similar.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/handlers/controllers/container.go backend/.env.example
git commit -m "feat: wire Fireworks.ai AI service into DI container"
```

⏸️ **STOP — Konfirmasi ke user sebelum Task 8**

---

### Task 8: Frontend — SharedListMakView Autocomplete + Nudge

**Files:**
- Modify: `frontend/src/views/SharedListMakView.vue`
- Modify: `frontend/src/api/share.js`

**Interfaces:**
- Consumes: `GET /share-links/:shareId/food-suggestions?q=X` → `{ data: string[] }` (Task 5)
- Produces: autocomplete dropdown on single-mode order_detail textarea, per-order ⚠️ warning in bulk preview

- [ ] **Step 1: Add `getFoodSuggestions` to share API**

In `frontend/src/api/share.js`, add inside the export default object (after `deleteShareLink`):

```js
getFoodSuggestions(shareId, query = '') {
  const params = query ? `?q=${encodeURIComponent(query)}` : ''
  return apiCall(`/share-links/${shareId}/food-suggestions${params}`, { method: 'GET' })
},
```

- [ ] **Step 2: Add suggestion state to `data()`**

In `SharedListMakView.vue` `data()`, add:

```js
suggestions: [],
showSuggestions: false,
suggestionTimer: null,
orderDetailNudge: false,
```

- [ ] **Step 3: Add suggestion methods**

In the `methods` section, add:

```js
hasLocationHint(detail) {
  const keywords = /\b(pak|bu|mas|mba|mbak|warung|resto|kedai|depot|kantin|gerobak)\b/i
  const dashSeparator = /[-–—]/
  return keywords.test(detail) || dashSeparator.test(detail)
},

async loadSuggestions(query) {
  try {
    const res = await share.getFoodSuggestions(this.shareId, query)
    this.suggestions = res.data || []
    this.showSuggestions = this.suggestions.length > 0
  } catch {
    this.showSuggestions = false
  }
},

onOrderDetailInput() {
  clearTimeout(this.suggestionTimer)
  const val = this.singleForm.order_detail
  this.orderDetailNudge = false
  if (val.length < 2) {
    this.showSuggestions = false
    return
  }
  this.suggestionTimer = setTimeout(() => this.loadSuggestions(val), 300)
},

onOrderDetailFocus() {
  if (this.singleForm.order_detail.length < 2) {
    this.loadSuggestions('')
  }
},

onOrderDetailBlur() {
  setTimeout(() => { this.showSuggestions = false }, 200)
  const val = this.singleForm.order_detail.trim()
  this.orderDetailNudge = val.length > 0 && !this.hasLocationHint(val)
},

selectSuggestion(item) {
  this.singleForm.order_detail = item
  this.showSuggestions = false
  this.orderDetailNudge = !this.hasLocationHint(item)
},
```

- [ ] **Step 4: Add `bulkOrdersWithWarnings` computed**

In the `computed` section of `SharedListMakView.vue`, add:

```js
bulkOrdersWithWarnings() {
  return this.parsedBulkOrders.map(o => ({
    ...o,
    hasLocation: this.hasLocationHint(o.order_detail)
  }))
},
```

- [ ] **Step 5: Update single-mode order_detail template**

Replace the existing single-mode `<div class="form-group">` block containing the order_detail textarea:

```html
<div class="form-group" style="position: relative;">
  <label class="form-label">Pesanan</label>
  <textarea
    v-model="singleForm.order_detail"
    class="form-textarea"
    rows="3"
    placeholder="Contoh: Nasi Ayam Madura Paha - Warung Bu Sari"
    :disabled="submitting"
    @input="onOrderDetailInput"
    @focus="onOrderDetailFocus"
    @blur="onOrderDetailBlur"
  ></textarea>

  <ul v-if="showSuggestions && suggestions.length" class="suggestions-dropdown">
    <li
      v-for="item in suggestions"
      :key="item"
      class="suggestion-item"
      @mousedown.prevent="selectSuggestion(item)"
    >
      {{ item }}
    </li>
  </ul>

  <p v-if="orderDetailNudge" class="nudge-hint">
    <i class="pi pi-info-circle"></i>
    Tulis juga belinya dimana, contoh: Ayam Goreng - Warung Bu Sari
  </p>
</div>
```

- [ ] **Step 6: Update bulk preview to show location warnings**

Find the `<ul class="bulk-preview-list">` block and update to use `bulkOrdersWithWarnings`:

```html
<ul class="bulk-preview-list">
  <li v-for="(o, i) in bulkOrdersWithWarnings" :key="i" class="bulk-preview-item">
    <span class="preview-name">{{ o.name }}</span>
    <span class="preview-detail">{{ o.order_detail }}</span>
    <span v-if="!o.hasLocation" class="location-warning">⚠️ Belum tulis lokasi</span>
  </li>
</ul>
```

Also update the submit button disabled check to use `bulkOrdersWithWarnings.length` instead of `parsedBulkOrders.length` — or keep using `parsedBulkOrders.length` (both work since they have the same length).

- [ ] **Step 7: Add CSS**

In `<style scoped>` of `SharedListMakView.vue`, add:

```css
.suggestions-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: #1e293b;
  border: 1px solid rgba(99, 102, 241, 0.3);
  border-radius: 0.625rem;
  list-style: none;
  margin: 0.25rem 0 0;
  padding: 0.25rem 0;
  z-index: 100;
  max-height: 220px;
  overflow-y: auto;
  box-shadow: 0 8px 24px rgba(0,0,0,0.4);
}

.suggestion-item {
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
  color: #e2e8f0;
  cursor: pointer;
  transition: background 0.1s;
}

.suggestion-item:hover {
  background: rgba(99, 102, 241, 0.15);
}

.nudge-hint {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.75rem;
  color: #94a3b8;
  margin-top: 0.375rem;
}

.nudge-hint i {
  color: #60a5fa;
  font-size: 0.75rem;
  flex-shrink: 0;
}

.location-warning {
  font-size: 0.6875rem;
  color: #f59e0b;
  margin-left: auto;
  flex-shrink: 0;
}
```

- [ ] **Step 8: Test in browser**

```bash
cd frontend && npm run dev
```

Open a valid share link. Test:
1. Tap Pesanan field → dropdown shows popular items
2. Type "nasi" → dropdown filters
3. Tap item → textarea fills
4. Type "ayam goreng" → click outside → nudge appears
5. Type "ayam goreng - warung bu sari" → click outside → no nudge
6. Switch to Banyak sekaligus, type `Budi: Nasi goreng` → preview shows ⚠️ Belum tulis lokasi
7. Type `Ani: Bakso pak donan` → preview shows no warning for Ani

- [ ] **Step 9: Commit**

```bash
git add frontend/src/views/SharedListMakView.vue frontend/src/api/share.js
git commit -m "feat: food autocomplete and location nudge on shared order form"
```

⏸️ **STOP — Konfirmasi ke user sebelum Task 9**

---

### Task 9: Frontend — OrderListView Vendor Grouping + Inline Edit

**Files:**
- Modify: `frontend/src/views/OrderListView.vue`
- Modify: `frontend/src/api/listmak.js`

**Interfaces:**
- Consumes: `orders[].vendor_name string` from existing `GET /listmaks/:id/orders` (field now present from Task 1)
- Consumes: `PATCH /api/v1/orders/:id/vendor` (Task 6)
- Produces: orders grouped by vendor with collapsible section headers, inline vendor edit chip per order

- [ ] **Step 1: Add `updateOrderVendor` to `listmak.js` API**

In `frontend/src/api/listmak.js`, add after `updateOrdersPaidByName`:

```js
updateOrderVendor(orderId, vendorName) {
  return apiCall(`/orders/${orderId}/vendor`, {
    method: 'PATCH',
    body: { vendor_name: vendorName }
  })
},
```

- [ ] **Step 2: Import listmak API in OrderListView**

At the top of `<script>` in `OrderListView.vue`, ensure `listmak` is imported. Find the existing import line (likely `import { listmak } from '../api'` or similar) — it should already be there since the view calls `listmak.getOrders`. If not, add:

```js
import { listmak } from '../api'
```

- [ ] **Step 3: Add vendor edit state to `data()`**

In `OrderListView.vue` `data()`, add:

```js
editingVendorId: null,
editingVendorValue: '',
```

- [ ] **Step 4: Add `ordersGroupedByVendor` and `existingVendors` computed**

Identify the computed property that filters orders (look for a computed that filters by `searchQuery` — it is likely called `filteredOrders` or uses orders directly). Add after existing computeds:

```js
ordersGroupedByVendor() {
  const source = this.filteredOrders ?? this.orders
  const groups = {}
  for (const order of source) {
    const key = order.vendor_name || ''
    if (!groups[key]) groups[key] = []
    groups[key].push(order)
  }
  const named = Object.entries(groups)
    .filter(([k]) => k !== '')
    .sort(([a], [b]) => a.localeCompare(b, 'id'))
  const unnamed = groups[''] ? [['', groups['']]] : []
  return [...named, ...unnamed]
},

existingVendors() {
  return [...new Set((this.orders || []).map(o => o.vendor_name).filter(Boolean))].sort((a, b) =>
    a.localeCompare(b, 'id')
  )
},
```

Note: if the filtered array is named differently (e.g. `filteredOrderGroups`), use the correct name. Check the existing computed section to find what property the template iterates over.

- [ ] **Step 5: Add vendor edit methods**

```js
startEditVendor(order) {
  this.editingVendorId = order.id
  this.editingVendorValue = order.vendor_name || ''
},

cancelEditVendor() {
  this.editingVendorId = null
  this.editingVendorValue = ''
},

async saveVendor(order) {
  const val = this.editingVendorValue.trim()
  try {
    await listmak.updateOrderVendor(order.id, val)
    order.vendor_name = val
    this.editingVendorId = null
  } catch {
    // Keep editing open on failure — user can retry
  }
},
```

- [ ] **Step 6: Update orders template to render vendor groups**

Find the section in `OrderListView.vue` that renders the orders list (the `v-for` over orders). Wrap it in vendor groups. The structure to add around the existing order card:

```html
<template v-for="[vendor, groupOrders] in ordersGroupedByVendor" :key="vendor || '__none__'">
  <!-- Vendor group header -->
  <div class="vendor-group-header">
    <span class="vendor-group-name">
      <i class="pi pi-map-marker"></i>
      {{ vendor || 'Belum dikategorikan' }}
    </span>
    <span class="vendor-group-count">{{ groupOrders.length }}</span>
  </div>

  <!-- Existing order card — now iterates groupOrders instead of all orders -->
  <div v-for="order in groupOrders" :key="order.id" class="order-card">
    <!-- ... all existing order card content unchanged ... -->

    <!-- Add vendor chip at bottom of existing card -->
    <div class="vendor-chip-row">
      <template v-if="editingVendorId === order.id">
        <input
          v-model="editingVendorValue"
          class="vendor-edit-input"
          :list="`vl-${order.id}`"
          placeholder="Nama lokasi beli"
          @keyup.enter="saveVendor(order)"
          @keyup.escape="cancelEditVendor"
        />
        <datalist :id="`vl-${order.id}`">
          <option v-for="v in existingVendors" :key="v" :value="v" />
        </datalist>
        <button class="vendor-save-btn" @click="saveVendor(order)">
          <i class="pi pi-check"></i>
        </button>
        <button class="vendor-cancel-btn" @click="cancelEditVendor">
          <i class="pi pi-times"></i>
        </button>
      </template>
      <button v-else class="vendor-chip" @click="startEditVendor(order)">
        <i class="pi pi-map-marker"></i>
        <span>{{ order.vendor_name || 'Tambah lokasi' }}</span>
        <i class="pi pi-pencil vendor-chip-edit-icon"></i>
      </button>
    </div>
  </div>
</template>
```

- [ ] **Step 7: Add CSS**

In `<style scoped>` of `OrderListView.vue`, add:

```css
.vendor-group-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.5rem 0.25rem;
  margin: 0.75rem 0 0.375rem;
  border-bottom: 1px solid rgba(129, 140, 248, 0.2);
}

.vendor-group-name {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.75rem;
  font-weight: 700;
  color: #818cf8;
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.vendor-group-name i {
  font-size: 0.75rem;
}

.vendor-group-count {
  font-size: 0.6875rem;
  color: #475569;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 999px;
  padding: 0.1rem 0.4rem;
}

.vendor-chip-row {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  margin-top: 0.5rem;
  padding-top: 0.375rem;
  border-top: 1px solid rgba(255, 255, 255, 0.04);
  flex-wrap: wrap;
}

.vendor-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.6875rem;
  color: #64748b;
  background: rgba(255, 255, 255, 0.04);
  border: 1px dashed rgba(255, 255, 255, 0.1);
  border-radius: 999px;
  padding: 0.2rem 0.6rem;
  cursor: pointer;
  transition: color 0.15s, border-color 0.15s;
  font-family: inherit;
}

.vendor-chip:hover {
  color: #94a3b8;
  border-color: rgba(99, 102, 241, 0.35);
}

.vendor-chip-edit-icon {
  font-size: 0.55rem;
  opacity: 0.4;
}

.vendor-edit-input {
  flex: 1;
  min-width: 120px;
  padding: 0.25rem 0.5rem;
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(99, 102, 241, 0.4);
  border-radius: 0.375rem;
  color: #f1f5f9;
  font-size: 0.8125rem;
  font-family: inherit;
}

.vendor-save-btn,
.vendor-cancel-btn {
  width: 1.75rem;
  height: 1.75rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.375rem;
  border: none;
  cursor: pointer;
  font-size: 0.75rem;
  flex-shrink: 0;
}

.vendor-save-btn {
  background: rgba(34, 197, 94, 0.15);
  color: #22c55e;
}

.vendor-cancel-btn {
  background: rgba(255, 255, 255, 0.05);
  color: #64748b;
}
```

- [ ] **Step 8: Test in browser**

```bash
cd frontend && npm run dev
```

Open a listmak with multiple orders (preferably listmak 2 which has varied vendors after AI backfill from Task 7):

1. Orders should be grouped under vendor headers (e.g., "AYAM MADURA", "SOTO", "BELUM DIKATEGORIKAN")
2. Tap any vendor chip → inline input + datalist autocomplete appears
3. Type a vendor name → press Enter → chip updates immediately, order moves to correct group on next page load
4. Press Escape → input closes, no change
5. Vendor with existing names appear in datalist suggestion

- [ ] **Step 9: Commit**

```bash
git add frontend/src/views/OrderListView.vue frontend/src/api/listmak.js
git commit -m "feat: group orders by vendor in OrderListView with inline vendor edit chip"
```

⏸️ **Semua task selesai! Lakukan review akhir bersama user.**

---

## Post-Implementation Notes

**Backfill existing orders:** Orders created before this feature have no `vendor_name`. To backfill, you can call the vendor override endpoint per order manually, or implement a batch endpoint later (see spec Future Considerations).

**Testing AI accuracy:** After deploy, spot-check `vendor_name` values in DB against `order_detail`. Adjust the prompt in `ai_service.go` if results are off.

**Token cost estimate:** At Fireworks.ai Llama 3.1 8B pricing (~$0.20/1M tokens), 1 order extraction ≈ 100 input + 10 output tokens ≈ $0.000022 per order. 1000 orders/day ≈ $0.022/day.
