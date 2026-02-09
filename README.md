# QuiverScore

Target archery score tracking application. Record scores across official round formats, track equipment setups, and monitor your progress over time.

## Features

- **Score Tracking** — Tap-based arrow entry with real-time running totals, X counts, and scorecard display
- **5 Official Round Templates** — WA Indoor 18m (Recurve & Compound), WA 720 70m, Vegas 300, NFAA Indoor 300
- **Equipment Management** — Track bows, risers, limbs, arrows, sights, and accessories with custom specs
- **Setup Profiles** — Group equipment into named setups and link them to scoring sessions
- **Dashboard Stats** — Personal best, average by round type, recent score trend, total arrows/X count
- **User Profiles** — Display name, bio, avatar (file upload or URL), bow type, classification
- **Session Metadata** — Location, weather, and notes on every scoring session

## Tech Stack

| Layer | Technology |
|-------|-----------|
| **API** | Python 3.14, FastAPI, SQLAlchemy 2 (async), Pydantic v2 |
| **Database** | PostgreSQL 16 |
| **Migrations** | Alembic |
| **Auth** | JWT (access + refresh tokens), bcrypt |
| **Frontend** | React 19, React Router 6, Tailwind CSS 4, Vite 7 |
| **HTTP Client** | Axios |
| **Containerization** | Docker, Docker Compose |
| **CI/CD** | GitHub Actions |

## Project Structure

```
quiverscore/
├── backend/
│   ├── app/
│   │   ├── api/v1/          # Route handlers
│   │   │   ├── auth.py      #   Registration, login, refresh, password change
│   │   │   ├── users.py     #   Profile, avatar
│   │   │   ├── rounds.py    #   Round template listing
│   │   │   ├── scoring.py   #   Sessions, ends, arrows, stats
│   │   │   ├── equipment.py #   Equipment CRUD
│   │   │   └── setups.py    #   Setup profiles, equipment linking
│   │   ├── models/          # SQLAlchemy models
│   │   ├── schemas/         # Pydantic request/response schemas
│   │   ├── core/            # Security (JWT, bcrypt), exceptions
│   │   ├── seed/            # Round template seed data
│   │   ├── config.py        # Settings via environment variables
│   │   ├── database.py      # Async engine + session factory
│   │   ├── dependencies.py  # Auth dependency (get_current_user)
│   │   └── main.py          # FastAPI app, CORS, lifespan
│   ├── alembic/             # Database migrations
│   ├── tests/               # pytest async tests (27 tests)
│   ├── Dockerfile
│   └── pyproject.toml
├── frontend/
│   ├── src/
│   │   ├── api/             # Axios API clients
│   │   ├── pages/           # Route components
│   │   │   ├── Dashboard    #   Stats cards, round averages, recent trend
│   │   │   ├── RoundSelect  #   Pick template, setup, location/weather
│   │   │   ├── ScoreSession #   Live scoring with tap-to-score grid
│   │   │   ├── SessionDetail#   Completed session scorecard
│   │   │   ├── History      #   All past sessions
│   │   │   ├── Equipment    #   Equipment list + CRUD
│   │   │   ├── Setups       #   Setup profile management
│   │   │   └── Profile      #   User profile + password change
│   │   ├── contexts/        # Auth context (token management)
│   │   ├── hooks/           # useAuth hook
│   │   └── components/      # Layout shell (nav, sidebar)
│   ├── Dockerfile           # Multi-stage build (Node → Nginx)
│   ├── nginx.conf           # SPA fallback + API reverse proxy
│   └── package.json
├── docker-compose.yml       # Local development
├── docker-compose.prod.yml  # Production (EC2)
├── .env.example             # Required environment variables
└── .github/workflows/
    ├── ci.yml               # Test + lint + build on PR and push
    └── deploy.yml           # SSH deploy to EC2 on merge to main
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/register` | Create account |
| POST | `/api/v1/auth/login` | Get access + refresh tokens |
| POST | `/api/v1/auth/refresh` | Refresh access token |
| POST | `/api/v1/auth/change-password` | Change password (authenticated) |
| GET | `/api/v1/users/me` | Get current user profile |
| PATCH | `/api/v1/users/me` | Update profile fields |
| POST | `/api/v1/users/me/avatar` | Upload avatar image |
| DELETE | `/api/v1/users/me/avatar` | Remove avatar |
| GET | `/api/v1/rounds` | List all round templates |
| GET | `/api/v1/rounds/:id` | Get round template detail |
| POST | `/api/v1/sessions` | Start a scoring session |
| GET | `/api/v1/sessions` | List user's sessions |
| GET | `/api/v1/sessions/stats` | Dashboard statistics |
| GET | `/api/v1/sessions/:id` | Get session detail with ends/arrows |
| POST | `/api/v1/sessions/:id/ends` | Submit an end (arrows) |
| POST | `/api/v1/sessions/:id/complete` | Mark session completed |
| POST | `/api/v1/equipment` | Create equipment item |
| GET | `/api/v1/equipment` | List user's equipment |
| GET | `/api/v1/equipment/:id` | Get equipment detail |
| PUT | `/api/v1/equipment/:id` | Update equipment |
| DELETE | `/api/v1/equipment/:id` | Delete equipment |
| POST | `/api/v1/setups` | Create setup profile |
| GET | `/api/v1/setups` | List setup profiles |
| GET | `/api/v1/setups/:id` | Get setup detail with equipment |
| PUT | `/api/v1/setups/:id` | Update setup profile |
| DELETE | `/api/v1/setups/:id` | Delete setup profile |
| POST | `/api/v1/setups/:id/equipment/:eqId` | Link equipment to setup |
| DELETE | `/api/v1/setups/:id/equipment/:eqId` | Unlink equipment from setup |

