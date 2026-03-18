# QuiverScore: Python → Go API Migration Roadmap

## Overview

Migrate the FastAPI backend to Go while keeping Python for PDF generation and Alembic migrations. This is an incremental migration — each phase follows the same workflow and both services coexist during the transition.

## Architecture

```
                    ┌─────────────────────────┐
                    │      Cloud Run          │
                    │                         │
  HTTP ──────────▶  │  Go API (primary)       │
                    │    ├── all routes        │
                    │    └── calls Python ───▶ │  Python sidecar
                    │         for PDF export   │    ├── PDF generation
                    │                         │    └── Alembic migrations
                    └─────────────────────────┘
                               │
                               ▼
                          PostgreSQL
```

**Go stack:** `chi` router, `sqlc` (type-safe SQL), `golang-migrate`, `golang-jwt`, `bcrypt`
**Kept in Python:** Alembic migrations, ReportLab PDF export

## Workflow (per phase)

1. **Write API contract tests** (pytest + httpx) against the current Python endpoint
2. **Run tests against Python** to establish the baseline
3. **Build the Go endpoint** matching the same routes and response shapes
4. **Run tests against Go** on a different port, compare with Python baseline
5. **Deploy** to Cloud Run
6. **Run smoke tests** against production URL
7. **Mark phase complete** below

---

## Phase 0: Foundation

### 0.1 — API Contract Test Infrastructure ✅
- [x] Create `tests/contract/` directory with shared fixtures
- [x] Configure pytest to run against a live server URL (env var `API_BASE_URL`)
- [x] Helper utilities: auth token acquisition, test user creation, cleanup
- [x] Verify tests pass against running Python API (25/25 passing)

### 0.2 — Go Project Scaffold ✅
- [x] Initialize Go module at `backend-go/`
- [x] Set up `chi` router with health check endpoint
- [x] Database connection pool (`pgxpool`) reading same `DATABASE_URL`
- [x] Dockerfile for the Go service (multi-stage, ~15MB final image)
- [x] docker-compose overlay (`docker-compose.go.yml`) to run Go on :8080 alongside Python on :8000
- [x] Verify `/health` returns 200 on both services

### 0.3 — Shared Auth Middleware (Go) ✅
- [x] JWT validation middleware (HS256, same `SECRET_KEY`)
- [x] `get_current_user` equivalent — extract user ID from token
- [x] Optional auth middleware for public endpoints
- [x] Bcrypt password hashing/verification
- [x] Token creation (access, refresh, reset, email verification)
- [x] Cross-compatibility verified: Python-generated JWTs decode in Go

---

## Phase 1: Auth Endpoints

### 1.1 — Contract Tests for Auth ✅
- [x] 24 contract tests covering all auth endpoints (written in Phase 0.1)

### 1.2 — Go Implementation: Auth ✅
- [x] All auth endpoints in Go (register, login, refresh, verify-email, resend-verification, change-password, forgot-password, reset-password, delete-account)
- [x] GET /api/v1/users/me implemented
- [x] Full account deletion cascade matching Python behavior
- [x] 24/24 contract tests passing against Go on :8080
- [x] 24/24 contract tests still passing against Python on :8000
- [x] Rate limiting middleware (token bucket per IP on auth routes, configurable via RATE_LIMIT_ENABLED)
- [x] SendGrid email integration (verification, password reset emails via SendGrid v3 API)

### 1.3 — Deploy Auth ✅
- [x] Reverse proxy: Go handles auth natively, proxies all other routes to Python
- [x] Cloud Run service-to-service auth (ID token via metadata server)
- [x] CI updated: Go tests run alongside Python tests
- [x] Deploy pipeline updated: builds both images, deploys Python as internal, Go as public
- [x] IAM: Go service granted `roles/run.invoker` on Python service
- [x] Local verification: 25/25 contract tests pass through Go proxy
- [x] Commit, push, and verify deploy succeeds
- [x] Smoke tests pass against production (25/25)
- [x] Verified working

---

## Phase 2: Rounds (Read-Heavy, Simple CRUD)

### 2.1 — Contract Tests for Rounds ✅
- [x] GET/POST/PUT/DELETE round templates
- [x] Round sharing endpoints
- [x] 12 contract tests covering list, create, get, update, delete (+ auth/validation edge cases)

