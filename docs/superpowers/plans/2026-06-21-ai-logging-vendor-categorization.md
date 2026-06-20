# AI Logging & Vendor Categorization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Log every AI call to PostgreSQL, add admin role system with dashboard, and add vendor scan+grouping feature to OrderListView.

**Architecture:** `AILogRepository` is injected into `fireworksAIService` at construction; logging happens inside `ExtractVendor`. Admin access is gated by `AdminOnly` middleware reading the `role` claim already present in JWT. Vendor scan is a new `OrderService.ScanVendors` method called from a new `POST /listmaks/:id/scan-vendors` endpoint.

**Tech Stack:** Go 1.21+, Gin, GORM, PostgreSQL, Vue 3 (Options API), PrimeVue/Aura

## Global Constraints

- All backend files use module path `github.com/muhali16/listmak-service`
- Backend runs on `http://localhost:9001`, frontend dev on `http://localhost:5173`
- `VITE_API_BASE_URL` points to backend; frontend uses `credentials: 'include'` for cookie auth
- JWT already encodes `role` field — no changes to `jwt.go` needed
- `User.Role` field already exists in model — no model change needed
- `AuthMiddleware` already sets `c.Get("role")` from JWT — no middleware change needed
- SQLite (`logs.db`) is for HTTP request logs only; `ai_logs` goes in PostgreSQL (main DB)
- `VendorName string` is the Go field name (json tag: `vendor_name`) in `models.Order`
- Run `cd backend && go build ./...` to verify compilation after each backend task
- Run `cd frontend && npm run build` to verify frontend after each frontend task

---

### Task 1: AILog model + AILogRepository

**Files:**
- Create: `backend/internal/models/ai_log_model.go`
- Modify: `backend/internal/models/registry.go`
- Create: `backend/internal/repository/ai_log_repository.go`

**Interfaces:**
- Produces: `models.AILog` struct; `repository.AILogRepository` interface with `Create(*models.AILog) error` and `GetAll(page, limit int) ([]models.AILog, int64)`

- [ ] **Step 1: Create `backend/internal/models/ai_log_model.go`**

```go
package models

import "time"

type AILog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OrderID   *uint     `gorm:"index" json:"order_id"`
	Input     string    `gorm:"type:text" json:"input"`
	Output    string    `gorm:"type:text" json:"output"`
	Model     string    `json:"model"`
	Provider  string    `json:"provider"`
	LatencyMs int64     `json:"latency_ms"`
	Status    string    `json:"status"`
	ErrorMsg  string    `json:"error_msg"`
	CreatedAt time.Time `json:"created_at"`
}
```

- [ ] **Step 2: Register AILog in `backend/internal/models/registry.go`**

Replace the existing `ModelRegistry` return with:

```go
func ModelRegistry() []interface{} {
	return []interface{}{
		&User{},
		&SystemLog{},
		&Listmak{},
		&Order{},
		&ShareLink{},
		&ViewShare{},
		&AILog{},
	}
}
```

- [ ] **Step 3: Create `backend/internal/repository/ai_log_repository.go`**

```go
package repository

import (
	"github.com/muhali16/listmak-service/internal/models"
	"gorm.io/gorm"
)

type AILogRepository interface {
	Create(log *models.AILog) error
	GetAll(page, limit int) ([]models.AILog, int64)
}

type aiLogRepository struct {
	db *gorm.DB
}

func NewAILogRepository(db *gorm.DB) AILogRepository {
	return &aiLogRepository{db: db}
}

func (r *aiLogRepository) Create(log *models.AILog) error {
	return r.db.Create(log).Error
}

func (r *aiLogRepository) GetAll(page, limit int) ([]models.AILog, int64) {
	var logs []models.AILog
	var total int64
	offset := (page - 1) * limit
	r.db.Model(&models.AILog{}).Count(&total)
	r.db.Order("created_at desc").Offset(offset).Limit(limit).Find(&logs)
	return logs, total
}
```

- [ ] **Step 4: Verify compilation**

```bash
cd backend && go build ./...
```

Expected: no errors. GORM AutoMigrate will create the `ai_logs` table when the server starts.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/models/ai_log_model.go \
        backend/internal/models/registry.go \
        backend/internal/repository/ai_log_repository.go
