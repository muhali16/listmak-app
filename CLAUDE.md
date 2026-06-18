# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Listmak is a full-stack collaborative ordering/list management app. The monorepo has two independent sub-projects:

- `backend/` — Go REST API (Gin + GORM + PostgreSQL)
- `frontend/` — Vue 3 SPA (Vite + PrimeVue + TanStack Query)

---

## Backend (`backend/`)

### Running

```bash
# Hot-reload (requires Air installed: go install github.com/air-verse/air@latest)
cd backend && air

# Direct run
cd backend && go run ./cmd/api
```

Server starts on `http://localhost:9001` (set by `PORT` in `.env`).

### Build & Swagger

```bash
# Build binary
cd backend && go build -o listmak-service ./cmd/api

# Regenerate Swagger docs (run after changing handler annotations)
cd backend && swag init -g cmd/api/main.go
```

Swagger UI: `http://localhost:9001/swagger/index.html`

### Environment

Copy `backend/.env.example` to `backend/.env`. Required vars:

```
PORT=9001
ENV=development
DB_HOST / DB_NAME / DB_PORT / DB_USER / DB_PASS  (PostgreSQL)
GOOGLE_CLIENT_ID / GOOGLE_CLIENT_SECRET / GOOGLE_REDIRECT_URL
JWT_SECRET
```

### Architecture

Strict layered pattern — each domain follows the same chain:

```
routes/ → handlers/controllers/ → services/ → repository/ → models/
```

`handlers/controllers/container.go` wires all dependencies (repo → service → controller) and passes the assembled `Container` to routes.

**Dual DB:** Primary data uses PostgreSQL (`configs/database.go`), request logs use a local SQLite file `logs.db` (`models/log_model.go` / `models.InitLogDB()`).

**Response shape:** All API responses go through `pkg/utils/response.go` → `SendResponse()`, which adds `request_id` and `latency` fields automatically from Gin context.

**Middlewares** (applied in `main.go`, order matters):
1. `CORSMiddleware` — must be first to handle preflight
2. `LoggerWithID` — injects `RequestID` and `StartTime` into Gin context

### Key Models

- `Listmak` — a dated order list; has `Orders` (cascade delete) and `ShareLinks`
- `Order` — a line item within a Listmak
- `ShareLink` — short-lived share token (`share_id` varchar 20, expiry-based)
- `ViewShare` — tracks external viewers of shared lists
- `User` — authenticated users (Google OAuth)
- `SystemLog` — request telemetry, stored in SQLite

---

## Frontend (`frontend/`)

### Running

```bash
cd frontend && npm install
cd frontend && npm run dev    # dev server (http://localhost:5173)
cd frontend && npm run build  # production build
```

### Environment

Create `frontend/.env`:

```
VITE_API_BASE_URL=http://localhost:9001/api/v1
```

Defaults to `http://localhost:8080/api/v1` if not set.

### Architecture

- **Router** (`src/main.js`) — routes have `meta: { requiresAuth, guest, hideNav, public }`. Navigation guard calls `verifyAuth()` on first load, then checks localStorage.
- **API layer** (`src/api/`) — `auth.js` handles Google OAuth redirect and localStorage user cache; `listmak.js` and `share.js` are thin fetch wrappers using `credentials: 'include'` for cookie-based JWT.
- **Server state** — TanStack Query (`@tanstack/vue-query`) with 5-minute stale time and 1 retry.
- **UI** — PrimeVue with the Aura preset, dark mode toggled via `.dark-mode` class; Tailwind CSS for layout utilities.
- **Nav** — `Sidebar.vue` (desktop) and `BottomNav.vue` (mobile); hidden on routes with `meta.hideNav`.

### Auth Flow

1. User clicks login → redirected to `GET /api/v1/auth/google/login` (backend handles OAuth dance)
2. Backend sets JWT as httpOnly cookie, redirects to frontend
3. Frontend's `verifyAuth()` calls `GET /api/v1/auth/user` to confirm session; saves user object to `localStorage`
4. Subsequent API calls pass the cookie automatically (`credentials: 'include'`)

### Public Routes

`/listmak/order/:shareId` and `/listmak/view/:viewId` are accessible without login (`meta.public`). These are the shared/view pages backed by `ShareLink` tokens.
