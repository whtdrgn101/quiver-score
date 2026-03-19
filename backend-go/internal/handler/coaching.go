package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
)

type CoachingRepository interface {
	Invite(ctx context.Context, coachID, athleteUsername string) (*repository.CoachAthleteLinkOut, error)
	Respond(ctx context.Context, athleteID, linkID string, accept bool) (*repository.CoachAthleteLinkOut, error)
	ListAthletes(ctx context.Context, coachID string) ([]repository.CoachAthleteLinkOut, error)
	ListCoaches(ctx context.Context, athleteID string) ([]repository.CoachAthleteLinkOut, error)
	GetAthleteSessions(ctx context.Context, coachID, athleteID string) ([]repository.AthleteSessionOut, error)
	CheckSessionAccess(ctx context.Context, userID, sessionID string) (string, error)
	AddAnnotation(ctx context.Context, sessionID, authorID string, endNumber, arrowNumber *int, text string) (*repository.AnnotationOut, error)
	ListAnnotations(ctx context.Context, sessionID string) ([]repository.AnnotationOut, error)
}

type CoachingHandler struct {
	Coaching CoachingRepository
	Cfg      *config.Config
}

func (h *CoachingHandler) Routes(r chi.Router) {
	r.Use(middleware.RequireAuth(h.Cfg.SecretKey))
	r.Post("/invite", h.Invite)
	r.Post("/respond", h.Respond)
	r.Get("/athletes", h.ListAthletes)
	r.Get("/coaches", h.ListCoaches)
	r.Get("/athletes/{athleteID}/sessions", h.ViewAthleteSessions)
	r.Post("/sessions/{sessionID}/annotations", h.AddAnnotation)
	r.Get("/sessions/{sessionID}/annotations", h.ListAnnotations)
}

// ── Invite ──────────────────────────────────────────────────────────────

type inviteRequest struct {
	AthleteUsername string `json:"athlete_username"`
}

func (h *CoachingHandler) Invite(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var body inviteRequest
	if !Decode(w, r, &body) {
		return
	}
	if body.AthleteUsername == "" {
		ValidationError(w, "athlete_username is required")
		return
	}

	link, err := h.Coaching.Invite(r.Context(), userID, body.AthleteUsername)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			Error(w, http.StatusNotFound, "User not found")
			return
		}
		if errors.Is(err, repository.ErrValidation) {
			Error(w, http.StatusUnprocessableEntity, "Cannot coach yourself")
			return
		}
		if errors.Is(err, repository.ErrAlreadyMember) {
			Error(w, http.StatusConflict, "Link already exists")
			return
		}
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusCreated, link)
}

// ── Respond ─────────────────────────────────────────────────────────────

type respondRequest struct {
	LinkID string `json:"link_id"`
	Accept bool   `json:"accept"`
}

func (h *CoachingHandler) Respond(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var body respondRequest
	if !Decode(w, r, &body) {
		return
	}

	link, err := h.Coaching.Respond(r.Context(), userID, body.LinkID, body.Accept)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			Error(w, http.StatusNotFound, "Invite not found")
			return
		}
		if errors.Is(err, repository.ErrValidation) {
			Error(w, http.StatusUnprocessableEntity, "Invite already responded to")
			return
		}
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, link)
}

// ── List Athletes / Coaches ─────────────────────────────────────────────

func (h *CoachingHandler) ListAthletes(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	links, err := h.Coaching.ListAthletes(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, links)
}

func (h *CoachingHandler) ListCoaches(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	links, err := h.Coaching.ListCoaches(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, links)
}

// ── Athlete Sessions ────────────────────────────────────────────────────

func (h *CoachingHandler) ViewAthleteSessions(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	athleteID := chi.URLParam(r, "athleteID")

	sessions, err := h.Coaching.GetAthleteSessions(r.Context(), userID, athleteID)
	if err != nil {
		if errors.Is(err, repository.ErrForbidden) {
			Error(w, http.StatusForbidden, "No active coaching link")
			return
		}
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, sessions)
}

// ── Annotations ─────────────────────────────────────────────────────────

type annotationCreate struct {
	EndNumber   *int   `json:"end_number"`
	ArrowNumber *int   `json:"arrow_number"`
	Text        string `json:"text"`
}

func (h *CoachingHandler) AddAnnotation(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	sessionID := chi.URLParam(r, "sessionID")

	var body annotationCreate
	if !Decode(w, r, &body) {
		return
	}
	if body.Text == "" || len(body.Text) > 2000 {
		ValidationError(w, "text must be 1-2000 characters")
		return
	}

	// Check session access
	_, err := h.Coaching.CheckSessionAccess(r.Context(), userID, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			Error(w, http.StatusNotFound, "Session not found")
			return
		}
		if errors.Is(err, repository.ErrForbidden) {
			Error(w, http.StatusForbidden, "Not authorized to annotate this session")
			return
		}
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	annotation, err := h.Coaching.AddAnnotation(r.Context(), sessionID, userID, body.EndNumber, body.ArrowNumber, body.Text)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusCreated, annotation)
}

func (h *CoachingHandler) ListAnnotations(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	sessionID := chi.URLParam(r, "sessionID")

	// Check session access
	_, err := h.Coaching.CheckSessionAccess(r.Context(), userID, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			Error(w, http.StatusNotFound, "Session not found")
			return
		}
		if errors.Is(err, repository.ErrForbidden) {
			Error(w, http.StatusForbidden, "Not authorized to view annotations")
			return
		}
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	annotations, err := h.Coaching.ListAnnotations(r.Context(), sessionID)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, annotations)
}
