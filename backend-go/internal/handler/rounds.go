package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
)

type RoundRepository interface {
	List(ctx context.Context, userID string) ([]repository.RoundTemplateOut, error)
	Get(ctx context.Context, id string) (*repository.RoundTemplateOut, error)
	Create(ctx context.Context, name, organization string, description *string, userID string, stages []repository.StageParams) (*repository.RoundTemplateOut, error)
	Update(ctx context.Context, id, name, organization string, description *string, stages []repository.StageParams) (*repository.RoundTemplateOut, error)
	Delete(ctx context.Context, id string) error
	GetPermissions(ctx context.Context, id string) (isOfficial bool, createdBy *string, err error)
	HasInProgressSessions(ctx context.Context, templateID string) (bool, error)
	IsMemberOfClub(ctx context.Context, clubID, userID string) (bool, error)
	IsSharedWithClub(ctx context.Context, clubID, templateID string) (bool, error)
	ShareWithClub(ctx context.Context, templateID, clubID, userID string) error
	UnshareFromClub(ctx context.Context, clubID, templateID string) (bool, error)
}

type RoundsHandler struct {
	Rounds RoundRepository
	Cfg    *config.Config
}

func (h *RoundsHandler) Routes(r chi.Router) {
	r.With(middleware.OptionalAuth(h.Cfg.SecretKey)).Get("/", h.List)
	r.Get("/{id}", h.Get)

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(h.Cfg.SecretKey))
		r.Post("/", h.Create)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
		r.Post("/{id}/share", h.Share)
		r.Delete("/{id}/share/{club_id}", h.Unshare)
	})
}

// ── Types ─────────────────────────────────────────────────────────────

type stageCreate struct {
	Name             string         `json:"name"`
	Distance         *string        `json:"distance"`
	NumEnds          int            `json:"num_ends"`
	ArrowsPerEnd     int            `json:"arrows_per_end"`
	AllowedValues    []string       `json:"allowed_values"`
	ValueScoreMap    map[string]int `json:"value_score_map"`
	MaxScorePerArrow int            `json:"max_score_per_arrow"`
}

type roundTemplateCreate struct {
	Name         string        `json:"name"`
	Organization string        `json:"organization"`
	Description  *string       `json:"description"`
	Stages       []stageCreate `json:"stages"`
}

// ── List ──────────────────────────────────────────────────────────────

func (h *RoundsHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	templates, err := h.Rounds.List(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusOK, templates)
}

// ── Get ───────────────────────────────────────────────────────────────

func (h *RoundsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		Error(w, http.StatusNotFound, "Round template not found")
		return
	}

	t, err := h.Rounds.Get(r.Context(), id)
	if err != nil {
		Error(w, http.StatusNotFound, "Round template not found")
		return
	}

	JSON(w, http.StatusOK, t)
}

// ── Create ────────────────────────────────────────────────────────────

func (h *RoundsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req roundTemplateCreate
	if !Decode(w, r, &req) {
		return
	}

	if req.Name == "" {
		ValidationError(w, "name is required")
		return
	}
	if req.Organization == "" {
		ValidationError(w, "organization is required")
		return
	}
	if len(req.Stages) == 0 {
		ValidationError(w, "At least one stage is required")
		return
	}

	userID := middleware.GetUserID(r.Context())

	stages := make([]repository.StageParams, len(req.Stages))
	for i, s := range req.Stages {
		stages[i] = repository.StageParams{
			Name: s.Name, Distance: s.Distance,
			NumEnds: s.NumEnds, ArrowsPerEnd: s.ArrowsPerEnd,
			AllowedValues: s.AllowedValues, ValueScoreMap: s.ValueScoreMap,
			MaxScorePerArrow: s.MaxScorePerArrow,
		}
	}

	t, err := h.Rounds.Create(r.Context(), req.Name, req.Organization, req.Description, userID, stages)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusCreated, t)
}

// ── Update ────────────────────────────────────────────────────────────