### 2.2 — Go Implementation: Rounds ✅
- [x] Direct SQL queries for round templates and stages (no sqlc needed)
- [x] All round CRUD endpoints (list, create, get, update, delete, share, unshare)
- [x] Ownership and official-round permission checks
- [x] In-progress session guard on update
- [x] 12/12 contract tests passing against Go on :8080
- [x] 36/36 total contract tests passing (12 rounds + 24 auth)

### 2.3 — Deploy Rounds ✅
- [x] Deploy succeeded (CI + deploy workflow green)
- [x] Smoke tests pass against production (37/37 — 12 rounds + 24 auth + 1 health)

---

## Phase 3: Equipment & Setups

### 3.1 — Contract Tests for Equipment & Setups ✅
- [x] Equipment CRUD (list, create, get, update, delete, stats) — 17 tests
- [x] Setup profiles with equipment linking (CRUD + add/remove equipment) — 19 tests
- [x] 36/36 contract tests passing against Python baseline

### 3.2 — Go Implementation: Equipment & Setups ✅
- [x] Equipment handler: full CRUD + stats endpoint with usage aggregation
- [x] Setups handler: full CRUD + equipment linking/unlinking
- [x] Partial update support (COALESCE pattern) for both resources
- [x] 36/36 contract tests passing against Go on :8080
- [x] 73/73 total contract tests passing (24 auth + 12 rounds + 17 equipment + 19 setups + 1 health)

### 3.3 — Deploy Equipment & Setups ✅
- [x] Deploy succeeded (CI + deploy workflow green)
- [x] Smoke tests pass against production (73/73)

---

## Phase 4: Sight Marks & Classifications

### 4.1 — Contract Tests ✅
- [x] Sight mark CRUD (list, create, update, delete + filters + auth) — 17 tests
- [x] Classification records (list all, current best, auth) — 4 tests
- [x] 21/21 contract tests passing against Python baseline

### 4.2 — Go Implementation ✅
- [x] Sight marks handler: full CRUD with equipment/setup filter support
- [x] Classifications handler: list all records + current best per system/round_type
- [x] Registered routes in main.go, users/me subrouter for classifications
- [x] 21/21 contract tests passing against Go on :8080
- [x] 94/94 total contract tests passing (24 auth + 12 rounds + 17 equipment + 19 setups + 17 sight marks + 4 classifications + 1 health)

### 4.3 — Deploy ✅
- [x] Deploy succeeded, smoke tests pass against production

---

## Phase 5: Scoring Sessions (Core Domain)

### 5.1 — Contract Tests for Scoring ✅
- [x] Session lifecycle: create, list, get, submit ends, complete, abandon, delete — 33 tests
- [x] Stats, trends, personal records — 9 tests
- [x] Undo last end — 3 tests
- [x] CSV export (single + bulk) — 5 tests
- [x] PDF export — 1 test
- [x] Session sharing (create, view, revoke) — 7 tests (under /api/v1/share)
- [x] 54/54 contract tests passing against Python baseline
- [x] 148/148 total contract tests passing

### 5.2 — Go Implementation: Scoring ✅
- [x] Session CRUD and scoring logic
- [x] Stats/trends queries
- [x] CSV export in Go
- [x] HTTP call to Python sidecar for PDF
- [x] Session sharing (create, view, revoke share links)
- [x] 148/148 contract tests passing against Go on :8080

### 5.3 — Deploy Scoring ✅
- [x] Deploy succeeded, smoke tests pass against production

---

## Phase 6: Users & Sharing

### 6.1 — Contract Tests ✅
- [x] User profile update (PATCH /users/me) — 4 tests
- [x] Avatar upload, delete — 5 tests (file upload, invalid type, delete, unauthenticated)
- [x] Public profile (GET /users/{username}) — 4 tests (public, private, not found, no auth required)
- [x] 13/13 contract tests passing against Python baseline
- [x] Session sharing already covered in Phase 5.1

### 6.2 — Go Implementation ✅
- [x] User profile update (PATCH) with partial field support
- [x] Avatar upload (multipart file), avatar-from-URL, avatar delete
- [x] Public profile with stats, personal best, recent sessions, clubs
- [x] Repository methods in internal/repository/user.go
- [x] 13/13 contract tests passing against Go on :8080
- [x] 161/161 total contract tests passing

### 6.3 — Deploy ✅
- [x] Deploy succeeded, smoke tests pass against production

---

## Phase 7: Clubs

