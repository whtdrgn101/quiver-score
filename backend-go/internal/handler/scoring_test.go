package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
)

var errNotFound = errors.New("not found")

// ── Mock ──────────────────────────────────────────────────────────────

type mockScoringRepo struct {
	createSessionErr error

	setupProfileExists    bool
	setupProfileExistsErr error

	loadSessionOutResult *repository.SessionOut
	loadSessionOutErr    error

	listSessionsResult []repository.SessionSummary
	listSessionsErr    error

	getSessionStatusResult string
	getSessionStatusErr    error

	isPersonalBestResult bool

	getStageInfoResult *repository.StageInfo
	getStageInfoErr    error

	getEndCountResult int

	submitEndResult *repository.EndOut
	submitEndErr    error

	undoLastEndErr error

	getSessionForCompleteTemplateID string
	getSessionForCompleteStatus     string
	getSessionForCompleteTotalScore int
	getSessionForCompleteErr        error

	completeSessionErr error

	upsertPersonalRecordResult bool
	upsertPersonalRecordErr    error

	getTemplateNameResult string

	insertClassificationErr error
	insertNotificationErr   error
	insertFeedItemErr       error

	abandonSessionErr error
	deleteSessionErr  error

	statsResult *repository.StatsOut
	statsErr    error

	personalRecordsResult []repository.PersonalRecordOut
	personalRecordsErr    error

	trendsResult []repository.TrendDataItem
	trendsErr    error

	exportBulkDataResult []repository.BulkExportRow
	exportBulkDataErr    error
}

func (m *mockScoringRepo) SetupProfileExists(_ context.Context, _, _ string) (bool, error) {
	return m.setupProfileExists, m.setupProfileExistsErr
}

func (m *mockScoringRepo) CreateSession(_ context.Context, _, _, _ string, _ *string, _, _, _ *string, _ time.Time) error {
	return m.createSessionErr
}

func (m *mockScoringRepo) LoadSessionOut(_ context.Context, _, _ string) (*repository.SessionOut, error) {
	return m.loadSessionOutResult, m.loadSessionOutErr
}

func (m *mockScoringRepo) ListSessions(_ context.Context, _ string, _, _, _, _ *string) ([]repository.SessionSummary, error) {
	return m.listSessionsResult, m.listSessionsErr
}

func (m *mockScoringRepo) GetSessionStatus(_ context.Context, _, _ string) (string, error) {
	return m.getSessionStatusResult, m.getSessionStatusErr
}

func (m *mockScoringRepo) IsPersonalBest(_ context.Context, _, _ string) bool {
	return m.isPersonalBestResult
}

func (m *mockScoringRepo) GetStageInfo(_ context.Context, _ string) (*repository.StageInfo, error) {
	return m.getStageInfoResult, m.getStageInfoErr
}

func (m *mockScoringRepo) GetEndCount(_ context.Context, _ string) int {
	return m.getEndCountResult
}

func (m *mockScoringRepo) SubmitEnd(_ context.Context, _, _ string, _ int, _ []repository.ArrowIn, _ map[string]int) (*repository.EndOut, error) {
	return m.submitEndResult, m.submitEndErr
}

func (m *mockScoringRepo) UndoLastEnd(_ context.Context, _ string) error {
	return m.undoLastEndErr
}

func (m *mockScoringRepo) GetSessionForComplete(_ context.Context, _, _ string) (string, string, int, error) {
	return m.getSessionForCompleteTemplateID, m.getSessionForCompleteStatus, m.getSessionForCompleteTotalScore, m.getSessionForCompleteErr
}

func (m *mockScoringRepo) CompleteSession(_ context.Context, _ string, _ time.Time, _, _, _ *string) error {
	return m.completeSessionErr
}

func (m *mockScoringRepo) UpsertPersonalRecord(_ context.Context, _, _, _ string, _ int, _ time.Time) (bool, error) {
	return m.upsertPersonalRecordResult, m.upsertPersonalRecordErr
}

func (m *mockScoringRepo) GetTemplateName(_ context.Context, _ string) string {
	return m.getTemplateNameResult
}

func (m *mockScoringRepo) InsertClassification(_ context.Context, _, _, _, _ string, _ int, _ time.Time, _ string) error {
	return m.insertClassificationErr
}

func (m *mockScoringRepo) InsertNotification(_ context.Context, _, _, _, _, _ string, _ time.Time) error {
	return m.insertNotificationErr
}

func (m *mockScoringRepo) InsertFeedItem(_ context.Context, _, _ string, _ map[string]any, _ time.Time) error {
	return m.insertFeedItemErr
}

func (m *mockScoringRepo) AbandonSession(_ context.Context, _ string) error {
	return m.abandonSessionErr
}

func (m *mockScoringRepo) DeleteSession(_ context.Context, _ string) error {
	return m.deleteSessionErr
}

