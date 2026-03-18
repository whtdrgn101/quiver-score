package handler

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
)

type ScoringHandler struct {
	DB  *pgxpool.Pool
	Cfg *config.Config
}

func (h *ScoringHandler) Routes(r chi.Router) {
	r.Use(middleware.RequireAuth(h.Cfg.SecretKey))

	r.Post("/", h.Create)
	r.Get("/", h.List)

	// Static paths MUST come before /{id} parameter routes
	r.Get("/export", h.ExportBulk)
	r.Get("/stats", h.Stats)
	r.Get("/personal-records", h.PersonalRecords)
	r.Get("/trends", h.Trends)

	r.Get("/{id}", h.Get)
	r.Delete("/{id}", h.Delete)
	r.Get("/{id}/export", h.ExportSingle)
	r.Post("/{id}/ends", h.SubmitEnd)
	r.Delete("/{id}/ends/last", h.UndoLastEnd)
	r.Post("/{id}/complete", h.Complete)
	r.Post("/{id}/abandon", h.Abandon)
}

// ── Types ─────────────────────────────────────────────────────────────

type arrowOut struct {
	ID           string   `json:"id"`
	ArrowNumber  int      `json:"arrow_number"`
	ScoreValue   string   `json:"score_value"`
	ScoreNumeric int      `json:"score_numeric"`
	XPos         *float64 `json:"x_pos"`
	YPos         *float64 `json:"y_pos"`
}

type endOut struct {
	ID        string     `json:"id"`
	EndNumber int        `json:"end_number"`
	EndTotal  int        `json:"end_total"`
	StageID   *string    `json:"stage_id"`
	Arrows    []arrowOut `json:"arrows"`
	CreatedAt time.Time  `json:"created_at"`
}

type sessionOut struct {
	ID               string            `json:"id"`
	TemplateID       string            `json:"template_id"`
	SetupProfileID   *string           `json:"setup_profile_id"`
	SetupProfileName *string           `json:"setup_profile_name"`
	Template         *roundTemplateOut `json:"template"`
	Status           string            `json:"status"`
	TotalScore       int               `json:"total_score"`
	TotalXCount      int               `json:"total_x_count"`
	TotalArrows      int               `json:"total_arrows"`
	Notes            *string           `json:"notes"`
	Location         *string           `json:"location"`
	Weather          *string           `json:"weather"`
	ShareToken       *string           `json:"share_token"`
	IsPersonalBest   bool              `json:"is_personal_best"`
	StartedAt        time.Time         `json:"started_at"`
	CompletedAt      *time.Time        `json:"completed_at"`
	Ends             []endOut          `json:"ends"`
}

type sessionSummary struct {
	ID               string     `json:"id"`
	TemplateID       string     `json:"template_id"`
	SetupProfileID   *string    `json:"setup_profile_id"`
	SetupProfileName *string    `json:"setup_profile_name"`
	Status           string     `json:"status"`
	TotalScore       int        `json:"total_score"`
	TotalXCount      int        `json:"total_x_count"`
	TotalArrows      int        `json:"total_arrows"`
	StartedAt        time.Time  `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at"`
	TemplateName     *string    `json:"template_name"`
}

type sessionCreateReq struct {
	TemplateID     string  `json:"template_id"`
	SetupProfileID *string `json:"setup_profile_id"`
	Notes          *string `json:"notes"`
	Location       *string `json:"location"`
	Weather        *string `json:"weather"`
}

type sessionCompleteReq struct {
	Notes    *string `json:"notes"`
	Location *string `json:"location"`
	Weather  *string `json:"weather"`
}

type arrowIn struct {
	ScoreValue string   `json:"score_value"`
	XPos       *float64 `json:"x_pos"`
	YPos       *float64 `json:"y_pos"`
}

type endIn struct {
	StageID string   `json:"stage_id"`
	Arrows  []arrowIn `json:"arrows"`
}

type personalRecordOut struct {
	TemplateName string    `json:"template_name"`
	Score        int       `json:"score"`
	MaxScore     int       `json:"max_score"`
	AchievedAt   time.Time `json:"achieved_at"`
	SessionID    string    `json:"session_id"`
}

type roundTypeAvg struct {
	TemplateName string  `json:"template_name"`
	AvgScore     float64 `json:"avg_score"`
	Count        int     `json:"count"`
}

type recentTrendItem struct {
	Score        int       `json:"score"`
	MaxScore     int       `json:"max_score"`
	TemplateName string    `json:"template_name"`
	Date         time.Time `json:"date"`
}

type trendDataItem struct {
	SessionID    string    `json:"session_id"`
	TemplateName string    `json:"template_name"`
	TotalScore   int       `json:"total_score"`
	MaxScore     int       `json:"max_score"`
	Percentage   float64   `json:"percentage"`
	CompletedAt  time.Time `json:"completed_at"`
}

