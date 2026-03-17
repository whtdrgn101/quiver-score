package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
)

type EquipmentHandler struct {
	DB  *pgxpool.Pool
	Cfg *config.Config
}

func (h *EquipmentHandler) Routes(r chi.Router) {
	r.Use(middleware.RequireAuth(h.Cfg.SecretKey))
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/stats", h.Stats)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
}

// ── Types ─────────────────────────────────────────────────────────────

var validCategories = map[string]bool{
	"riser": true, "limbs": true, "arrows": true, "sight": true,
	"stabilizer": true, "rest": true, "release": true, "scope": true,
	"string": true, "other": true,
}

type equipmentOut struct {
	ID        string          `json:"id"`
	Category  string          `json:"category"`
	Name      string          `json:"name"`
	Brand     *string         `json:"brand"`
	Model     *string         `json:"model"`
	Specs     json.RawMessage `json:"specs"`
	Notes     *string         `json:"notes"`
	CreatedAt time.Time       `json:"created_at"`
}

type equipmentCreate struct {
	Category string          `json:"category"`
	Name     string          `json:"name"`
	Brand    *string         `json:"brand"`
	Model    *string         `json:"model"`
	Specs    json.RawMessage `json:"specs"`
	Notes    *string         `json:"notes"`
}

type equipmentUpdate struct {
	Category *string         `json:"category"`
	Name     *string         `json:"name"`
	Brand    *string         `json:"brand"`
	Model    *string         `json:"model"`
	Specs    json.RawMessage `json:"specs"`
	Notes    *string         `json:"notes"`
}

type equipmentUsageOut struct {
	ItemID        string     `json:"item_id"`
	ItemName      string     `json:"item_name"`
	Category      string     `json:"category"`
	SessionsCount int        `json:"sessions_count"`
	TotalArrows   int        `json:"total_arrows"`
	LastUsed      *time.Time `json:"last_used"`
}

// ── List ──────────────────────────────────────────────────────────────

func (h *EquipmentHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	rows, err := h.DB.Query(ctx, `
		SELECT id, category, name, brand, model, specs, notes, created_at
		FROM equipment
		WHERE user_id = $1
		ORDER BY category, name`, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()

	items := []equipmentOut{}
	for rows.Next() {
		var e equipmentOut
		if err := rows.Scan(&e.ID, &e.Category, &e.Name, &e.Brand, &e.Model,
			&e.Specs, &e.Notes, &e.CreatedAt); err != nil {
			Error(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		if e.Specs == nil {
			e.Specs = json.RawMessage("null")
		}
		items = append(items, e)
	}
	if err := rows.Err(); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusOK, items)
}

// ── Create ────────────────────────────────────────────────────────────

func (h *EquipmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req equipmentCreate
	if !Decode(w, r, &req) {
		return
	}

	if req.Name == "" {
		ValidationError(w, "name is required")
		return
	}
	if !validCategories[req.Category] {
		ValidationError(w, "Invalid category")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	id := uuid.New().String()
	specsBytes := req.Specs
	if specsBytes == nil {
		specsBytes = json.RawMessage("null")
	}

	var e equipmentOut
	err := h.DB.QueryRow(ctx, `
		INSERT INTO equipment (id, user_id, category, name, brand, model, specs, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, category, name, brand, model, specs, notes, created_at`,
		id, userID, req.Category, req.Name, req.Brand, req.Model, specsBytes, req.Notes,
	).Scan(&e.ID, &e.Category, &e.Name, &e.Brand, &e.Model, &e.Specs, &e.Notes, &e.CreatedAt)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if e.Specs == nil {
		e.Specs = json.RawMessage("null")
	}

	JSON(w, http.StatusCreated, e)
}

// ── Get ───────────────────────────────────────────────────────────────

func (h *EquipmentHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		Error(w, http.StatusNotFound, "Equipment not found")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	var e equipmentOut
	err := h.DB.QueryRow(ctx, `
		SELECT id, category, name, brand, model, specs, notes, created_at
		FROM equipment
		WHERE id = $1 AND user_id = $2`, id, userID,
	).Scan(&e.ID, &e.Category, &e.Name, &e.Brand, &e.Model, &e.Specs, &e.Notes, &e.CreatedAt)
	if err != nil {
		Error(w, http.StatusNotFound, "Equipment not found")
		return
	}
	if e.Specs == nil {
		e.Specs = json.RawMessage("null")
	}

	JSON(w, http.StatusOK, e)
}

// ── Update ────────────────────────────────────────────────────────────

func (h *EquipmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		Error(w, http.StatusNotFound, "Equipment not found")
		return
	}

	var req equipmentUpdate
	if !Decode(w, r, &req) {
		return
	}

	if req.Category != nil && !validCategories[*req.Category] {
		ValidationError(w, "Invalid category")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	// Verify exists and belongs to user
	var exists bool
	err := h.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM equipment WHERE id = $1 AND user_id = $2)",
		id, userID,
	).Scan(&exists)
	if err != nil || !exists {
		Error(w, http.StatusNotFound, "Equipment not found")
		return
	}

	// Build partial update
	var e equipmentOut
	err = h.DB.QueryRow(ctx, `
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
		req.Category, req.Name,
		req.Brand != nil, req.Brand,
		req.Model != nil, req.Model,
		req.Specs != nil, req.Specs,
		req.Notes != nil, req.Notes,
	).Scan(&e.ID, &e.Category, &e.Name, &e.Brand, &e.Model, &e.Specs, &e.Notes, &e.CreatedAt)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if e.Specs == nil {
		e.Specs = json.RawMessage("null")
	}

	JSON(w, http.StatusOK, e)
}

// ── Delete ────────────────────────────────────────────────────────────

func (h *EquipmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		Error(w, http.StatusNotFound, "Equipment not found")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	tag, err := h.DB.Exec(ctx,
		"DELETE FROM equipment WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil || tag.RowsAffected() == 0 {
		Error(w, http.StatusNotFound, "Equipment not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ── Stats ─────────────────────────────────────────────────────────────

func (h *EquipmentHandler) Stats(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	// Get all user equipment
	eqRows, err := h.DB.Query(ctx, `
		SELECT id, name, category FROM equipment WHERE user_id = $1`, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer eqRows.Close()

	type eqInfo struct {
		id       string
		name     string
		category string
	}
	var allEquipment []eqInfo
	for eqRows.Next() {
		var e eqInfo
		if err := eqRows.Scan(&e.id, &e.name, &e.category); err != nil {
			Error(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		allEquipment = append(allEquipment, e)
	}
	if err := eqRows.Err(); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Get usage stats via sessions linked through setup profiles
	usageRows, err := h.DB.Query(ctx, `
		SELECT se.equipment_id, ss.id, ss.total_arrows,
		       COALESCE(ss.completed_at, ss.started_at) as last_ts
		FROM scoring_sessions ss
		JOIN setup_equipment se ON se.setup_id = ss.setup_profile_id
		WHERE ss.user_id = $1`, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
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
			Error(w, http.StatusInternalServerError, "Internal server error")
			return
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
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	out := make([]equipmentUsageOut, 0, len(allEquipment))
	for _, eq := range allEquipment {
		stat := equipmentUsageOut{
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

	JSON(w, http.StatusOK, out)
}