## Getting Started

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose
- [uv](https://docs.astral.sh/uv/) (for running backend locally outside Docker)
- [Node.js 22+](https://nodejs.org/) (for running frontend locally)

### Local Development

1. **Clone the repo**

   ```bash
   git clone https://github.com/whtdrgn101/quiver-score.git
   cd quiver-score
   ```

2. **Start the backend** (API + PostgreSQL)

   ```bash
   docker compose up -d
   ```

   The API is now running at `http://localhost:8000`. Round templates are seeded automatically on startup. The backend source is volume-mounted with hot reload enabled.

3. **Start the frontend**

   ```bash
   cd frontend
   npm install
   npm run dev
   ```

   The frontend is now running at `http://localhost:5173` with Vite's dev server proxying API requests.

4. **Run backend tests**

   ```bash
   cd backend
   uv sync
   uv run pytest -x -v
   ```

   Tests use SQLite in-memory so they don't require Docker.

### Database Migrations

Migrations are managed by Alembic and run inside the Docker container:

```bash
# Generate a new migration after model changes
docker compose exec api uv run alembic revision --autogenerate -m "description"

# Apply pending migrations
docker compose exec api uv run alembic upgrade head

# Check current migration version
docker compose exec api uv run alembic current
```

## Production Deployment (EC2 + Docker Compose)

### EC2 Setup

1. Launch an EC2 instance (Amazon Linux 2023 or Ubuntu 22.04, t3.small or larger)
2. Install Docker and Docker Compose:
   ```bash
   # Amazon Linux 2023
   sudo yum install -y docker
   sudo systemctl enable --now docker
   sudo usermod -aG docker $USER
   sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
   sudo chmod +x /usr/local/bin/docker-compose
   ```
3. Clone the repo:
   ```bash
   git clone https://github.com/whtdrgn101/quiver-score.git ~/quiver-score
   ```
4. Create a `.env` file from the example:
   ```bash
   cd ~/quiver-score
   cp .env.example .env
   # Edit .env with real values:
   #   POSTGRES_PASSWORD — strong random password
   #   SECRET_KEY — generate with: python3 -c "import secrets; print(secrets.token_urlsafe(32))"
   #   CORS_ORIGINS — your domain, e.g. ["https://quiverscore.com"]
   ```
5. Start the application:
   ```bash
   docker compose -f docker-compose.prod.yml up -d
   ```
6. Open port 80 (and optionally 443) in the EC2 security group.

### CI/CD Pipeline

Automated deployment is configured via GitHub Actions. On every push to `main`:

1. **CI job** runs backend tests (pytest) and frontend build (lint + vite build)
2. **Deploy job** SSHs into the EC2 instance, pulls latest code, and rebuilds containers

To enable deployment, add these **GitHub repository secrets** (Settings > Secrets and variables > Actions):

| Secret | Description |
|--------|-------------|
| `EC2_HOST` | Public IP or hostname of your EC2 instance |
| `EC2_USER` | SSH username (`ec2-user` for Amazon Linux, `ubuntu` for Ubuntu) |
| `EC2_SSH_KEY` | Private SSH key (the full PEM file contents) |

The deploy job uses `docker compose -f docker-compose.prod.yml` which builds the frontend into an Nginx container that also reverse-proxies `/api/` requests to the backend.

### Stopping to Save Costs

Since this runs on EC2, you can stop the instance from the AWS console when not in use. Your data persists in the Docker volume. Just start the instance and `docker compose -f docker-compose.prod.yml up -d` again.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgresql+asyncpg://...` | Full async database connection string |
| `SECRET_KEY` | `dev-secret-key-...` | JWT signing key — **change in production** |
| `ACCESS_TOKEN_EXPIRE_MINUTES` | `15` | Access token TTL |
| `REFRESH_TOKEN_EXPIRE_DAYS` | `30` | Refresh token TTL |
| `CORS_ORIGINS` | `["http://localhost:5173"]` | Allowed origins (JSON array) |
| `POSTGRES_USER` | — | PostgreSQL username (prod compose) |
| `POSTGRES_PASSWORD` | — | PostgreSQL password (prod compose) |
| `POSTGRES_DB` | — | PostgreSQL database name (prod compose) |

## License

This project is licensed under the Apache License 2.0 — see [LICENSE](LICENSE) for details.