type statsOut struct {
	TotalSessions     int                 `json:"total_sessions"`
	CompletedSessions int                 `json:"completed_sessions"`
	TotalArrows       int                 `json:"total_arrows"`
	TotalXCount       int                 `json:"total_x_count"`
	PersonalBestScore *int                `json:"personal_best_score"`
	PersonalBestTmpl  *string             `json:"personal_best_template"`
	AvgByRoundType    []roundTypeAvg      `json:"avg_by_round_type"`
	RecentTrend       []recentTrendItem   `json:"recent_trend"`
	PersonalRecords   []personalRecordOut `json:"personal_records"`
}

// ── Create ────────────────────────────────────────────────────────────

func (h *ScoringHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req sessionCreateReq
	if !Decode(w, r, &req) {
		return
	}

	if req.TemplateID == "" {
		ValidationError(w, "template_id is required")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	// Validate setup profile if provided
	if req.SetupProfileID != nil && *req.SetupProfileID != "" {
		var exists bool
		err := h.DB.QueryRow(ctx,
			"SELECT EXISTS(SELECT 1 FROM setup_profiles WHERE id = $1 AND user_id = $2)",
			*req.SetupProfileID, userID,
		).Scan(&exists)
		if err != nil || !exists {
			Error(w, http.StatusNotFound, "Setup profile not found")
			return
		}
	}

	id := uuid.New().String()
	now := time.Now().UTC()

	_, err := h.DB.Exec(ctx, `
		INSERT INTO scoring_sessions (id, user_id, template_id, setup_profile_id,
			notes, location, weather, status, total_score, total_x_count, total_arrows, started_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'in_progress', 0, 0, 0, $8)`,
		id, userID, req.TemplateID, req.SetupProfileID,
		req.Notes, req.Location, req.Weather, now,
	)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	out, err := loadSessionOut(ctx, h.DB, id, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusCreated, out)
}

// ── List ──────────────────────────────────────────────────────────────

func (h *ScoringHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	query := `
		SELECT ss.id, ss.template_id, ss.setup_profile_id, ss.status,
		       ss.total_score, ss.total_x_count, ss.total_arrows,
		       ss.started_at, ss.completed_at,
		       rt.name AS template_name,
		       sp.name AS setup_profile_name
		FROM scoring_sessions ss
		LEFT JOIN round_templates rt ON rt.id = ss.template_id
		LEFT JOIN setup_profiles sp ON sp.id = ss.setup_profile_id
		WHERE ss.user_id = $1`
	args := []any{userID}
	argN := 2

	if tid := r.URL.Query().Get("template_id"); tid != "" {
		query += fmt.Sprintf(" AND ss.template_id = $%d", argN)
		args = append(args, tid)
		argN++
	}
	if df := r.URL.Query().Get("date_from"); df != "" {
		query += fmt.Sprintf(" AND ss.started_at >= $%d::date", argN)
		args = append(args, df)
		argN++
	}
	if dt := r.URL.Query().Get("date_to"); dt != "" {
		query += fmt.Sprintf(" AND ss.started_at <= ($%d::date + interval '1 day' - interval '1 second')", argN)
		args = append(args, dt)
		argN++
	}
	if search := r.URL.Query().Get("search"); search != "" {
		pattern := "%" + search + "%"
		query += fmt.Sprintf(" AND (ss.notes ILIKE $%d OR ss.location ILIKE $%d)", argN, argN)
		args = append(args, pattern)
		argN++
	}

	query += " ORDER BY ss.started_at DESC"

	rows, err := h.DB.Query(ctx, query, args...)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()

	items := []sessionSummary{}
	for rows.Next() {
		var s sessionSummary
		if err := rows.Scan(&s.ID, &s.TemplateID, &s.SetupProfileID, &s.Status,
			&s.TotalScore, &s.TotalXCount, &s.TotalArrows,
			&s.StartedAt, &s.CompletedAt,
			&s.TemplateName, &s.SetupProfileName); err != nil {
			Error(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		items = append(items, s)
	}
	if err := rows.Err(); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusOK, items)
}

// ── Get ───────────────────────────────────────────────────────────────

func (h *ScoringHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	out, err := loadSessionOut(ctx, h.DB, id, userID)
	if err != nil {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}

	// Check personal best
	var isPB bool
	h.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM personal_records WHERE user_id = $1 AND session_id = $2)",
		userID, id,
	).Scan(&isPB)
	out.IsPersonalBest = isPB

	JSON(w, http.StatusOK, out)
}

// ── Submit End ────────────────────────────────────────────────────────

