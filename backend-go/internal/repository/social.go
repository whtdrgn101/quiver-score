package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SocialRepo struct {
	DB *pgxpool.Pool
}

// ── Types ───────────────────────────────────────────────────────────────

type FollowOut struct {
	ID                string    `json:"id"`
	FollowerID        string    `json:"follower_id"`
	FollowingID       string    `json:"following_id"`
	FollowerUsername  *string   `json:"follower_username"`
	FollowingUsername *string   `json:"following_username"`
	CreatedAt         time.Time `json:"created_at"`
}

type FeedItemOut struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Username  *string                `json:"username"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
}

// ── Follow ──────────────────────────────────────────────────────────────

func (r *SocialRepo) Follow(ctx context.Context, followerID, followingID string) (*FollowOut, error) {
	// Check target exists
	var exists bool
	err := r.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, followingID).Scan(&exists)
	if err != nil || !exists {
		return nil, ErrNotFound
	}

	// Check not already following
	var alreadyFollowing bool
	err = r.DB.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = $2)`,
		followerID, followingID,
	).Scan(&alreadyFollowing)
	if err != nil {
		return nil, err
	}
	if alreadyFollowing {
		return nil, ErrAlreadyMember
	}

	id := uuid.New().String()
	now := time.Now().UTC()
	_, err = r.DB.Exec(ctx,
		`INSERT INTO follows (id, follower_id, following_id, created_at) VALUES ($1, $2, $3, $4)`,
		id, followerID, followingID, now,
	)
	if err != nil {
		return nil, err
	}

	// Fetch with usernames
	return r.getFollow(ctx, id)
}

func (r *SocialRepo) Unfollow(ctx context.Context, followerID, followingID string) error {
	tag, err := r.DB.Exec(ctx,
		`DELETE FROM follows WHERE follower_id = $1 AND following_id = $2`,
		followerID, followingID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SocialRepo) ListFollowers(ctx context.Context, userID string) ([]FollowOut, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT f.id, f.follower_id, f.following_id,
		       u1.username AS follower_username,
		       u2.username AS following_username,
		       f.created_at
		FROM follows f
		LEFT JOIN users u1 ON u1.id = f.follower_id
		LEFT JOIN users u2 ON u2.id = f.following_id
		WHERE f.following_id = $1
		ORDER BY f.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFollows(rows)
}

func (r *SocialRepo) ListFollowing(ctx context.Context, userID string) ([]FollowOut, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT f.id, f.follower_id, f.following_id,
		       u1.username AS follower_username,
		       u2.username AS following_username,
		       f.created_at
		FROM follows f
		LEFT JOIN users u1 ON u1.id = f.follower_id
		LEFT JOIN users u2 ON u2.id = f.following_id
		WHERE f.follower_id = $1
		ORDER BY f.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFollows(rows)
}

// ── Feed ────────────────────────────────────────────────────────────────

func (r *SocialRepo) GetFeed(ctx context.Context, userID string, limit, offset int) ([]FeedItemOut, error) {
	// Get IDs of users I follow
	rows, err := r.DB.Query(ctx,
		`SELECT following_id FROM follows WHERE follower_id = $1`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var followingIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		followingIDs = append(followingIDs, id)
	}

	if len(followingIDs) == 0 {
		return []FeedItemOut{}, nil
	}

	feedRows, err := r.DB.Query(ctx, `
		SELECT fi.id, fi.user_id, u.username, fi.type, fi.data, fi.created_at
		FROM feed_items fi
		LEFT JOIN users u ON u.id = fi.user_id
		WHERE fi.user_id = ANY($1)
		ORDER BY fi.created_at DESC
		OFFSET $2 LIMIT $3
	`, followingIDs, offset, limit)
	if err != nil {
		return nil, err
	}
	defer feedRows.Close()

	var items []FeedItemOut
	for feedRows.Next() {
		var item FeedItemOut
		if err := feedRows.Scan(&item.ID, &item.UserID, &item.Username, &item.Type, &item.Data, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []FeedItemOut{}
	}
	return items, nil
}

// ── Helpers ─────────────────────────────────────────────────────────────

func (r *SocialRepo) getFollow(ctx context.Context, id string) (*FollowOut, error) {
	var f FollowOut
	err := r.DB.QueryRow(ctx, `
		SELECT f.id, f.follower_id, f.following_id,
		       u1.username AS follower_username,
		       u2.username AS following_username,
		       f.created_at
		FROM follows f
		LEFT JOIN users u1 ON u1.id = f.follower_id
		LEFT JOIN users u2 ON u2.id = f.following_id
		WHERE f.id = $1
	`, id).Scan(&f.ID, &f.FollowerID, &f.FollowingID, &f.FollowerUsername, &f.FollowingUsername, &f.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func scanFollows(rows pgx.Rows) ([]FollowOut, error) {
	var out []FollowOut
	for rows.Next() {
		var f FollowOut
		if err := rows.Scan(&f.ID, &f.FollowerID, &f.FollowingID, &f.FollowerUsername, &f.FollowingUsername, &f.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	if out == nil {
		out = []FollowOut{}
	}
	return out, nil
}
