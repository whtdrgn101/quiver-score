package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SightMarkRepo struct {
	DB *pgxpool.Pool
}

// ── Types ─────────────────────────────────────────────────────────────

type SightMarkOut struct {
	ID           string    `json:"id"`
	EquipmentID  *string   `json:"equipment_id"`
	SetupID      *string   `json:"setup_id"`
	Distance     string    `json:"distance"`
	Setting      string    `json:"setting"`
	Notes        *string   `json:"notes"`
	DateRecorded time.Time `json:"date_recorded"`
	CreatedAt    time.Time `json:"created_at"`
}

// ── Methods ───────────────────────────────────────────────────────────

func (r *SightMarkRepo) List(ctx context.Context, userID string, equipmentID, setupID *string) ([]SightMarkOut, error) {
	query := `
		SELECT id, equipment_id, setup_id, distance, setting, notes, date_recorded, created_at
		FROM sight_marks
		WHERE user_id = $1`
	args := []any{userID}
	argN := 2

	if equipmentID != nil {
		query += fmt.Sprintf(" AND equipment_id = $%d", argN)
		args = append(args, *equipmentID)
		argN++
	}

	if setupID != nil {
		query += fmt.Sprintf(" AND setup_id = $%d", argN)
		args = append(args, *setupID)
		argN++
	}

	query += ` ORDER BY distance, date_recorded DESC`

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []SightMarkOut{}
	for rows.Next() {
		var sm SightMarkOut
		if err := rows.Scan(&sm.ID, &sm.EquipmentID, &sm.SetupID,
			&sm.Distance, &sm.Setting, &sm.Notes,
			&sm.DateRecorded, &sm.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, sm)
	}
	return items, rows.Err()
}

func (r *SightMarkRepo) Create(ctx context.Context, id, userID string, equipmentID, setupID *string, distance, setting string, notes *string, dateRecorded time.Time) (*SightMarkOut, error) {
	var sm SightMarkOut
	err := r.DB.QueryRow(ctx, `
		INSERT INTO sight_marks (id, user_id, equipment_id, setup_id, distance, setting, notes, date_recorded)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, equipment_id, setup_id, distance, setting, notes, date_recorded, created_at`,
		id, userID, equipmentID, setupID, distance, setting, notes, dateRecorded,
	).Scan(&sm.ID, &sm.EquipmentID, &sm.SetupID,
		&sm.Distance, &sm.Setting, &sm.Notes,
		&sm.DateRecorded, &sm.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &sm, nil
}

func (r *SightMarkRepo) Exists(ctx context.Context, id, userID string) (bool, error) {
	var exists bool
	err := r.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM sight_marks WHERE id = $1 AND user_id = $2)",
		id, userID,
	).Scan(&exists)
	return exists, err
}

func (r *SightMarkRepo) Update(ctx context.Context, id, userID string, distance, setting *string, notes *string, notesSet bool, dateRecorded *time.Time, dateSet bool, equipmentID *string, eqSet bool, setupID *string, setupSet bool) (*SightMarkOut, error) {
	var sm SightMarkOut
	err := r.DB.QueryRow(ctx, `
		UPDATE sight_marks
		SET distance      = COALESCE($3, distance),
		    setting       = COALESCE($4, setting),
		    notes         = CASE WHEN $5::boolean THEN $6 ELSE notes END,
		    date_recorded = CASE WHEN $7::boolean THEN $8 ELSE date_recorded END,
		    equipment_id  = CASE WHEN $9::boolean THEN $10::uuid ELSE equipment_id END,
		    setup_id      = CASE WHEN $11::boolean THEN $12::uuid ELSE setup_id END
		WHERE id = $1 AND user_id = $2
		RETURNING id, equipment_id, setup_id, distance, setting, notes, date_recorded, created_at`,
		id, userID,
		distance, setting,
		notesSet, notes,
		dateSet, dateRecorded,
		eqSet, equipmentID,
		setupSet, setupID,
	).Scan(&sm.ID, &sm.EquipmentID, &sm.SetupID,
		&sm.Distance, &sm.Setting, &sm.Notes,
		&sm.DateRecorded, &sm.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &sm, nil
}

func (r *SightMarkRepo) Delete(ctx context.Context, id, userID string) (bool, error) {
	tag, err := r.DB.Exec(ctx,
		"DELETE FROM sight_marks WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}
