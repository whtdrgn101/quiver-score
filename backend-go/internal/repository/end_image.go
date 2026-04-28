package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EndImageRepo struct {
	DB *pgxpool.Pool
}

type EndImageOut struct {
	ID          string    `json:"id"`
	EndID       string    `json:"end_id"`
	SessionID   string    `json:"session_id"`
	UserID      string    `json:"user_id"`
	ContentType string    `json:"content_type"`
	FileSize    int       `json:"file_size"`
	CreatedAt   time.Time `json:"created_at"`
}

func (r *EndImageRepo) Upload(ctx context.Context, id, endID, sessionID, userID, contentType string, fileSize int, imageData []byte) (*EndImageOut, error) {
	_, err := r.DB.Exec(ctx,
		`INSERT INTO end_images (id, end_id, session_id, user_id, image_data, content_type, file_size)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		id, endID, sessionID, userID, imageData, contentType, fileSize,
	)
	if err != nil {
		return nil, err
	}

	return r.GetMeta(ctx, id, userID)
}

func (r *EndImageRepo) GetMeta(ctx context.Context, imageID, userID string) (*EndImageOut, error) {
	var out EndImageOut
	err := r.DB.QueryRow(ctx,
		`SELECT id, end_id, session_id, user_id, content_type, file_size, created_at
		 FROM end_images WHERE id = $1 AND user_id = $2`,
		imageID, userID,
	).Scan(&out.ID, &out.EndID, &out.SessionID, &out.UserID, &out.ContentType, &out.FileSize, &out.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *EndImageRepo) GetImageData(ctx context.Context, imageID, userID string) ([]byte, string, error) {
	var data []byte
	var contentType string
	err := r.DB.QueryRow(ctx,
		`SELECT image_data, content_type FROM end_images WHERE id = $1 AND user_id = $2`,
		imageID, userID,
	).Scan(&data, &contentType)
	if err != nil {
		return nil, "", err
	}
	return data, contentType, nil
}

func (r *EndImageRepo) ListByEnd(ctx context.Context, endID, userID string) ([]EndImageOut, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT id, end_id, session_id, user_id, content_type, file_size, created_at
		 FROM end_images WHERE end_id = $1 AND user_id = $2
		 ORDER BY created_at`,
		endID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []EndImageOut
	for rows.Next() {
		var img EndImageOut
		if err := rows.Scan(&img.ID, &img.EndID, &img.SessionID, &img.UserID, &img.ContentType, &img.FileSize, &img.CreatedAt); err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	return images, nil
}

func (r *EndImageRepo) ListBySession(ctx context.Context, sessionID, userID string) ([]EndImageOut, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT id, end_id, session_id, user_id, content_type, file_size, created_at
		 FROM end_images WHERE session_id = $1 AND user_id = $2
		 ORDER BY created_at`,
		sessionID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []EndImageOut
	for rows.Next() {
		var img EndImageOut
		if err := rows.Scan(&img.ID, &img.EndID, &img.SessionID, &img.UserID, &img.ContentType, &img.FileSize, &img.CreatedAt); err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	return images, nil
}

func (r *EndImageRepo) Delete(ctx context.Context, imageID, userID string) error {
	tag, err := r.DB.Exec(ctx,
		"DELETE FROM end_images WHERE id = $1 AND user_id = $2",
		imageID, userID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *EndImageRepo) EndBelongsToSession(ctx context.Context, endID, sessionID string) (bool, error) {
	var exists bool
	err := r.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM ends WHERE id = $1 AND session_id = $2)",
		endID, sessionID,
	).Scan(&exists)
	return exists, err
}

func (r *EndImageRepo) SessionBelongsToUser(ctx context.Context, sessionID, userID string) (bool, error) {
	var exists bool
	err := r.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM scoring_sessions WHERE id = $1 AND user_id = $2)",
		sessionID, userID,
	).Scan(&exists)
	return exists, err
}
