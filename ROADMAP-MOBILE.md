# QuiverScore: Mobile App & Platform Roadmap

## Overview

Build a Flutter mobile app for offline-first round scoring and tournament play, paired with API enhancements for end-of-end target photography, user profiles, and social features. The mobile app lives in `mobile/` and targets both Android and iOS.

## Architecture

```
                  ┌──────────────────────┐
  Mobile App ──▶  │  Go API (Cloud Run)  │
  (Flutter)       │  quiverscore.com     │
  offline-first   │    /api/v1/sessions  │──▶ PostgreSQL
  SQLite local    │    /api/v1/users/me  │      ├── end_images (bytea)
                  │    /api/v1/scoring/  │      └── (future: GCS bucket)
                  │      {id}/images     │
                  └──────────────────────┘
```

---

## Phase 1: End Images API ✅

### 1.1 — Database Migration

- [x] Alembic migration: `end_images` table (id, end_id, session_id, user_id, image_data bytea, content_type, file_size, created_at)
- [x] Foreign keys: end_id → ends, session_id → scoring_sessions, user_id → users (all CASCADE delete)
- [x] Indexes on end_id, session_id, user_id

### 1.2 — Go Repository

- [x] `EndImageRepo` with: Upload, GetMeta, GetImageData, ListByEnd, ListBySession, Delete
- [x] Ownership verification helpers: EndBelongsToSession, SessionBelongsToUser

### 1.3 — Go Handler & Routes

- [x] `EndImagesHandler` with routes under `/api/v1/scoring`:
  - `POST /{sessionId}/ends/{endId}/images` — multipart upload (jpeg, png, webp, heic; 10 MB max)
  - `GET /{sessionId}/ends/{endId}/images` — list images for an end
  - `GET /{sessionId}/images` — list all images for a session
  - `GET /{sessionId}/images/{imageId}` — download image binary
  - `DELETE /{sessionId}/images/{imageId}` — delete image
- [x] Auth required, session ownership verified on all routes

### 1.4 — Unit Tests

- [x] 14 handler tests covering: upload success, session not owned, end not in session, missing field, get image success/not found/not owned, list by end (with results + empty), list by session (success + not owned), delete (success + not found + not owned)
- [x] Verify all existing tests still pass after wiring (265/265 pass)

### 1.5 — Contract Tests

- [x] 12 pytest contract tests covering: upload (success, not found, wrong end, no auth, wrong user), get image (success, not found), list by end (with results, empty), list by session, delete (success, not found)
- [x] Run against Go API on :8080 — 265/265 pass

### 1.6 — Deploy & Verify

- [x] Run migration locally (alembic upgrade e9f0a1b2c3d4 -> f0a1b2c3d4e5)
- [ ] Push to main → CI/CD deploys migration + Go API to production
- [ ] Smoke test image upload/download in production

---

## Phase 2: Web UI — End Images ✅

### 2.1 — Session Detail View

- [x] Display end images in the session detail page (thumbnail per end)
- [x] Click thumbnail to view full-size image (lightbox modal)
- [x] Upload button per end (file picker with camera icon)
- [x] Delete button per image (in lightbox modal)

### 2.2 — Scoring Flow Integration

- [x] After submitting an end, optional prompt to attach a photo
- [x] Photo thumbnail visible in the end history during scoring

---

## Phase 3: Mobile App — Core Scoring ✅

### 3.1 — Project Scaffold

- [x] Flutter project at `mobile/` with Android + iOS targets
- [x] Dependencies: Riverpod, Dio, Drift (SQLite), connectivity_plus, image_picker, flutter_secure_storage
- [x] Core architecture: api client, secure storage, connectivity service, sync engine, local DB tables

### 3.2 — Auth Flow

- [x] Login screen (username + password)
- [x] Registration screen (email, username, password, display name)
- [x] Token persistence in secure storage
- [x] Auto token refresh interceptor on Dio
- [x] Auth-gated routing (login vs home)
- [x] "Logged in as" indicator + logout (More tab with user profile card)
- [x] Handle token expiry gracefully (redirect to login)

### 3.3 — Round Templates (Read-Only Sync)

- [x] Pull round templates from API on login/app open
- [x] Cache in local SQLite (Drift)
- [x] Round selection screen for new session
- [x] Show template details (stages, distances, arrows per end)
- [x] Pull-to-refresh

### 3.4 — Scoring UX

- [x] New session screen (select round, optional location/notes)
- [x] End-by-end arrow input pad (dynamic based on allowed_values)
- [x] Running score display (total, arrows, Xs, current end)
- [x] End history with arrow values and photo indicators
- [x] Complete / abandon session
- [x] Final end flow: photo capture before complete dialog
- [x] Undo last end
- [x] Multi-stage support (auto-advance between stages with distance change indicator)
- [x] Haptic feedback on arrow input

### 3.5 — Offline Sync Engine

- [x] SyncQueue table for pending mutations
- [x] Connectivity listener triggers sync on reconnect
- [x] Session create → enqueue for API sync
- [x] End submit → enqueue for API sync
- [x] Session complete/abandon → enqueue for API sync
- [x] Handle server ID mapping (client UUID → server UUID for sessions and ends)
- [x] Actionable sync button with success/failure feedback
- [x] Debug logging throughout sync pipeline
- [x] Sync status indicator in UI (pending count badge on sync button)
- [x] Error handling: retry with exponential backoff

### 3.6 — Dashboard & History

- [x] Dashboard tab with stats from API (rounds, arrows, Xs, personal best, personal records)
- [x] History tab merging local + server sessions (deduplicated)
- [x] Session detail screen for local sessions (end/arrow breakdown, photo indicators)
- [x] Server session detail screen for cloud-only sessions (fetched from API)
- [x] Tappable photo icons to view full-size images with pinch-to-zoom
- [x] Web links to quiverscore.com features (More tab)