func (h *ScoringHandler) SubmitEnd(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(sessionID); err != nil {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}

	var req endIn
	if !Decode(w, r, &req) {
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	// Load session
	var status string
	err := h.DB.QueryRow(ctx,
		"SELECT status FROM scoring_sessions WHERE id = $1 AND user_id = $2",
		sessionID, userID,
	).Scan(&status)
	if err != nil {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}
	if status != "in_progress" {
		ValidationError(w, "Session is not in progress")
		return
	}

	// Load stage for validation
	var arrowsPerEnd int
	var allowedJSON, scoreMapJSON []byte
	err = h.DB.QueryRow(ctx,
		"SELECT arrows_per_end, allowed_values, value_score_map FROM round_template_stages WHERE id = $1",
		req.StageID,
	).Scan(&arrowsPerEnd, &allowedJSON, &scoreMapJSON)
	if err != nil {
		Error(w, http.StatusNotFound, "Stage not found")
		return
	}

	var allowedValues []string
	var valueScoreMap map[string]int
	json.Unmarshal(allowedJSON, &allowedValues)
	json.Unmarshal(scoreMapJSON, &valueScoreMap)

	// Validate arrow count
	if len(req.Arrows) != arrowsPerEnd {
		ValidationError(w, fmt.Sprintf("Expected %d arrows, got %d", arrowsPerEnd, len(req.Arrows)))
		return
	}

	// Validate arrow values
	allowed := map[string]bool{}
	for _, v := range allowedValues {
		allowed[v] = true
	}
	for _, a := range req.Arrows {
		if !allowed[a.ScoreValue] {
			ValidationError(w, fmt.Sprintf("Invalid arrow value '%s'. Allowed: %v", a.ScoreValue, allowedValues))
			return
		}
	}

	// Determine end number
	var endCount int
	h.DB.QueryRow(ctx, "SELECT COUNT(*) FROM ends WHERE session_id = $1", sessionID).Scan(&endCount)
	endNumber := endCount + 1

	tx, err := h.DB.Begin(ctx)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer tx.Rollback(ctx)

	// Create end
	endID := uuid.New().String()
	now := time.Now().UTC()
	_, err = tx.Exec(ctx, `
		INSERT INTO ends (id, session_id, stage_id, end_number, end_total, created_at)
		VALUES ($1, $2, $3, $4, 0, $5)`,
		endID, sessionID, req.StageID, endNumber, now,
	)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Create arrows and calculate totals
	endTotal := 0
	xCount := 0
	arrows := []arrowOut{}
	for i, a := range req.Arrows {
		arrowID := uuid.New().String()
		numeric := valueScoreMap[a.ScoreValue]
		_, err = tx.Exec(ctx, `
			INSERT INTO arrows (id, end_id, arrow_number, score_value, score_numeric, x_pos, y_pos)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			arrowID, endID, i+1, a.ScoreValue, numeric, a.XPos, a.YPos,
		)
		if err != nil {
			Error(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		endTotal += numeric
		if a.ScoreValue == "X" {
			xCount++
		}
		arrows = append(arrows, arrowOut{
			ID: arrowID, ArrowNumber: i + 1,
			ScoreValue: a.ScoreValue, ScoreNumeric: numeric,
			XPos: a.XPos, YPos: a.YPos,
		})
	}

	// Update end total
	_, err = tx.Exec(ctx, "UPDATE ends SET end_total = $1 WHERE id = $2", endTotal, endID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Update session totals
	_, err = tx.Exec(ctx, `
		UPDATE scoring_sessions
		SET total_score = total_score + $2,
		    total_x_count = total_x_count + $3,
		    total_arrows = total_arrows + $4
		WHERE id = $1`,
		sessionID, endTotal, xCount, len(req.Arrows),
	)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if err := tx.Commit(ctx); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	stageID := req.StageID
	JSON(w, http.StatusCreated, endOut{
		ID: endID, EndNumber: endNumber, EndTotal: endTotal,
		StageID: &stageID, Arrows: arrows, CreatedAt: now,
	})
}

// ── Undo Last End ────────────────────────────────────────────────────

func (h *ScoringHandler) UndoLastEnd(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(sessionID); err != nil {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	var status string
	err := h.DB.QueryRow(ctx,
		"SELECT status FROM scoring_sessions WHERE id = $1 AND user_id = $2",
		sessionID, userID,
	).Scan(&status)
	if err != nil {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}
	if status != "in_progress" {
		ValidationError(w, "Session is not in progress")
		return
	}

	// Find the last end
	var endID string
	var endTotal int
	err = h.DB.QueryRow(ctx, `
		SELECT id, end_total FROM ends
		WHERE session_id = $1
		ORDER BY end_number DESC LIMIT 1`, sessionID,
	).Scan(&endID, &endTotal)
	if err != nil {
		ValidationError(w, "No ends to undo")
		return
	}

	// Count arrows and X's in last end
	var arrowCount, xCount int
	h.DB.QueryRow(ctx,
		"SELECT COUNT(*), COALESCE(SUM(CASE WHEN score_value = 'X' THEN 1 ELSE 0 END), 0) FROM arrows WHERE end_id = $1",
		endID,
	).Scan(&arrowCount, &xCount)

	tx, err := h.DB.Begin(ctx)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer tx.Rollback(ctx)

	// Delete arrows then end
	tx.Exec(ctx, "DELETE FROM arrows WHERE end_id = $1", endID)
	tx.Exec(ctx, "DELETE FROM ends WHERE id = $1", endID)

	// Update session totals
	tx.Exec(ctx, `
		UPDATE scoring_sessions
		SET total_score = total_score - $2,
		    total_x_count = total_x_count - $3,
		    total_arrows = total_arrows - $4
		WHERE id = $1`,
		sessionID, endTotal, xCount, arrowCount,
	)

	if err := tx.Commit(ctx); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	out, err := loadSessionOut(ctx, h.DB, sessionID, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusOK, out)
}

// ── Complete ──────────────────────────────────────────────────────────

func (h *ScoringHandler) Complete(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(sessionID); err != nil {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	// Check session exists
	var templateID, sStatus string
	var totalScore int
	err := h.DB.QueryRow(ctx,
		"SELECT template_id, status, total_score FROM scoring_sessions WHERE id = $1 AND user_id = $2",
		sessionID, userID,
	).Scan(&templateID, &sStatus, &totalScore)
	if err != nil {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}

	// Parse optional body (don't use Decode — it writes 422 on empty body)
	var req sessionCompleteReq
	json.NewDecoder(r.Body).Decode(&req) // ignore error — body is optional

	now := time.Now().UTC()
	_, err = h.DB.Exec(ctx, `
		UPDATE scoring_sessions
		SET status = 'completed',
		    completed_at = $2,
		    notes = CASE WHEN $3::boolean THEN $4 ELSE notes END,
		    location = CASE WHEN $5::boolean THEN $6 ELSE location END,
		    weather = CASE WHEN $7::boolean THEN $8 ELSE weather END
		WHERE id = $1`,
		sessionID, now,
		req.Notes != nil, req.Notes,
		req.Location != nil, req.Location,
		req.Weather != nil, req.Weather,
	)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Check / update personal record
	isPersonalBest := false
	var existingPRID *string
	var existingScore int
	err = h.DB.QueryRow(ctx,
		"SELECT id, score FROM personal_records WHERE user_id = $1 AND template_id = $2",
		userID, templateID,
	).Scan(&existingPRID, &existingScore)

	if err != nil {
		// No existing PR — create one
		prID := uuid.New().String()
		h.DB.Exec(ctx, `
			INSERT INTO personal_records (id, user_id, template_id, session_id, score, achieved_at)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			prID, userID, templateID, sessionID, totalScore, now,
		)
		isPersonalBest = true
	} else if totalScore > existingScore {
		h.DB.Exec(ctx, `
			UPDATE personal_records SET session_id = $1, score = $2, achieved_at = $3 WHERE id = $4`,
			sessionID, totalScore, now, *existingPRID,
		)
		isPersonalBest = true
	}

	// Classification
	var templateName string
	h.DB.QueryRow(ctx, "SELECT name FROM round_templates WHERE id = $1", templateID).Scan(&templateName)
	if system, classification := calculateClassification(totalScore, templateName); system != "" {
		crID := uuid.New().String()
		h.DB.Exec(ctx, `
			INSERT INTO classification_records (id, user_id, system, classification, round_type, score, achieved_at, session_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			crID, userID, system, classification, templateName, totalScore, now, sessionID,
		)
	}

	// Notification for personal record
	if isPersonalBest {
		notifID := uuid.New().String()
		h.DB.Exec(ctx, `
			INSERT INTO notifications (id, user_id, type, title, message, link, is_read, created_at)
			VALUES ($1, $2, 'personal_record', 'New Personal Record!', $3, $4, false, $5)`,
			notifID, userID,
			fmt.Sprintf("You scored %d on %s — a new personal best!", totalScore, templateName),
			fmt.Sprintf("/sessions/%s", sessionID),
			now,
		)
	}

	// Feed item
	feedID := uuid.New().String()
	feedType := "session_completed"
	if isPersonalBest {
		feedType = "personal_record"
	}
	feedData, _ := json.Marshal(map[string]any{
		"template_name": templateName,
		"total_score":   totalScore,
		"session_id":    sessionID,
	})
	h.DB.Exec(ctx, `
		INSERT INTO feed_items (id, user_id, type, data, created_at)
		VALUES ($1, $2, $3, $4, $5)`,
		feedID, userID, feedType, feedData, now,
	)

	out, err := loadSessionOut(ctx, h.DB, sessionID, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	out.IsPersonalBest = isPersonalBest

	JSON(w, http.StatusOK, out)
}

// ── Abandon ───────────────────────────────────────────────────────────

func (h *ScoringHandler) Abandon(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(sessionID); err != nil {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	var status string
	err := h.DB.QueryRow(ctx,
		"SELECT status FROM scoring_sessions WHERE id = $1 AND user_id = $2",
		sessionID, userID,
	).Scan(&status)
	if err != nil {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}
	if status != "in_progress" {
		ValidationError(w, "Only in-progress sessions can be abandoned")
		return
	}

	h.DB.Exec(ctx, "UPDATE scoring_sessions SET status = 'abandoned' WHERE id = $1", sessionID)
	JSON(w, http.StatusOK, map[string]string{"detail": "Session abandoned"})
}

// ── Delete ────────────────────────────────────────────────────────────

func (h *ScoringHandler) Delete(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(sessionID); err != nil {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	var status string
	err := h.DB.QueryRow(ctx,
		"SELECT status FROM scoring_sessions WHERE id = $1 AND user_id = $2",
		sessionID, userID,
	).Scan(&status)
	if err != nil {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}
	if status != "abandoned" {
		ValidationError(w, "Only abandoned sessions can be deleted")
		return
	}

	tx, err := h.DB.Begin(ctx)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer tx.Rollback(ctx)

	// Delete arrows → ends → session
	tx.Exec(ctx, `DELETE FROM arrows WHERE end_id IN (SELECT id FROM ends WHERE session_id = $1)`, sessionID)
	tx.Exec(ctx, "DELETE FROM ends WHERE session_id = $1", sessionID)
	tx.Exec(ctx, "DELETE FROM scoring_sessions WHERE id = $1", sessionID)

	if err := tx.Commit(ctx); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ── Stats ─────────────────────────────────────────────────────────────

func (h *ScoringHandler) Stats(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	// Load all sessions with template info
	rows, err := h.DB.Query(ctx, `
		SELECT ss.id, ss.status, ss.total_score, ss.total_arrows, ss.total_x_count,
		       ss.completed_at, ss.started_at, rt.name AS template_name, ss.template_id
		FROM scoring_sessions ss
		LEFT JOIN round_templates rt ON rt.id = ss.template_id
		WHERE ss.user_id = $1
		ORDER BY ss.completed_at DESC NULLS LAST`, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()

	type sessionInfo struct {
		id           string
		status       string
		totalScore   int
		totalArrows  int
		totalXCount  int
		completedAt  *time.Time
		startedAt    time.Time
		templateName *string
		templateID   string
	}

	var sessions []sessionInfo
	for rows.Next() {
		var s sessionInfo
		if err := rows.Scan(&s.id, &s.status, &s.totalScore, &s.totalArrows, &s.totalXCount,
			&s.completedAt, &s.startedAt, &s.templateName, &s.templateID); err != nil {
			Error(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		sessions = append(sessions, s)
	}

	totalArrows := 0
	totalXCount := 0
	for _, s := range sessions {
		totalArrows += s.totalArrows
		totalXCount += s.totalXCount
	}

	// Completed sessions
	var completed []sessionInfo
	for _, s := range sessions {
		if s.status == "completed" {
			completed = append(completed, s)
		}
	}

	// Personal best
	var bestScore *int
	var bestTemplate *string
	for _, s := range completed {
		if bestScore == nil || s.totalScore > *bestScore {
			score := s.totalScore
			bestScore = &score
			name := "Unknown"
			if s.templateName != nil {
				name = *s.templateName
			}
			bestTemplate = &name
		}
	}

	// Avg by round type
	byType := map[string][]int{}
	for _, s := range completed {
		name := "Unknown"
		if s.templateName != nil {
			name = *s.templateName
		}
		byType[name] = append(byType[name], s.totalScore)
	}
	avgByRound := []roundTypeAvg{}
	for name, scores := range byType {
		sum := 0
		for _, sc := range scores {
			sum += sc
		}
		avg := math.Round(float64(sum)/float64(len(scores))*10) / 10
		avgByRound = append(avgByRound, roundTypeAvg{
			TemplateName: name, AvgScore: avg, Count: len(scores),
		})
	}

	// Recent trend (last 10 completed)
	recentTrend := []recentTrendItem{}
	limit := 10
	if len(completed) < limit {
		limit = len(completed)
	}
	for _, s := range completed[:limit] {
		maxScore := getTemplateMaxScore(ctx, h.DB, s.templateID)
		name := "Unknown"
		if s.templateName != nil {
			name = *s.templateName
		}
		date := s.startedAt
		if s.completedAt != nil {
			date = *s.completedAt
		}
		recentTrend = append(recentTrend, recentTrendItem{
			Score: s.totalScore, MaxScore: maxScore, TemplateName: name, Date: date,
		})
	}

	// Personal records
	prRows, err := h.DB.Query(ctx, `
		SELECT pr.template_id, pr.score, pr.achieved_at, pr.session_id, rt.name
		FROM personal_records pr
		LEFT JOIN round_templates rt ON rt.id = pr.template_id
		WHERE pr.user_id = $1`, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer prRows.Close()

	prList := []personalRecordOut{}
	for prRows.Next() {
		var templateID, sessionID string
		var score int
		var achievedAt time.Time
		var tname *string
		if err := prRows.Scan(&templateID, &score, &achievedAt, &sessionID, &tname); err != nil {
			continue
		}
		name := "Unknown"
		if tname != nil {
			name = *tname
		}
		maxScore := getTemplateMaxScore(ctx, h.DB, templateID)
		prList = append(prList, personalRecordOut{
			TemplateName: name, Score: score, MaxScore: maxScore,
			AchievedAt: achievedAt, SessionID: sessionID,
		})
	}

	JSON(w, http.StatusOK, statsOut{
		TotalSessions:     len(sessions),
		CompletedSessions: len(completed),
		TotalArrows:       totalArrows,
		TotalXCount:       totalXCount,
		PersonalBestScore: bestScore,
		PersonalBestTmpl:  bestTemplate,
		AvgByRoundType:    avgByRound,
		RecentTrend:       recentTrend,
		PersonalRecords:   prList,
	})
}

// ── Personal Records ──────────────────────────────────────────────────

func (h *ScoringHandler) PersonalRecords(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	rows, err := h.DB.Query(ctx, `
		SELECT pr.template_id, pr.score, pr.achieved_at, pr.session_id, rt.name
		FROM personal_records pr
		LEFT JOIN round_templates rt ON rt.id = pr.template_id
		WHERE pr.user_id = $1`, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()

	items := []personalRecordOut{}
	for rows.Next() {
		var templateID, sessionID string
		var score int
		var achievedAt time.Time
		var tname *string
		if err := rows.Scan(&templateID, &score, &achievedAt, &sessionID, &tname); err != nil {
			continue
		}
		name := "Unknown"
		if tname != nil {
			name = *tname
		}
		maxScore := getTemplateMaxScore(ctx, h.DB, templateID)
		items = append(items, personalRecordOut{
			TemplateName: name, Score: score, MaxScore: maxScore,
			AchievedAt: achievedAt, SessionID: sessionID,
		})
	}

	JSON(w, http.StatusOK, items)
}

// ── Trends ────────────────────────────────────────────────────────────

func (h *ScoringHandler) Trends(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	rows, err := h.DB.Query(ctx, `
		SELECT ss.id, ss.total_score, ss.completed_at, ss.started_at,
		       ss.template_id, rt.name
		FROM scoring_sessions ss
		LEFT JOIN round_templates rt ON rt.id = ss.template_id
		WHERE ss.user_id = $1 AND ss.status = 'completed'
		ORDER BY ss.completed_at DESC`, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()

	items := []trendDataItem{}
	for rows.Next() {
		var sid, templateID string
		var totalScore int
		var completedAt *time.Time
		var startedAt time.Time
		var tname *string
		if err := rows.Scan(&sid, &totalScore, &completedAt, &startedAt, &templateID, &tname); err != nil {
			continue
		}
		name := "Unknown"
		if tname != nil {
			name = *tname
		}
		maxScore := getTemplateMaxScore(ctx, h.DB, templateID)
		pct := 0.0
		if maxScore > 0 {
			pct = math.Round(float64(totalScore)/float64(maxScore)*1000) / 10
		}
		date := startedAt
		if completedAt != nil {
			date = *completedAt
		}
		items = append(items, trendDataItem{
			SessionID: sid, TemplateName: name, TotalScore: totalScore,
			MaxScore: maxScore, Percentage: pct, CompletedAt: date,
		})
	}

	JSON(w, http.StatusOK, items)
}

// ── Export Single Session ─────────────────────────────────────────────

func (h *ScoringHandler) ExportSingle(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(sessionID); err != nil {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "csv"
	}

	// For PDF, proxy to Python
	if format == "pdf" {
		h.proxyToPython(w, r)
		return
	}

	// Load session for CSV
	out, err := loadSessionOut(ctx, h.DB, sessionID, userID)
	if err != nil {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}

	csvContent := generateSessionCSV(out)
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=session-%s.csv", sessionID))
	w.Write([]byte(csvContent))
}

// ── Export Bulk ───────────────────────────────────────────────────────

func (h *ScoringHandler) ExportBulk(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	query := `
		SELECT ss.id, ss.status, ss.total_score, ss.total_x_count, ss.total_arrows,
		       ss.started_at, ss.completed_at, ss.location, ss.notes,
		       rt.name AS template_name
		FROM scoring_sessions ss
		LEFT JOIN round_templates rt ON rt.id = ss.template_id
		WHERE ss.user_id = $1`
	args := []any{userID}
	argN := 2

	if tid := r.URL.Query().Get("template_id"); tid != "" {
		query += fmt.Sprintf(" AND ss.template_id = $%d", argN)
		args = append(args, tid)
		argN++
	}
	if df := r.URL.Query().Get("date_from"); df != "" {
		query += fmt.Sprintf(" AND ss.started_at >= $%d::date", argN)
		args = append(args, df)
		argN++
	}
	if dt := r.URL.Query().Get("date_to"); dt != "" {
		query += fmt.Sprintf(" AND ss.started_at <= ($%d::date + interval '1 day' - interval '1 second')", argN)
		args = append(args, dt)
		argN++
	}
	if search := r.URL.Query().Get("search"); search != "" {
		pattern := "%" + search + "%"
		query += fmt.Sprintf(" AND (ss.notes ILIKE $%d OR ss.location ILIKE $%d)", argN, argN)
		args = append(args, pattern)
		argN++
	}
	query += " ORDER BY ss.started_at DESC"

	rows, err := h.DB.Query(ctx, query, args...)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	writer.Write([]string{"Date", "Round", "Status", "Score", "X Count", "Arrows", "Location", "Notes"})

	for rows.Next() {
		var id, status string
		var totalScore, totalXCount, totalArrows int
		var startedAt time.Time
		var completedAt *time.Time
		var location, notes, templateName *string
		if err := rows.Scan(&id, &status, &totalScore, &totalXCount, &totalArrows,
			&startedAt, &completedAt, &location, &notes, &templateName); err != nil {
			continue
		}
		dateVal := startedAt
		if completedAt != nil {
			dateVal = *completedAt
		}
		tname := "Unknown"
		if templateName != nil {
			tname = *templateName
		}
		loc := ""
		if location != nil {
			loc = *location
		}
		n := ""
		if notes != nil {
			n = *notes
		}
		writer.Write([]string{
			dateVal.Format("2006-01-02"),
			tname, status,
			fmt.Sprintf("%d", totalScore),
			fmt.Sprintf("%d", totalXCount),
			fmt.Sprintf("%d", totalArrows),
			loc, n,
		})
	}
	writer.Flush()

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=sessions.csv")
	w.Write(buf.Bytes())
}

// ── PDF Proxy ─────────────────────────────────────────────────────────

func (h *ScoringHandler) proxyToPython(w http.ResponseWriter, r *http.Request) {
	targetURL := strings.TrimRight(h.Cfg.PythonAPIURL, "/") + r.URL.Path + "?" + r.URL.RawQuery

	proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
	if err != nil {
		Error(w, http.StatusBadGateway, "Upstream service unavailable")
		return
	}
	// Copy auth header
	proxyReq.Header.Set("Authorization", r.Header.Get("Authorization"))

	resp, err := http.DefaultClient.Do(proxyReq)
	if err != nil {
		Error(w, http.StatusBadGateway, "Upstream service unavailable")
		return
	}
	defer resp.Body.Close()

	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// ── Helpers ───────────────────────────────────────────────────────────

func loadSessionOut(ctx context.Context, db *pgxpool.Pool, sessionID, userID string) (*sessionOut, error) {
	var s sessionOut
	var templateID string
	err := db.QueryRow(ctx, `
		SELECT ss.id, ss.template_id, ss.setup_profile_id, ss.status,
		       ss.total_score, ss.total_x_count, ss.total_arrows,
		       ss.notes, ss.location, ss.weather, ss.share_token,
		       ss.started_at, ss.completed_at,
		       sp.name AS setup_profile_name
		FROM scoring_sessions ss
		LEFT JOIN setup_profiles sp ON sp.id = ss.setup_profile_id
		WHERE ss.id = $1 AND ss.user_id = $2`,
		sessionID, userID,
	).Scan(&s.ID, &templateID, &s.SetupProfileID, &s.Status,
		&s.TotalScore, &s.TotalXCount, &s.TotalArrows,
		&s.Notes, &s.Location, &s.Weather, &s.ShareToken,
		&s.StartedAt, &s.CompletedAt,
		&s.SetupProfileName)
	if err != nil {
		return nil, err
	}
	s.TemplateID = templateID

	// Load template with stages
	var t roundTemplateOut
	err = db.QueryRow(ctx, `
		SELECT id, name, organization, description, is_official, created_by
		FROM round_templates WHERE id = $1`, templateID,
	).Scan(&t.ID, &t.Name, &t.Organization, &t.Description, &t.IsOfficial, &t.CreatedBy)
	if err == nil {
		stages, err := loadStages(ctx, db, t.ID)
		if err == nil {
			t.Stages = stages
		}
		s.Template = &t
	}

	// Load ends with arrows
	s.Ends, _ = loadEnds(ctx, db, sessionID)

	return &s, nil
}

func loadEnds(ctx context.Context, db *pgxpool.Pool, sessionID string) ([]endOut, error) {
	rows, err := db.Query(ctx, `
		SELECT id, end_number, end_total, stage_id, created_at
		FROM ends WHERE session_id = $1
		ORDER BY end_number`, sessionID)
	if err != nil {
		return []endOut{}, err
	}
	defer rows.Close()

	ends := []endOut{}
	for rows.Next() {
		var e endOut
		if err := rows.Scan(&e.ID, &e.EndNumber, &e.EndTotal, &e.StageID, &e.CreatedAt); err != nil {
			continue
		}
		e.Arrows, _ = loadArrows(ctx, db, e.ID)
		ends = append(ends, e)
	}
	return ends, nil
}

func loadArrows(ctx context.Context, db *pgxpool.Pool, endID string) ([]arrowOut, error) {
	rows, err := db.Query(ctx, `
		SELECT id, arrow_number, score_value, score_numeric, x_pos, y_pos
		FROM arrows WHERE end_id = $1
		ORDER BY arrow_number`, endID)
	if err != nil {
		return []arrowOut{}, err
	}
	defer rows.Close()

	arrows := []arrowOut{}
	for rows.Next() {
		var a arrowOut
		if err := rows.Scan(&a.ID, &a.ArrowNumber, &a.ScoreValue, &a.ScoreNumeric, &a.XPos, &a.YPos); err != nil {
			continue
		}
		arrows = append(arrows, a)
	}
	return arrows, nil
}

func getTemplateMaxScore(ctx context.Context, db *pgxpool.Pool, templateID string) int {
	rows, err := db.Query(ctx, `
		SELECT num_ends, arrows_per_end, max_score_per_arrow
		FROM round_template_stages WHERE template_id = $1`, templateID)
	if err != nil {
		return 0
	}
	defer rows.Close()

	total := 0
	for rows.Next() {
		var numEnds, arrowsPerEnd, maxScore int
		if err := rows.Scan(&numEnds, &arrowsPerEnd, &maxScore); err != nil {
			continue
		}
		total += numEnds * arrowsPerEnd * maxScore
	}
	return total
}

func generateSessionCSV(s *sessionOut) string {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	templateName := "Unknown"
	if s.Template != nil {
		templateName = s.Template.Name
	}

	w.Write([]string{"Session Detail"})
	w.Write([]string{"Round", templateName})
	w.Write([]string{"Status", s.Status})
	w.Write([]string{"Total Score", fmt.Sprintf("%d", s.TotalScore)})
	w.Write([]string{"Total X Count", fmt.Sprintf("%d", s.TotalXCount)})
	w.Write([]string{"Total Arrows", fmt.Sprintf("%d", s.TotalArrows)})
	w.Write([]string{"Location", deref(s.Location)})
	w.Write([]string{"Weather", deref(s.Weather)})
	w.Write([]string{"Notes", deref(s.Notes)})
	w.Write([]string{"Started", s.StartedAt.Format(time.RFC3339)})
	completed := ""
	if s.CompletedAt != nil {
		completed = s.CompletedAt.Format(time.RFC3339)
	}
	w.Write([]string{"Completed", completed})
	w.Write([]string{""})

	w.Write([]string{"End", "Arrow", "Value", "Score"})
	for _, end := range s.Ends {
		for _, arrow := range end.Arrows {
			w.Write([]string{
				fmt.Sprintf("%d", end.EndNumber),
				fmt.Sprintf("%d", arrow.ArrowNumber),
				arrow.ScoreValue,
				fmt.Sprintf("%d", arrow.ScoreNumeric),
			})
		}
	}

	w.Flush()
	return buf.String()
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ── Classification ────────────────────────────────────────────────────

type classificationThreshold struct {
	score          int
	classification string
}

var classificationTables = map[string]struct {
	system     string
	thresholds []classificationThreshold
}{
	"WA 720 (70m)": {"ArcheryGB", []classificationThreshold{
		{625, "Grand Master Bowman"}, {575, "Master Bowman"}, {525, "Bowman 1st Class"},
		{475, "Bowman 2nd Class"}, {400, "Bowman 3rd Class"}, {300, "Archer 1st Class"},
		{200, "Archer 2nd Class"}, {100, "Archer 3rd Class"},
	}},
	"WA 720 (60m)": {"ArcheryGB", []classificationThreshold{
		{640, "Grand Master Bowman"}, {590, "Master Bowman"}, {540, "Bowman 1st Class"},
		{490, "Bowman 2nd Class"}, {420, "Bowman 3rd Class"}, {320, "Archer 1st Class"},
		{220, "Archer 2nd Class"}, {120, "Archer 3rd Class"},
	}},
	"WA 18m Round (60 arrows)": {"ArcheryGB", []classificationThreshold{
		{550, "Grand Master Bowman"}, {510, "Master Bowman"}, {470, "Bowman 1st Class"},
		{420, "Bowman 2nd Class"}, {350, "Bowman 3rd Class"}, {270, "Archer 1st Class"},
		{180, "Archer 2nd Class"}, {90, "Archer 3rd Class"},
	}},
	"NFAA 300 Indoor": {"NFAA", []classificationThreshold{
		{290, "Expert"}, {270, "Sharpshooter"}, {240, "Marksman"}, {200, "Bowman"},
	}},
	"NFAA 300 Outdoor": {"NFAA", []classificationThreshold{
		{280, "Expert"}, {260, "Sharpshooter"}, {230, "Marksman"}, {190, "Bowman"},
	}},
}

func calculateClassification(score int, templateName string) (string, string) {
	entry, ok := classificationTables[templateName]
	if !ok {
		return "", ""
	}
	for _, t := range entry.thresholds {
		if score >= t.score {
			return entry.system, t.classification
		}
	}
	return "", ""
}
