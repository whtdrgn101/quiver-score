# QuiverScore Platform Roadmap

## Overview

QuiverScore is a modern, privacy-focused platform for target archery — tracking scores, building community, and running tournaments. The platform spans a React web app, a Flutter mobile app (iOS + Android), and a Go API backend.

## Architecture

```
                  ┌──────────────────────┐
  HTTP ────────▶  │  Firebase Hosting     │
                  │  (static CDN)         │
                  │    ├── SPA assets     │
                  │    └── /api/** ───────┼──▶  Cloud Run: Go API
                  └──────────────────────┘     ├── all routes + PDF export
                                               └──▶ PostgreSQL (Cloud SQL)
  Mobile App ──▶  Go API (same)
  (Flutter)       offline-first
  SQLite local    sync on reconnect

  On deploy:  Cloud Run Job (Python/Alembic) → run migrations → exit
```

**Go API:** `chi` router, `pgx` (SQL), `golang-jwt`, `bcrypt`, `go-pdf/fpdf`
**Web:** React 19 + Vite SPA with PWA, hosted on Firebase
**Mobile:** Flutter (Riverpod, Dio, Drift/SQLite, offline-first sync engine)
**Migrations:** Python/Alembic (deploy-only, no runtime server)
**CI/CD:** GitHub Actions → Go tests + frontend build → deploy API to Cloud Run + frontend to Firebase
**Email:** SendGrid (verification + password reset)

---

## Completed Work

### API Migration (Phases 0–10) ✅

Full migration from Python/FastAPI to Go. 253 contract tests covering all endpoints. Python stripped to Alembic migrations only.

- **Phase 0** — Foundation: contract test infrastructure, Go scaffold, shared auth middleware
- **Phase 1** — Auth: register, login, refresh, verify email, password reset, delete account, rate limiting, SendGrid
- **Phase 2** — Rounds: round template CRUD, sharing, ownership checks
- **Phase 3** — Equipment & Setups: CRUD with partial updates, equipment linking
- **Phase 4** — Sight Marks & Classifications: CRUD with filters, classification records
- **Phase 5** — Scoring Sessions: session lifecycle, stats/trends, CSV/PDF export, session sharing
- **Phase 6** — Users & Sharing: profile update, avatar upload, public profiles
- **Phase 7** — Clubs: CRUD, invites, member RBAC, events/RSVP, teams, tournaments, leaderboards
- **Phase 8** — Social & Coaching: follow/unfollow, activity feed, coaching invites, session annotations
- **Phase 9** — Notifications: list, unread count, mark read
- **Phase 10** — Cleanup & Cutover: stripped Python, updated CI/CD, production verified

### End Images (Phases 11–12) ✅

- **Phase 11** — End Images API: `end_images` table, Go handler for upload/download/delete (multipart, 10 MB max), 12 contract tests, 14 unit tests
- **Phase 12** — Web UI: thumbnails in session detail, lightbox modal, photo capture during scoring flow

### Mobile App (Phases 13–14) ✅

- **Phase 13** — Core Scoring: Flutter scaffold, auth flow, round template sync, end-by-end scoring UX, offline sync engine (dependency-aware ordering, exponential backoff, download-and-resume), dashboard & history
- **Phase 14** — Target Photos: camera/gallery capture, image compression, photo sync via queue, server end ID tracking

### User Profiles & Social Links (Phase 15) ✅

- **Phase 15** — Social Links: `social_links` JSONB column, web profile editor + public display, mobile profile edit screen with avatar upload

### Multi-Round Tournament Brackets — API (Phase 16) ✅

- **Phase 16** — Tournament rounds API: `tournament_rounds` and `tournament_round_scores` tables, 6 endpoints (add/list/start/complete rounds, submit round score, round leaderboard), advancement logic with tie handling, 11 unit tests + 8 contract tests

---

## Up Next

### Phase 17: Multi-Round Tournament Management — Web

Wire the existing tournament rounds API into the web frontend. The backend endpoints are built and tested — this phase is purely frontend.

#### 17.1 — API Wrappers & Rounds Section

- [ ] Add tournament round API functions to `frontend/src/api/tournaments.js` (addRound, listRounds, startRound, completeRound, submitRoundScore, getRoundLeaderboard)
- [ ] Add rounds section to `TournamentDetail.jsx` — list rounds with status, round number, and template name
- [ ] Organizer controls: "Add Round" button with round name, template selection, and advancement count

#### 17.2 — Organizer Round Lifecycle

- [ ] Start round button (only when tournament is in_progress and round is pending)
- [ ] Complete round button (ranks participants, advances top N, handles ties at cutoff)
- [ ] Visual round status indicators (pending, in_progress, completed)
- [ ] Show which participants advanced vs eliminated per round

#### 17.3 — Participant Round Scoring

- [ ] "Score This Round" navigates with round context (roundId + tournamentId + clubId)
- [ ] Update `RoundSelect.jsx` and `ScoreSession.jsx` to pass roundId through the scoring flow
- [ ] On session complete, submit to round-level endpoint (`/rounds/{roundId}/submit-score`) instead of tournament-level
- [ ] Banner in scoring flow indicating which tournament round is being scored

#### 17.4 — Per-Round Leaderboards & Bracket View

- [ ] Per-round leaderboard tab/accordion in `TournamentDetail.jsx`
- [ ] Advancement indicators (advanced, eliminated) on leaderboard entries
- [ ] Overall tournament leaderboard remains as summary view
- [ ] Visual bracket progression showing round-over-round results

---

### Phase 18: Tournament Play — Mobile

Bring tournament participation to the Flutter app, building on the web tournament flow.

#### 18.1 — Tournament List & Detail

- [ ] Pull active tournaments user is registered for (via clubs API)
- [ ] Tournament detail screen (name, template, dates, status, participants)
- [ ] "Score This Round" button → starts session with tournament's template

#### 18.2 — Score Submission

- [ ] After completing a tournament round, prompt to submit score
- [ ] Call `POST /api/v1/clubs/{clubId}/tournaments/{tournamentId}/rounds/{roundId}/submit-score?session_id=X`
- [ ] Show submission confirmation with score + ranking

#### 18.3 — Leaderboard

- [ ] View tournament leaderboard from mobile (overall + per-round)
- [ ] Highlight user's own position

---

## Future

### Phase 19: Image Storage Migration

- [ ] Add GCS bucket for end images
- [ ] Go API: write to GCS, store URL in `end_images.storage_url` column
- [ ] Migration: add `storage_url` column, make `image_data` nullable
- [ ] Backfill: move existing bytea data to GCS
- [ ] Update GET endpoint: serve from GCS (signed URLs or proxy)
- [ ] Drop `image_data` column after backfill verified
- [ ] CDN via Cloud CDN or Firebase Hosting proxy
- [ ] Generate thumbnails on upload for list views
- [ ] Consider WebP conversion for bandwidth savings

### Phase 20: Challenge Friends

- [ ] Challenge a friend to shoot the same round
- [ ] Real-time or async comparison
- [ ] Leverage existing social/follow infrastructure

### Phase 21: Push Notifications

- [ ] Firebase Cloud Messaging integration
- [ ] Tournament reminders, challenge notifications, personal record alerts

### Phase 22: Head-to-Head Matchups

- [ ] Bracket-style head-to-head pairing within tournament rounds
- [ ] Matchup table: round_id, participant_a, participant_b, winner_id
- [ ] Auto-generate pairings from round leaderboard (1 vs N, 2 vs N-1, etc.)
- [ ] Visual bracket display on web and mobile
- [ ] Support byes for non-power-of-2 participant counts