git commit -m "feat: add AILog model and AILogRepository"
```

---

### Task 2: Update AIService to log calls + fix callers

**Files:**
- Modify: `backend/internal/services/ai_service.go`
- Modify: `backend/internal/services/order_service.go`

**Interfaces:**
- Consumes: `repository.AILogRepository` (from Task 1)
- Produces: `AIService.ExtractVendor(orderDetail string, orderID *uint) (string, error)` — updated signature

- [ ] **Step 1: Rewrite `backend/internal/services/ai_service.go`**

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

	"github.com/muhali16/listmak-service/internal/models"
	"github.com/muhali16/listmak-service/internal/repository"
)

type AIService interface {
	ExtractVendor(orderDetail string, orderID *uint) (string, error)
}

type fireworksAIService struct {
	apiKey  string
	model   string
	client  *http.Client
	logRepo repository.AILogRepository
}

func NewFireworksAIService(apiKey, model string, logRepo repository.AILogRepository) AIService {
	return &fireworksAIService{
		apiKey:  apiKey,
		model:   model,
		client:  &http.Client{Timeout: 10 * time.Second},
		logRepo: logRepo,
	}
}

type noopAIService struct{}

func NewNoopAIService() AIService { return &noopAIService{} }

func (n *noopAIService) ExtractVendor(_ string, _ *uint) (string, error) { return "", nil }

func (s *fireworksAIService) ExtractVendor(orderDetail string, orderID *uint) (string, error) {
	start := time.Now()

	prompt := fmt.Sprintf(
		"Dari pesanan makanan berikut, ekstrak nama resto atau kategori tempatnya. "+
			"Jika ada nama tempat atau nama orang (contoh: 'pak donan', 'warung bu sari') gunakan itu. "+
			"Jika tidak ada nama tempat eksplisit, infer dari jenis makanannya "+
			"(contoh: 'soto daging' → 'Soto', 'dimsum' → 'Dimsum', 'ayam madura' → 'Ayam Madura'). "+
			"Jawab HANYA nama tempatnya saja, tanpa penjelasan, tanpa tanda kutip, singkat. "+
			"Pesanan: %s", orderDetail,
	)

	reqBody, _ := json.Marshal(map[string]any{
		"model": s.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":  30,
		"temperature": 0,
	})

	req, err := http.NewRequest("POST", "https://api.fireworks.ai/inference/v1/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		s.writeLog(orderID, orderDetail, "", time.Since(start).Milliseconds(), "failed", err.Error())
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		s.writeLog(orderID, orderDetail, "", time.Since(start).Milliseconds(), "failed", err.Error())
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("fireworks API %d: %s", resp.StatusCode, string(b))
		s.writeLog(orderID, orderDetail, "", time.Since(start).Milliseconds(), "failed", errMsg)
		return "", fmt.Errorf("%s", errMsg)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.writeLog(orderID, orderDetail, "", time.Since(start).Milliseconds(), "failed", err.Error())
		return "", err
	}
	if len(result.Choices) == 0 {
		s.writeLog(orderID, orderDetail, "", time.Since(start).Milliseconds(), "success", "")
		return "", nil
	}

	output := strings.TrimSpace(result.Choices[0].Message.Content)
	s.writeLog(orderID, orderDetail, output, time.Since(start).Milliseconds(), "success", "")
	return output, nil
}

func (s *fireworksAIService) writeLog(orderID *uint, input, output string, latencyMs int64, status, errMsg string) {
	if s.logRepo == nil {
		return
	}
	s.logRepo.Create(&models.AILog{
		OrderID:   orderID,
		Input:     input,
		Output:    output,
		Model:     s.model,
		Provider:  "fireworks",
		LatencyMs: latencyMs,
		Status:    status,
		ErrorMsg:  errMsg,
	})
}
```

- [ ] **Step 2: Fix `fillVendorAsync` in `backend/internal/services/order_service.go`**

Change the `fillVendorAsync` function from:
```go
func (s *orderService) fillVendorAsync(id uint, orderDetail string) {
	vendor, err := s.ai.ExtractVendor(orderDetail)
```
to:
```go
func (s *orderService) fillVendorAsync(id uint, orderDetail string) {
	vendor, err := s.ai.ExtractVendor(orderDetail, &id)
```

- [ ] **Step 3: Verify compilation**

```bash
cd backend && go build ./...
```

Expected: no errors. Note: container.go will fail until Task 6 since `NewFireworksAIService` now requires 3 args. That's acceptable — fix it in Task 6.

Actually, Task 6 must be done before this compiles cleanly. To avoid breaking build, do Task 6 Step 1 (update container.go) right after this step before committing. Or accept the build break and fix in Task 6. **Recommended: do Task 6 immediately after this task.**

- [ ] **Step 4: Commit**

```bash
git add backend/internal/services/ai_service.go \
        backend/internal/services/order_service.go
git commit -m "feat: add AI call logging to ExtractVendor"
```

---

### Task 3: UpdateRole in UserRepo + UserService + ADMIN_EMAILS in auth

**Files:**
- Modify: `backend/internal/repository/user_repository.go`
- Modify: `backend/internal/services/user_service.go`
- Modify: `backend/internal/handlers/controllers/auth_controller.go`

**Interfaces:**
- Produces: `UserRepository.UpdateRole(id uint, role string) error`; `UserService.UpdateRole(id uint, role string) error`

- [ ] **Step 1: Add `UpdateRole` to `UserRepository` interface and implementation**

In `backend/internal/repository/user_repository.go`, add to the `UserRepository` interface:
```go
UpdateRole(id uint, role string) error
```

Add implementation after `DeleteUser`:
```go
func (ur *userRepository) UpdateRole(id uint, role string) error {
	return ur.db.Model(&models.User{}).Where("id = ?", id).Update("role", role).Error
}
```

- [ ] **Step 2: Add `UpdateRole` to `UserService` interface and implementation**

In `backend/internal/services/user_service.go`, add to the `UserService` interface:
```go
UpdateRole(id uint, role string) error
```

