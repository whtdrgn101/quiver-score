package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/repository"
)

// ── Mock repositories ────────────────────────────────────────────────

type mockSharingScoringRepo struct {
	getShareTokenResult *string
	getShareTokenErr    error

	setShareTokenErr error

	revokeResult bool
	revokeErr    error

	getSharedSessionResult *repository.SharedSessionData
	getSharedSessionErr    error

	loadEndsResult []repository.EndOut
	loadEndsErr    error
}

func (m *mockSharingScoringRepo) GetShareToken(_ context.Context, _, _ string) (*string, error) {
	return m.getShareTokenResult, m.getShareTokenErr
}

func (m *mockSharingScoringRepo) SetShareToken(_ context.Context, _, _ string) error {
	return m.setShareTokenErr
}

func (m *mockSharingScoringRepo) RevokeShareToken(_ context.Context, _, _ string) (bool, error) {
	return m.revokeResult, m.revokeErr
}

func (m *mockSharingScoringRepo) GetSharedSession(_ context.Context, _ string) (*repository.SharedSessionData, error) {
	return m.getSharedSessionResult, m.getSharedSessionErr
}

func (m *mockSharingScoringRepo) LoadEnds(_ context.Context, _ string) ([]repository.EndOut, error) {
	return m.loadEndsResult, m.loadEndsErr
}

type mockSharingUserRepo struct {
	username    string
	displayName *string
	avatar      *string
	err         error
}

func (m *mockSharingUserRepo) GetArcherInfo(_ context.Context, _ string) (string, *string, *string, error) {
	return m.username, m.displayName, m.avatar, m.err
}

type mockSharingRoundRepo struct {
	result *repository.RoundTemplateOut
	err    error
}

func (m *mockSharingRoundRepo) Get(_ context.Context, _ string) (*repository.RoundTemplateOut, error) {
	return m.result, m.err
}

// ── Helpers ──────────────────────────────────────────────────────────

func sharingHandler(scoring *mockSharingScoringRepo, users *mockSharingUserRepo, rounds *mockSharingRoundRepo) *SharingHandler {
	return &SharingHandler{
		Scoring: scoring,
		Users:   users,
		Rounds:  rounds,
		Cfg: &config.Config{
			FrontendURL: "https://example.com",
		},
	}
}

func withChiURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func strPtr(s string) *string { return &s }

// ── CreateShareLink ──────────────────────────────────────────────────

