package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
)

func withChiURLParams(r *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// ── Mock ──────────────────────────────────────────────────────────────

type mockRoundRepo struct {
	listResult []repository.RoundTemplateOut
	listErr    error

	getResult *repository.RoundTemplateOut
	getErr    error

	createResult *repository.RoundTemplateOut
	createErr    error

	updateResult *repository.RoundTemplateOut
	updateErr    error

	deleteErr error

	permIsOfficial bool
	permCreatedBy  *string
	permErr        error

	hasInProgress    bool
	hasInProgressErr error

	isMember    bool
	isMemberErr error

	isShared    bool
	isSharedErr error

	shareErr error

	unshareOK  bool
	unshareErr error
}

func (m *mockRoundRepo) List(_ context.Context, _ string) ([]repository.RoundTemplateOut, error) {
	return m.listResult, m.listErr
}

func (m *mockRoundRepo) Get(_ context.Context, _ string) (*repository.RoundTemplateOut, error) {
	return m.getResult, m.getErr
}

func (m *mockRoundRepo) Create(_ context.Context, _, _ string, _ *string, _ string, _ []repository.StageParams) (*repository.RoundTemplateOut, error) {
	return m.createResult, m.createErr
}

func (m *mockRoundRepo) Update(_ context.Context, _, _, _ string, _ *string, _ []repository.StageParams) (*repository.RoundTemplateOut, error) {
	return m.updateResult, m.updateErr
}

func (m *mockRoundRepo) Delete(_ context.Context, _ string) error {
	return m.deleteErr
}

func (m *mockRoundRepo) GetPermissions(_ context.Context, _ string) (bool, *string, error) {
	return m.permIsOfficial, m.permCreatedBy, m.permErr
}

func (m *mockRoundRepo) HasInProgressSessions(_ context.Context, _ string) (bool, error) {
	return m.hasInProgress, m.hasInProgressErr
}

func (m *mockRoundRepo) IsMemberOfClub(_ context.Context, _, _ string) (bool, error) {
	return m.isMember, m.isMemberErr
}

func (m *mockRoundRepo) IsSharedWithClub(_ context.Context, _, _ string) (bool, error) {
	return m.isShared, m.isSharedErr
}

func (m *mockRoundRepo) ShareWithClub(_ context.Context, _, _, _ string) error {
	return m.shareErr
}

func (m *mockRoundRepo) UnshareFromClub(_ context.Context, _, _ string) (bool, error) {
	return m.unshareOK, m.unshareErr
}

// ── Helpers ───────────────────────────────────────────────────────────

func roundRequest(method, path, userID string, body any) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	return req.WithContext(ctx)
}

// sampleStages returns a minimal valid stage list for create/update requests.
func sampleStages() []map[string]any {
	return []map[string]any{
		{
			"name":               "Stage 1",
			"distance":           "18m",
			"num_ends":           10,
			"arrows_per_end":     3,
			"allowed_values":     []string{"X", "10", "9"},
			"value_score_map":    map[string]int{"X": 10, "10": 10, "9": 9},
			"max_score_per_arrow": 10,
		},
	}
}

func sampleTemplate() *repository.RoundTemplateOut {
	return &repository.RoundTemplateOut{
		ID:           uuid.New().String(),
		Name:         "WA 18m",
		Organization: "World Archery",
		IsOfficial:   false,
		CreatedBy:    strPtr("user-1"),
		Stages: []repository.StageOut{
			{
				ID:               uuid.New().String(),
				StageOrder:       1,
				Name:             "Stage 1",
				Distance:         strPtr("18m"),
				NumEnds:          10,
				ArrowsPerEnd:     3,
				AllowedValues:    []string{"X", "10", "9"},
				ValueScoreMap:    map[string]int{"X": 10, "10": 10, "9": 9},
				MaxScorePerArrow: 10,
			},
		},
	}
}

// ── List ──────────────────────────────────────────────────────────────