Add implementation after `GetUserById`:
```go
func (s *userService) UpdateRole(id uint, role string) error {
	return s.UserRepository.UpdateRole(id, role)
}
```

- [ ] **Step 3: Update `auth_controller.go` to enforce ADMIN_EMAILS**

Add import `"strings"` if not already present.

Add helper function before `GoogleLogin`:
```go
func isAdminEmail(email string) bool {
	adminEmails := os.Getenv("ADMIN_EMAILS")
	if adminEmails == "" {
		return false
	}
	for _, e := range strings.Split(adminEmails, ",") {
		if strings.EqualFold(strings.TrimSpace(e), email) {
			return true
		}
	}
	return false
}
```

In `GoogleCallback`, change the new-user creation block from:
```go
user, err := ac.UserService.GetUserByGoogleId(googleUser.GoogleID)
if err != nil {
    user = models.User{
        GoogleID: googleUser.GoogleID,
        Email:    googleUser.Email,
        Name:     googleUser.Name,
        Avatar:   googleUser.Picture,
        Role:     "user",
    }
    user, err = ac.UserService.CreateUser(user)
    if err != nil {
        utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to create user", nil)
        return
    }
}
```
to:
```go
user, err := ac.UserService.GetUserByGoogleId(googleUser.GoogleID)
if err != nil {
    role := "user"
    if isAdminEmail(googleUser.Email) {
        role = "admin"
    }
    user = models.User{
        GoogleID: googleUser.GoogleID,
        Email:    googleUser.Email,
        Name:     googleUser.Name,
        Avatar:   googleUser.Picture,
        Role:     role,
    }
    user, err = ac.UserService.CreateUser(user)
    if err != nil {
        utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to create user", nil)
        return
    }
} else {
    // Existing user: enforce ADMIN_EMAILS promotion on every login
    if isAdminEmail(user.Email) && user.Role != "admin" {
        ac.UserService.UpdateRole(user.ID, "admin")
        user.Role = "admin"
    }
}
```

- [ ] **Step 4: Add `ADMIN_EMAILS` to `.env.example`**

In `backend/.env.example`, add:
```
ADMIN_EMAILS=youremail@gmail.com
```

- [ ] **Step 5: Verify compilation**

```bash
cd backend && go build ./...
```

Expected: no errors (container.go still broken from Task 2 — fix in Task 6).

- [ ] **Step 6: Commit**

```bash
git add backend/internal/repository/user_repository.go \
        backend/internal/services/user_service.go \
        backend/internal/handlers/controllers/auth_controller.go \
        backend/.env.example
git commit -m "feat: add UpdateRole and ADMIN_EMAILS auto-promotion on login"
```

---

### Task 4: AdminOnly middleware + AdminController + admin routes

**Files:**
- Create: `backend/internal/handlers/middlewares/admin_middleware.go`
- Create: `backend/internal/handlers/controllers/admin_controller.go`
- Create: `backend/internal/routes/admin_routes.go`
- Modify: `backend/internal/routes/routes.go`
- Modify: `backend/internal/handlers/controllers/container.go` (add AdminController field only — full wiring in Task 6)

**Interfaces:**
- Consumes: `repository.AILogRepository.GetAll`, `repository.UserRepository.UpdateRole` (Tasks 1 & 3)
- Produces: `GET /api/v1/admin/ai-logs`, `PATCH /api/v1/admin/users/:id/role`

- [ ] **Step 1: Create `backend/internal/handlers/middlewares/admin_middleware.go`**

```go
package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/pkg/utils"
)

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			utils.SendResponse(c, http.StatusForbidden, false, "Admin access required", nil)
			c.Abort()
			return
		}
		c.Next()
	}
}
```

- [ ] **Step 2: Create `backend/internal/handlers/controllers/admin_controller.go`**

```go
package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/internal/repository"
	"github.com/muhali16/listmak-service/pkg/utils"
)

type AdminController interface {
	GetAILogs(c *gin.Context)
	UpdateUserRole(c *gin.Context)
}

type adminController struct {
	aiLogRepo repository.AILogRepository
	userRepo  repository.UserRepository
}

func NewAdminController(aiLogRepo repository.AILogRepository, userRepo repository.UserRepository) AdminController {
	return &adminController{aiLogRepo: aiLogRepo, userRepo: userRepo}
}

func (ac *adminController) GetAILogs(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	logs, total := ac.aiLogRepo.GetAll(page, 50)
	utils.SendResponse(c, http.StatusOK, true, "AI logs retrieved", gin.H{
		"logs":  logs,
		"total": total,
		"page":  page,
	})
}

func (ac *adminController) UpdateUserRole(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid user ID", nil)
		return
	}
	var body struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid request body", nil)
		return
	}
	if body.Role != "admin" && body.Role != "user" {
		utils.SendResponse(c, http.StatusBadRequest, false, "Role must be 'admin' or 'user'", nil)
		return
	}
	if err := ac.userRepo.UpdateRole(uint(userID), body.Role); err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to update role", nil)
		return
	}
	utils.SendResponse(c, http.StatusOK, true, "Role updated", nil)
}
```

