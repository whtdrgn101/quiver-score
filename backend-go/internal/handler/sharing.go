package handler

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
)

type SharingHandler struct {
	DB  *pgxpool.Pool
	Cfg *config.Config
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
	ArcherName   string           `json:"archer_name"`
	ArcherAvatar *string          `json:"archer_avatar"`
	TemplateName string           `json:"template_name"`
	Template     *roundTemplateOut `json:"template"`
	TotalScore   int              `json:"total_score"`
	TotalXCount  int              `json:"total_x_count"`
	TotalArrows  int              `json:"total_arrows"`
	Notes        *string          `json:"notes"`
	Location     *string          `json:"location"`
	Weather      *string          `json:"weather"`
	StartedAt    time.Time        `json:"started_at"`
	CompletedAt  *time.Time       `json:"completed_at"`
	Ends         []endOut         `json:"ends"`
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

	var shareToken *string
	err := h.DB.QueryRow(ctx,
		"SELECT share_token FROM scoring_sessions WHERE id = $1 AND user_id = $2",
		sessionID, userID,
	).Scan(&shareToken)
	if err != nil {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}

	// If no share token yet, generate one
	if shareToken == nil || *shareToken == "" {
		token := generateURLSafeToken(16)
		_, err := h.DB.Exec(ctx,
			"UPDATE scoring_sessions SET share_token = $1 WHERE id = $2",
			token, sessionID,
		)
		if err != nil {
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
	ctx := r.Context()

	tag, err := h.DB.Exec(ctx,
		"UPDATE scoring_sessions SET share_token = NULL WHERE id = $1 AND user_id = $2",
		sessionID, userID,
	)
	if err != nil || tag.RowsAffected() == 0 {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}

	JSON(w, http.StatusOK, map[string]string{"detail": "Share link revoked"})
}

// ── Get Shared Session ────────────────────────────────────────────────

func (h *SharingHandler) GetSharedSession(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	ctx := r.Context()

	var sessionID, templateID string
	var totalScore, totalXCount, totalArrows int
	var notes, location, weather *string
	var startedAt time.Time
	var completedAt *time.Time
	var userID string

	err := h.DB.QueryRow(ctx, `
		SELECT ss.id, ss.template_id, ss.total_score, ss.total_x_count, ss.total_arrows,
		       ss.notes, ss.location, ss.weather, ss.started_at, ss.completed_at, ss.user_id
		FROM scoring_sessions ss
		WHERE ss.share_token = $1`, token,
	).Scan(&sessionID, &templateID, &totalScore, &totalXCount, &totalArrows,
		&notes, &location, &weather, &startedAt, &completedAt, &userID)
	if err != nil {
		Error(w, http.StatusNotFound, "Shared session not found")
		return
	}

	// Load archer info
	var username string
	var displayName, avatar *string
	h.DB.QueryRow(ctx,
		"SELECT username, display_name, avatar FROM users WHERE id = $1", userID,
	).Scan(&username, &displayName, &avatar)

	archerName := username
	if displayName != nil && *displayName != "" {
		archerName = *displayName
	}

	// Load template
	var templateName string
	var template *roundTemplateOut
	var t roundTemplateOut
	err = h.DB.QueryRow(ctx, `
		SELECT id, name, organization, description, is_official, created_by
		FROM round_templates WHERE id = $1`, templateID,
	).Scan(&t.ID, &t.Name, &t.Organization, &t.Description, &t.IsOfficial, &t.CreatedBy)
	if err == nil {
		templateName = t.Name
		stages, err := loadStages(ctx, h.DB, t.ID)
		if err == nil {
			t.Stages = stages
		}
		template = &t
	} else {
		templateName = "Unknown"
	}

	// Load ends
	ends, _ := loadEnds(ctx, h.DB, sessionID)

	JSON(w, http.StatusOK, sharedSessionOut{
		ArcherName:   archerName,
		ArcherAvatar: avatar,
		TemplateName: templateName,
		Template:     template,
		TotalScore:   totalScore,
		TotalXCount:  totalXCount,
		TotalArrows:  totalArrows,
		Notes:        notes,
		Location:     location,
		Weather:      weather,
		StartedAt:    startedAt,
		CompletedAt:  completedAt,
		Ends:         ends,
	})
}

// ── Helpers ───────────────────────────────────────────────────────────

func generateURLSafeToken(nBytes int) string {
	b := make([]byte, nBytes)
	rand.Read(b)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
}