---

## Phase 4: Mobile App — Target Photos ✅

### 4.1 — Camera Capture

- [x] Camera screen after end submission (snackbar for non-final ends, direct navigation for final end)
- [x] Save photo to app documents directory
- [x] Store metadata in local SQLite (EndImages table)
- [x] Photo indicator on end rows (tappable to view full image)
- [x] Gallery picker as alternative to camera
- [x] Image compression before upload (JPEG quality 70, max 1920px)

### 4.2 — Photo Sync

- [x] Enqueue image upload in sync queue
- [x] Multipart upload to `POST /api/v1/scoring/{sessionId}/ends/{endId}/images`
- [x] Server end ID tracking (EndsLocal.serverId, schema v2) for image upload mapping
- [x] Track sync status per image (pending, uploading, synced, failed)

---

## Phase 5: User Profiles & Social Links ✅

### 5.1 — Social Links — Database & API ✅

- [x] Alembic migration: add `social_links` JSONB column to `users` table
- [x] Update `UserOut` struct to include `social_links`
- [x] Update `profileUpdate` struct to accept `social_links`
- [x] Update `UpdateProfile` repository method to persist social links
- [x] Unit tests for social links CRUD
- [x] Contract tests for social links via profile update

### 5.2 — Web — Profile Social Links ✅

- [x] Add social links section to Profile.jsx (Instagram, X/Twitter, Facebook, YouTube, TikTok, website)
- [x] Display social links on PublicProfile.jsx
- [x] Icon buttons linking to external profiles

### 5.3 — Mobile — Profile Edit Screen ✅

- [x] Full `UserInfo` model (add bio, avatar, classification, profile_public, social_links)
- [x] Profile edit screen: display name, bio, bow type, classification, profile public toggle
- [x] Avatar upload (camera or gallery → `POST /api/v1/users/me/avatar`)
- [x] Avatar display on More tab and profile screen
- [x] Social links editor (add/remove platforms with URL input)
- [x] Navigate to profile edit from More tab

---

## Phase 5.5: Multi-Round Tournament Brackets ✅

### 5.5.1 — Database Migration ✅

- [x] Alembic migration: `tournament_rounds` table (id, tournament_id, round_number, name, template_id, advancement, status, started_at, completed_at, created_at)
- [x] Alembic migration: `tournament_round_scores` table (id, round_id, participant_id, session_id, score, x_count, rank_in_round, advanced)
- [x] Foreign keys with CASCADE deletes, unique constraints, indexes

### 5.5.2 — Go API Endpoints ✅

- [x] `POST /api/v1/clubs/{clubId}/tournaments/{tournamentId}/rounds` — add round
- [x] `GET /api/v1/clubs/{clubId}/tournaments/{tournamentId}/rounds` — list rounds
- [x] `POST .../rounds/{roundId}/start` — start round
- [x] `POST .../rounds/{roundId}/submit-score?session_id=X` — submit round score
- [x] `GET .../rounds/{roundId}/leaderboard` — round leaderboard
- [x] `POST .../rounds/{roundId}/complete` — complete round (ranks, advances top N)
- [x] Advancement logic: top N advance, ties at cutoff boundary advance all tied

### 5.5.3 — Tests ✅

- [x] 11 unit tests for tournament round handlers
- [x] 8 contract tests for round lifecycle
- [x] Delete cascade updated for tournament_round_scores and tournament_rounds

---

## Phase 6: Mobile App — Tournament Play

### 6.1 — Tournament List

- [ ] Pull active tournaments user is registered for (via clubs API)
- [ ] Tournament detail screen (name, template, dates, status)
- [ ] "Score This Round" button → starts session with tournament's template

### 6.2 — Score Submission

- [ ] After completing a tournament round, prompt to submit score
- [ ] Call `POST /api/v1/clubs/{clubId}/tournaments/{tournamentId}/rounds/{roundId}/submit-score?session_id=X`
- [ ] Show submission confirmation with score + ranking

### 6.3 — Leaderboard

- [ ] View tournament leaderboard from mobile (overall + per-round)
- [ ] Highlight user's own position

---

## Phase 7: Image Storage Migration (Future)

### 7.1 — Object Storage Backend

- [ ] Add GCS bucket for end images
- [ ] Go API: write to GCS, store URL in `end_images.storage_url` column
- [ ] Migration: add `storage_url` column, make `image_data` nullable
- [ ] Backfill: move existing bytea data to GCS
- [ ] Update GET endpoint: serve from GCS (signed URLs or proxy)
- [ ] Drop `image_data` column after backfill verified

### 7.2 — CDN & Optimization

- [ ] Serve images via CDN (Cloud CDN or Firebase Hosting proxy)
- [ ] Generate thumbnails on upload (for list views)
- [ ] Consider WebP conversion for bandwidth savings

---

## Phase 8: Social & Future Features (Future)

### 8.1 — Challenge Friends (Future)

- [ ] Challenge a friend to shoot the same round
- [ ] Real-time or async comparison
- [ ] Leverage existing social/follow infrastructure

### 8.2 — Push Notifications (Future)

- [ ] Firebase Cloud Messaging integration
- [ ] Tournament reminders, challenge notifications, personal record alerts

### 8.3 — Head-to-Head Matchups (Future)

- [ ] Bracket-style head-to-head pairing within tournament rounds
- [ ] Matchup table: round_id, participant_a, participant_b, winner_id
- [ ] Auto-generate pairings from round leaderboard (1 vs N, 2 vs N-1, etc.)
- [ ] Visual bracket display on web and mobile
- [ ] Support byes for non-power-of-2 participant counts
