package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
)

type SharingScoringRepository interface {
	GetShareToken(ctx context.Context, sessionID, userID string) (*string, error)
	SetShareToken(ctx context.Context, sessionID, token string) error
	RevokeShareToken(ctx context.Context, sessionID, userID string) (bool, error)
	GetSharedSession(ctx context.Context, token string) (*repository.SharedSessionData, error)
	LoadEnds(ctx context.Context, sessionID string) ([]repository.EndOut, error)
}

type SharingUserRepository interface {
	GetArcherInfo(ctx context.Context, userID string) (username string, displayName, avatar *string, err error)
}

type SharingRoundRepository interface {
	Get(ctx context.Context, id string) (*repository.RoundTemplateOut, error)
}

type SharingHandler struct {
	Scoring SharingScoringRepository
	Users   SharingUserRepository
	Rounds  SharingRoundRepository
	Cfg     *config.Config
}

func (h *SharingHandler) Routes(r chi.Router) {
	r.With(middleware.RequireAuth(h.Cfg.SecretKey)).Post("/sessions/{id}", h.CreateShareLink)
	r.With(middleware.RequireAuth(h.Cfg.SecretKey)).Delete("/sessions/{id}", h.RevokeShareLink)
	r.Get("/s/{token}", h.GetSharedSession)
}

// ── Types ─────────────────────────────────────────────────────────────

type shareLinkOut struct {
	ShareToken string `json:"share_token"`
	URL        string `json:"url"`
}

type sharedSessionOut struct {
	ArcherName   string                   `json:"archer_name"`
	ArcherAvatar *string                  `json:"archer_avatar"`
	TemplateName string                   `json:"template_name"`
	Template     *repository.RoundTemplateOut `json:"template"`
	TotalScore   int                      `json:"total_score"`
	TotalXCount  int                      `json:"total_x_count"`
	TotalArrows  int                      `json:"total_arrows"`
	Notes        *string                  `json:"notes"`
	Location     *string                  `json:"location"`
	Weather      *string                  `json:"weather"`
	StartedAt    time.Time                `json:"started_at"`
	CompletedAt  *time.Time               `json:"completed_at"`
	Ends         []repository.EndOut      `json:"ends"`
}

// ── Create Share Link ─────────────────────────────────────────────────

func (h *SharingHandler) CreateShareLink(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(sessionID); err != nil {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	shareToken, err := h.Scoring.GetShareToken(ctx, sessionID, userID)
	if err != nil {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}

	if shareToken == nil || *shareToken == "" {
		token := generateURLSafeToken(16)
		if err := h.Scoring.SetShareToken(ctx, sessionID, token); err != nil {
			Error(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		shareToken = &token
	}

	JSON(w, http.StatusOK, shareLinkOut{
		ShareToken: *shareToken,
		URL:        h.Cfg.FrontendURL + "/shared/" + *shareToken,
	})
}

// ── Revoke Share Link ─────────────────────────────────────────────────

func (h *SharingHandler) RevokeShareLink(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(sessionID); err != nil {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}

	userID := middleware.GetUserID(r.Context())

	ok, err := h.Scoring.RevokeShareToken(r.Context(), sessionID, userID)
	if err != nil || !ok {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}

	JSON(w, http.StatusOK, map[string]string{"detail": "Share link revoked"})
}

// ── Get Shared Session ────────────────────────────────────────────────

func (h *SharingHandler) GetSharedSession(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	ctx := r.Context()

	data, err := h.Scoring.GetSharedSession(ctx, token)
	if err != nil {
		Error(w, http.StatusNotFound, "Shared session not found")
		return
	}

	// Load archer info
	username, displayName, avatar, _ := h.Users.GetArcherInfo(ctx, data.UserID)
	archerName := username
	if displayName != nil && *displayName != "" {
		archerName = *displayName
	}

	// Load template
	var templateName string
	var template *repository.RoundTemplateOut
	t, err := h.Rounds.Get(ctx, data.TemplateID)
	if err == nil {
		templateName = t.Name
		template = t
	} else {
		templateName = "Unknown"
	}

	// Load ends
	ends, _ := h.Scoring.LoadEnds(ctx, data.SessionID)

	JSON(w, http.StatusOK, sharedSessionOut{
		ArcherName:   archerName,
		ArcherAvatar: avatar,
		TemplateName: templateName,
		Template:     template,
		TotalScore:   data.TotalScore,
		TotalXCount:  data.TotalXCount,
		TotalArrows:  data.TotalArrows,
		Notes:        data.Notes,
		Location:     data.Location,
		Weather:      data.Weather,
		StartedAt:    data.StartedAt,
		CompletedAt:  data.CompletedAt,
		Ends:         ends,
	})
}

// ── Helpers ───────────────────────────────────────────────────────────

func generateURLSafeToken(nBytes int) string {
	b := make([]byte, nBytes)
	rand.Read(b)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
}