- [ ] **Step 3: Create `backend/internal/routes/admin_routes.go`**

```go
package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/muhali16/listmak-service/internal/handlers/controllers"
	"github.com/muhali16/listmak-service/internal/handlers/middlewares"
)

func AdminRoutes(r *gin.RouterGroup, ac controllers.AdminController) {
	admin := r.Group("/admin")
	admin.Use(middlewares.AuthMiddleware(), middlewares.AdminOnly())
	{
		admin.GET("/ai-logs", ac.GetAILogs)
		admin.PATCH("/users/:id/role", ac.UpdateUserRole)
	}
}
```

- [ ] **Step 4: Register admin routes in `backend/internal/routes/routes.go`**

Add `AdminRoutes(v1, container.AdminController)` after `LogRoutes(v1)`:

```go
func Routes(r *gin.Engine) {
	container := controllers.InitContainer(configs.GetDB())

	r.NoRoute(func(c *gin.Context) {
		utils.SendResponse(c, http.StatusNotFound, false, "What you looking for?", nil)
	})
	r.HandleMethodNotAllowed = true
	r.NoMethod(func(c *gin.Context) {
		utils.SendResponse(c, http.StatusMethodNotAllowed, false, "Illegal method", nil)
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/api/v1")
	UserRoutes(v1, container.UserController)
	AuthRoute(v1, container.AuthController)
	ListmakRoutes(v1, container.ListmakController, container.OrderController)
	ShareRoutes(v1, container.ShareController)
	LogRoutes(v1)
	AdminRoutes(v1, container.AdminController)
}
```

- [ ] **Step 5: Add `AdminController` field to `Container` struct in `container.go`**

In `backend/internal/handlers/controllers/container.go`, add field to the `Container` struct:

```go
type Container struct {
	UserController    UserController
	AuthController    AuthController
	ListmakController ListmakController
	OrderController   OrderController
	ShareController   ShareController
	AdminController   AdminController
}
```

(Do not wire it yet — that happens in Task 6.)

- [ ] **Step 6: Verify compilation**

```bash
cd backend && go build ./...
```

Expected: error about `container.AdminController` being nil / `NewAdminController` not called yet — that's fine, it'll be wired in Task 6.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/handlers/middlewares/admin_middleware.go \
        backend/internal/handlers/controllers/admin_controller.go \
        backend/internal/routes/admin_routes.go \
        backend/internal/routes/routes.go \
        backend/internal/handlers/controllers/container.go
git commit -m "feat: add AdminOnly middleware, admin controller, and admin routes"
```

---

### Task 5: ScanVendors endpoint

**Files:**
- Modify: `backend/internal/services/order_service.go`
- Modify: `backend/internal/handlers/controllers/order_controller.go`
- Modify: `backend/internal/routes/listmak_routes.go`

**Interfaces:**
- Consumes: `AIService.ExtractVendor(string, *uint)` (Task 2), `OrderRepository.GetOrdersByListmakId`, `OrderRepository.UpdateVendorName`
- Produces: `OrderService.ScanVendors(listmakID uint) ([]models.Order, error)`; `POST /api/v1/listmaks/:id/scan-vendors`

- [ ] **Step 1: Add `ScanVendors` to `OrderService` interface and implementation**

In `backend/internal/services/order_service.go`, add to the `OrderService` interface:
```go
ScanVendors(listmakID uint) ([]models.Order, error)
```

Add implementation before `DeleteOrder`:
```go
func (s *orderService) ScanVendors(listmakID uint) ([]models.Order, error) {
	orders, err := s.orderRepo.GetOrdersByListmakId(listmakID, nil, "")
	if err != nil {
		return nil, err
	}
	for _, o := range orders {
		if o.VendorName != "" {
			continue
		}
		vendor, err := s.ai.ExtractVendor(o.OrderDetail, &o.ID)
		if err != nil || vendor == "" {
			continue
		}
		s.orderRepo.UpdateVendorName(o.ID, vendor)
	}
	return s.orderRepo.GetOrdersByListmakId(listmakID, nil, "")
}
```

- [ ] **Step 2: Add `ScanVendors` to `OrderController` interface and implementation**

In `backend/internal/handlers/controllers/order_controller.go`, add to the `OrderController` interface:
```go
ScanVendors(c *gin.Context)
```

Add implementation after `UpdateOrderVendor`:
```go
func (oc *orderController) ScanVendors(c *gin.Context) {
	listmakId, err := strconv.Atoi(c.Param("id"))
	if err != nil || listmakId <= 0 {
		utils.SendResponse(c, http.StatusBadRequest, false, "Invalid listmak ID", nil)
		return
	}
	orders, err := oc.orderService.ScanVendors(uint(listmakId))
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, false, "Failed to scan vendors", nil)
		return
	}
	utils.SendResponse(c, http.StatusOK, true, "Vendor scan complete", orders)
}
```

- [ ] **Step 3: Register route in `backend/internal/routes/listmak_routes.go`**

Inside the `listmakGroup` block, add after `listmakGroup.PATCH("/:id/orders/paid", oc.UpdateOrdersPaidByName)`:

```go
listmakGroup.POST("/:id/scan-vendors", oc.ScanVendors)
```

- [ ] **Step 4: Verify compilation**

```bash
cd backend && go build ./...
```

Expected: still fails on container.go due to Tasks 2 & 6. Acceptable — wired in next task.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/services/order_service.go \
        backend/internal/handlers/controllers/order_controller.go \
        backend/internal/routes/listmak_routes.go
git commit -m "feat: add ScanVendors endpoint POST /listmaks/:id/scan-vendors"
```

