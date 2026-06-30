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
**Email:** Resend (verification + password reset)

---

## Completed Work

### API Migration (Phases 0–10) ✅

Full migration from Python/FastAPI to Go. 253 contract tests covering all endpoints. Python stripped to Alembic migrations only.

- **Phase 0** — Foundation: contract test infrastructure, Go scaffold, shared auth middleware
- **Phase 1** — Auth: register, login, refresh, verify email, password reset, delete account, rate limiting, Resend
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
- _Superseded by Phase 19 — `end_images` table and `/scoring/.../images` endpoints removed; image storage now lives in GCS via the generic `attachments` table._

### Mobile App (Phases 13–14) ✅

- **Phase 13** — Core Scoring: Flutter scaffold, auth flow, round template sync, end-by-end scoring UX, offline sync engine (dependency-aware ordering, exponential backoff, download-and-resume), dashboard & history
- **Phase 14** — Target Photos: camera/gallery capture, image compression, photo sync via queue, server end ID tracking

### User Profiles & Social Links (Phase 15) ✅

- **Phase 15** — Social Links: `social_links` JSONB column, web profile editor + public display, mobile profile edit screen with avatar upload

### Multi-Round Tournament Brackets — API (Phase 16) ✅

- **Phase 16** — Tournament rounds API: `tournament_rounds` and `tournament_round_scores` tables, 6 endpoints (add/list/start/complete rounds, submit round score, round leaderboard), advancement logic with tie handling, 11 unit tests + 8 contract tests

### Mobile Feature Parity — Equipment & Clubs (Phase 16.5) ✅

- **Phase 16.5** — Mobile equipment & clubs: Drift schema v3 with cache tables, equipment & setups full CRUD with online-first caching, clubs view/join/leave/leaderboard/activity/events with RSVP/teams, 4-tab bottom nav (Dashboard, History, Clubs, More), 29 model serialization tests

### Multi-Round Tournament Management — Web (Phase 17) ✅

- **Phase 17** — Web tournament rounds UI: API wrappers for 6 round endpoints, organizer controls (add/start/complete rounds with advancement), participant round scoring flow threaded through RoundSelect → ScoreSession, per-round leaderboards with Advanced/Eliminated badges, scored indicators on round rows, re-score warning, user highlighting in leaderboards, post-submission confirmation with auto-navigate back to tournament

### Image Storage Migration (Phase 19) ✅

- **Infra** — `quiverscore-images-prod` and `-dev` GCS buckets in us-central1 with UBLA + Public Access Prevention enforced; default 7-day soft-delete; Cloud Run SA granted `roles/storage.objectAdmin`
- **Go API** — `internal/storage` ObjectStore interface with GCS / local / memory backends + conformance suite; `internal/imaging` JPEG processor (1920px Q80 full + 320px Q70 thumb, decodes JPEG/PNG/WebP, rejects HEIC); generic polymorphic `attachments` table with owner_type ∈ {`session_end`, `equipment`, `setup`}; `/api/v1/attachments` handler with OwnerVerifier registry, per-user-per-owner-type rate limits, and per-owner caps; storage layout `users/{userID}/attachments/{attachmentID}/{full,thumb}.jpg`; account-deletion prefix wipe; 21 unit + 18 contract tests
- **Web** — `<AttachmentImage>` (auth-fetch + blob URL) and `<AttachmentGallery>` components; SessionDetail and ScoreSession cut over to render thumbs from `attachment_ids` embedded on each end; equipment + setup expanded views got photo galleries; Playwright e2e for upload/view/delete
- **Mobile (1.4.0+6)** — Drift schema v4 with `server_attachment_id` on `end_images` for sync idempotency; `_syncImage` posts to `/api/v1/attachments`; honors `Retry-After` on 429
- **Cleanup** — legacy `end_images` table dropped; legacy `/scoring/.../images` endpoints + handler + repo + backfill command removed

---

## Up Next

### Phase 18: Tournament Play — Mobile ✅

Bring tournament participation to the Flutter app, building on the web tournament flow.

#### 18.1 — Tournament List & Detail

- [x] Pull active tournaments user is registered for (via clubs API)
- [x] Tournament detail screen (name, template, dates, status, participants)
- [x] "Score This Round" button → starts session with tournament's template
- [x] Tournaments tab added to club detail screen (6 tabs)
- [x] Dashboard active tournaments navigate to tournament detail

#### 18.2 — Score Submission

- [x] After completing a tournament round, prompt to submit score
- [x] Call `POST /api/v1/clubs/{clubId}/tournaments/{tournamentId}/rounds/{roundId}/submit-score?session_id=X`
- [x] Show submission confirmation with score + ranking
- [x] Force sync session before tournament score submission
- [x] Graceful error handling when offline

#### 18.3 — Leaderboard

- [x] View tournament leaderboard from mobile (overall + per-round)
- [x] Highlight user's own position
- [x] Per-round leaderboards with advancement badges (Advanced/Eliminated)

---

## Future

### Phase 20: Challenge Friends

- [ ] Challenge a friend to shoot the same round
- [ ] Real-time or async comparison
- [ ] Leverage existing social/follow infrastructure

### Phase 21: Push Notifications

- [ ] Firebase Cloud Messaging integration
- [ ] Tournament reminders, challenge notifications, personal record alerts

### Phase 22: Head-to-Head Matchups

- [x] Bracket-style head-to-head pairing within tournament rounds
- [x] Matchup table: round_id, participant_a, participant_b, winner_id
- [x] Auto-generate pairings from round leaderboard (1 vs N, 2 vs N-1, etc.)
- [x] Visual bracket display on web and mobile (Mobile view completed; Web view deferred)
- [x] Support byes for non-power-of-2 participant counts

### Phase 23: Tournament Bracket Visualization

- [ ] Visual bracket progression showing round-over-round results (web)
- [x] Bracket view on mobile
- [x] Show advancement flow across rounds with scores

### Phase 24: Biometric App Lock ✅

- [x] Integrate `local_auth` Flutter package for Face ID / fingerprint
- [x] Optional biometric lock on app launch (user-configurable in settings)
- [x] Biometric unlock gates app access, not token refresh — complements offline auth
- [x] Fallback to device PIN/pattern when biometrics unavailable
- [x] Persist biometric preference in local storage