func TestRounds_List_Success(t *testing.T) {
	mock := &mockRoundRepo{
		listResult: []repository.RoundTemplateOut{*sampleTemplate()},
	}
	h := &RoundsHandler{Rounds: mock}

	rr := httptest.NewRecorder()
	h.List(rr, authedRequest(http.MethodGet, "/", "user-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result []repository.RoundTemplateOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result))
	}
	if result[0].Name != "WA 18m" {
		t.Errorf("expected 'WA 18m', got '%s'", result[0].Name)
	}
}

// ── Get ───────────────────────────────────────────────────────────────

func TestRounds_Get_Success(t *testing.T) {
	tmpl := sampleTemplate()
	mock := &mockRoundRepo{getResult: tmpl}
	h := &RoundsHandler{Rounds: mock}

	validID := uuid.New().String()
	req := authedRequest(http.MethodGet, "/"+validID, "user-1")
	req = withChiURLParam(req, "id", validID)

	rr := httptest.NewRecorder()
	h.Get(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result repository.RoundTemplateOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.Name != "WA 18m" {
		t.Errorf("expected 'WA 18m', got '%s'", result.Name)
	}
}

func TestRounds_Get_InvalidUUID(t *testing.T) {
	mock := &mockRoundRepo{}
	h := &RoundsHandler{Rounds: mock}

	req := authedRequest(http.MethodGet, "/not-a-uuid", "user-1")
	req = withChiURLParam(req, "id", "not-a-uuid")

	rr := httptest.NewRecorder()
	h.Get(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestRounds_Get_NotFound(t *testing.T) {
	mock := &mockRoundRepo{getErr: errors.New("no rows")}
	h := &RoundsHandler{Rounds: mock}

	validID := uuid.New().String()
	req := authedRequest(http.MethodGet, "/"+validID, "user-1")
	req = withChiURLParam(req, "id", validID)

	rr := httptest.NewRecorder()
	h.Get(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

// ── Create ────────────────────────────────────────────────────────────

func TestRounds_Create_Success(t *testing.T) {
	tmpl := sampleTemplate()
	mock := &mockRoundRepo{createResult: tmpl}
	h := &RoundsHandler{Rounds: mock}

	body := map[string]any{
		"name":         "WA 18m",
		"organization": "World Archery",
		"stages":       sampleStages(),
	}
	req := roundRequest(http.MethodPost, "/", "user-1", body)

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}

	var result repository.RoundTemplateOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.Name != "WA 18m" {
		t.Errorf("expected 'WA 18m', got '%s'", result.Name)
	}
}

func TestRounds_Create_MissingName(t *testing.T) {
	mock := &mockRoundRepo{}
	h := &RoundsHandler{Rounds: mock}

	body := map[string]any{
		"organization": "World Archery",
		"stages":       sampleStages(),
	}
	req := roundRequest(http.MethodPost, "/", "user-1", body)

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rr.Code)
	}
}

func TestRounds_Create_MissingStages(t *testing.T) {
	mock := &mockRoundRepo{}
	h := &RoundsHandler{Rounds: mock}

	body := map[string]any{
		"name":         "WA 18m",
		"organization": "World Archery",
	}
	req := roundRequest(http.MethodPost, "/", "user-1", body)

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rr.Code)
	}
}

// ── Update ────────────────────────────────────────────────────────────

func TestRounds_Update_Success(t *testing.T) {
	userID := "user-1"
	tmpl := sampleTemplate()
	mock := &mockRoundRepo{
		updateResult:   tmpl,
		permIsOfficial: false,
		permCreatedBy:  strPtr(userID),
		hasInProgress:  false,
	}
	h := &RoundsHandler{Rounds: mock}

	validID := uuid.New().String()
	body := map[string]any{
		"name":         "WA 18m Updated",
		"organization": "World Archery",
		"stages":       sampleStages(),
	}
	req := roundRequest(http.MethodPut, "/"+validID, userID, body)
	req = withChiURLParam(req, "id", validID)

	rr := httptest.NewRecorder()
	h.Update(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestRounds_Update_OfficialRound(t *testing.T) {
	userID := "user-1"
	mock := &mockRoundRepo{
		permIsOfficial: true,
		permCreatedBy:  strPtr(userID),
	}
	h := &RoundsHandler{Rounds: mock}

	validID := uuid.New().String()
	body := map[string]any{
		"name":         "WA 18m",
		"organization": "World Archery",
		"stages":       sampleStages(),
	}
	req := roundRequest(http.MethodPut, "/"+validID, userID, body)
	req = withChiURLParam(req, "id", validID)

	rr := httptest.NewRecorder()
	h.Update(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestRounds_Update_NotOwner(t *testing.T) {
	mock := &mockRoundRepo{
		permIsOfficial: false,
		permCreatedBy:  strPtr("other-user"),
	}
	h := &RoundsHandler{Rounds: mock}

	validID := uuid.New().String()
	body := map[string]any{
		"name":         "WA 18m",
		"organization": "World Archery",
		"stages":       sampleStages(),
	}
	req := roundRequest(http.MethodPut, "/"+validID, "user-1", body)
	req = withChiURLParam(req, "id", validID)

	rr := httptest.NewRecorder()
	h.Update(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestRounds_Update_InProgressSession(t *testing.T) {
	userID := "user-1"
	mock := &mockRoundRepo{
		permIsOfficial: false,
		permCreatedBy:  strPtr(userID),
		hasInProgress:  true,
	}
	h := &RoundsHandler{Rounds: mock}

	validID := uuid.New().String()
	body := map[string]any{
		"name":         "WA 18m",
		"organization": "World Archery",
		"stages":       sampleStages(),
	}
	req := roundRequest(http.MethodPut, "/"+validID, userID, body)
	req = withChiURLParam(req, "id", validID)

	rr := httptest.NewRecorder()
	h.Update(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rr.Code)
	}
}

// ── Delete ────────────────────────────────────────────────────────────

func TestRounds_Delete_Success(t *testing.T) {
	userID := "user-1"
	mock := &mockRoundRepo{
		permIsOfficial: false,
		permCreatedBy:  strPtr(userID),
	}
	h := &RoundsHandler{Rounds: mock}

	validID := uuid.New().String()
	req := authedRequest(http.MethodDelete, "/"+validID, userID)
	req = withChiURLParam(req, "id", validID)

	rr := httptest.NewRecorder()
	h.Delete(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

func TestRounds_Delete_OfficialRound(t *testing.T) {
	userID := "user-1"
	mock := &mockRoundRepo{
		permIsOfficial: true,
		permCreatedBy:  strPtr(userID),
	}
	h := &RoundsHandler{Rounds: mock}

	validID := uuid.New().String()
	req := authedRequest(http.MethodDelete, "/"+validID, userID)
	req = withChiURLParam(req, "id", validID)

	rr := httptest.NewRecorder()
	h.Delete(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestRounds_Delete_NotOwner(t *testing.T) {
	mock := &mockRoundRepo{
		permIsOfficial: false,
		permCreatedBy:  strPtr("other-user"),
	}
	h := &RoundsHandler{Rounds: mock}

	validID := uuid.New().String()
	req := authedRequest(http.MethodDelete, "/"+validID, "user-1")
	req = withChiURLParam(req, "id", validID)

	rr := httptest.NewRecorder()
	h.Delete(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

// ── Share ─────────────────────────────────────────────────────────────

func TestRounds_Share_Success(t *testing.T) {
	userID := "user-1"
	clubID := uuid.New().String()
	mock := &mockRoundRepo{
		permIsOfficial: false,
		permCreatedBy:  strPtr(userID),
		isMember:       true,
		isShared:       false,
		isSharedErr:    errors.New("not found"),
	}
	h := &RoundsHandler{Rounds: mock}

	validID := uuid.New().String()
	body := map[string]any{"club_id": clubID}
	req := roundRequest(http.MethodPost, "/"+validID+"/share", userID, body)
	req = withChiURLParam(req, "id", validID)

	rr := httptest.NewRecorder()
	h.Share(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}
}

func TestRounds_Share_OfficialRound(t *testing.T) {
	userID := "user-1"
	clubID := uuid.New().String()
	mock := &mockRoundRepo{
		permIsOfficial: true,
		permCreatedBy:  strPtr(userID),
	}
	h := &RoundsHandler{Rounds: mock}

	validID := uuid.New().String()
	body := map[string]any{"club_id": clubID}
	req := roundRequest(http.MethodPost, "/"+validID+"/share", userID, body)
	req = withChiURLParam(req, "id", validID)

	rr := httptest.NewRecorder()
	h.Share(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestRounds_Share_AlreadyShared(t *testing.T) {
	userID := "user-1"
	clubID := uuid.New().String()
	mock := &mockRoundRepo{
		permIsOfficial: false,
		permCreatedBy:  strPtr(userID),
		isMember:       true,
		isShared:       true,
	}
	h := &RoundsHandler{Rounds: mock}

	validID := uuid.New().String()
	body := map[string]any{"club_id": clubID}
	req := roundRequest(http.MethodPost, "/"+validID+"/share", userID, body)
	req = withChiURLParam(req, "id", validID)

	rr := httptest.NewRecorder()
	h.Share(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", rr.Code)
	}
}

// ── Unshare ───────────────────────────────────────────────────────────

func TestRounds_Unshare_Success(t *testing.T) {
	userID := "user-1"
	mock := &mockRoundRepo{
		permIsOfficial: false,
		permCreatedBy:  strPtr(userID),
		unshareOK:      true,
	}
	h := &RoundsHandler{Rounds: mock}

	validID := uuid.New().String()
	clubID := uuid.New().String()
	req := authedRequest(http.MethodDelete, "/"+validID+"/share/"+clubID, userID)
	req = withChiURLParams(req, map[string]string{"id": validID, "club_id": clubID})

	rr := httptest.NewRecorder()
	h.Unshare(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

func TestRounds_Unshare_NotFound(t *testing.T) {
	userID := "user-1"
	mock := &mockRoundRepo{
		permIsOfficial: false,
		permCreatedBy:  strPtr(userID),
		unshareOK:      false,
	}
	h := &RoundsHandler{Rounds: mock}

	validID := uuid.New().String()
	clubID := uuid.New().String()
	req := authedRequest(http.MethodDelete, "/"+validID+"/share/"+clubID, userID)
	req = withChiURLParams(req, map[string]string{"id": validID, "club_id": clubID})

	rr := httptest.NewRecorder()
	h.Unshare(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}
