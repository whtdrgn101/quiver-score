package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
)

type SetupsHandler struct {
	DB  *pgxpool.Pool
	Cfg *config.Config
}

func (h *SetupsHandler) Routes(r chi.Router) {
	r.Use(middleware.RequireAuth(h.Cfg.SecretKey))
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/equipment/{equipment_id}", h.AddEquipment)
	r.Delete("/{id}/equipment/{equipment_id}", h.RemoveEquipment)
}

// ── Types ─────────────────────────────────────────────────────────────

type setupProfileOut struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description *string        `json:"description"`
	BraceHeight *float64       `json:"brace_height"`
	Tiller      *float64       `json:"tiller"`
	DrawWeight  *float64       `json:"draw_weight"`
	DrawLength  *float64       `json:"draw_length"`
	ArrowFOC    *float64       `json:"arrow_foc"`
	Equipment   []equipmentOut `json:"equipment"`
	CreatedAt   time.Time      `json:"created_at"`
}

type setupProfileSummary struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    *string   `json:"description"`
	EquipmentCount int       `json:"equipment_count"`
	CreatedAt      time.Time `json:"created_at"`
}

type setupProfileCreate struct {
	Name        string   `json:"name"`
	Description *string  `json:"description"`
	BraceHeight *float64 `json:"brace_height"`
	Tiller      *float64 `json:"tiller"`
	DrawWeight  *float64 `json:"draw_weight"`
	DrawLength  *float64 `json:"draw_length"`
	ArrowFOC    *float64 `json:"arrow_foc"`
}

type setupProfileUpdate struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	BraceHeight *float64 `json:"brace_height"`
	Tiller      *float64 `json:"tiller"`
	DrawWeight  *float64 `json:"draw_weight"`
	DrawLength  *float64 `json:"draw_length"`
	ArrowFOC    *float64 `json:"arrow_foc"`
}

// ── Helpers ───────────────────────────────────────────────────────────