func (m *mockScoringRepo) Stats(_ context.Context, _ string) (*repository.StatsOut, error) {
	return m.statsResult, m.statsErr
}

func (m *mockScoringRepo) PersonalRecords(_ context.Context, _ string) ([]repository.PersonalRecordOut, error) {
	return m.personalRecordsResult, m.personalRecordsErr
}

func (m *mockScoringRepo) Trends(_ context.Context, _ string) ([]repository.TrendDataItem, error) {
	return m.trendsResult, m.trendsErr
}

func (m *mockScoringRepo) ExportBulkData(_ context.Context, _ string, _, _, _, _ *string) ([]repository.BulkExportRow, error) {
	return m.exportBulkDataResult, m.exportBulkDataErr
}

// ── Helpers ───────────────────────────────────────────────────────────

func scoringHandler(mock *mockScoringRepo) *ScoringHandler {
	return &ScoringHandler{
		Scoring: mock,
		Cfg:     &config.Config{SecretKey: "test-secret"},
	}
}

func withURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func sessionOut() *repository.SessionOut {
	return &repository.SessionOut{
		ID:          uuid.New().String(),
		TemplateID:  uuid.New().String(),
		Status:      "in_progress",
		TotalScore:  100,
		TotalXCount: 3,
		TotalArrows: 30,
		StartedAt:   time.Now().UTC(),
		Ends:        []repository.EndOut{},
	}
}

func scoringAuthedReq(method, path string, body string, userID string) *http.Request {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	return req.WithContext(ctx)
}

// ── CreateSession ─────────────────────────────────────────────────────

func TestCreateSession_Success(t *testing.T) {
	out := sessionOut()
	mock := &mockScoringRepo{
		loadSessionOutResult: out,
	}
	h := scoringHandler(mock)

	body := `{"template_id":"` + uuid.New().String() + `"}`
	req := scoringAuthedReq(http.MethodPost, "/", body, "user-1")

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestCreateSession_MissingTemplateID(t *testing.T) {
	mock := &mockScoringRepo{}
	h := scoringHandler(mock)

	req := scoringAuthedReq(http.MethodPost, "/", `{}`, "user-1")

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d: %s", rr.Code, rr.Body.String())
	}
}

// ── ListSessions ──────────────────────────────────────────────────────

func TestListSessions_Success(t *testing.T) {
	now := time.Now().UTC()
	mock := &mockScoringRepo{
		listSessionsResult: []repository.SessionSummary{
			{
				ID:         uuid.New().String(),
				TemplateID: uuid.New().String(),
				Status:     "completed",
				TotalScore: 250,
				StartedAt:  now,
			},
		},
	}
	h := scoringHandler(mock)

	rr := httptest.NewRecorder()
	h.List(rr, authedRequest(http.MethodGet, "/", "user-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result []repository.SessionSummary
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result))
	}
	if result[0].TotalScore != 250 {
		t.Errorf("expected score 250, got %d", result[0].TotalScore)
	}
}