---

### Task 6: Wire container.go (fix compilation)

**Files:**
- Modify: `backend/internal/handlers/controllers/container.go`

**Interfaces:**
- Consumes: `repository.NewAILogRepository` (Task 1), `services.NewFireworksAIService` with 3-arg signature (Task 2), `NewAdminController` (Task 4)

- [ ] **Step 1: Rewrite `backend/internal/handlers/controllers/container.go`**

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
	AdminController   AdminController
}

func InitContainer(db *gorm.DB) *Container {
	// init repositories
	userRepo := repository.NewUserRepository(db)
	listmakRepo := repository.NewListmakRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	shareRepo := repository.NewShareLinkRepository(db)
	viewShareRepo := repository.NewViewShareRepository(db)
	aiLogRepo := repository.NewAILogRepository(db)

	// init AI service
	var aiService services.AIService
	apiKey := os.Getenv("FIREWORKS_API_KEY")
	model := os.Getenv("FIREWORKS_MODEL")
	if apiKey != "" && model != "" {
		aiService = services.NewFireworksAIService(apiKey, model, aiLogRepo)
	} else {
		aiService = services.NewNoopAIService()
	}

	// init services
	userService := services.NewUserService(userRepo)
	listmakService := services.NewListmakService(listmakRepo)
	orderService := services.NewOrderService(orderRepo, listmakRepo, aiService)
	shareService := services.NewShareService(shareRepo, viewShareRepo, listmakRepo)

	// init controllers
	userController := NewUserController(userService)
	authController := NewAuthController(userService)
	listmakController := NewListmakController(listmakService)
	orderController := NewOrderController(orderService)
	shareController := NewShareController(shareService, orderService)
	adminController := NewAdminController(aiLogRepo, userRepo)

	return &Container{
		UserController:    userController,
		AuthController:    authController,
		ListmakController: listmakController,
		OrderController:   orderController,
		ShareController:   shareController,
		AdminController:   adminController,
	}
}
```

- [ ] **Step 2: Verify full compilation**

```bash
cd backend && go build ./...
```

Expected: no errors.

- [ ] **Step 3: Start server and smoke-test**

```bash
cd backend && go run ./cmd/api
```

Then in another terminal:
```bash
# Should return 401
curl -s http://localhost:9001/api/v1/admin/ai-logs | jq .

# Start server, login via browser, then test with cookie
# curl with cookie: curl -s -b "X-User-Authentication-Token=<token>" http://localhost:9001/api/v1/admin/ai-logs | jq .
```

- [ ] **Step 4: Commit**

```bash
git add backend/internal/handlers/controllers/container.go
git commit -m "feat: wire aiLogRepo and AdminController into DI container"
```

---

### Task 7: Frontend admin API + dashboard view

**Files:**
- Create: `frontend/src/api/admin.js`
- Create: `frontend/src/views/AdminAILogsView.vue`
- Modify: `frontend/src/main.js`

**Interfaces:**
- Consumes: `GET /api/v1/admin/ai-logs?page=N`, `PATCH /api/v1/admin/users/:id/role`
- Produces: route `/admin/ai-logs`, guarded by `user.role === 'admin'` check in component

- [ ] **Step 1: Create `frontend/src/api/admin.js`**

```js
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'

async function apiCall(endpoint, options = {}) {
  const url = `${API_BASE_URL}${endpoint}`
  const config = {
    ...options,
    credentials: 'include',
    headers: { 'Content-Type': 'application/json', ...options.headers }
  }
  if (config.body && typeof config.body === 'object') {
    config.body = JSON.stringify(config.body)
  }
  const response = await fetch(url, config)
  const data = await response.json()
  if (!response.ok) throw new Error(data.message || 'API request failed')
  return data
}

