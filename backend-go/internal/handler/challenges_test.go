package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
)

type mockChallengesRepo struct {
	createFn      func(ctx context.Context, challengerID, challengeeID, templateID string, expiresAt *time.Time) (*repository.ChallengeOut, error)
	getFn         func(ctx context.Context, id string) (*repository.ChallengeOut, error)
	listFn        func(ctx context.Context, userID string) ([]repository.ChallengeOut, error)
	acceptFn      func(ctx context.Context, id, userID string) (*repository.ChallengeOut, error)
	declineFn     func(ctx context.Context, id, userID string) (*repository.ChallengeOut, error)
	submitScoreFn func(ctx context.Context, id, userID, sessionID string) (*repository.ChallengeOut, error)
}

func (m *mockChallengesRepo) CreateChallenge(ctx context.Context, challengerID, challengeeID, templateID string, expiresAt *time.Time) (*repository.ChallengeOut, error) {
	return m.createFn(ctx, challengerID, challengeeID, templateID, expiresAt)
}

func (m *mockChallengesRepo) GetChallenge(ctx context.Context, challengeID string) (*repository.ChallengeOut, error) {
	return m.getFn(ctx, challengeID)
}

func (m *mockChallengesRepo) ListChallengesForUser(ctx context.Context, userID string) ([]repository.ChallengeOut, error) {
	return m.listFn(ctx, userID)
}

func (m *mockChallengesRepo) AcceptChallenge(ctx context.Context, challengeID, userID string) (*repository.ChallengeOut, error) {
	return m.acceptFn(ctx, challengeID, userID)
}

func (m *mockChallengesRepo) DeclineChallenge(ctx context.Context, challengeID, userID string) (*repository.ChallengeOut, error) {
	return m.declineFn(ctx, challengeID, userID)
}

func (m *mockChallengesRepo) SubmitChallengeScore(ctx context.Context, challengeID, userID, sessionID string) (*repository.ChallengeOut, error) {
	return m.submitScoreFn(ctx, challengeID, userID, sessionID)
}

func authedBody(method, path, userID string, body any) *http.Request {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	return req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))
}

