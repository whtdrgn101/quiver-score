package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
)

type ChallengesRepository interface {
	CreateChallenge(ctx context.Context, challengerID, challengeeID, templateID string, expiresAt *time.Time) (*repository.ChallengeOut, error)
	GetChallenge(ctx context.Context, challengeID string) (*repository.ChallengeOut, error)
	ListChallengesForUser(ctx context.Context, userID string) ([]repository.ChallengeOut, error)
	AcceptChallenge(ctx context.Context, challengeID, userID string) (*repository.ChallengeOut, error)
	DeclineChallenge(ctx context.Context, challengeID, userID string) (*repository.ChallengeOut, error)
	SubmitChallengeScore(ctx context.Context, challengeID, userID, sessionID string) (*repository.ChallengeOut, error)
}

type ChallengesHandler struct {
	Challenges ChallengesRepository
	Cfg        *config.Config
}

func (h *ChallengesHandler) Routes(r chi.Router) {
	r.Use(middleware.RequireAuth(h.Cfg.SecretKey))
	r.Post("/", h.CreateChallenge)
	r.Get("/", h.ListChallenges)
	r.Get("/{challengeID}", h.GetChallenge)
	r.Post("/{challengeID}/accept", h.AcceptChallenge)
	r.Post("/{challengeID}/decline", h.DeclineChallenge)
	r.Post("/{challengeID}/submit-score", h.SubmitChallengeScore)
}

func (h *ChallengesHandler) CreateChallenge(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	var req struct {
		ChallengeeID   string `json:"challengee_id"`
		TemplateID     string `json:"template_id"`
		ExpiresInHours *int   `json:"expires_in_hours"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusUnprocessableEntity, "invalid request body")
		return
	}
	if req.ChallengeeID == "" || req.TemplateID == "" {
		Error(w, http.StatusUnprocessableEntity, "missing challengee_id or template_id")
		return
	}
	if userID == req.ChallengeeID {
		Error(w, http.StatusUnprocessableEntity, "cannot challenge yourself")
		return
	}

	var expiresAt *time.Time
	if req.ExpiresInHours != nil {
		t := time.Now().Add(time.Duration(*req.ExpiresInHours) * time.Hour)
		expiresAt = &t
	}

	chall, err := h.Challenges.CreateChallenge(r.Context(), userID, req.ChallengeeID, req.TemplateID, expiresAt)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusCreated, chall)
}

func (h *ChallengesHandler) ListChallenges(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	list, err := h.Challenges.ListChallengesForUser(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, list)
}

func (h *ChallengesHandler) GetChallenge(w http.ResponseWriter, r *http.Request) {
	challengeID := chi.URLParam(r, "challengeID")
	chall, err := h.Challenges.GetChallenge(r.Context(), challengeID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			Error(w, http.StatusNotFound, "challenge not found")
			return
		}
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, chall)
}

func (h *ChallengesHandler) AcceptChallenge(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	challengeID := chi.URLParam(r, "challengeID")
	chall, err := h.Challenges.AcceptChallenge(r.Context(), challengeID, userID)
	if err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}
	JSON(w, http.StatusOK, chall)
}

func (h *ChallengesHandler) DeclineChallenge(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	challengeID := chi.URLParam(r, "challengeID")
	chall, err := h.Challenges.DeclineChallenge(r.Context(), challengeID, userID)
	if err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}
	JSON(w, http.StatusOK, chall)
}

func (h *ChallengesHandler) SubmitChallengeScore(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	challengeID := chi.URLParam(r, "challengeID")
	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusUnprocessableEntity, "invalid request body")
		return
	}
	if req.SessionID == "" {
		Error(w, http.StatusUnprocessableEntity, "missing session_id")
		return
	}

	chall, err := h.Challenges.SubmitChallengeScore(r.Context(), challengeID, userID, req.SessionID)
	if err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}
	JSON(w, http.StatusOK, chall)
}