export default {
  getAILogs(page = 1) {
    return apiCall(`/admin/ai-logs?page=${page}`)
  },
  updateUserRole(userId, role) {
    return apiCall(`/admin/users/${userId}/role`, {
      method: 'PATCH',
      body: { role }
    })
  }
}
```

- [ ] **Step 2: Create `frontend/src/views/AdminAILogsView.vue`**

```vue
<template>
  <div class="admin-container">
    <div class="admin-header">
      <h1 class="admin-title">AI Request Logs</h1>
      <span class="admin-badge">Admin</span>
    </div>

    <div v-if="loading" class="state-block">
      <i class="pi pi-spin pi-spinner"></i>
      <p>Memuat logs...</p>
    </div>

    <div v-else-if="error" class="state-block state-error">
      <i class="pi pi-exclamation-triangle"></i>
      <p>{{ error }}</p>
      <button class="retry-btn" @click="loadLogs">Coba lagi</button>
    </div>

    <template v-else>
      <div class="stats-row">
        <span class="stats-label">Total: {{ total }} logs</span>
        <span class="stats-label">Halaman {{ page }}</span>
      </div>

      <div class="logs-table-wrap">
        <table class="logs-table">
          <thead>
            <tr>
              <th>Waktu</th>
              <th>Input</th>
              <th>Output</th>
              <th>Provider</th>
              <th>Latency</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="log in logs" :key="log.id">
              <td class="td-time">{{ formatTime(log.created_at) }}</td>
              <td class="td-input" :title="log.input">{{ truncate(log.input, 60) }}</td>
              <td class="td-output">{{ log.output || '—' }}</td>
              <td class="td-provider">{{ log.provider }}</td>
              <td class="td-latency">{{ log.latency_ms }}ms</td>
              <td>
                <span class="status-badge" :class="log.status === 'success' ? 'status-ok' : 'status-err'">
                  {{ log.status }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="logs.length === 0" class="state-block">
        <p>Belum ada AI logs.</p>
      </div>

      <div class="pagination">
        <button class="page-btn" :disabled="page <= 1" @click="changePage(page - 1)">
          <i class="pi pi-chevron-left"></i> Sebelumnya
        </button>
        <button class="page-btn" :disabled="logs.length < 50" @click="changePage(page + 1)">
          Berikutnya <i class="pi pi-chevron-right"></i>
        </button>
      </div>
    </template>
  </div>
</template>

<script>
import admin from '../api/admin'

export default {
  name: 'AdminAILogsView',

  data() {
    return {
      logs: [],
      total: 0,
      page: 1,
      loading: false,
      error: ''
    }
  },

  created() {
    const user = JSON.parse(localStorage.getItem('user') || '{}')
    if (user.role !== 'admin') {
      this.$router.replace('/today')
      return
    }
    this.loadLogs()
  },

  methods: {
    async loadLogs() {
      this.loading = true
      this.error = ''
      try {
        const res = await admin.getAILogs(this.page)
        if (res.success && res.data) {
          this.logs = res.data.logs || []
          this.total = res.data.total || 0
        }
      } catch (err) {
        this.error = err.message || 'Gagal memuat logs.'
      } finally {
        this.loading = false
      }
    },

    changePage(newPage) {
      this.page = newPage
      this.loadLogs()
    },

    formatTime(iso) {
      if (!iso) return '—'
      return new Date(iso).toLocaleString('id-ID', {
        day: '2-digit', month: 'short', year: 'numeric',
        hour: '2-digit', minute: '2-digit', second: '2-digit'
      })
    },

    truncate(str, max) {
      if (!str) return '—'
      return str.length > max ? str.slice(0, max) + '…' : str
    }
  }
}
</script>

<style scoped>
.admin-container {
  padding: 1.5rem;
  max-width: 1100px;
}

.admin-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1.5rem;
}

.admin-title {
  font-size: 1.375rem;
  font-weight: 700;
  color: #f1f5f9;
  margin: 0;
}

.admin-badge {
  padding: 0.2rem 0.6rem;
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 700;
  color: #ef4444;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.stats-row {
  display: flex;
  gap: 1.5rem;
  margin-bottom: 1rem;
}

.stats-label {
  font-size: 0.8125rem;
  color: #64748b;
}

.logs-table-wrap {
  overflow-x: auto;
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 0.75rem;
}

.logs-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.8125rem;
}

.logs-table th {
  padding: 0.75rem 1rem;
  text-align: left;
  font-size: 0.75rem;
  font-weight: 700;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  background: rgba(15, 23, 42, 0.6);
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.logs-table td {
  padding: 0.625rem 1rem;
  color: #cbd5e1;
  border-bottom: 1px solid rgba(255, 255, 255, 0.03);
  vertical-align: top;
}

.logs-table tbody tr:last-child td {
  border-bottom: none;
}

.logs-table tbody tr:hover td {
  background: rgba(255, 255, 255, 0.02);
}

.td-time {
  white-space: nowrap;
  color: #64748b;
  font-size: 0.75rem;
}

.td-input {
  max-width: 280px;
  color: #94a3b8;
  cursor: default;
}

.td-output {
  max-width: 160px;
  color: #e2e8f0;
}

.td-provider {
  white-space: nowrap;
  color: #64748b;
  font-size: 0.75rem;
}

.td-latency {
  white-space: nowrap;
  color: #64748b;
  font-size: 0.75rem;
}

.status-badge {
  display: inline-block;
  padding: 0.15rem 0.5rem;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 600;
}

.status-ok {
  background: rgba(34, 197, 94, 0.12);
  color: #22c55e;
}

.status-err {
  background: rgba(239, 68, 68, 0.12);
  color: #ef4444;
}

.pagination {
  display: flex;
  gap: 0.75rem;
  margin-top: 1.25rem;
  justify-content: flex-end;
}

.page-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.5rem 0.875rem;
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.07);
  border-radius: 0.5rem;
  color: #94a3b8;
  font-size: 0.875rem;
  cursor: pointer;
  transition: background 0.15s;
}

