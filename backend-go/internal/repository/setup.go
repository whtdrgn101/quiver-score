package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SetupRepo struct {
	DB *pgxpool.Pool
}

// ── Types ─────────────────────────────────────────────────────────────

type SetupProfileOut struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description *string        `json:"description"`
	BraceHeight *float64       `json:"brace_height"`
	Tiller      *float64       `json:"tiller"`
	DrawWeight  *float64       `json:"draw_weight"`
	DrawLength  *float64       `json:"draw_length"`
	ArrowFOC    *float64       `json:"arrow_foc"`
	Equipment   []EquipmentOut `json:"equipment"`
	CreatedAt   time.Time      `json:"created_at"`
}

type SetupProfileSummary struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    *string   `json:"description"`
	EquipmentCount int       `json:"equipment_count"`
	CreatedAt      time.Time `json:"created_at"`
}

// ── Methods ───────────────────────────────────────────────────────────

func (r *SetupRepo) LoadWithEquipment(ctx context.Context, setupID, userID string) (*SetupProfileOut, error) {
	var s SetupProfileOut
	err := r.DB.QueryRow(ctx, `
		SELECT id, name, description, brace_height, tiller, draw_weight, draw_length, arrow_foc, created_at
		FROM setup_profiles
		WHERE id = $1 AND user_id = $2`, setupID, userID,
	).Scan(&s.ID, &s.Name, &s.Description, &s.BraceHeight, &s.Tiller,
		&s.DrawWeight, &s.DrawLength, &s.ArrowFOC, &s.CreatedAt)
	if err != nil {
		return nil, err
	}

	rows, err := r.DB.Query(ctx, `
		SELECT e.id, e.category, e.name, e.brand, e.model, e.specs, e.notes, e.created_at
		FROM equipment e
		JOIN setup_equipment se ON se.equipment_id = e.id
		WHERE se.setup_id = $1
		ORDER BY e.category, e.name`, setupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	s.Equipment = []EquipmentOut{}
	for rows.Next() {
		var e EquipmentOut
		if err := rows.Scan(&e.ID, &e.Category, &e.Name, &e.Brand, &e.Model,
			&e.Specs, &e.Notes, &e.CreatedAt); err != nil {
			return nil, err
		}
		if e.Specs == nil {
			e.Specs = json.RawMessage("null")
		}
		s.Equipment = append(s.Equipment, e)
	}

	return &s, rows.Err()
}

func (r *SetupRepo) List(ctx context.Context, userID string) ([]SetupProfileSummary, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT sp.id, sp.name, sp.description, sp.created_at,
		       COUNT(se.id) as equipment_count
		FROM setup_profiles sp
		LEFT JOIN setup_equipment se ON se.setup_id = sp.id
		WHERE sp.user_id = $1
		GROUP BY sp.id
		ORDER BY sp.name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []SetupProfileSummary{}
	for rows.Next() {
		var s SetupProfileSummary
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.CreatedAt, &s.EquipmentCount); err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

func (r *SetupRepo) Create(ctx context.Context, userID, name string, description *string, braceHeight, tiller, drawWeight, drawLength, arrowFOC *float64) (string, error) {
	id := uuid.New().String()
	_, err := r.DB.Exec(ctx, `
		INSERT INTO setup_profiles (id, user_id, name, description, brace_height, tiller, draw_weight, draw_length, arrow_foc)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		id, userID, name, description, braceHeight, tiller, drawWeight, drawLength, arrowFOC,
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *SetupRepo) Exists(ctx context.Context, id, userID string) (bool, error) {
	var exists bool
	err := r.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM setup_profiles WHERE id = $1 AND user_id = $2)",
		id, userID,
	).Scan(&exists)
	return exists, err
}

func (r *SetupRepo) Update(ctx context.Context, id, userID string, name *string, description *string, descSet bool, braceHeight *float64, bhSet bool, tiller *float64, tillerSet bool, drawWeight *float64, dwSet bool, drawLength *float64, dlSet bool, arrowFOC *float64, focSet bool) (bool, error) {
	tag, err := r.DB.Exec(ctx, `
		UPDATE setup_profiles
		SET name        = COALESCE($3, name),
		    description = CASE WHEN $4::boolean THEN $5 ELSE description END,
		    brace_height = CASE WHEN $6::boolean THEN $7 ELSE brace_height END,
		    tiller       = CASE WHEN $8::boolean THEN $9 ELSE tiller END,
		    draw_weight  = CASE WHEN $10::boolean THEN $11 ELSE draw_weight END,
		    draw_length  = CASE WHEN $12::boolean THEN $13 ELSE draw_length END,
		    arrow_foc    = CASE WHEN $14::boolean THEN $15 ELSE arrow_foc END
		WHERE id = $1 AND user_id = $2`,
		id, userID,
		name,
		descSet, description,
		bhSet, braceHeight,
		tillerSet, tiller,
		dwSet, drawWeight,
		dlSet, drawLength,
		focSet, arrowFOC,
	)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *SetupRepo) Delete(ctx context.Context, id, userID string) (bool, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "DELETE FROM setup_equipment WHERE setup_id = $1", id); err != nil {
		return false, err
	}

	tag, err := tx.Exec(ctx, "DELETE FROM setup_profiles WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		return false, err
	}
	if tag.RowsAffected() == 0 {
		return false, nil
	}

	return true, tx.Commit(ctx)
}

func (r *SetupRepo) EquipmentExists(ctx context.Context, equipmentID, userID string) (bool, error) {
	var exists bool
	err := r.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM equipment WHERE id = $1 AND user_id = $2)",
		equipmentID, userID,
	).Scan(&exists)
	return exists, err
}

func (r *SetupRepo) EquipmentLinked(ctx context.Context, setupID, equipmentID string) (bool, error) {
	var linked bool
	err := r.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM setup_equipment WHERE setup_id = $1 AND equipment_id = $2)",
		setupID, equipmentID,
	).Scan(&linked)
	return linked, err
}

func (r *SetupRepo) AddEquipment(ctx context.Context, setupID, equipmentID string) error {
	linkID := uuid.New().String()
	_, err := r.DB.Exec(ctx,
		"INSERT INTO setup_equipment (id, setup_id, equipment_id) VALUES ($1, $2, $3)",
		linkID, setupID, equipmentID,
	)
	return err
}

func (r *SetupRepo) RemoveEquipment(ctx context.Context, setupID, equipmentID string) (bool, error) {
	tag, err := r.DB.Exec(ctx,
		"DELETE FROM setup_equipment WHERE setup_id = $1 AND equipment_id = $2",
		setupID, equipmentID,
	)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}
