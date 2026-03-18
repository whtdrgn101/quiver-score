package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
)

type SightMarksHandler struct {
	SightMarks *repository.SightMarkRepo
	Cfg        *config.Config
}

func (h *SightMarksHandler) Routes(r chi.Router) {
	r.Use(middleware.RequireAuth(h.Cfg.SecretKey))
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
}

// ── Types ─────────────────────────────────────────────────────────────

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

	var equipmentID, setupID *string

	if eqID := r.URL.Query().Get("equipment_id"); eqID != "" {
		if _, err := uuid.Parse(eqID); err != nil {
			JSON(w, http.StatusOK, []repository.SightMarkOut{})
			return
		}
		equipmentID = &eqID
	}

	if sID := r.URL.Query().Get("setup_id"); sID != "" {
		if _, err := uuid.Parse(sID); err != nil {
			JSON(w, http.StatusOK, []repository.SightMarkOut{})
			return
		}
		setupID = &sID
	}

	items, err := h.SightMarks.List(r.Context(), userID, equipmentID, setupID)
	if err != nil {
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
	id := uuid.New().String()

	sm, err := h.SightMarks.Create(r.Context(), id, userID, req.EquipmentID, req.SetupID,
		req.Distance, req.Setting, req.Notes, dateRecorded)
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

	exists, err := h.SightMarks.Exists(ctx, id, userID)
	if err != nil || !exists {
		Error(w, http.StatusNotFound, "Sight mark not found")
		return
	}

	var dateRecorded *time.Time
	if req.DateRecorded != nil {
		t, err := time.Parse(time.RFC3339, *req.DateRecorded)
		if err != nil {
			ValidationError(w, "date_recorded must be a valid ISO 8601 datetime")
			return
		}
		dateRecorded = &t
	}

	sm, err := h.SightMarks.Update(ctx, id, userID,
		req.Distance, req.Setting,
		req.Notes, req.Notes != nil,
		dateRecorded, dateRecorded != nil,
		req.EquipmentID, req.EquipmentID != nil,
		req.SetupID, req.SetupID != nil,
	)
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

	ok, err := h.SightMarks.Delete(r.Context(), id, userID)
	if err != nil || !ok {
		Error(w, http.StatusNotFound, "Sight mark not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