.page-btn:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

.page-btn:not(:disabled):hover {
  background: rgba(30, 41, 59, 0.9);
  color: #cbd5e1;
}

.state-block {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2.5rem 1rem;
  color: #64748b;
  gap: 0.75rem;
}

.state-block > i {
  font-size: 2rem;
  color: #3b82f6;
}

.state-error > i {
  color: #ef4444;
}

.retry-btn {
  padding: 0.5rem 1rem;
  background: rgba(59, 130, 246, 0.15);
  border: 1px solid rgba(59, 130, 246, 0.3);
  border-radius: 0.5rem;
  color: #3b82f6;
  font-size: 0.875rem;
  cursor: pointer;
}
</style>
```

- [ ] **Step 3: Add admin route to `frontend/src/main.js`**

Add import at the top with other view imports:
```js
import AdminAILogsView from './views/AdminAILogsView.vue'
```

Add route to the `routes` array (before the closing bracket):
```js
{
  path: '/admin/ai-logs',
  name: 'AdminAILogs',
  component: AdminAILogsView,
  meta: { requiresAuth: true }
},
```

- [ ] **Step 4: Verify frontend build**

```bash
cd frontend && npm run build
```

Expected: no errors.

- [ ] **Step 5: Test admin dashboard**

Start both servers:
```bash
# Terminal 1
cd backend && go run ./cmd/api
# Terminal 2
cd frontend && npm run dev
```

1. Login with email in `ADMIN_EMAILS`
2. Navigate to `http://localhost:5173/admin/ai-logs`
3. Verify table loads (may be empty if no AI calls yet)
4. Login with non-admin email → navigate to `/admin/ai-logs` → should redirect to `/today`

- [ ] **Step 6: Commit**

```bash
git add frontend/src/api/admin.js \
        frontend/src/views/AdminAILogsView.vue \
        frontend/src/main.js
git commit -m "feat: add admin AI logs dashboard at /admin/ai-logs"
```

---

### Task 8: Frontend OrderListView — vendor scan button + grouping

**Files:**
- Modify: `frontend/src/api/listmak.js`
- Modify: `frontend/src/views/OrderListView.vue`

**Interfaces:**
- Consumes: `POST /api/v1/listmaks/:id/scan-vendors` (Task 5)
- Produces: `groupMode` toggle (`'name'` | `'vendor'`), vendor-grouped card view

- [ ] **Step 1: Add `scanVendors` to `frontend/src/api/listmak.js`**

In the `export default { ... }` block, add after `updateOrderVendor`:

```js
scanVendors(listmakId) {
  return apiCall(`/listmaks/${listmakId}/scan-vendors`, {
    method: 'POST'
  })
},
```

- [ ] **Step 2: Add `groupMode` and `scanningVendors` to data in `OrderListView.vue`**

In the `data()` return object, add after `searchQuery`:
```js
groupMode: 'name',
scanningVendors: false,
scanVendorError: '',
```

- [ ] **Step 3: Add `vendorGroups` and `activeGroups` computed properties**

In the `computed` section, add after `existingVendors`:

```js
vendorGroups() {
  const map = new Map()
  for (const order of this.orders) {
    const key = order.vendor_name || ''
    const display = order.vendor_name || 'Belum ada lokasi'
    if (!map.has(key)) {
      map.set(key, { key, name: display, orders: [] })
    }
    map.get(key).orders.push(order)
  }
  return Array.from(map.values()).sort((a, b) => {
    if (!a.key) return 1
    if (!b.key) return -1
    return a.name.localeCompare(b.name, 'id')
  })
},

vendorFilteredGroups() {
  if (!this.searchQuery.trim()) return this.vendorGroups
  const q = this.searchQuery.trim().toLowerCase()
  return this.vendorGroups.filter(
    g =>
      g.name.toLowerCase().includes(q) ||
      g.orders.some(o => o.order_detail.toLowerCase().includes(q))
  )
},
```

- [ ] **Step 4: Add `scanVendors` method**

In the `methods` section, add after `cancelEditVendor`:

```js
async scanVendors() {
  if (this.groupMode === 'vendor' && !this.scanningVendors) {
    this.groupMode = 'name'
    return
  }
  this.scanningVendors = true
  this.scanVendorError = ''
  try {
    const res = await listmak.scanVendors(this.listmakId)
    if (res.success && res.data) {
      this.orders = Array.isArray(res.data) ? res.data : []
    }
    this.groupMode = 'vendor'
  } catch (err) {
    this.scanVendorError = err.message || 'Gagal scan vendor.'
    this.$toast.add({
      severity: 'error',
      summary: 'Gagal scan lokasi',
      detail: this.scanVendorError,
      life: 3000
    })
  } finally {
    this.scanningVendors = false
  }
},
```

- [ ] **Step 5: Add vendor scan button to template**

In the template, add between the `summary-card` div and the `add-btn` button (before line `<!-- Add order — full-width, explicit label -->`):