func TestListSessions_Empty(t *testing.T) {
	mock := &mockScoringRepo{
		listSessionsResult: []repository.SessionSummary{},
	}
	h := scoringHandler(mock)

	rr := httptest.NewRecorder()
	h.List(rr, authedRequest(http.MethodGet, "/", "user-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result []repository.SessionSummary
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 items, got %d", len(result))
	}
}

// ── GetSession ────────────────────────────────────────────────────────

func TestGetSession_Success(t *testing.T) {
	out := sessionOut()
	mock := &mockScoringRepo{
		loadSessionOutResult: out,
		isPersonalBestResult: true,
	}
	h := scoringHandler(mock)

	sessionID := uuid.New().String()
	req := authedRequest(http.MethodGet, "/"+sessionID, "user-1")
	req = withURLParam(req, "id", sessionID)

	rr := httptest.NewRecorder()
	h.Get(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result repository.SessionOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if !result.IsPersonalBest {
		t.Error("expected IsPersonalBest to be true")
	}
}

func TestGetSession_NotFound(t *testing.T) {
	mock := &mockScoringRepo{
		loadSessionOutErr: errNotFound,
	}
	h := scoringHandler(mock)

	sessionID := uuid.New().String()
	req := authedRequest(http.MethodGet, "/"+sessionID, "user-1")
	req = withURLParam(req, "id", sessionID)

	rr := httptest.NewRecorder()
	h.Get(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

// ── SubmitEnd ─────────────────────────────────────────────────────────

func TestSubmitEnd_Success(t *testing.T) {
	stageID := uuid.New().String()
	now := time.Now().UTC()
	endOut := &repository.EndOut{
		ID:        uuid.New().String(),
		EndNumber: 1,
		EndTotal:  27,
		StageID:   &stageID,
		Arrows: []repository.ArrowOut{
			{ID: uuid.New().String(), ArrowNumber: 1, ScoreValue: "10", ScoreNumeric: 10},
			{ID: uuid.New().String(), ArrowNumber: 2, ScoreValue: "9", ScoreNumeric: 9},
			{ID: uuid.New().String(), ArrowNumber: 3, ScoreValue: "8", ScoreNumeric: 8},
		},
		CreatedAt: now,
	}
	mock := &mockScoringRepo{
		getSessionStatusResult: "in_progress",
		getStageInfoResult: &repository.StageInfo{
			ArrowsPerEnd:  3,
			AllowedValues: []string{"X", "10", "9", "8", "7", "6", "5", "4", "3", "2", "1", "M"},
			ValueScoreMap: map[string]int{"X": 10, "10": 10, "9": 9, "8": 8, "7": 7, "6": 6, "5": 5, "4": 4, "3": 3, "2": 2, "1": 1, "M": 0},
		},
		getEndCountResult: 0,
		submitEndResult:   endOut,
	}
	h := scoringHandler(mock)

	sessionID := uuid.New().String()
	body := `{"stage_id":"` + stageID + `","arrows":[{"score_value":"10"},{"score_value":"9"},{"score_value":"8"}]}`
	req := scoringAuthedReq(http.MethodPost, "/"+sessionID+"/ends", body, "user-1")
	req = withURLParam(req, "id", sessionID)

	rr := httptest.NewRecorder()
	h.SubmitEnd(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var result repository.EndOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if result.EndTotal != 27 {
		t.Errorf("expected end total 27, got %d", result.EndTotal)
	}
}

// ── CompleteSession ───────────────────────────────────────────────────

func TestCompleteSession_Success(t *testing.T) {
	out := sessionOut()
	out.Status = "completed"
	mock := &mockScoringRepo{
		getSessionForCompleteTemplateID: uuid.New().String(),
		getSessionForCompleteStatus:     "in_progress",
		getSessionForCompleteTotalScore: 250,
		upsertPersonalRecordResult:      false,
		getTemplateNameResult:           "Practice Round",
		loadSessionOutResult:            out,
	}
	h := scoringHandler(mock)

	sessionID := uuid.New().String()
	req := scoringAuthedReq(http.MethodPost, "/"+sessionID+"/complete", "", "user-1")
	req = withURLParam(req, "id", sessionID)

	rr := httptest.NewRecorder()
	h.Complete(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

// ── AbandonSession ───────────────────────────────────────────────────

func TestAbandonSession_Success(t *testing.T) {
	mock := &mockScoringRepo{
		getSessionStatusResult: "in_progress",
	}
	h := scoringHandler(mock)

	sessionID := uuid.New().String()
	req := authedRequest(http.MethodPost, "/"+sessionID+"/abandon", "user-1")
	req = withURLParam(req, "id", sessionID)

	rr := httptest.NewRecorder()
	h.Abandon(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if result["detail"] != "Session abandoned" {
		t.Errorf("expected 'Session abandoned', got '%s'", result["detail"])
	}
}

// ── DeleteSession ────────────────────────────────────────────────────

func TestDeleteSession_Success(t *testing.T) {
	mock := &mockScoringRepo{
		getSessionStatusResult: "abandoned",
	}
	h := scoringHandler(mock)

	sessionID := uuid.New().String()
	req := authedRequest(http.MethodDelete, "/"+sessionID, "user-1")
	req = withURLParam(req, "id", sessionID)

	rr := httptest.NewRecorder()
	h.Delete(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestDeleteSession_NotFound(t *testing.T) {
	mock := &mockScoringRepo{
		getSessionStatusErr: errNotFound,
	}
	h := scoringHandler(mock)

	sessionID := uuid.New().String()
	req := authedRequest(http.MethodDelete, "/"+sessionID, "user-1")
	req = withURLParam(req, "id", sessionID)

	rr := httptest.NewRecorder()
	h.Delete(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

// ── UndoEnd ──────────────────────────────────────────────────────────

func TestUndoEnd_Success(t *testing.T) {
	out := sessionOut()
	mock := &mockScoringRepo{
		getSessionStatusResult: "in_progress",
		loadSessionOutResult:   out,
	}
	h := scoringHandler(mock)

	sessionID := uuid.New().String()
	req := authedRequest(http.MethodDelete, "/"+sessionID+"/ends/last", "user-1")
	req = withURLParam(req, "id", sessionID)

	rr := httptest.NewRecorder()
	h.UndoLastEnd(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

// ── Stats ────────────────────────────────────────────────────────────

func TestGetStats_Success(t *testing.T) {
	mock := &mockScoringRepo{
		statsResult: &repository.StatsOut{
			TotalSessions:     5,
			CompletedSessions: 3,
			TotalArrows:       150,
			TotalXCount:       12,
			AvgByRoundType:    []repository.RoundTypeAvg{},
			RecentTrend:       []repository.RecentTrendItem{},
			PersonalRecords:   []repository.PersonalRecordOut{},
		},
	}
	h := scoringHandler(mock)

	rr := httptest.NewRecorder()
	h.Stats(rr, authedRequest(http.MethodGet, "/stats", "user-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result repository.StatsOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if result.TotalSessions != 5 {
		t.Errorf("expected 5 sessions, got %d", result.TotalSessions)
	}
	if result.TotalArrows != 150 {
		t.Errorf("expected 150 arrows, got %d", result.TotalArrows)
	}
}

// ── PersonalRecords ──────────────────────────────────────────────────

func TestGetPersonalRecords_Success(t *testing.T) {
	mock := &mockScoringRepo{
		personalRecordsResult: []repository.PersonalRecordOut{
			{
				TemplateName: "WA 720 (70m)",
				Score:        580,
				MaxScore:     720,
				AchievedAt:   time.Now().UTC(),
				SessionID:    uuid.New().String(),
			},
		},
	}
	h := scoringHandler(mock)

	rr := httptest.NewRecorder()
	h.PersonalRecords(rr, authedRequest(http.MethodGet, "/personal-records", "user-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result []repository.PersonalRecordOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 record, got %d", len(result))
	}
	if result[0].Score != 580 {
		t.Errorf("expected score 580, got %d", result[0].Score)
	}
}

// ── Trends ───────────────────────────────────────────────────────────

func TestGetTrends_Success(t *testing.T) {
	mock := &mockScoringRepo{
		trendsResult: []repository.TrendDataItem{
			{
				SessionID:    uuid.New().String(),
				TemplateName: "WA 720 (70m)",
				TotalScore:   550,
				MaxScore:     720,
				Percentage:   76.4,
				CompletedAt:  time.Now().UTC(),
			},
		},
	}
	h := scoringHandler(mock)

	rr := httptest.NewRecorder()
	h.Trends(rr, authedRequest(http.MethodGet, "/trends", "user-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result []repository.TrendDataItem
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 trend item, got %d", len(result))
	}
	if result[0].TotalScore != 550 {
		t.Errorf("expected score 550, got %d", result[0].TotalScore)
	}
}

// ── ExportCSV (Single Session) ───────────────────────────────────────

func TestExportCSV_Success(t *testing.T) {
	out := sessionOut()
	out.Status = "completed"
	out.Template = &repository.RoundTemplateOut{
		ID:           out.TemplateID,
		Name:         "WA 720 (70m)",
		Organization: "World Archery",
		IsOfficial:   true,
	}
	out.Ends = []repository.EndOut{
		{
			ID:        uuid.New().String(),
			EndNumber: 1,
			EndTotal:  27,
			Arrows: []repository.ArrowOut{
				{ID: uuid.New().String(), ArrowNumber: 1, ScoreValue: "10", ScoreNumeric: 10},
				{ID: uuid.New().String(), ArrowNumber: 2, ScoreValue: "9", ScoreNumeric: 9},
				{ID: uuid.New().String(), ArrowNumber: 3, ScoreValue: "8", ScoreNumeric: 8},
			},
		},
	}
	mock := &mockScoringRepo{
		loadSessionOutResult: out,
	}
	h := scoringHandler(mock)

	sessionID := uuid.New().String()
	req := authedRequest(http.MethodGet, "/"+sessionID+"/export?format=csv", "user-1")
	req = withURLParam(req, "id", sessionID)

	rr := httptest.NewRecorder()
	h.ExportSingle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("expected Content-Type text/csv, got %s", ct)
	}
	if !strings.Contains(rr.Body.String(), "WA 720 (70m)") {
		t.Error("expected CSV to contain template name")
	}
}

// ── ExportBulkCSV ────────────────────────────────────────────────────

func TestExportBulkCSV_Success(t *testing.T) {
	now := time.Now().UTC()
	tname := "WA 720 (70m)"
	mock := &mockScoringRepo{
		exportBulkDataResult: []repository.BulkExportRow{
			{
				StartedAt:    now,
				TemplateName: &tname,
				Status:       "completed",
				TotalScore:   580,
				TotalXCount:  5,
				TotalArrows:  72,
			},
		},
	}
	h := scoringHandler(mock)

	rr := httptest.NewRecorder()
	h.ExportBulk(rr, authedRequest(http.MethodGet, "/export", "user-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("expected Content-Type text/csv, got %s", ct)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "WA 720 (70m)") {
		t.Error("expected CSV to contain template name")
	}
	if !strings.Contains(body, "580") {
		t.Error("expected CSV to contain score")
	}
}
