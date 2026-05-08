package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AttachmentRepo handles the attachments table plus the cross-cutting owner
// verification queries Postgres can't enforce as polymorphic FKs.
type AttachmentRepo struct {
	DB *pgxpool.Pool
}

// AttachmentRow mirrors a row in the attachments table. JSON tags are tuned for
// the API: storage keys and legacy_id are internal and not exposed to clients.
type AttachmentRow struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	OwnerType   string    `json:"owner_type"`
	OwnerID     string    `json:"owner_id"`
	StorageKey  string    `json:"-"`
	ThumbKey    string    `json:"-"`
	ContentType string    `json:"content_type"`
	FullSize    int       `json:"full_size"`
	ThumbSize   int       `json:"thumb_size"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	LegacyID    *string   `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
}

func (r *AttachmentRepo) Insert(ctx context.Context, a *AttachmentRow) error {
	_, err := r.DB.Exec(ctx,
		`INSERT INTO attachments (id, user_id, owner_type, owner_id, storage_key, thumb_key, content_type, full_size, thumb_size, width, height, legacy_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		a.ID, a.UserID, a.OwnerType, a.OwnerID, a.StorageKey, a.ThumbKey, a.ContentType, a.FullSize, a.ThumbSize, a.Width, a.Height, a.LegacyID,
	)
	return err
}

func (r *AttachmentRepo) Get(ctx context.Context, id, userID string) (*AttachmentRow, error) {
	var a AttachmentRow
	err := r.DB.QueryRow(ctx,
		`SELECT id, user_id, owner_type, owner_id, storage_key, thumb_key, content_type, full_size, thumb_size, width, height, legacy_id, created_at
		 FROM attachments WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&a.ID, &a.UserID, &a.OwnerType, &a.OwnerID, &a.StorageKey, &a.ThumbKey, &a.ContentType, &a.FullSize, &a.ThumbSize, &a.Width, &a.Height, &a.LegacyID, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AttachmentRepo) ListByOwner(ctx context.Context, ownerType, ownerID, userID string) ([]AttachmentRow, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT id, user_id, owner_type, owner_id, storage_key, thumb_key, content_type, full_size, thumb_size, width, height, legacy_id, created_at
		 FROM attachments
		 WHERE owner_type = $1 AND owner_id = $2 AND user_id = $3
		 ORDER BY created_at`,
		ownerType, ownerID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AttachmentRow
	for rows.Next() {
		var a AttachmentRow
		if err := rows.Scan(&a.ID, &a.UserID, &a.OwnerType, &a.OwnerID, &a.StorageKey, &a.ThumbKey, &a.ContentType, &a.FullSize, &a.ThumbSize, &a.Width, &a.Height, &a.LegacyID, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// Delete removes the row and returns it so the caller knows which storage keys
// to wipe from object storage. Returns ErrNotFound if the attachment doesn't
// exist or doesn't belong to userID.
func (r *AttachmentRepo) Delete(ctx context.Context, id, userID string) (*AttachmentRow, error) {
	var a AttachmentRow
	err := r.DB.QueryRow(ctx,
		`DELETE FROM attachments WHERE id = $1 AND user_id = $2
		 RETURNING id, user_id, owner_type, owner_id, storage_key, thumb_key, content_type, full_size, thumb_size, width, height, legacy_id, created_at`,
		id, userID,
	).Scan(&a.ID, &a.UserID, &a.OwnerType, &a.OwnerID, &a.StorageKey, &a.ThumbKey, &a.ContentType, &a.FullSize, &a.ThumbSize, &a.Width, &a.Height, &a.LegacyID, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AttachmentRepo) CountByOwner(ctx context.Context, ownerType, ownerID string) (int, error) {
	var n int
	err := r.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM attachments WHERE owner_type = $1 AND owner_id = $2`,
		ownerType, ownerID,
	).Scan(&n)
	return n, err
}

// ── Owner-verification queries ────────────────────────────────────────────
// Polymorphic FKs aren't a thing in Postgres, so the application checks each
// owner_type's parent table to confirm the owner exists and belongs to userID.

func (r *AttachmentRepo) EndBelongsToUser(ctx context.Context, endID, userID string) (bool, error) {
	var ok bool
	err := r.DB.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM ends e
			JOIN scoring_sessions s ON s.id = e.session_id
			WHERE e.id = $1 AND s.user_id = $2
		)`,
		endID, userID,
	).Scan(&ok)
	return ok, err
}

func (r *AttachmentRepo) EquipmentBelongsToUser(ctx context.Context, equipmentID, userID string) (bool, error) {
	var ok bool
	err := r.DB.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM equipment WHERE id = $1 AND user_id = $2)`,
		equipmentID, userID,
	).Scan(&ok)
	return ok, err
}

func (r *AttachmentRepo) SetupBelongsToUser(ctx context.Context, setupID, userID string) (bool, error) {
	var ok bool
	err := r.DB.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM setup_profiles WHERE id = $1 AND user_id = $2)`,
		setupID, userID,
	).Scan(&ok)
	return ok, err
}