```html
<!-- Vendor scan / group toggle -->
<button
  class="vendor-scan-btn"
  :class="{ 'vendor-scan-btn--active': groupMode === 'vendor' }"
  :disabled="scanningVendors"
  @click="scanVendors"
>
  <i v-if="scanningVendors" class="pi pi-spin pi-spinner"></i>
  <i v-else class="pi pi-map-marker"></i>
  <span>{{
    scanningVendors
      ? 'Mendeteksi lokasi...'
      : groupMode === 'vendor'
      ? 'Kembali ke nama pemesan'
      : 'Kelompokkan per lokasi'
  }}</span>
</button>
```

- [ ] **Step 6: Update the groups list section to support vendor mode**

In the template, find the `<!-- Grouped list -->` section. Change the `v-for` on `<article>` from:
```html
<article
  v-for="group in visibleGroups"
  :key="group.key"
  class="group-card"
>
```
to:
```html
<article
  v-for="group in groupMode === 'name' ? visibleGroups : vendorFilteredGroups"
  :key="group.key"
  class="group-card"
>
```

On the paid toggle button inside the group card, add `v-if="groupMode === 'name'"` so it only shows in name mode:
```html
<button
  v-if="groupMode === 'name'"
  class="paid-btn"
  ...
>
```

Also update the empty state check from:
```html
<div
  v-if="filteredGroups.length === 0"
  class="state-block empty"
>
```
to:
```html
<div
  v-if="(groupMode === 'name' ? filteredGroups : vendorFilteredGroups).length === 0"
  class="state-block empty"
>
```

And update "Muat lebih banyak" button — only show in name mode (vendor mode shows all):
```html
<button
  v-if="groupMode === 'name' && hasMore"
  class="load-more-btn"
  @click="loadMore"
>
```

- [ ] **Step 7: Add vendor scan button styles**

In the `<style scoped>` block, add after `.add-btn` styles:

```css
/* Vendor scan toggle button */
.vendor-scan-btn {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  margin-bottom: 0.75rem;
  background: rgba(99, 102, 241, 0.1);
  border: 1px solid rgba(99, 102, 241, 0.25);
  border-radius: 0.875rem;
  color: #818cf8;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s;
}

.vendor-scan-btn:hover:not(:disabled) {
  background: rgba(99, 102, 241, 0.18);
  border-color: rgba(99, 102, 241, 0.4);
}

.vendor-scan-btn--active {
  background: rgba(99, 102, 241, 0.18);
  border-color: rgba(99, 102, 241, 0.45);
}

.vendor-scan-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
```

- [ ] **Step 8: Verify frontend build**

```bash
cd frontend && npm run build
```

Expected: no errors.

- [ ] **Step 9: Manual test**

1. Open `http://localhost:5173/listmak/<id>` for a listmak with orders
2. Click "Kelompokkan per lokasi"
3. Verify loading spinner shows while AI scans
4. Verify orders regroup by vendor (or "Belum ada lokasi" for uncategorized)
5. Verify paid toggles are hidden in vendor mode
6. Click button again → verify back to name grouping
7. Open `http://localhost:9001/api/v1/admin/ai-logs` (with admin cookie) → verify new log entries appeared

- [ ] **Step 10: Commit**

```bash
git add frontend/src/api/listmak.js \
        frontend/src/views/OrderListView.vue
git commit -m "feat: add vendor scan button and vendor-grouped view to OrderListView"
```

---

## Self-Review

### Spec Coverage

| Spec Requirement | Task |
|---|---|
| `ai_logs` table in PostgreSQL | Task 1 |
| `AILogRepository.Create` + logging in `ExtractVendor` | Tasks 1, 2 |
| `ADMIN_EMAILS` env auto-promotion on login | Task 3 |
| Admin manual role promotion via `PATCH /admin/users/:id/role` | Task 4 |
| `AdminOnly` middleware (403 for non-admin) | Task 4 |
| `GET /api/v1/admin/ai-logs` paginated | Task 4 |
| `POST /listmaks/:id/scan-vendors` (skip existing vendor) | Task 5 |
| Wire aiLogRepo + AdminController into DI container | Task 6 |
| Frontend admin dashboard `/admin/ai-logs` | Task 7 |
| Admin role guard in frontend component | Task 7 |
| Vendor scan button on OrderListView | Task 8 |
| Vendor-grouped computed view | Task 8 |
| No paid toggle in vendor mode | Task 8 |
| `noopAIService` updated signature | Task 2 |
| `fillVendorAsync` caller updated | Task 2 |

### Placeholder Scan
None found — all steps have complete code.

### Type Consistency
- `AILogRepository` defined in Task 1, consumed in Tasks 2, 4, 6 ✓
- `ExtractVendor(string, *uint)` defined in Task 2, called in Tasks 2 (fillVendorAsync) and 5 (ScanVendors) ✓
- `UpdateRole(uint, string)` defined in Task 3, called in auth_controller (Task 3) and admin_controller (Task 4) ✓
- `Container.AdminController` field added in Task 4 Step 5, populated in Task 6 ✓
- `OrderController.ScanVendors` interface + implementation in Task 5, route wired in Task 5 ✓
- `listmak.scanVendors(id)` defined in Task 8 Step 1, called in Task 8 Step 4 ✓
