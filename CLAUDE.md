# QuiverScore Development Notes

This project serves three purposes. The first is I have a love and passion for target archery and want a solid platform to track scores and create a community on a modern and privacy-focused platform. Second, I use this use-case to learn and stretch my technical skills by experimenting with new programming languages, patterns, practices, frameworks, etc. By having a functioning platform with users, that added pressure helps me really learn and understand real-world applications for technology. Finally, this also serves as a mechanism for me to develop my skill at working with AI coding agents and tools and learning how to operate in a new world for developers with my 28 years of experience.

## Guiding Principles

- **Quality first.** Every new feature ships with tests. Untested code is unfinished code.
- **Smooth, low-friction UX.** Users should never fight the interface — whether on web or mobile. Flows should feel intuitive, errors should be clear and actionable, and offline scenarios should be handled gracefully.
- **Long-term maintainability.** This is a product, not a prototype. Favor clear architecture, consistent patterns, and well-documented decisions over clever shortcuts. Code should be easy for a new developer (or buyer) to pick up.
- **Saleable product.** Feature recommendations and technology choices should consider what makes this a viable, transferable product — clean codebase, reliable test suites, standard tooling, and no undocumented tribal knowledge.

## Architecture

- **Go API** (chi router) is the sole runtime service — handles all routes including PDF export
- **React web app** (Vite + React 19) hosted on Firebase as a PWA
- **Flutter mobile app** (iOS + Android) with offline-first architecture and sync engine
- **Python** kept only for Alembic database migrations (runs on deploy then exits, no runtime server)
- **PostgreSQL** (Cloud SQL) is the single source of truth
- PDF generation uses go-pdf/fpdf (pure Go)

## Go API (backend-go/)

- Router: chi, handlers in `internal/handler/`, one file per resource
- Each handler follows: struct with DB+Cfg → Routes() method → CRUD methods
- Auth: `middleware.RequireAuth(secretKey)` injects user ID into context
- SQL via pgx (no ORM), UUIDs as strings, COALESCE for partial updates
- **Repository pattern**: SQL queries live in `internal/repository/`, handlers call repository methods for all database access. This separates HTTP concerns from data access and improves testability.
- Run: `docker compose -f docker-compose.yml -f docker-compose.go.yml up -d`

## Web App (frontend/)

- React 19 + Vite, deployed to Firebase Hosting (global CDN)
- PWA with offline support
- API client in `src/api/`, pages in `src/pages/`, shared hooks in `src/hooks/`

## Mobile App (mobile/)

- Flutter with Riverpod (state), Dio (HTTP), Drift (SQLite)
- Offline-first: all scoring works without connectivity, sync engine handles upload on reconnect
- Dependency-aware sync ordering: session create → ends → images → session complete

## Python (backend/) — Migrations Only

- Use `uv` to run all Python commands (e.g., `uv run alembic upgrade head`)
- Always run from `backend/` directory for Python commands
- Alembic migrations run on startup then the container exits

## Services (Docker)

- Go API: localhost:8080
- Postgres: localhost:5432
- Start: `docker compose -f docker-compose.yml -f docker-compose.go.yml up -d`
- Rebuild Go after changes: same command with `--build api-go`

## Quality & Testing

Testing is non-negotiable. New features must include appropriate test coverage before they are considered complete.

### Go API
- **Unit tests**: minimum 80% coverage per handler/repository. Run with `go test ./...` from `backend-go/`.
- **Contract tests** (pytest + httpx): validate API behavior end-to-end. Run with:
  `API_BASE_URL=http://localhost:8080 uv run pytest tests/contract/ -v`
- Each contract test creates its own user via `register_user` fixture (function-scoped)
- New resources need a factory fixture in conftest.py (see `create_sight_mark`)
- Currently 253+ contract tests covering all endpoints

### Web App
- UI features should have automated tests around key functionality and workflows
- Test the golden path and edge cases in a browser before reporting work as complete
- Monitor for regressions in existing features when making changes

### Mobile App
- Widget and integration tests for key flows (scoring, sync, auth)
- Offline scenarios must be tested — sync engine correctness is critical

### General
- All new API endpoints need both unit tests and contract tests
- Bug fixes should include a regression test when practical
- Don't skip tests to ship faster — test debt compounds
