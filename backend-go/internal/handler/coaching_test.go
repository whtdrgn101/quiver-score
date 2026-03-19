package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
)

// ── Mock ────────────────────────────────────────────────────────────────

type mockCoachingRepo struct {
	inviteResult          *repository.CoachAthleteLinkOut
	inviteErr             error
	respondResult         *repository.CoachAthleteLinkOut
	respondErr            error
	listAthletesResult    []repository.CoachAthleteLinkOut
	listAthletesErr       error
	listCoachesResult     []repository.CoachAthleteLinkOut
	listCoachesErr        error
	athleteSessionsResult []repository.AthleteSessionOut
	athleteSessionsErr    error
	checkAccessOwner      string
	checkAccessErr        error
	addAnnotationResult   *repository.AnnotationOut
	addAnnotationErr      error
	listAnnotationsResult []repository.AnnotationOut
	listAnnotationsErr    error
}

func (m *mockCoachingRepo) Invite(_ context.Context, _, _ string) (*repository.CoachAthleteLinkOut, error) {
	return m.inviteResult, m.inviteErr
}

func (m *mockCoachingRepo) Respond(_ context.Context, _, _ string, _ bool) (*repository.CoachAthleteLinkOut, error) {
	return m.respondResult, m.respondErr
}

func (m *mockCoachingRepo) ListAthletes(_ context.Context, _ string) ([]repository.CoachAthleteLinkOut, error) {
	return m.listAthletesResult, m.listAthletesErr
}

func (m *mockCoachingRepo) ListCoaches(_ context.Context, _ string) ([]repository.CoachAthleteLinkOut, error) {
	return m.listCoachesResult, m.listCoachesErr
}

func (m *mockCoachingRepo) GetAthleteSessions(_ context.Context, _, _ string) ([]repository.AthleteSessionOut, error) {
	return m.athleteSessionsResult, m.athleteSessionsErr
}

func (m *mockCoachingRepo) CheckSessionAccess(_ context.Context, _, _ string) (string, error) {
	return m.checkAccessOwner, m.checkAccessErr
}

func (m *mockCoachingRepo) AddAnnotation(_ context.Context, _, _ string, _, _ *int, _ string) (*repository.AnnotationOut, error) {
	return m.addAnnotationResult, m.addAnnotationErr
}

func (m *mockCoachingRepo) ListAnnotations(_ context.Context, _ string) ([]repository.AnnotationOut, error) {
	return m.listAnnotationsResult, m.listAnnotationsErr
}

// ── Helpers ─────────────────────────────────────────────────────────────

func coachingPostRequest(method, path, userID, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	return req.WithContext(ctx)
}

func coachingSessionRequest(method, path, userID, sessionID string) *http.Request {
	req := authedRequest(method, path, userID)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("sessionID", sessionID)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func coachingSessionPostRequest(method, path, userID, sessionID, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("sessionID", sessionID)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	return req.WithContext(ctx)
}

func coachingAthleteRequest(method, path, userID, athleteID string) *http.Request {
	req := authedRequest(method, path, userID)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("athleteID", athleteID)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

// ── Invite ──────────────────────────────────────────────────────────────

func TestCoaching_Invite_Success(t *testing.T) {
	athleteUsername := "archer1"
	mock := &mockCoachingRepo{
		inviteResult: &repository.CoachAthleteLinkOut{
			ID:             "link-1",
			CoachID:        "coach-1",
			AthleteID:      "athlete-1",
			AthleteUsername: &athleteUsername,
			Status:         "pending",
			CreatedAt:      time.Now(),
		},
	}
	h := &CoachingHandler{Coaching: mock}

	rr := httptest.NewRecorder()
	h.Invite(rr, coachingPostRequest(http.MethodPost, "/invite", "coach-1", `{"athlete_username":"archer1"}`))

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}

	var result repository.CoachAthleteLinkOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.ID != "link-1" {
		t.Errorf("expected link ID 'link-1', got '%s'", result.ID)
	}
	if result.Status != "pending" {
		t.Errorf("expected status 'pending', got '%s'", result.Status)
	}
}

func TestCoaching_Invite_MissingUsername(t *testing.T) {
	h := &CoachingHandler{Coaching: &mockCoachingRepo{}}

	rr := httptest.NewRecorder()
	h.Invite(rr, coachingPostRequest(http.MethodPost, "/invite", "coach-1", `{"athlete_username":""}`))

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rr.Code)
	}
}

