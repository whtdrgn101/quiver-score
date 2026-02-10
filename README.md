# QuiverScore

Target archery score tracking application. Record scores across official round formats, track equipment setups, and monitor your progress over time. Installable as a PWA on mobile devices.

## Features

- **Score Tracking** — Tap-based arrow entry with real-time running totals, X counts, and scorecard display
- **6 Official Round Templates** — WA Indoor 18m (Recurve & Compound), WA 720 70m, WA 1440 (multi-stage), Vegas 300, NFAA Indoor 300
- **Multi-Stage Rounds** — Dynamic stage progression for rounds like WA 1440 with multiple distances
- **Equipment Management** — Track bows, risers, limbs, arrows, sights, and accessories with custom specs
- **Setup Profiles** — Group equipment into named setups and link them to scoring sessions
- **Dashboard Stats** — Personal best, average by round type, recent score trend, total arrows/X count
- **Session Resume** — Banner notification to resume in-progress scoring sessions
- **User Profiles** — Display name, bio, avatar (file upload or URL), bow type, classification
- **Session Metadata** — Location, weather, and notes on every scoring session
- **PWA Support** — Install on mobile, offline app shell caching, auto-updating service worker

## Tech Stack

| Layer | Technology |
|-------|-----------|
| **API** | Python 3.14, FastAPI, SQLAlchemy 2 (async), Pydantic v2 |
| **Database** | PostgreSQL 16 |
| **Migrations** | Alembic |
| **Auth** | JWT (access + refresh tokens), bcrypt |
| **Frontend** | React 19, React Router 6, Tailwind CSS 4, Vite 7 |
| **PWA** | vite-plugin-pwa, Workbox |
| **HTTP Client** | Axios |
| **Containerization** | Docker, Docker Compose |
| **CI/CD** | GitHub Actions |
| **Hosting** | Google Cloud Run (scale-to-zero) |

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
│   ├── tests/               # pytest async tests
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
│   │   └── components/      # Layout shell (nav, resume banner)
│   ├── Dockerfile           # Multi-stage build (Node → Nginx)
│   ├── nginx.conf           # SPA fallback + API proxy (local compose)
│   ├── nginx.conf.template  # Parameterized config (Cloud Run)
│   └── package.json
├── docker-compose.yml       # Local development
├── docker-compose.prod.yml  # Local prod testing
├── .env.example             # Required environment variables
└── .github/workflows/
    ├── ci.yml               # Test + lint + build on PR and push
    └── deploy.yml           # Build + deploy to Cloud Run on merge to main
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

## Production Deployment (Google Cloud Run)

Cloud Run provides scale-to-zero hosting — **$0 when idle**.

### GCP Setup

1. **Create a GCP project** and enable the Cloud Run and Artifact Registry APIs:

   ```bash
   gcloud services enable run.googleapis.com artifactregistry.googleapis.com
   ```

2. **Create an Artifact Registry repository**:

   ```bash
   gcloud artifacts repositories create quiverscore \
     --repository-format=docker \
     --location=us-central1
   ```

3. **Create a service account** for GitHub Actions deployment:

   ```bash
   gcloud iam service-accounts create github-deploy
   # Grant necessary roles
   gcloud projects add-iam-policy-binding $PROJECT_ID \
     --member="serviceAccount:github-deploy@$PROJECT_ID.iam.gserviceaccount.com" \
     --role="roles/run.admin"
   gcloud projects add-iam-policy-binding $PROJECT_ID \
     --member="serviceAccount:github-deploy@$PROJECT_ID.iam.gserviceaccount.com" \
     --role="roles/artifactregistry.writer"
   gcloud projects add-iam-policy-binding $PROJECT_ID \
     --member="serviceAccount:github-deploy@$PROJECT_ID.iam.gserviceaccount.com" \
     --role="roles/iam.serviceAccountUser"
   ```

4. **Export a service account key** and add it as a GitHub secret:

   ```bash
   gcloud iam service-accounts keys create key.json \
     --iam-account=github-deploy@$PROJECT_ID.iam.gserviceaccount.com
   # Add contents of key.json as GCP_SA_KEY secret in GitHub
   ```

### Database

Use a managed PostgreSQL service for the database:

- **[Neon](https://neon.tech)** — Free tier with generous limits, scale-to-zero
- **[Supabase](https://supabase.com)** — Free tier PostgreSQL with dashboard

Set the `DATABASE_URL` secret to the async connection string:
```
postgresql+asyncpg://user:pass@host:5432/dbname
```

### GitHub Secrets

Add these secrets in **Settings > Secrets and variables > Actions**:

| Secret | Description |
|--------|-------------|
| `GCP_PROJECT_ID` | Your GCP project ID |
| `GCP_REGION` | Deployment region (e.g., `us-central1`) |
| `GCP_SA_KEY` | Service account key JSON |
| `DATABASE_URL` | PostgreSQL async connection string |
| `SECRET_KEY` | JWT signing key — generate with `python3 -c "import secrets; print(secrets.token_urlsafe(32))"` |
| `CORS_ORIGINS` | Allowed origins JSON array (e.g., `["https://quiverscore-frontend-xxx.run.app"]`) |

### Deploying

Deployment is fully automated. Push to `main` and GitHub Actions will:

1. Run CI tests
2. Build and push Docker images to Artifact Registry
3. Deploy API and frontend services to Cloud Run

### Local Prod Testing

To test the production Docker setup locally:

```bash
cp .env.example .env
# Edit .env with real values
docker compose -f docker-compose.prod.yml up --build
# App available at http://localhost
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgresql+asyncpg://...` | Full async database connection string |
| `SECRET_KEY` | `dev-secret-key-...` | JWT signing key — **change in production** |
| `ACCESS_TOKEN_EXPIRE_MINUTES` | `15` | Access token TTL |
| `REFRESH_TOKEN_EXPIRE_DAYS` | `30` | Refresh token TTL |
| `CORS_ORIGINS` | `["http://localhost:5173"]` | Allowed origins (JSON array) |
| `PORT` | `8000` (API) / `8080` (frontend) | Server port (set by Cloud Run) |

## License

This project is licensed under the Apache License 2.0 — see [LICENSE](LICENSE) for details.
