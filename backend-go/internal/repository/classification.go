package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ClassificationRepo struct {
	DB *pgxpool.Pool
}

// ── Types ─────────────────────────────────────────────────────────────

type ClassificationRecordOut struct {
	ID             string    `json:"id"`
	System         string    `json:"system"`
	Classification string    `json:"classification"`
	RoundType      string    `json:"round_type"`
	Score          int       `json:"score"`
	AchievedAt     time.Time `json:"achieved_at"`
	SessionID      string    `json:"session_id"`
}

type CurrentClassificationOut struct {
	System         string    `json:"system"`
	Classification string    `json:"classification"`
	RoundType      string    `json:"round_type"`
	Score          int       `json:"score"`
	AchievedAt     time.Time `json:"achieved_at"`
}

// ── Methods ───────────────────────────────────────────────────────────

func (r *ClassificationRepo) List(ctx context.Context, userID string) ([]ClassificationRecordOut, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id, system, classification, round_type, score, achieved_at, session_id
		FROM classification_records
		WHERE user_id = $1
		ORDER BY achieved_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []ClassificationRecordOut{}
	for rows.Next() {
		var c ClassificationRecordOut
		if err := rows.Scan(&c.ID, &c.System, &c.Classification,
			&c.RoundType, &c.Score, &c.AchievedAt, &c.SessionID); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

func (r *ClassificationRepo) Current(ctx context.Context, userID string) ([]CurrentClassificationOut, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id, system, classification, round_type, score, achieved_at, session_id
		FROM classification_records
		WHERE user_id = $1
		ORDER BY achieved_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	best := map[string]CurrentClassificationOut{}
	for rows.Next() {
		var c ClassificationRecordOut
		if err := rows.Scan(&c.ID, &c.System, &c.Classification,
			&c.RoundType, &c.Score, &c.AchievedAt, &c.SessionID); err != nil {
			return nil, err
		}
		key := c.System + ":" + c.RoundType
		if _, exists := best[key]; !exists {
			best[key] = CurrentClassificationOut{
				System:         c.System,
				Classification: c.Classification,
				RoundType:      c.RoundType,
				Score:          c.Score,
				AchievedAt:     c.AchievedAt,
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	items := make([]CurrentClassificationOut, 0, len(best))
	for _, v := range best {
		items = append(items, v)
	}
	return items, nil
}