func TestCoaching_Invite_NotFound(t *testing.T) {
	mock := &mockCoachingRepo{inviteErr: repository.ErrNotFound}
	h := &CoachingHandler{Coaching: mock}

	rr := httptest.NewRecorder()
	h.Invite(rr, coachingPostRequest(http.MethodPost, "/invite", "coach-1", `{"athlete_username":"archer1"}`))

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestCoaching_Invite_Self(t *testing.T) {
	mock := &mockCoachingRepo{inviteErr: repository.ErrValidation}
	h := &CoachingHandler{Coaching: mock}

	rr := httptest.NewRecorder()
	h.Invite(rr, coachingPostRequest(http.MethodPost, "/invite", "coach-1", `{"athlete_username":"myself"}`))

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rr.Code)
	}
}

func TestCoaching_Invite_AlreadyExists(t *testing.T) {
	mock := &mockCoachingRepo{inviteErr: repository.ErrAlreadyMember}
	h := &CoachingHandler{Coaching: mock}

	rr := httptest.NewRecorder()
	h.Invite(rr, coachingPostRequest(http.MethodPost, "/invite", "coach-1", `{"athlete_username":"archer1"}`))

	if rr.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", rr.Code)
	}
}

// ── Respond ─────────────────────────────────────────────────────────────

func TestCoaching_Respond_Success(t *testing.T) {
	mock := &mockCoachingRepo{
		respondResult: &repository.CoachAthleteLinkOut{
			ID:        "link-1",
			CoachID:   "coach-1",
			AthleteID: "athlete-1",
			Status:    "active",
			CreatedAt: time.Now(),
		},
	}
	h := &CoachingHandler{Coaching: mock}

	rr := httptest.NewRecorder()
	h.Respond(rr, coachingPostRequest(http.MethodPost, "/respond", "athlete-1", `{"link_id":"link-1","accept":true}`))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result repository.CoachAthleteLinkOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.Status != "active" {
		t.Errorf("expected status 'active', got '%s'", result.Status)
	}
}

