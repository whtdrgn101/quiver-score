package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepo struct {
	DB *pgxpool.Pool
}

type NotificationOut struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Read      bool      `json:"read"`
	Link      *string   `json:"link"`
	CreatedAt time.Time `json:"created_at"`
}

func (r *NotificationRepo) List(ctx context.Context, userID string) ([]NotificationOut, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id, type, title, message, read, link, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 50
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []NotificationOut
	for rows.Next() {
		var n NotificationOut
		if err := rows.Scan(&n.ID, &n.Type, &n.Title, &n.Message, &n.Read, &n.Link, &n.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	if out == nil {
		out = []NotificationOut{}
	}
	return out, nil
}

func (r *NotificationRepo) UnreadCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read = false`,
		userID,
	).Scan(&count)
	return count, err
}

func (r *NotificationRepo) MarkRead(ctx context.Context, userID, notificationID string) (*NotificationOut, error) {
	tag, err := r.DB.Exec(ctx,
		`UPDATE notifications SET read = true WHERE id = $1 AND user_id = $2`,
		notificationID, userID,
	)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrNotFound
	}

	var n NotificationOut
	err = r.DB.QueryRow(ctx,
		`SELECT id, type, title, message, read, link, created_at FROM notifications WHERE id = $1`,
		notificationID,
	).Scan(&n.ID, &n.Type, &n.Title, &n.Message, &n.Read, &n.Link, &n.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *NotificationRepo) MarkAllRead(ctx context.Context, userID string) error {
	_, err := r.DB.Exec(ctx,
		`UPDATE notifications SET read = true WHERE user_id = $1 AND read = false`,
		userID,
	)
	return err
}
