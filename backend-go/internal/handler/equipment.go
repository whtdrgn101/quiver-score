package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
)

type EquipmentRepository interface {
	List(ctx context.Context, userID string) ([]repository.EquipmentOut, error)
	Create(ctx context.Context, id, userID, category, name string, brand, model *string, specs json.RawMessage, notes *string) (*repository.EquipmentOut, error)
	Get(ctx context.Context, id, userID string) (*repository.EquipmentOut, error)
	Update(ctx context.Context, id, userID string, category, name *string, brand, model *string, brandSet, modelSet bool, specs json.RawMessage, specsSet bool, notes *string, notesSet bool) (*repository.EquipmentOut, error)
	Delete(ctx context.Context, id, userID string) (bool, error)
	Stats(ctx context.Context, userID string) ([]repository.EquipmentUsageOut, error)
}

type EquipmentHandler struct {
	Equipment EquipmentRepository
	Cfg       *config.Config
}

var validCategories = map[string]bool{
	"riser": true, "limbs": true, "arrows": true, "sight": true,
	"stabilizer": true, "rest": true, "release": true, "scope": true,
	"string": true, "other": true,
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

func (h *EquipmentHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	items, err := h.Equipment.List(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusOK, items)
}

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
	id := uuid.New().String()

	e, err := h.Equipment.Create(r.Context(), id, userID, req.Category, req.Name, req.Brand, req.Model, req.Specs, req.Notes)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusCreated, e)
}

func (h *EquipmentHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		Error(w, http.StatusNotFound, "Equipment not found")
		return
	}

	userID := middleware.GetUserID(r.Context())

	e, err := h.Equipment.Get(r.Context(), id, userID)
	if err != nil {
		Error(w, http.StatusNotFound, "Equipment not found")
		return
	}

	JSON(w, http.StatusOK, e)
}

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

	e, err := h.Equipment.Update(r.Context(), id, userID,
		req.Category, req.Name,
		req.Brand, req.Model, req.Brand != nil, req.Model != nil,
		req.Specs, req.Specs != nil,
		req.Notes, req.Notes != nil,
	)
	if err != nil || e == nil {
		Error(w, http.StatusNotFound, "Equipment not found")
		return
	}

	JSON(w, http.StatusOK, e)
}

func (h *EquipmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		Error(w, http.StatusNotFound, "Equipment not found")
		return
	}

	userID := middleware.GetUserID(r.Context())

	ok, err := h.Equipment.Delete(r.Context(), id, userID)
	if err != nil || !ok {
		Error(w, http.StatusNotFound, "Equipment not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *EquipmentHandler) Stats(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	stats, err := h.Equipment.Stats(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusOK, stats)
}