func TestCoaching_Respond_NotFound(t *testing.T) {
	mock := &mockCoachingRepo{respondErr: repository.ErrNotFound}
	h := &CoachingHandler{Coaching: mock}

	rr := httptest.NewRecorder()
	h.Respond(rr, coachingPostRequest(http.MethodPost, "/respond", "athlete-1", `{"link_id":"link-1","accept":true}`))

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

// ── ListAthletes / ListCoaches ──────────────────────────────────────────

func TestCoaching_ListAthletes_Success(t *testing.T) {
	mock := &mockCoachingRepo{
		listAthletesResult: []repository.CoachAthleteLinkOut{
			{ID: "link-1", CoachID: "coach-1", AthleteID: "athlete-1", Status: "active", CreatedAt: time.Now()},
			{ID: "link-2", CoachID: "coach-1", AthleteID: "athlete-2", Status: "pending", CreatedAt: time.Now()},
		},
	}
	h := &CoachingHandler{Coaching: mock}

	rr := httptest.NewRecorder()
	h.ListAthletes(rr, authedRequest(http.MethodGet, "/athletes", "coach-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result []repository.CoachAthleteLinkOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}
}

func TestCoaching_ListCoaches_Success(t *testing.T) {
	mock := &mockCoachingRepo{
		listCoachesResult: []repository.CoachAthleteLinkOut{
			{ID: "link-1", CoachID: "coach-1", AthleteID: "athlete-1", Status: "active", CreatedAt: time.Now()},
		},
	}
	h := &CoachingHandler{Coaching: mock}

	rr := httptest.NewRecorder()
	h.ListCoaches(rr, authedRequest(http.MethodGet, "/coaches", "athlete-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result []repository.CoachAthleteLinkOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 item, got %d", len(result))
	}
}

// ── ViewAthleteSessions ─────────────────────────────────────────────────

func TestCoaching_ViewAthleteSessions_Success(t *testing.T) {
	completedAt := time.Now()
	mock := &mockCoachingRepo{
		athleteSessionsResult: []repository.AthleteSessionOut{
			{ID: "s-1", TemplateName: "WA 18m", TotalScore: 540, TotalXCount: 10, TotalArrows: 60, CompletedAt: &completedAt},
		},
	}
	h := &CoachingHandler{Coaching: mock}

	rr := httptest.NewRecorder()
	h.ViewAthleteSessions(rr, coachingAthleteRequest(http.MethodGet, "/athletes/athlete-1/sessions", "coach-1", "athlete-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result []repository.AthleteSessionOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 session, got %d", len(result))
	}
	if result[0].TemplateName != "WA 18m" {
		t.Errorf("expected 'WA 18m', got '%s'", result[0].TemplateName)
	}
}

func TestCoaching_ViewAthleteSessions_Forbidden(t *testing.T) {
	mock := &mockCoachingRepo{athleteSessionsErr: repository.ErrForbidden}
	h := &CoachingHandler{Coaching: mock}

	rr := httptest.NewRecorder()
	h.ViewAthleteSessions(rr, coachingAthleteRequest(http.MethodGet, "/athletes/athlete-1/sessions", "coach-1", "athlete-1"))

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

// ── AddAnnotation ───────────────────────────────────────────────────────

func TestCoaching_AddAnnotation_Success(t *testing.T) {
	authorUsername := "coach1"
	mock := &mockCoachingRepo{
		checkAccessOwner: "athlete-1",
		addAnnotationResult: &repository.AnnotationOut{
			ID:             "ann-1",
			SessionID:      "session-1",
			AuthorID:       "coach-1",
			AuthorUsername: &authorUsername,
			Text:           "Great shot!",
			CreatedAt:      time.Now(),
		},
	}
	h := &CoachingHandler{Coaching: mock}

	rr := httptest.NewRecorder()
	h.AddAnnotation(rr, coachingSessionPostRequest(http.MethodPost, "/sessions/session-1/annotations", "coach-1", "session-1", `{"text":"Great shot!"}`))

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}

	var result repository.AnnotationOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.Text != "Great shot!" {
		t.Errorf("expected 'Great shot!', got '%s'", result.Text)
	}
}

func TestCoaching_AddAnnotation_EmptyText(t *testing.T) {
	h := &CoachingHandler{Coaching: &mockCoachingRepo{}}

	rr := httptest.NewRecorder()
	h.AddAnnotation(rr, coachingSessionPostRequest(http.MethodPost, "/sessions/session-1/annotations", "coach-1", "session-1", `{"text":""}`))

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rr.Code)
	}
}

func TestCoaching_AddAnnotation_SessionNotFound(t *testing.T) {
	mock := &mockCoachingRepo{checkAccessErr: repository.ErrNotFound}
	h := &CoachingHandler{Coaching: mock}

	rr := httptest.NewRecorder()
	h.AddAnnotation(rr, coachingSessionPostRequest(http.MethodPost, "/sessions/session-1/annotations", "coach-1", "session-1", `{"text":"Great shot!"}`))

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestCoaching_AddAnnotation_Forbidden(t *testing.T) {
	mock := &mockCoachingRepo{checkAccessErr: repository.ErrForbidden}
	h := &CoachingHandler{Coaching: mock}

	rr := httptest.NewRecorder()
	h.AddAnnotation(rr, coachingSessionPostRequest(http.MethodPost, "/sessions/session-1/annotations", "coach-1", "session-1", `{"text":"Great shot!"}`))

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

// ── ListAnnotations ─────────────────────────────────────────────────────

func TestCoaching_ListAnnotations_Success(t *testing.T) {
	authorUsername := "coach1"
	mock := &mockCoachingRepo{
		checkAccessOwner: "athlete-1",
		listAnnotationsResult: []repository.AnnotationOut{
			{ID: "ann-1", SessionID: "session-1", AuthorID: "coach-1", AuthorUsername: &authorUsername, Text: "Nice grouping", CreatedAt: time.Now()},
		},
	}
	h := &CoachingHandler{Coaching: mock}

	rr := httptest.NewRecorder()
	h.ListAnnotations(rr, coachingSessionRequest(http.MethodGet, "/sessions/session-1/annotations", "coach-1", "session-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result []repository.AnnotationOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 annotation, got %d", len(result))
	}
}

func TestCoaching_ListAnnotations_SessionNotFound(t *testing.T) {
	mock := &mockCoachingRepo{checkAccessErr: repository.ErrNotFound}
	h := &CoachingHandler{Coaching: mock}

	rr := httptest.NewRecorder()
	h.ListAnnotations(rr, coachingSessionRequest(http.MethodGet, "/sessions/session-1/annotations", "coach-1", "session-1"))

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestCoaching_ListAnnotations_Forbidden(t *testing.T) {
	mock := &mockCoachingRepo{checkAccessErr: repository.ErrForbidden}
	h := &CoachingHandler{Coaching: mock}

	rr := httptest.NewRecorder()
	h.ListAnnotations(rr, coachingSessionRequest(http.MethodGet, "/sessions/session-1/annotations", "coach-1", "session-1"))

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}
