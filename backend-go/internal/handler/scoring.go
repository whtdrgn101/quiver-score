package handler

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/pdf"
	"github.com/quiverscore/backend-go/internal/repository"
)

type ScoringRepository interface {
	SetupProfileExists(ctx context.Context, setupID, userID string) (bool, error)
	CreateSession(ctx context.Context, id, userID, templateID string, setupProfileID *string, notes, location, weather *string, now time.Time) error
	LoadSessionOut(ctx context.Context, sessionID, userID string) (*repository.SessionOut, error)
	ListSessions(ctx context.Context, userID string, templateID, dateFrom, dateTo, search *string) ([]repository.SessionSummary, error)
	GetSessionStatus(ctx context.Context, sessionID, userID string) (string, error)
	IsPersonalBest(ctx context.Context, userID, sessionID string) bool
	GetStageInfo(ctx context.Context, stageID string) (*repository.StageInfo, error)
	GetEndCount(ctx context.Context, sessionID string) int
	SubmitEnd(ctx context.Context, sessionID, stageID string, endNumber int, arrows []repository.ArrowIn, scoreMap map[string]int) (*repository.EndOut, error)
	UndoLastEnd(ctx context.Context, sessionID string) error
	GetSessionForComplete(ctx context.Context, sessionID, userID string) (templateID, status string, totalScore int, err error)
	CompleteSession(ctx context.Context, sessionID string, now time.Time, notes, location, weather *string) error
	UpsertPersonalRecord(ctx context.Context, userID, templateID, sessionID string, totalScore int, now time.Time) (bool, error)
	GetTemplateName(ctx context.Context, templateID string) string
	InsertClassification(ctx context.Context, userID, system, classification, roundType string, score int, now time.Time, sessionID string) error
	InsertNotification(ctx context.Context, userID, nType, title, message, link string, now time.Time) error
	InsertFeedItem(ctx context.Context, userID, feedType string, data map[string]any, now time.Time) error
	AbandonSession(ctx context.Context, sessionID string) error
	DeleteSession(ctx context.Context, sessionID string) error
	Stats(ctx context.Context, userID string) (*repository.StatsOut, error)
	PersonalRecords(ctx context.Context, userID string) ([]repository.PersonalRecordOut, error)
	Trends(ctx context.Context, userID string) ([]repository.TrendDataItem, error)
	ExportBulkData(ctx context.Context, userID string, templateID, dateFrom, dateTo, search *string) ([]repository.BulkExportRow, error)
}

type ScoringHandler struct {
	Scoring ScoringRepository
	Cfg     *config.Config
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

type endIn struct {
	StageID string               `json:"stage_id"`
	Arrows  []repository.ArrowIn `json:"arrows"`
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
		exists, err := h.Scoring.SetupProfileExists(ctx, *req.SetupProfileID, userID)
		if err != nil || !exists {
			Error(w, http.StatusNotFound, "Setup profile not found")
			return
		}
	}

	id := uuid.New().String()
	now := time.Now().UTC()

	if err := h.Scoring.CreateSession(ctx, id, userID, req.TemplateID, req.SetupProfileID,
		req.Notes, req.Location, req.Weather, now); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	out, err := h.Scoring.LoadSessionOut(ctx, id, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusCreated, out)
}

// ── List ──────────────────────────────────────────────────────────────

