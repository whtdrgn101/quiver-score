package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
)

type SetupsHandler struct {
	Setups *repository.SetupRepo
	Cfg    *config.Config
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

// ── List ──────────────────────────────────────────────────────────────

func (h *SetupsHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	items, err := h.Setups.List(r.Context(), userID)
	if err != nil {
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

	id, err := h.Setups.Create(ctx, userID, req.Name, req.Description,
		req.BraceHeight, req.Tiller, req.DrawWeight, req.DrawLength, req.ArrowFOC)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	s, err := h.Setups.LoadWithEquipment(ctx, id, userID)
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

	s, err := h.Setups.LoadWithEquipment(r.Context(), id, userID)
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

	exists, err := h.Setups.Exists(ctx, id, userID)
	if err != nil || !exists {
		Error(w, http.StatusNotFound, "Setup profile not found")
		return
	}

	ok, err := h.Setups.Update(ctx, id, userID,
		req.Name,
		req.Description, req.Description != nil,
		req.BraceHeight, req.BraceHeight != nil,
		req.Tiller, req.Tiller != nil,
		req.DrawWeight, req.DrawWeight != nil,
		req.DrawLength, req.DrawLength != nil,
		req.ArrowFOC, req.ArrowFOC != nil,
	)
	if err != nil || !ok {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	s, err := h.Setups.LoadWithEquipment(ctx, id, userID)
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

	ok, err := h.Setups.Delete(r.Context(), id, userID)
	if err != nil || !ok {
		Error(w, http.StatusNotFound, "Setup profile not found")
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

	setupExists, err := h.Setups.Exists(ctx, setupID, userID)
	if err != nil || !setupExists {
		Error(w, http.StatusNotFound, "Setup profile not found")
		return
	}

	eqExists, err := h.Setups.EquipmentExists(ctx, equipmentID, userID)
	if err != nil || !eqExists {
		Error(w, http.StatusNotFound, "Equipment not found")
		return
	}

	linked, err := h.Setups.EquipmentLinked(ctx, setupID, equipmentID)
	if err == nil && linked {
		Error(w, http.StatusConflict, "Equipment already linked to this setup")
		return
	}

	if err := h.Setups.AddEquipment(ctx, setupID, equipmentID); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	s, err := h.Setups.LoadWithEquipment(ctx, setupID, userID)
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

	setupExists, err := h.Setups.Exists(ctx, setupID, userID)
	if err != nil || !setupExists {
		Error(w, http.StatusNotFound, "Setup profile not found")
		return
	}

	ok, err := h.Setups.RemoveEquipment(ctx, setupID, equipmentID)
	if err != nil || !ok {
		Error(w, http.StatusNotFound, "Equipment not linked to this setup")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
