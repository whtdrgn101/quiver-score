package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
)

type ClassificationsHandler struct {
	DB  *pgxpool.Pool
	Cfg *config.Config
}

func (h *ClassificationsHandler) Routes(r chi.Router) {
	r.Use(middleware.RequireAuth(h.Cfg.SecretKey))
	r.Get("/", h.List)
	r.Get("/current", h.Current)
}

// ── Types ─────────────────────────────────────────────────────────────

type classificationRecordOut struct {
	ID             string    `json:"id"`
	System         string    `json:"system"`
	Classification string    `json:"classification"`
	RoundType      string    `json:"round_type"`
	Score          int       `json:"score"`
	AchievedAt     time.Time `json:"achieved_at"`
	SessionID      string    `json:"session_id"`
}

type currentClassificationOut struct {
	System         string    `json:"system"`
	Classification string    `json:"classification"`
	RoundType      string    `json:"round_type"`
	Score          int       `json:"score"`
	AchievedAt     time.Time `json:"achieved_at"`
}

// ── List ──────────────────────────────────────────────────────────────

func (h *ClassificationsHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	rows, err := h.DB.Query(ctx, `
		SELECT id, system, classification, round_type, score, achieved_at, session_id
		FROM classification_records
		WHERE user_id = $1
		ORDER BY achieved_at DESC`, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()

	items := []classificationRecordOut{}
	for rows.Next() {
		var c classificationRecordOut
		if err := rows.Scan(&c.ID, &c.System, &c.Classification,
			&c.RoundType, &c.Score, &c.AchievedAt, &c.SessionID); err != nil {
			Error(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		items = append(items, c)
	}
	if err := rows.Err(); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusOK, items)
}

// ── Current ───────────────────────────────────────────────────────────

func (h *ClassificationsHandler) Current(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	rows, err := h.DB.Query(ctx, `
		SELECT id, system, classification, round_type, score, achieved_at, session_id
		FROM classification_records
		WHERE user_id = $1
		ORDER BY achieved_at DESC`, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()

	// Get the most recent classification per system:round_type
	best := map[string]currentClassificationOut{}
	for rows.Next() {
		var c classificationRecordOut
		if err := rows.Scan(&c.ID, &c.System, &c.Classification,
			&c.RoundType, &c.Score, &c.AchievedAt, &c.SessionID); err != nil {
			Error(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		key := c.System + ":" + c.RoundType
		if _, exists := best[key]; !exists {
			best[key] = currentClassificationOut{
				System:         c.System,
				Classification: c.Classification,
				RoundType:      c.RoundType,
				Score:          c.Score,
				AchievedAt:     c.AchievedAt,
			}
		}
	}
	if err := rows.Err(); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	items := make([]currentClassificationOut, 0, len(best))
	for _, v := range best {
		items = append(items, v)
	}

	JSON(w, http.StatusOK, items)
}
