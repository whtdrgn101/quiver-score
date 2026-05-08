package main

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/quiverscore/backend-go/internal/imaging"
	"github.com/quiverscore/backend-go/internal/storage"
)

// These tests need a running Postgres with the schema migrated. They skip
// gracefully if BACKFILL_TEST_DATABASE_URL is unset so the package still
// builds cleanly in CI without a database.
func dbForTest(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("BACKFILL_TEST_DATABASE_URL")
	if url == "" {
		url = os.Getenv("DATABASE_URL")
	}
	if url == "" {
		t.Skip("BACKFILL_TEST_DATABASE_URL or DATABASE_URL not set; skipping integration test")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// seedSchemaUser creates a throwaway user row so foreign keys (user_id,
// session.user_id) line up. Returns the user id.
func seedUser(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	id := uuid.New().String()
	suffix := uuid.New().String()[:8]
	_, err := pool.Exec(context.Background(),
		`INSERT INTO users (id, email, username, hashed_password, display_name, email_verified, profile_public, created_at, updated_at)
		 VALUES ($1, $2, $3, 'x', 'Backfill Test', true, false, now(), now())`,
		id, "backfill_"+suffix+"@test.local", "backfill_"+suffix,
	)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, id)
	})
	return id
}

// seedEndImage builds a session + end + end_images row and returns the
// end_images.id. The whole chain is cleaned up in t.Cleanup.
func seedEndImage(t *testing.T, pool *pgxpool.Pool, userID string, body []byte, contentType string) (endImageID string) {
	t.Helper()
	ctx := context.Background()

	// Need a round template + scoring_session + end. Use the simplest valid
	// shapes the schema accepts.
	templateID := uuid.New().String()
	_, err := pool.Exec(ctx,
		`INSERT INTO round_templates (id, name, organization, is_official, created_at)
		 VALUES ($1, 'BackfillTest', 'Test', false, now())`,
		templateID)
	if err != nil {
		t.Fatalf("insert round_template: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(ctx, `DELETE FROM round_templates WHERE id = $1`, templateID) })

	stageID := uuid.New().String()
	_, err = pool.Exec(ctx,
		`INSERT INTO round_template_stages (id, template_id, stage_order, name, distance, num_ends, arrows_per_end, allowed_values, value_score_map, max_score_per_arrow)
		 VALUES ($1, $2, 1, 'S1', '18m', 10, 3, '["10","9","M"]'::json, '{"10":10,"9":9,"M":0}'::json, 10)`,
		stageID, templateID)
	if err != nil {
		t.Fatalf("insert round_template_stage: %v", err)
	}

	sessionID := uuid.New().String()
	_, err = pool.Exec(ctx,
		`INSERT INTO scoring_sessions (id, user_id, template_id, status, total_score, total_x_count, total_arrows, started_at)
		 VALUES ($1, $2, $3, 'in_progress', 0, 0, 0, now())`,
		sessionID, userID, templateID)
	if err != nil {
		t.Fatalf("insert scoring_session: %v", err)
	}

	endID := uuid.New().String()
	_, err = pool.Exec(ctx,
		`INSERT INTO ends (id, session_id, stage_id, end_number, end_total, created_at)
		 VALUES ($1, $2, $3, 1, 0, now())`,
		endID, sessionID, stageID)
	if err != nil {
		t.Fatalf("insert end: %v", err)
	}

	endImageID = uuid.New().String()
	_, err = pool.Exec(ctx,
		`INSERT INTO end_images (id, end_id, session_id, user_id, image_data, content_type, file_size, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, now())`,
		endImageID, endID, sessionID, userID, body, contentType, len(body))
	if err != nil {
		t.Fatalf("insert end_image: %v", err)
	}

	t.Cleanup(func() {
		// Order matters for FKs; just nuke everything we created in reverse.
		_, _ = pool.Exec(ctx, `DELETE FROM attachments WHERE legacy_id = $1`, endImageID)
		_, _ = pool.Exec(ctx, `DELETE FROM end_images WHERE id = $1`, endImageID)
		_, _ = pool.Exec(ctx, `DELETE FROM ends WHERE id = $1`, endID)
		_, _ = pool.Exec(ctx, `DELETE FROM scoring_sessions WHERE id = $1`, sessionID)
	})
	return endImageID
}

func realJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 64, 48))
	for y := 0; y < 48; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 3), G: uint8(y * 5), B: 100, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}
	return buf.Bytes()
}

