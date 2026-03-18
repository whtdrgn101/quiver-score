package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EquipmentRepo struct {
	DB *pgxpool.Pool
}

// ── Types ─────────────────────────────────────────────────────────────

type EquipmentOut struct {
	ID        string          `json:"id"`
	Category  string          `json:"category"`
	Name      string          `json:"name"`
	Brand     *string         `json:"brand"`
	Model     *string         `json:"model"`
	Specs     json.RawMessage `json:"specs"`
	Notes     *string         `json:"notes"`
	CreatedAt time.Time       `json:"created_at"`
}

type EquipmentUsageOut struct {
	ItemID        string     `json:"item_id"`
	ItemName      string     `json:"item_name"`
	Category      string     `json:"category"`
	SessionsCount int        `json:"sessions_count"`
	TotalArrows   int        `json:"total_arrows"`
	LastUsed      *time.Time `json:"last_used"`
}

// ── Methods ───────────────────────────────────────────────────────────

func (r *EquipmentRepo) List(ctx context.Context, userID string) ([]EquipmentOut, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id, category, name, brand, model, specs, notes, created_at
		FROM equipment
		WHERE user_id = $1
		ORDER BY category, name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []EquipmentOut{}
	for rows.Next() {
		var e EquipmentOut
		if err := rows.Scan(&e.ID, &e.Category, &e.Name, &e.Brand, &e.Model,
			&e.Specs, &e.Notes, &e.CreatedAt); err != nil {
			return nil, err
		}
		if e.Specs == nil {
			e.Specs = json.RawMessage("null")
		}
		items = append(items, e)
	}
	return items, rows.Err()
}

func (r *EquipmentRepo) Create(ctx context.Context, id, userID, category, name string, brand, model *string, specs json.RawMessage, notes *string) (*EquipmentOut, error) {
	if specs == nil {
		specs = json.RawMessage("null")
	}
	var e EquipmentOut
	err := r.DB.QueryRow(ctx, `
		INSERT INTO equipment (id, user_id, category, name, brand, model, specs, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, category, name, brand, model, specs, notes, created_at`,
		id, userID, category, name, brand, model, specs, notes,
	).Scan(&e.ID, &e.Category, &e.Name, &e.Brand, &e.Model, &e.Specs, &e.Notes, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	if e.Specs == nil {
		e.Specs = json.RawMessage("null")
	}
	return &e, nil
}

func (r *EquipmentRepo) Get(ctx context.Context, id, userID string) (*EquipmentOut, error) {
	var e EquipmentOut
	err := r.DB.QueryRow(ctx, `
		SELECT id, category, name, brand, model, specs, notes, created_at
		FROM equipment
		WHERE id = $1 AND user_id = $2`, id, userID,
	).Scan(&e.ID, &e.Category, &e.Name, &e.Brand, &e.Model, &e.Specs, &e.Notes, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	if e.Specs == nil {
		e.Specs = json.RawMessage("null")
	}
	return &e, nil
}

func (r *EquipmentRepo) Update(ctx context.Context, id, userID string, category, name *string, brand, model *string, brandSet, modelSet bool, specs json.RawMessage, specsSet bool, notes *string, notesSet bool) (*EquipmentOut, error) {
	// Verify exists
	var exists bool
	err := r.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM equipment WHERE id = $1 AND user_id = $2)",
		id, userID,
	).Scan(&exists)
	if err != nil || !exists {
		return nil, err
	}

	var e EquipmentOut
	err = r.DB.QueryRow(ctx, `
		UPDATE equipment
		SET category = COALESCE($3, category),
		    name     = COALESCE($4, name),
		    brand    = CASE WHEN $5::boolean THEN $6 ELSE brand END,
		    model    = CASE WHEN $7::boolean THEN $8 ELSE model END,
		    specs    = CASE WHEN $9::boolean THEN $10 ELSE specs END,
		    notes    = CASE WHEN $11::boolean THEN $12 ELSE notes END
		WHERE id = $1 AND user_id = $2
		RETURNING id, category, name, brand, model, specs, notes, created_at`,
		id, userID,
		category, name,
		brandSet, brand,
		modelSet, model,
		specsSet, specs,
		notesSet, notes,
	).Scan(&e.ID, &e.Category, &e.Name, &e.Brand, &e.Model, &e.Specs, &e.Notes, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	if e.Specs == nil {
		e.Specs = json.RawMessage("null")
	}
	return &e, nil
}

func (r *EquipmentRepo) Delete(ctx context.Context, id, userID string) (bool, error) {
	tag, err := r.DB.Exec(ctx,
		"DELETE FROM equipment WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *EquipmentRepo) Stats(ctx context.Context, userID string) ([]EquipmentUsageOut, error) {
	eqRows, err := r.DB.Query(ctx,
		`SELECT id, name, category FROM equipment WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer eqRows.Close()

	type eqInfo struct {
		id, name, category string
	}
	var allEquipment []eqInfo
	for eqRows.Next() {
		var e eqInfo
		if err := eqRows.Scan(&e.id, &e.name, &e.category); err != nil {
			return nil, err
		}
		allEquipment = append(allEquipment, e)
	}
	if err := eqRows.Err(); err != nil {
		return nil, err
	}

	usageRows, err := r.DB.Query(ctx, `
		SELECT se.equipment_id, ss.id, ss.total_arrows,
		       COALESCE(ss.completed_at, ss.started_at) as last_ts
		FROM scoring_sessions ss
		JOIN setup_equipment se ON se.setup_id = ss.setup_profile_id
		WHERE ss.user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer usageRows.Close()

	type usageData struct {
		sessions map[string]bool
		arrows   int
		lastUsed *time.Time
	}
	usage := map[string]*usageData{}
	for usageRows.Next() {
		var eqID, sessionID string
		var arrows int
		var lastTS time.Time
		if err := usageRows.Scan(&eqID, &sessionID, &arrows, &lastTS); err != nil {
			return nil, err
		}
		u, ok := usage[eqID]
		if !ok {
			u = &usageData{sessions: map[string]bool{}}
			usage[eqID] = u
		}
		u.sessions[sessionID] = true
		u.arrows += arrows
		if u.lastUsed == nil || lastTS.After(*u.lastUsed) {
			t := lastTS
			u.lastUsed = &t
		}
	}
	if err := usageRows.Err(); err != nil {
		return nil, err
	}

	out := make([]EquipmentUsageOut, 0, len(allEquipment))
	for _, eq := range allEquipment {
		stat := EquipmentUsageOut{
			ItemID:   eq.id,
			ItemName: eq.name,
			Category: eq.category,
		}
		if u, ok := usage[eq.id]; ok {
			stat.SessionsCount = len(u.sessions)
			stat.TotalArrows = u.arrows
			stat.LastUsed = u.lastUsed
		}
		out = append(out, stat)
	}

	return out, nil
}
