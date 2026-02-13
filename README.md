# QuiverScore

Target archery score tracking application. Record scores across official and custom round formats, manage equipment and sight marks, join clubs, compete in tournaments, train with coaches, and connect with the archery community. Free, privacy-focused, and installable as a PWA on mobile devices.

## Features

### Score Tracking
- Tap-based arrow entry with real-time running totals, X counts, and scorecard display
- Multi-stage rounds with dynamic stage progression (e.g., WA 1440 with multiple distances)
- 6 official round templates: WA Indoor 18m (Recurve & Compound), WA 720 70m, WA 1440, Vegas 300, NFAA Indoor 300
- Custom round templates — create, edit, and share with clubs
- Session metadata: location, weather, notes
- Abandon and delete sessions
- Resume in-progress sessions from a banner notification
- Export completed sessions to CSV or PDF
- Share sessions via public link
- Compare two sessions side by side

### Equipment & Sight Marks
- Track bows, risers, limbs, arrows, sights, and accessories with custom specs
- Build named setup profiles and link equipment to them
- Log sight marks by distance for each setup

### Clubs & Teams
- Create or join clubs with invite codes
- Owner/admin/member role hierarchy
- Club events with participant tracking
- Teams within clubs
- Share custom round templates with clubs

### Tournaments
- Organize tournaments within clubs
- Leaderboards with automatic ranking
- Link scoring sessions to tournament participation

### Coaching
- Coach/athlete linking with invite system
- Coaches can view athlete sessions and scores
- Session annotations by coaches
- Track athlete progress over time

### Social
- Follow other archers
- Activity feed with recent sessions and achievements
- Public user profiles

### Analytics
- Dashboard: personal best, average by round type, recent score trend, total arrows/X count
- Personal records tracked automatically
- Classification tracking for World Archery and NFAA
- Scoring history with filter and search

### Account & Privacy
- Email verification
- Password reset via email
- Full account deletion with all associated data
- No tracking, no ads, no data selling

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
│   │   ├── api/v1/              # Route handlers
│   │   │   ├── auth.py          #   Registration, login, refresh, email verify, password reset, account deletion
│   │   │   ├── users.py         #   Profile, avatar, public profiles
│   │   │   ├── rounds.py        #   Round templates: list, create, edit, share with clubs
│   │   │   ├── scoring.py       #   Sessions, ends, arrows, stats, abandon, delete, export
│   │   │   ├── sharing.py       #   Session sharing via public links
│   │   │   ├── equipment.py     #   Equipment CRUD
│   │   │   ├── setups.py        #   Setup profiles, equipment linking
│   │   │   ├── clubs.py         #   Clubs, members, events, teams, shared rounds
│   │   │   ├── notifications.py #   User notifications
│   │   │   ├── classifications.py # Classification records
│   │   │   ├── sight_marks.py   #   Sight mark CRUD by distance
│   │   │   ├── coaching.py      #   Coach/athlete links, session annotations
│   │   │   ├── social.py        #   Follow/unfollow, activity feed
│   │   │   └── __init__.py      #   Router aggregation
│   │   ├── models/              # SQLAlchemy models
│   │   ├── schemas/             # Pydantic request/response schemas
│   │   ├── core/                # Security (JWT, bcrypt), email, rate limiting, exceptions
│   │   ├── seed/                # Round template seed data
│   │   ├── config.py            # Settings via environment variables
│   │   ├── database.py          # Async engine + session factory
│   │   ├── dependencies.py      # Auth dependency (get_current_user)
│   │   └── main.py              # FastAPI app, CORS, lifespan
│   ├── alembic/                 # Database migrations
│   ├── tests/                   # pytest async tests (189 tests)
│   ├── Dockerfile
│   └── pyproject.toml
├── frontend/
│   ├── src/
│   │   ├── api/                 # Axios API clients
│   │   ├── pages/               # Route components
│   │   │   ├── Dashboard        #   Stats cards, round averages, recent trend
│   │   │   ├── RoundSelect      #   Pick template, setup, location/weather
│   │   │   ├── CreateRound      #   Create or edit custom round templates
│   │   │   ├── ScoreSession     #   Live scoring with tap-to-score grid
│   │   │   ├── SessionDetail    #   Completed session scorecard + export
│   │   │   ├── SharedSession    #   Public shared session view
│   │   │   ├── CompareSession   #   Side-by-side session comparison
│   │   │   ├── History          #   All past sessions with filter/search
│   │   │   ├── Equipment        #   Equipment list + CRUD
│   │   │   ├── Profile          #   User profile, password change, danger zone
│   │   │   ├── PublicProfile    #   Public user profile page
│   │   │   ├── Clubs            #   Club listing + creation
│   │   │   ├── ClubDetail       #   Club members, events, teams, rounds, tournaments
│   │   │   ├── ClubSettings     #   Club admin settings
│   │   │   ├── TournamentCreate #   Create tournament within club
│   │   │   ├── TournamentDetail #   Tournament leaderboard + participation
│   │   │   ├── CoachDashboard   #   Coach's athlete list + overview
│   │   │   ├── AthleteView      #   Coach's view of an athlete's data
│   │   │   ├── Feed             #   Social activity feed
│   │   │   └── About            #   About the project
│   │   ├── contexts/            # Auth + theme contexts
│   │   ├── hooks/               # useAuth, useTheme hooks
│   │   └── components/          # Layout shell (nav, resume banner), error boundary
│   ├── Dockerfile               # Multi-stage build (Node → Nginx)
│   ├── nginx.conf               # SPA fallback + API proxy (local compose)
│   ├── nginx.conf.template      # Parameterized config (Cloud Run)
│   └── package.json
├── docker-compose.yml           # Local development
├── docker-compose.prod.yml      # Local prod testing
├── .env.example                 # Required environment variables
└── .github/workflows/
    ├── ci.yml                   # Test + lint + build on PR and push
    └── deploy.yml               # Build + deploy to Cloud Run on merge to main
```

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