func TestCreateShareLink_ExistingToken(t *testing.T) {
	existingToken := "existing-token-abc"
	scoring := &mockSharingScoringRepo{
		getShareTokenResult: &existingToken,
	}
	h := sharingHandler(scoring, nil, nil)

	sessionID := uuid.New().String()
	req := authedRequest(http.MethodPost, "/sessions/"+sessionID, "user-1")
	req = withChiURLParam(req, "id", sessionID)

	rr := httptest.NewRecorder()
	h.CreateShareLink(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var out shareLinkOut
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if out.ShareToken != existingToken {
		t.Errorf("expected token %q, got %q", existingToken, out.ShareToken)
	}
	if out.URL != "https://example.com/shared/"+existingToken {
		t.Errorf("expected URL %q, got %q", "https://example.com/shared/"+existingToken, out.URL)
	}
}

func TestCreateShareLink_NewToken(t *testing.T) {
	scoring := &mockSharingScoringRepo{
		getShareTokenResult: nil, // no existing token
	}
	h := sharingHandler(scoring, nil, nil)

	sessionID := uuid.New().String()
	req := authedRequest(http.MethodPost, "/sessions/"+sessionID, "user-1")
	req = withChiURLParam(req, "id", sessionID)

	rr := httptest.NewRecorder()
	h.CreateShareLink(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var out shareLinkOut
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if out.ShareToken == "" {
		t.Error("expected a non-empty share token")
	}
	if out.URL == "" {
		t.Error("expected a non-empty URL")
	}
}

func TestCreateShareLink_SessionNotFound(t *testing.T) {
	scoring := &mockSharingScoringRepo{
		getShareTokenErr: errors.New("no rows"),
	}
	h := sharingHandler(scoring, nil, nil)

	sessionID := uuid.New().String()
	req := authedRequest(http.MethodPost, "/sessions/"+sessionID, "user-1")
	req = withChiURLParam(req, "id", sessionID)

	rr := httptest.NewRecorder()
	h.CreateShareLink(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestCreateShareLink_InvalidUUID(t *testing.T) {
	h := sharingHandler(&mockSharingScoringRepo{}, nil, nil)

	req := authedRequest(http.MethodPost, "/sessions/not-a-uuid", "user-1")
	req = withChiURLParam(req, "id", "not-a-uuid")

	rr := httptest.NewRecorder()
	h.CreateShareLink(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

// ── RevokeShareLink ──────────────────────────────────────────────────

func TestRevokeShareLink_Success(t *testing.T) {
	scoring := &mockSharingScoringRepo{
		revokeResult: true,
	}
	h := sharingHandler(scoring, nil, nil)

	sessionID := uuid.New().String()
	req := authedRequest(http.MethodDelete, "/sessions/"+sessionID, "user-1")
	req = withChiURLParam(req, "id", sessionID)

	rr := httptest.NewRecorder()
	h.RevokeShareLink(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var out map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if out["detail"] != "Share link revoked" {
		t.Errorf("expected 'Share link revoked', got %q", out["detail"])
	}
}

func TestRevokeShareLink_NotFound(t *testing.T) {
	scoring := &mockSharingScoringRepo{
		revokeResult: false,
	}
	h := sharingHandler(scoring, nil, nil)

	sessionID := uuid.New().String()
	req := authedRequest(http.MethodDelete, "/sessions/"+sessionID, "user-1")
	req = withChiURLParam(req, "id", sessionID)

	rr := httptest.NewRecorder()
	h.RevokeShareLink(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

// ── GetSharedSession ─────────────────────────────────────────────────

func TestGetSharedSession_Success(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	displayName := "Robin Hood"

	scoring := &mockSharingScoringRepo{
		getSharedSessionResult: &repository.SharedSessionData{
			SessionID:   "sess-1",
			TemplateID:  "tmpl-1",
			TotalScore:  280,
			TotalXCount: 5,
			TotalArrows: 30,
			Notes:       strPtr("Good session"),
			Location:    strPtr("Range A"),
			Weather:     strPtr("Sunny"),
			StartedAt:   now,
			CompletedAt: &now,
			UserID:      "user-1",
		},
		loadEndsResult: []repository.EndOut{
			{
				ID:        "end-1",
				EndNumber: 1,
				EndTotal:  50,
				CreatedAt: now,
				Arrows:    []repository.ArrowOut{},
			},
		},
	}

	users := &mockSharingUserRepo{
		username:    "robinhood",
		displayName: &displayName,
		avatar:      strPtr("https://example.com/avatar.png"),
	}

	rounds := &mockSharingRoundRepo{
		result: &repository.RoundTemplateOut{
			ID:           "tmpl-1",
			Name:         "WA 70m",
			Organization: "World Archery",
			IsOfficial:   true,
			Stages:       []repository.StageOut{},
		},
	}

	h := sharingHandler(scoring, users, rounds)

	req := httptest.NewRequest(http.MethodGet, "/s/some-token", nil)
	req = withChiURLParam(req, "token", "some-token")

	rr := httptest.NewRecorder()
	h.GetSharedSession(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var out sharedSessionOut
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if out.ArcherName != "Robin Hood" {
		t.Errorf("expected archer name 'Robin Hood', got %q", out.ArcherName)
	}
	if out.ArcherAvatar == nil || *out.ArcherAvatar != "https://example.com/avatar.png" {
		t.Errorf("unexpected avatar: %v", out.ArcherAvatar)
	}
	if out.TemplateName != "WA 70m" {
		t.Errorf("expected template name 'WA 70m', got %q", out.TemplateName)
	}
	if out.Template == nil {
		t.Fatal("expected template to be non-nil")
	}
	if out.TotalScore != 280 {
		t.Errorf("expected total score 280, got %d", out.TotalScore)
	}
	if out.TotalXCount != 5 {
		t.Errorf("expected total x count 5, got %d", out.TotalXCount)
	}
	if out.TotalArrows != 30 {
		t.Errorf("expected total arrows 30, got %d", out.TotalArrows)
	}
	if len(out.Ends) != 1 {
		t.Fatalf("expected 1 end, got %d", len(out.Ends))
	}
	if out.Ends[0].EndNumber != 1 {
		t.Errorf("expected end number 1, got %d", out.Ends[0].EndNumber)
	}
}

func TestGetSharedSession_NotFound(t *testing.T) {
	scoring := &mockSharingScoringRepo{
		getSharedSessionErr: errors.New("not found"),
	}
	h := sharingHandler(scoring, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/s/bad-token", nil)
	req = withChiURLParam(req, "token", "bad-token")

	rr := httptest.NewRecorder()
	h.GetSharedSession(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}