func (h *ScoringHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var templateID, dateFrom, dateTo, search *string
	if v := r.URL.Query().Get("template_id"); v != "" {
		templateID = &v
	}
	if v := r.URL.Query().Get("date_from"); v != "" {
		dateFrom = &v
	}
	if v := r.URL.Query().Get("date_to"); v != "" {
		dateTo = &v
	}
	if v := r.URL.Query().Get("search"); v != "" {
		search = &v
	}

	items, err := h.Scoring.ListSessions(r.Context(), userID, templateID, dateFrom, dateTo, search)
	if err != nil {
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

	out, err := h.Scoring.LoadSessionOut(ctx, id, userID)
	if err != nil {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}

	out.IsPersonalBest = h.Scoring.IsPersonalBest(ctx, userID, id)

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

	status, err := h.Scoring.GetSessionStatus(ctx, sessionID, userID)
	if err != nil {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}
	if status != "in_progress" {
		ValidationError(w, "Session is not in progress")
		return
	}

	stageInfo, err := h.Scoring.GetStageInfo(ctx, req.StageID)
	if err != nil {
		Error(w, http.StatusNotFound, "Stage not found")
		return
	}

	// Validate arrow count
	if len(req.Arrows) != stageInfo.ArrowsPerEnd {
		ValidationError(w, fmt.Sprintf("Expected %d arrows, got %d", stageInfo.ArrowsPerEnd, len(req.Arrows)))
		return
	}

	// Validate arrow values
	allowed := map[string]bool{}
	for _, v := range stageInfo.AllowedValues {
		allowed[v] = true
	}
	for _, a := range req.Arrows {
		if !allowed[a.ScoreValue] {
			ValidationError(w, fmt.Sprintf("Invalid arrow value '%s'. Allowed: %v", a.ScoreValue, stageInfo.AllowedValues))
			return
		}
	}

	endNumber := h.Scoring.GetEndCount(ctx, sessionID) + 1

	end, err := h.Scoring.SubmitEnd(ctx, sessionID, req.StageID, endNumber, req.Arrows, stageInfo.ValueScoreMap)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusCreated, end)
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

	status, err := h.Scoring.GetSessionStatus(ctx, sessionID, userID)
	if err != nil {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}
	if status != "in_progress" {
		ValidationError(w, "Session is not in progress")
		return
	}

	if err := h.Scoring.UndoLastEnd(ctx, sessionID); err != nil {
		ValidationError(w, "No ends to undo")
		return
	}

	out, err := h.Scoring.LoadSessionOut(ctx, sessionID, userID)
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

	templateID, _, totalScore, err := h.Scoring.GetSessionForComplete(ctx, sessionID, userID)
	if err != nil {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}

	// Parse optional body (don't use Decode — it writes 422 on empty body)
	var req sessionCompleteReq
	json.NewDecoder(r.Body).Decode(&req) // ignore error — body is optional

	now := time.Now().UTC()
	if err := h.Scoring.CompleteSession(ctx, sessionID, now, req.Notes, req.Location, req.Weather); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Check / update personal record
	isPersonalBest, _ := h.Scoring.UpsertPersonalRecord(ctx, userID, templateID, sessionID, totalScore, now)

	// Classification
	templateName := h.Scoring.GetTemplateName(ctx, templateID)
	if system, classification := calculateClassification(totalScore, templateName); system != "" {
		h.Scoring.InsertClassification(ctx, userID, system, classification, templateName, totalScore, now, sessionID)
	}

	// Notification for personal record
	if isPersonalBest {
		h.Scoring.InsertNotification(ctx, userID, "personal_record", "New Personal Record!",
			fmt.Sprintf("You scored %d on %s — a new personal best!", totalScore, templateName),
			fmt.Sprintf("/sessions/%s", sessionID), now,
		)
	}

	// Feed item
	feedType := "session_completed"
	if isPersonalBest {
		feedType = "personal_record"
	}
	h.Scoring.InsertFeedItem(ctx, userID, feedType, map[string]any{
		"template_name": templateName,
		"total_score":   totalScore,
		"session_id":    sessionID,
	}, now)

	out, err := h.Scoring.LoadSessionOut(ctx, sessionID, userID)
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

	status, err := h.Scoring.GetSessionStatus(ctx, sessionID, userID)
	if err != nil {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}
	if status != "in_progress" {
		ValidationError(w, "Only in-progress sessions can be abandoned")
		return
	}

	h.Scoring.AbandonSession(ctx, sessionID)
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

	status, err := h.Scoring.GetSessionStatus(ctx, sessionID, userID)
	if err != nil {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}
	if status != "abandoned" {
		ValidationError(w, "Only abandoned sessions can be deleted")
		return
	}

	if err := h.Scoring.DeleteSession(ctx, sessionID); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ── Stats ─────────────────────────────────────────────────────────────

func (h *ScoringHandler) Stats(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	stats, err := h.Scoring.Stats(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusOK, stats)
}

// ── Personal Records ──────────────────────────────────────────────────

func (h *ScoringHandler) PersonalRecords(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	items, err := h.Scoring.PersonalRecords(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusOK, items)
}

// ── Trends ────────────────────────────────────────────────────────────

func (h *ScoringHandler) Trends(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	items, err := h.Scoring.Trends(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
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

	out, err := h.Scoring.LoadSessionOut(ctx, sessionID, userID)
	if err != nil {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}

	if format == "pdf" {
		pdfBytes, err := pdf.GenerateSessionPDF(out)
		if err != nil {
			Error(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=session-%s.pdf", sessionID))
		w.Write(pdfBytes)
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

	var templateID, dateFrom, dateTo, search *string
	if v := r.URL.Query().Get("template_id"); v != "" {
		templateID = &v
	}
	if v := r.URL.Query().Get("date_from"); v != "" {
		dateFrom = &v
	}
	if v := r.URL.Query().Get("date_to"); v != "" {
		dateTo = &v
	}
	if v := r.URL.Query().Get("search"); v != "" {
		search = &v
	}

	rows, err := h.Scoring.ExportBulkData(r.Context(), userID, templateID, dateFrom, dateTo, search)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	writer.Write([]string{"Date", "Round", "Status", "Score", "X Count", "Arrows", "Location", "Notes"})

	for _, b := range rows {
		dateVal := b.StartedAt
		if b.CompletedAt != nil {
			dateVal = *b.CompletedAt
		}
		tname := "Unknown"
		if b.TemplateName != nil {
			tname = *b.TemplateName
		}
		loc := ""
		if b.Location != nil {
			loc = *b.Location
		}
		n := ""
		if b.Notes != nil {
			n = *b.Notes
		}
		writer.Write([]string{
			dateVal.Format("2006-01-02"),
			tname, b.Status,
			fmt.Sprintf("%d", b.TotalScore),
			fmt.Sprintf("%d", b.TotalXCount),
			fmt.Sprintf("%d", b.TotalArrows),
			loc, n,
		})
	}
	writer.Flush()

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=sessions.csv")
	w.Write(buf.Bytes())
}

// ── Helpers ───────────────────────────────────────────────────────────

func generateSessionCSV(s *repository.SessionOut) string {
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
