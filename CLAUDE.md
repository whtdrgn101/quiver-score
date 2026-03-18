# QuiverScore Development Notes
This project serves three purposes. The first is I have a love and passion for target archery and want a solid platform to track scores and create a community on a modern and privacy-focused platform. Second, I use this use-case to learn and stretch my technical skills by experimenting with new programming languages, patterns, practices, frameworks, etc. By having a functioning platform with users, that added pressure helps me really learn and understand real-world applications for technology. Finally, this also serves as a mechanism for me to develop my skill at working with AI coding agents and tools and learning how to operate in a new world for developers with my 28 years of experience.

## Architecture
- Python → Go migration in progress (see ROADMAP.md for current phase)
- Go API (chi router) is the public entrypoint, proxies unhandled routes to Python
- Python kept for: Alembic migrations, ReportLab PDF generation
- Both services share the same PostgreSQL database and JWT secret

## Go API (backend-go/)
- Router: chi, handlers in `internal/handler/`, one file per resource
- Each handler follows: struct with DB+Cfg → Routes() method → CRUD methods
- Auth: `middleware.RequireAuth(secretKey)` injects user ID into context
- Direct SQL via pgx (no ORM), UUIDs as strings, COALESCE for partial updates
- Run: `docker compose -f docker-compose.yml -f docker-compose.go.yml up -d`

## Python API (backend/)
- Use `uv` to run all Python commands (e.g., `uv run pytest`, `uv run python`)
- Always run from `backend/` directory for Python/pytest commands
- FastAPI + SQLAlchemy async, Pydantic schemas

## Services (Docker)
- Python API: localhost:8001 (mapped from container 8000)
- Go API: localhost:8080
- Postgres: localhost:5432
- Start both: `docker compose -f docker-compose.yml -f docker-compose.go.yml up -d`
- Rebuild Go after changes: same command with `--build api-go`

## Contract Tests
- Run against Go (avoids Python rate limits on registration):
  `API_BASE_URL=http://localhost:8080 uv run pytest tests/contract/ -v`
- Each test creates its own user via `register_user` fixture (function-scoped)
- New resources need a factory fixture in conftest.py (see `create_sight_mark`)
- Python's register endpoint has 5/min rate limit — always test via Go proxy

## Quality
- All code should have minimum 80% unit test coverage
- All APIs should have automated contract testing (validates migration parity)
- The UI should also have a degree of automated testing around key functionality and workflows

## Migration Workflow (per phase)
1. Write contract tests (pytest) against Python baseline
2. Run tests against Python to establish baseline
3. Build Go handler matching same routes/response shapes
4. Run tests against Go on :8080
5. Deploy, smoke test, mark phase complete in ROADMAP.md