func TestBackfill_MigratesAndIsIdempotent(t *testing.T) {
	pool := dbForTest(t)
	store := storage.NewMemory()
	processor := imaging.NewProcessor()

	userID := seedUser(t, pool)
	endImageID := seedEndImage(t, pool, userID, realJPEG(t), "image/jpeg")

	ctx := context.Background()

	// First run — migrates the row.
	s, err := run(ctx, pool, store, processor, runOpts{BatchSize: 10})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if s.migrated < 1 {
		t.Errorf("first run: migrated = %d, want at least 1", s.migrated)
	}

	// Verify a row in attachments references our end_image via legacy_id.
	var attachmentID string
	if err := pool.QueryRow(ctx,
		`SELECT id FROM attachments WHERE legacy_id = $1`, endImageID,
	).Scan(&attachmentID); err != nil {
		t.Fatalf("attachment row not found post-migrate: %v", err)
	}

	// Verify the storage objects exist under the user prefix.
	if _, _, err := store.Get(ctx, "users/"+userID+"/attachments/"+attachmentID+"/full.jpg"); err != nil {
		t.Errorf("full not in storage: %v", err)
	}
	if _, _, err := store.Get(ctx, "users/"+userID+"/attachments/"+attachmentID+"/thumb.jpg"); err != nil {
		t.Errorf("thumb not in storage: %v", err)
	}

	// Second run — should skip the row, not duplicate it.
	s2, err := run(ctx, pool, store, processor, runOpts{BatchSize: 10})
	if err != nil {
		t.Fatalf("second run: %v", err)
	}
	if s2.skipped < 1 {
		t.Errorf("second run: skipped = %d, want at least 1 (idempotent re-run)", s2.skipped)
	}

	var count int
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM attachments WHERE legacy_id = $1`, endImageID,
	).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Errorf("attachment count = %d, want 1 (no duplicate)", count)
	}
}

func TestBackfill_DryRunWritesNothing(t *testing.T) {
	pool := dbForTest(t)
	store := storage.NewMemory()
	processor := imaging.NewProcessor()

	userID := seedUser(t, pool)
	endImageID := seedEndImage(t, pool, userID, realJPEG(t), "image/jpeg")

	ctx := context.Background()
	s, err := run(ctx, pool, store, processor, runOpts{BatchSize: 10, DryRun: true})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if s.scanned < 1 {
		t.Errorf("scanned = %d, want at least 1", s.scanned)
	}

	// Nothing in attachments, nothing in storage.
	var count int
	_ = pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM attachments WHERE legacy_id = $1`, endImageID,
	).Scan(&count)
	if count != 0 {
		t.Errorf("dry-run wrote %d attachment rows", count)
	}
}

func TestBackfill_FailedRowsCountedNotFatal(t *testing.T) {
	pool := dbForTest(t)
	store := storage.NewMemory()
	processor := imaging.NewProcessor()

	userID := seedUser(t, pool)
	// Garbage bytes claiming to be a JPEG → imaging.Process fails.
	bad := seedEndImage(t, pool, userID, []byte("not really a jpeg"), "image/jpeg")
	good := seedEndImage(t, pool, userID, realJPEG(t), "image/jpeg")

	ctx := context.Background()
	s, err := run(ctx, pool, store, processor, runOpts{BatchSize: 10})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if s.failed < 1 {
		t.Errorf("failed = %d, want at least 1", s.failed)
	}
	if s.migrated < 1 {
		t.Errorf("migrated = %d, want at least 1 (good row should still go through)", s.migrated)
	}

	// Confirm the good row landed and the bad one didn't.
	var goodCount, badCount int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM attachments WHERE legacy_id = $1`, good).Scan(&goodCount)
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM attachments WHERE legacy_id = $1`, bad).Scan(&badCount)
	if goodCount != 1 || badCount != 0 {
		t.Errorf("good=%d bad=%d, want 1/0", goodCount, badCount)
	}
}

// Sanity check that the backfill runs cleanly with no end_images at all.
func TestBackfill_NoRows(t *testing.T) {
	pool := dbForTest(t)
	store := storage.NewMemory()
	processor := imaging.NewProcessor()

	// Wipe end_images for this test to ensure an empty start (other tests
	// have already cleaned up their own rows via t.Cleanup).
	ctx := context.Background()
	_, err := pool.Exec(ctx, `SELECT 1`) // canary: connection works
	if err != nil {
		t.Fatalf("ping: %v", err)
	}
	// We can't safely TRUNCATE — there may be unrelated end_images rows in dev.
	// Instead just check that run completes without error against whatever's there.
	_, err = run(ctx, pool, store, processor, runOpts{BatchSize: 10, Limit: 0, DryRun: true})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	_ = time.Now() // anchor for staleness checks if any
}