func (h *SetupsHandler) loadSetupWithEquipment(ctx context.Context, setupID, userID string) (*setupProfileOut, error) {
	var s setupProfileOut
	err := h.DB.QueryRow(ctx, `
		SELECT id, name, description, brace_height, tiller, draw_weight, draw_length, arrow_foc, created_at
		FROM setup_profiles
		WHERE id = $1 AND user_id = $2`, setupID, userID,
	).Scan(&s.ID, &s.Name, &s.Description, &s.BraceHeight, &s.Tiller,
		&s.DrawWeight, &s.DrawLength, &s.ArrowFOC, &s.CreatedAt)
	if err != nil {
		return nil, err
	}

	rows, err := h.DB.Query(ctx, `
		SELECT e.id, e.category, e.name, e.brand, e.model, e.specs, e.notes, e.created_at
		FROM equipment e
		JOIN setup_equipment se ON se.equipment_id = e.id
		WHERE se.setup_id = $1
		ORDER BY e.category, e.name`, setupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	s.Equipment = []equipmentOut{}
	for rows.Next() {
		var e equipmentOut
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

// ── List ──────────────────────────────────────────────────────────────

func (h *SetupsHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	rows, err := h.DB.Query(ctx, `
		SELECT sp.id, sp.name, sp.description, sp.created_at,
		       COUNT(se.id) as equipment_count
		FROM setup_profiles sp
		LEFT JOIN setup_equipment se ON se.setup_id = sp.id
		WHERE sp.user_id = $1
		GROUP BY sp.id
		ORDER BY sp.name`, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()

	items := []setupProfileSummary{}
	for rows.Next() {
		var s setupProfileSummary
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.CreatedAt, &s.EquipmentCount); err != nil {
			Error(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		items = append(items, s)
	}
	if err := rows.Err(); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusOK, items)
}

// ── Create ────────────────────────────────────────────────────────────

func (h *SetupsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req setupProfileCreate
	if !Decode(w, r, &req) {
		return
	}

	if req.Name == "" {
		ValidationError(w, "name is required")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	id := uuid.New().String()
	_, err := h.DB.Exec(ctx, `
		INSERT INTO setup_profiles (id, user_id, name, description, brace_height, tiller, draw_weight, draw_length, arrow_foc)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		id, userID, req.Name, req.Description, req.BraceHeight, req.Tiller,
		req.DrawWeight, req.DrawLength, req.ArrowFOC,
	)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	s, err := h.loadSetupWithEquipment(ctx, id, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusCreated, s)
}

// ── Get ───────────────────────────────────────────────────────────────

func (h *SetupsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		Error(w, http.StatusNotFound, "Setup profile not found")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	s, err := h.loadSetupWithEquipment(ctx, id, userID)
	if err != nil {
		Error(w, http.StatusNotFound, "Setup profile not found")
		return
	}

	JSON(w, http.StatusOK, s)
}

// ── Update ────────────────────────────────────────────────────────────

func (h *SetupsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		Error(w, http.StatusNotFound, "Setup profile not found")
		return
	}

	var req setupProfileUpdate
	if !Decode(w, r, &req) {
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	// Verify exists
	var exists bool
	err := h.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM setup_profiles WHERE id = $1 AND user_id = $2)",
		id, userID,
	).Scan(&exists)
	if err != nil || !exists {
		Error(w, http.StatusNotFound, "Setup profile not found")
		return
	}

	tag, err := h.DB.Exec(ctx, `
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
		req.Name,
		req.Description != nil, req.Description,
		req.BraceHeight != nil, req.BraceHeight,
		req.Tiller != nil, req.Tiller,
		req.DrawWeight != nil, req.DrawWeight,
		req.DrawLength != nil, req.DrawLength,
		req.ArrowFOC != nil, req.ArrowFOC,
	)
	if err != nil || tag.RowsAffected() == 0 {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	s, err := h.loadSetupWithEquipment(ctx, id, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusOK, s)
}

// ── Delete ────────────────────────────────────────────────────────────

func (h *SetupsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		Error(w, http.StatusNotFound, "Setup profile not found")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	tx, err := h.DB.Begin(ctx)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer tx.Rollback(ctx)

	// Delete equipment links first
	if _, err := tx.Exec(ctx, "DELETE FROM setup_equipment WHERE setup_id = $1", id); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	tag, err := tx.Exec(ctx, "DELETE FROM setup_profiles WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil || tag.RowsAffected() == 0 {
		Error(w, http.StatusNotFound, "Setup profile not found")
		return
	}

	if err := tx.Commit(ctx); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ── Add Equipment ─────────────────────────────────────────────────────

func (h *SetupsHandler) AddEquipment(w http.ResponseWriter, r *http.Request) {
	setupID := chi.URLParam(r, "id")
	equipmentID := chi.URLParam(r, "equipment_id")

	if _, err := uuid.Parse(setupID); err != nil {
		Error(w, http.StatusNotFound, "Setup profile not found")
		return
	}
	if _, err := uuid.Parse(equipmentID); err != nil {
		Error(w, http.StatusNotFound, "Equipment not found")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	// Verify setup exists and belongs to user
	var setupExists bool
	err := h.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM setup_profiles WHERE id = $1 AND user_id = $2)",
		setupID, userID,
	).Scan(&setupExists)
	if err != nil || !setupExists {
		Error(w, http.StatusNotFound, "Setup profile not found")
		return
	}

	// Verify equipment exists and belongs to user
	var eqExists bool
	err = h.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM equipment WHERE id = $1 AND user_id = $2)",
		equipmentID, userID,
	).Scan(&eqExists)
	if err != nil || !eqExists {
		Error(w, http.StatusNotFound, "Equipment not found")
		return
	}

	// Check not already linked
	var alreadyLinked bool
	err = h.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM setup_equipment WHERE setup_id = $1 AND equipment_id = $2)",
		setupID, equipmentID,
	).Scan(&alreadyLinked)
	if err == nil && alreadyLinked {
		Error(w, http.StatusConflict, "Equipment already linked to this setup")
		return
	}

	linkID := uuid.New().String()
	_, err = h.DB.Exec(ctx,
		"INSERT INTO setup_equipment (id, setup_id, equipment_id) VALUES ($1, $2, $3)",
		linkID, setupID, equipmentID,
	)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	s, err := h.loadSetupWithEquipment(ctx, setupID, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusCreated, s)
}

// ── Remove Equipment ──────────────────────────────────────────────────

func (h *SetupsHandler) RemoveEquipment(w http.ResponseWriter, r *http.Request) {
	setupID := chi.URLParam(r, "id")
	equipmentID := chi.URLParam(r, "equipment_id")

	if _, err := uuid.Parse(setupID); err != nil {
		Error(w, http.StatusNotFound, "Setup profile not found")
		return
	}
	if _, err := uuid.Parse(equipmentID); err != nil {
		Error(w, http.StatusNotFound, "Equipment not found")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	// Verify setup exists and belongs to user
	var setupExists bool
	err := h.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM setup_profiles WHERE id = $1 AND user_id = $2)",
		setupID, userID,
	).Scan(&setupExists)
	if err != nil || !setupExists {
		Error(w, http.StatusNotFound, "Setup profile not found")
		return
	}

	tag, err := h.DB.Exec(ctx,
		"DELETE FROM setup_equipment WHERE setup_id = $1 AND equipment_id = $2",
		setupID, equipmentID,
	)
	if err != nil || tag.RowsAffected() == 0 {
		Error(w, http.StatusNotFound, "Equipment not linked to this setup")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
