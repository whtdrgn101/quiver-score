// Command backfill_images migrates rows from end_images (image bytes stored
// in Postgres BYTEA) into the new attachments + GCS layout.
//
// It is idempotent — re-running picks up where it left off via the legacy_id
// column on attachments. The unique partial index ix_attachments_legacy_id
// guarantees no duplicate writes even under concurrent runs.
//
// Usage (Cloud Run Job ready — env-only config):
//
//	STORAGE_BACKEND=gcs GCS_BUCKET=quiverscore-images-prod \
//	DATABASE_URL=... ./backfill_images [flags]
//
// Flags:
//
//	-dry-run        Report what would be migrated without writing.
//	-batch-size N   Page size for cursor walk over end_images (default 100).
//	-limit N        Stop after N rows (0 = no limit).
package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/database"
	"github.com/quiverscore/backend-go/internal/imaging"
	"github.com/quiverscore/backend-go/internal/storage"
)

type stats struct {
	scanned   int
	migrated  int
	skipped   int // already migrated
	failed    int
}

func (s stats) String() string {
	return fmt.Sprintf("scanned=%d migrated=%d skipped=%d failed=%d", s.scanned, s.migrated, s.skipped, s.failed)
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	dryRun := flag.Bool("dry-run", false, "report only; do not write to GCS or attachments")
	batchSize := flag.Int("batch-size", 100, "rows per cursor page")
	limit := flag.Int("limit", 0, "stop after N rows (0 = no limit)")
	flag.Parse()

	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := database.Connect(ctx, cfg.NormalizeDatabaseURL())
	if err != nil {
		slog.Error("connect db", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	store, err := storage.FromConfig(ctx, cfg)
	if err != nil {
		slog.Error("init storage", "error", err, "backend", cfg.StorageBackend)
		os.Exit(1)
	}

	processor := imaging.NewProcessor()

	// Use a long-lived context for the actual walk; the 30s timeout above is
	// just for setup. The walk can take minutes against a real prod table.
	runCtx := context.Background()

	s, err := run(runCtx, pool, store, processor, runOpts{
		DryRun:    *dryRun,
		BatchSize: *batchSize,
		Limit:     *limit,
	})
	slog.Info("backfill complete", "stats", s.String(), "dry_run", *dryRun)
	if err != nil {
		slog.Error("backfill encountered error", "error", err)
		os.Exit(1)
	}
}

type runOpts struct {
	DryRun    bool
	BatchSize int
	Limit     int
}

// run is the core backfill loop, factored out of main so tests can drive it
// against a synthetic end_images table.
func run(ctx context.Context, pool *pgxpool.Pool, store storage.ObjectStore, processor *imaging.Processor, opts runOpts) (stats, error) {
	if opts.BatchSize <= 0 {
		opts.BatchSize = 100
	}
	var s stats

	// Cursor walk: ordered by created_at, id so paging is stable. We use a
	// keyset cursor (last seen created_at, id) rather than OFFSET to avoid
	// re-scanning earlier rows as we go.
	var cur *batchCursor

	for {
		if opts.Limit > 0 && s.scanned >= opts.Limit {
			return s, nil
		}

		batch, err := fetchBatch(ctx, pool, cur, opts.BatchSize)
		if err != nil {
			return s, fmt.Errorf("fetch batch: %w", err)
		}
		if len(batch) == 0 {
			return s, nil
		}

		for _, row := range batch {
			if opts.Limit > 0 && s.scanned >= opts.Limit {
				return s, nil
			}
			s.scanned++

			if err := migrateRow(ctx, pool, store, processor, row, opts.DryRun); err != nil {
				if errors.Is(err, errAlreadyMigrated) {
					s.skipped++
					continue
				}
				s.failed++
				slog.Error("migrate row failed", "end_image_id", row.id, "error", err)
				continue
			}
			s.migrated++
		}

		last := batch[len(batch)-1]
		cur = &batchCursor{createdAt: last.createdAt, id: last.id}
	}
}

type batchCursor struct {
	createdAt time.Time
	id        string
}

type endImageRow struct {
	id          string
	endID       string
	userID      string
	imageData   []byte
	contentType string
	createdAt   time.Time
}

func fetchBatch(ctx context.Context, pool *pgxpool.Pool, after *batchCursor, batchSize int) ([]endImageRow, error) {
	var (
		rows pgx.Rows
		err  error
	)
	const baseSelect = `SELECT id, end_id, user_id, image_data, content_type, created_at
		FROM end_images`
	if after == nil {
		rows, err = pool.Query(ctx, baseSelect+`
			ORDER BY created_at, id LIMIT $1`, batchSize)
	} else {
		rows, err = pool.Query(ctx, baseSelect+`
			WHERE (created_at, id) > ($1, $2)
			ORDER BY created_at, id LIMIT $3`, after.createdAt, after.id, batchSize)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []endImageRow
	for rows.Next() {
		var r endImageRow
		if err := rows.Scan(&r.id, &r.endID, &r.userID, &r.imageData, &r.contentType, &r.createdAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

var errAlreadyMigrated = errors.New("already migrated")

func migrateRow(ctx context.Context, pool *pgxpool.Pool, store storage.ObjectStore, processor *imaging.Processor, row endImageRow, dryRun bool) error {
	// Idempotency check: if a row with this legacy_id already exists, skip.
	var existing string
	err := pool.QueryRow(ctx,
		`SELECT id FROM attachments WHERE legacy_id = $1`, row.id,
	).Scan(&existing)
	if err == nil {
		return errAlreadyMigrated
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("idempotency check: %w", err)
	}

	processed, err := processor.Process(row.imageData, row.contentType)
	if err != nil {
		return fmt.Errorf("imaging: %w", err)
	}

	if dryRun {
		return nil
	}

	id := uuid.New().String()
	fullKey := fmt.Sprintf("users/%s/attachments/%s/full.jpg", row.userID, id)
	thumbKey := fmt.Sprintf("users/%s/attachments/%s/thumb.jpg", row.userID, id)

	if err := store.Put(ctx, fullKey, processed.ContentType, bytes.NewReader(processed.Full)); err != nil {
		return fmt.Errorf("put full: %w", err)
	}
	if err := store.Put(ctx, thumbKey, processed.ContentType, bytes.NewReader(processed.Thumb)); err != nil {
		_ = store.Delete(ctx, fullKey)
		return fmt.Errorf("put thumb: %w", err)
	}

	legacyID := row.id
	_, err = pool.Exec(ctx,
		`INSERT INTO attachments
		   (id, user_id, owner_type, owner_id, storage_key, thumb_key, content_type,
		    full_size, thumb_size, width, height, legacy_id, created_at)
		 VALUES ($1, $2, 'session_end', $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		id, row.userID, row.endID, fullKey, thumbKey, processed.ContentType,
		len(processed.Full), len(processed.Thumb), processed.Width, processed.Height,
		legacyID, row.createdAt,
	)
	if err != nil {
		// Roll back GCS so we don't leave orphan objects.
		_ = store.Delete(ctx, fullKey)
		_ = store.Delete(ctx, thumbKey)
		return fmt.Errorf("insert attachment: %w", err)
	}
	return nil
}
