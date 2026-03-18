package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
)

type SightMarksHandler struct {
	DB  *pgxpool.Pool
	Cfg *config.Config
}

func (h *SightMarksHandler) Routes(r chi.Router) {
	r.Use(middleware.RequireAuth(h.Cfg.SecretKey))
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
}

// ── Types ─────────────────────────────────────────────────────────────

type sightMarkOut struct {
	ID           string    `json:"id"`
	EquipmentID  *string   `json:"equipment_id"`
	SetupID      *string   `json:"setup_id"`
	Distance     string    `json:"distance"`
	Setting      string    `json:"setting"`
	Notes        *string   `json:"notes"`
	DateRecorded time.Time `json:"date_recorded"`
	CreatedAt    time.Time `json:"created_at"`
}

type sightMarkCreate struct {
	EquipmentID  *string `json:"equipment_id"`
	SetupID      *string `json:"setup_id"`
	Distance     string  `json:"distance"`
	Setting      string  `json:"setting"`
	Notes        *string `json:"notes"`
	DateRecorded *string `json:"date_recorded"`
}

type sightMarkUpdate struct {
	EquipmentID  *string `json:"equipment_id"`
	SetupID      *string `json:"setup_id"`
	Distance     *string `json:"distance"`
	Setting      *string `json:"setting"`
	Notes        *string `json:"notes"`
	DateRecorded *string `json:"date_recorded"`
}

// ── List ──────────────────────────────────────────────────────────────

func (h *SightMarksHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	query := `
		SELECT id, equipment_id, setup_id, distance, setting, notes, date_recorded, created_at
		FROM sight_marks
		WHERE user_id = $1`
	args := []any{userID}
	argN := 2

	if eqID := r.URL.Query().Get("equipment_id"); eqID != "" {
		if _, err := uuid.Parse(eqID); err != nil {
			JSON(w, http.StatusOK, []sightMarkOut{})
			return
		}
		query += ` AND equipment_id = $` + itoa(argN)
		args = append(args, eqID)
		argN++
	}

	if sID := r.URL.Query().Get("setup_id"); sID != "" {
		if _, err := uuid.Parse(sID); err != nil {
			JSON(w, http.StatusOK, []sightMarkOut{})
			return
		}
		query += ` AND setup_id = $` + itoa(argN)
		args = append(args, sID)
		argN++
	}

	query += ` ORDER BY distance, date_recorded DESC`

	rows, err := h.DB.Query(ctx, query, args...)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()

	items := []sightMarkOut{}
	for rows.Next() {
		var sm sightMarkOut
		if err := rows.Scan(&sm.ID, &sm.EquipmentID, &sm.SetupID,
			&sm.Distance, &sm.Setting, &sm.Notes,
			&sm.DateRecorded, &sm.CreatedAt); err != nil {
			Error(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		items = append(items, sm)
	}
	if err := rows.Err(); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusOK, items)
}

// ── Create ────────────────────────────────────────────────────────────

func (h *SightMarksHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req sightMarkCreate
	if !Decode(w, r, &req) {
		return
	}

	if req.Distance == "" || req.Setting == "" {
		ValidationError(w, "distance and setting are required")
		return
	}
	if req.DateRecorded == nil {
		ValidationError(w, "date_recorded is required")
		return
	}

	dateRecorded, err := time.Parse(time.RFC3339, *req.DateRecorded)
	if err != nil {
		ValidationError(w, "date_recorded must be a valid ISO 8601 datetime")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()
	id := uuid.New().String()

	var sm sightMarkOut
	err = h.DB.QueryRow(ctx, `
		INSERT INTO sight_marks (id, user_id, equipment_id, setup_id, distance, setting, notes, date_recorded)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, equipment_id, setup_id, distance, setting, notes, date_recorded, created_at`,
		id, userID, req.EquipmentID, req.SetupID,
		req.Distance, req.Setting, req.Notes, dateRecorded,
	).Scan(&sm.ID, &sm.EquipmentID, &sm.SetupID,
		&sm.Distance, &sm.Setting, &sm.Notes,
		&sm.DateRecorded, &sm.CreatedAt)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusCreated, sm)
}

// ── Update ────────────────────────────────────────────────────────────

func (h *SightMarksHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		Error(w, http.StatusNotFound, "Sight mark not found")
		return
	}

	var req sightMarkUpdate
	if !Decode(w, r, &req) {
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	// Verify exists and belongs to user
	var exists bool
	err := h.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM sight_marks WHERE id = $1 AND user_id = $2)",
		id, userID,
	).Scan(&exists)
	if err != nil || !exists {
		Error(w, http.StatusNotFound, "Sight mark not found")
		return
	}

	// Parse date_recorded if provided
	var dateRecorded *time.Time
	if req.DateRecorded != nil {
		t, err := time.Parse(time.RFC3339, *req.DateRecorded)
		if err != nil {
			ValidationError(w, "date_recorded must be a valid ISO 8601 datetime")
			return
		}
		dateRecorded = &t
	}

	var sm sightMarkOut
	err = h.DB.QueryRow(ctx, `
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
		req.Distance, req.Setting,
		req.Notes != nil, req.Notes,
		dateRecorded != nil, dateRecorded,
		req.EquipmentID != nil, req.EquipmentID,
		req.SetupID != nil, req.SetupID,
	).Scan(&sm.ID, &sm.EquipmentID, &sm.SetupID,
		&sm.Distance, &sm.Setting, &sm.Notes,
		&sm.DateRecorded, &sm.CreatedAt)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusOK, sm)
}

// ── Delete ────────────────────────────────────────────────────────────

func (h *SightMarksHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		Error(w, http.StatusNotFound, "Sight mark not found")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	tag, err := h.DB.Exec(ctx,
		"DELETE FROM sight_marks WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil || tag.RowsAffected() == 0 {
		Error(w, http.StatusNotFound, "Sight mark not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// itoa converts a small int to its string form (avoids importing strconv).
func itoa(n int) string {
	return string(rune('0'+n)) // works for 1-9
}