func (h *RoundsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		Error(w, http.StatusNotFound, "Round template not found")
		return
	}

	var req roundTemplateCreate
	if !Decode(w, r, &req) {
		return
	}

	if req.Name == "" {
		ValidationError(w, "name is required")
		return
	}
	if req.Organization == "" {
		ValidationError(w, "organization is required")
		return
	}
	if len(req.Stages) == 0 {
		ValidationError(w, "At least one stage is required")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	isOfficial, createdBy, err := h.Rounds.GetPermissions(ctx, id)
	if err != nil {
		Error(w, http.StatusNotFound, "Round template not found")
		return
	}
	if isOfficial {
		Error(w, http.StatusForbidden, "Cannot edit official round templates")
		return
	}
	if createdBy == nil || *createdBy != userID {
		Error(w, http.StatusForbidden, "You can only edit your own custom rounds")
		return
	}

	hasInProgress, err := h.Rounds.HasInProgressSessions(ctx, id)
	if err == nil && hasInProgress {
		ValidationError(w, "Cannot edit a round template while a scoring session is in progress")
		return
	}

	stages := make([]repository.StageParams, len(req.Stages))
	for i, s := range req.Stages {
		stages[i] = repository.StageParams{
			Name: s.Name, Distance: s.Distance,
			NumEnds: s.NumEnds, ArrowsPerEnd: s.ArrowsPerEnd,
			AllowedValues: s.AllowedValues, ValueScoreMap: s.ValueScoreMap,
			MaxScorePerArrow: s.MaxScorePerArrow,
		}
	}

	t, err := h.Rounds.Update(ctx, id, req.Name, req.Organization, req.Description, stages)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusOK, t)
}

// ── Delete ────────────────────────────────────────────────────────────

func (h *RoundsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		Error(w, http.StatusNotFound, "Round template not found")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	isOfficial, createdBy, err := h.Rounds.GetPermissions(ctx, id)
	if err != nil {
		Error(w, http.StatusNotFound, "Round template not found")
		return
	}
	if isOfficial {
		Error(w, http.StatusForbidden, "Cannot delete official round templates")
		return
	}
	if createdBy == nil || *createdBy != userID {
		Error(w, http.StatusForbidden, "You can only delete your own custom rounds")
		return
	}

	if err := h.Rounds.Delete(ctx, id); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ── Share ─────────────────────────────────────────────────────────────

type shareRoundRequest struct {
	ClubID string `json:"club_id"`
}

func (h *RoundsHandler) Share(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		Error(w, http.StatusNotFound, "Round template not found")
		return
	}

	var req shareRoundRequest
	if !Decode(w, r, &req) {
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	isOfficial, createdBy, err := h.Rounds.GetPermissions(ctx, id)
	if err != nil {
		Error(w, http.StatusNotFound, "Round template not found")
		return
	}
	if isOfficial {
		Error(w, http.StatusForbidden, "Cannot share official round templates")
		return
	}
	if createdBy == nil || *createdBy != userID {
		Error(w, http.StatusForbidden, "You can only share your own custom rounds")
		return
	}

	isMember, err := h.Rounds.IsMemberOfClub(ctx, req.ClubID, userID)
	if err != nil || !isMember {
		Error(w, http.StatusUnauthorized, "You are not a member of this club")
		return
	}

	alreadyShared, err := h.Rounds.IsSharedWithClub(ctx, req.ClubID, id)
	if err == nil && alreadyShared {
		Error(w, http.StatusConflict, "Round is already shared with this club")
		return
	}

	if err := h.Rounds.ShareWithClub(ctx, id, req.ClubID, userID); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusCreated, map[string]string{"detail": "Round shared with club"})
}

// ── Unshare ───────────────────────────────────────────────────────────

func (h *RoundsHandler) Unshare(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	clubID := chi.URLParam(r, "club_id")

	if _, err := uuid.Parse(id); err != nil {
		Error(w, http.StatusNotFound, "Round template not found")
		return
	}
	if _, err := uuid.Parse(clubID); err != nil {
		Error(w, http.StatusNotFound, "Share not found")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	_, createdBy, err := h.Rounds.GetPermissions(ctx, id)
	if err != nil {
		Error(w, http.StatusNotFound, "Round template not found")
		return
	}
	if createdBy == nil || *createdBy != userID {
		Error(w, http.StatusForbidden, "You can only unshare your own custom rounds")
		return
	}

	ok, err := h.Rounds.UnshareFromClub(ctx, clubID, id)
	if err != nil || !ok {
		Error(w, http.StatusNotFound, "Share not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