func chiParams(req *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestCreateChallenge_Success(t *testing.T) {
	challengerID := uuid.New().String()
	challengeeID := uuid.New().String()
	templateID := uuid.New().String()

	mock := &mockChallengesRepo{
		createFn: func(ctx context.Context, challengerID, challengeeID, templateID string, expiresAt *time.Time) (*repository.ChallengeOut, error) {
			return &repository.ChallengeOut{
				ID:                 "c1",
				ChallengerID:       challengerID,
				ChallengerUsername: "Alice",
				ChallengeeID:       challengeeID,
				ChallengeeUsername: "Bob",
				TemplateID:         templateID,
				TemplateName:       "WA 720",
				Status:             "pending",
				CreatedAt:          time.Now(),
			}, nil
		},
	}

	h := &ChallengesHandler{Challenges: mock}

	req := authedBody(http.MethodPost, "/", challengerID, map[string]any{
		"challengee_id":    challengeeID,
		"template_id":      templateID,
		"expires_in_hours": 48,
	})
	rr := httptest.NewRecorder()

	h.CreateChallenge(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var res repository.ChallengeOut
	json.NewDecoder(rr.Body).Decode(&res)
	if res.ID != "c1" || res.Status != "pending" || res.ChallengerUsername != "Alice" {
		t.Errorf("unexpected output: %v", res)
	}
}

func TestListChallenges_Success(t *testing.T) {
	userID := uuid.New().String()
	mock := &mockChallengesRepo{
		listFn: func(ctx context.Context, uID string) ([]repository.ChallengeOut, error) {
			return []repository.ChallengeOut{
				{
					ID:                 "c1",
					ChallengerID:       uID,
					ChallengerUsername: "Alice",
					ChallengeeID:       "bob-id",
					ChallengeeUsername: "Bob",
					TemplateID:         "t1",
					TemplateName:       "WA 720",
					Status:             "pending",
				},
			}, nil
		},
	}

	h := &ChallengesHandler{Challenges: mock}

	req := authedRequest(http.MethodGet, "/", userID)
	rr := httptest.NewRecorder()

	h.ListChallenges(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var res []repository.ChallengeOut
	json.NewDecoder(rr.Body).Decode(&res)
	if len(res) != 1 || res[0].ID != "c1" {
		t.Errorf("unexpected output: %v", res)
	}
}

func TestGetChallenge_Success(t *testing.T) {
	mock := &mockChallengesRepo{
		getFn: func(ctx context.Context, id string) (*repository.ChallengeOut, error) {
			return &repository.ChallengeOut{
				ID:     id,
				Status: "pending",
			}, nil
		},
	}

	h := &ChallengesHandler{Challenges: mock}

	req := authedRequest(http.MethodGet, "/c1", "user-1")
	req = chiParams(req, map[string]string{"challengeID": "c1"})
	rr := httptest.NewRecorder()

	h.GetChallenge(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var res repository.ChallengeOut
	json.NewDecoder(rr.Body).Decode(&res)
	if res.ID != "c1" {
		t.Errorf("unexpected output: %v", res)
	}
}

func TestGetChallenge_NotFound(t *testing.T) {
	mock := &mockChallengesRepo{
		getFn: func(ctx context.Context, id string) (*repository.ChallengeOut, error) {
			return nil, repository.ErrNotFound
		},
	}

	h := &ChallengesHandler{Challenges: mock}

	req := authedRequest(http.MethodGet, "/c1", "user-1")
	req = chiParams(req, map[string]string{"challengeID": "c1"})
	rr := httptest.NewRecorder()

	h.GetChallenge(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestAcceptChallenge_Success(t *testing.T) {
	userID := uuid.New().String()
	mock := &mockChallengesRepo{
		acceptFn: func(ctx context.Context, id, uID string) (*repository.ChallengeOut, error) {
			return &repository.ChallengeOut{
				ID:     id,
				Status: "accepted",
			}, nil
		},
	}

	h := &ChallengesHandler{Challenges: mock}

	req := authedRequest(http.MethodPost, "/c1/accept", userID)
	req = chiParams(req, map[string]string{"challengeID": "c1"})
	rr := httptest.NewRecorder()

	h.AcceptChallenge(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var res repository.ChallengeOut
	json.NewDecoder(rr.Body).Decode(&res)
	if res.Status != "accepted" {
		t.Errorf("unexpected status: %s", res.Status)
	}
}

func TestDeclineChallenge_Success(t *testing.T) {
	userID := uuid.New().String()
	mock := &mockChallengesRepo{
		declineFn: func(ctx context.Context, id, uID string) (*repository.ChallengeOut, error) {
			return &repository.ChallengeOut{
				ID:     id,
				Status: "declined",
			}, nil
		},
	}

	h := &ChallengesHandler{Challenges: mock}

	req := authedRequest(http.MethodPost, "/c1/decline", userID)
	req = chiParams(req, map[string]string{"challengeID": "c1"})
	rr := httptest.NewRecorder()

	h.DeclineChallenge(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var res repository.ChallengeOut
	json.NewDecoder(rr.Body).Decode(&res)
	if res.Status != "declined" {
		t.Errorf("unexpected status: %s", res.Status)
	}
}

func TestSubmitChallengeScore_Success(t *testing.T) {
	userID := uuid.New().String()
	sessionID := uuid.New().String()
	mock := &mockChallengesRepo{
		submitScoreFn: func(ctx context.Context, id, uID, sID string) (*repository.ChallengeOut, error) {
			score := 335
			return &repository.ChallengeOut{
				ID:                  id,
				ChallengerSessionID: &sID,
				ChallengerScore:     &score,
				Status:              "accepted",
			}, nil
		},
	}

	h := &ChallengesHandler{Challenges: mock}

	req := authedBody(http.MethodPost, "/c1/submit-score", userID, map[string]any{
		"session_id": sessionID,
	})
	req = chiParams(req, map[string]string{"challengeID": "c1"})
	rr := httptest.NewRecorder()

	h.SubmitChallengeScore(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var res repository.ChallengeOut
	json.NewDecoder(rr.Body).Decode(&res)
	if res.ChallengerScore == nil || *res.ChallengerScore != 335 {
		t.Errorf("unexpected score: %v", res)
	}
}