### 7.1 — Contract Tests ✅
- [x] Club CRUD (create, list, get, update, delete) — 11 tests
- [x] Invites (create, list, preview, join, deactivate) — 8 tests
- [x] Member management (promote, demote, remove, leave) — 4 tests
- [x] Leaderboard and activity feed — 4 tests
- [x] Events with RSVP (CRUD + participants) — 8 tests
- [x] Teams (CRUD + member management) — 8 tests
- [x] Shared rounds — 1 test
- [x] Tournaments (create, register, start, score, leaderboard, complete, withdraw) — 10 tests
- [x] 54/54 contract tests passing against Python baseline

### 7.2 — Go Implementation ✅
- [x] Club handler: full CRUD + invite management + member RBAC (owner/admin/member)
- [x] Events handler: CRUD + RSVP upsert with DB constraint compliance
- [x] Teams handler: CRUD + member add/remove with team leader permissions
- [x] Tournaments handler: full lifecycle (create → register → start → score → complete)
- [x] Leaderboard, activity feed, shared rounds
- [x] Repository pattern: all SQL in `internal/repository/club.go`
- [x] 54/54 contract tests passing against Go on :8080
- [x] 215/215 total contract tests passing

### 7.3 — Deploy ✅
- [x] Deploy succeeded, smoke tests pass against production

---

## Phase 8: Social & Coaching

### 8.1 — Contract Tests ✅
- [x] Follow/unfollow, list followers/following — 8 tests
- [x] Activity feed (empty, unauthenticated) — 2 tests
- [x] Coaching invites (create, accept, reject, errors) — 8 tests
- [x] Coach athlete session viewing — 2 tests
- [x] Session annotations (create, list, authorization) — 7 tests
- [x] 27/27 contract tests passing against Python baseline

### 8.2 — Go Implementation ✅
- [x] Social handler: follow/unfollow, list followers/following, activity feed
- [x] Coaching handler: invite, respond, list athletes/coaches, view sessions, annotations
- [x] Repository pattern: `internal/repository/social.go` and `internal/repository/coaching.go`
- [x] Shared sentinel errors in `internal/repository/errors.go`
- [x] 27/27 contract tests passing against Go on :8080
- [x] 242/242 total contract tests passing

### 8.3 — Deploy ✅
- [x] Deploy succeeded, smoke tests pass against production

---

## Phase 9: Notifications

### 9.1 — Contract Tests ✅
- [x] List notifications (empty, with items after PR, unauthenticated) — 3 tests
- [x] Unread count (zero, after PR, unauthenticated) — 3 tests
- [x] Mark single notification read (success, not found, unauthenticated) — 3 tests
- [x] Mark all notifications read (success, unauthenticated) — 2 tests
- [x] 11/11 contract tests passing against Python baseline

### 9.2 — Go Implementation ✅
- [x] Notification handler: list, unread count, mark read, mark all read
- [x] Repository pattern: `internal/repository/notification.go`
- [x] Fixed `is_read` → `read` column bug in scoring repo's `InsertNotification`
- [x] 11/11 contract tests passing against Go on :8080
- [x] 253/253 total contract tests passing

### 9.3 — Deploy ✅
- [x] Deploy succeeded, smoke tests pass against production

---

## Phase 10: Cleanup & Cutover ✅

- [x] Strip Python to PDF sidecar only (single export endpoint + Alembic migrations)
- [x] Remove all Python route handlers (auth, users, rounds, scoring, equipment, setups, clubs, sight marks, notifications, classifications, coaching, social, sharing)
- [x] Remove Python unit tests for migrated endpoints (contract tests are the test suite now)
- [x] Remove CORS middleware and rate limiting from Python sidecar (Go handles these)
- [x] Update CI pipeline: removed Python backend-test job, Go tests + frontend build only
- [x] Update deploy pipeline: renamed CI dependency
- [x] Update docker-compose files with sidecar documentation
- [x] 253/253 contract tests passing locally (including PDF export via Go→Python proxy)
- [x] Deploy, production smoke test, verify

---

## Current Status

**Migration complete.** Go API handles all routes. Python retained as sidecar for PDF export and Alembic migrations only.

### Post-Migration Hardening
- [x] Python sidecar locked down: `--ingress internal` + `--no-allow-unauthenticated`
- [x] Go API: `--min-instances 1` to eliminate cold starts
- [x] Rate limiting on auth routes (in-memory token bucket, 10 req/min per IP)
- [x] SendGrid email integration restored (verification + password reset)
- [x] Deploy pipeline passes `SENDGRID_API_KEY`, `SENDGRID_FROM_EMAIL`, `FRONTEND_URL` to Go service
