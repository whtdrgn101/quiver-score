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

	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
)

// ── Mock ──────────────────────────────────────────────────────────────

type mockSetupRepo struct {
	listResult          []repository.SetupProfileSummary
	listErr             error
	createID            string
	createErr           error
	loadResult          *repository.SetupProfileOut
	loadErr             error
	existsResult        bool
	existsErr           error
	updateResult        bool
	updateErr           error
	deleteResult        bool
	deleteErr           error
	equipExistsResult   bool
	equipExistsErr      error
	equipLinkedResult   bool
	equipLinkedErr      error
	addEquipErr         error
	removeEquipResult   bool
	removeEquipErr      error
}

func (m *mockSetupRepo) List(_ context.Context, _ string) ([]repository.SetupProfileSummary, error) {
	return m.listResult, m.listErr
}

func (m *mockSetupRepo) Create(_ context.Context, _, _ string, _ *string, _, _, _, _, _ *float64) (string, error) {
	return m.createID, m.createErr
}

func (m *mockSetupRepo) LoadWithEquipment(_ context.Context, _, _ string) (*repository.SetupProfileOut, error) {
	return m.loadResult, m.loadErr
}

func (m *mockSetupRepo) Exists(_ context.Context, _, _ string) (bool, error) {
	return m.existsResult, m.existsErr
}

func (m *mockSetupRepo) Update(_ context.Context, _, _ string, _ *string, _ *string, _ bool, _ *float64, _ bool, _ *float64, _ bool, _ *float64, _ bool, _ *float64, _ bool, _ *float64, _ bool) (bool, error) {
	return m.updateResult, m.updateErr
}

func (m *mockSetupRepo) Delete(_ context.Context, _, _ string) (bool, error) {
	return m.deleteResult, m.deleteErr
}

func (m *mockSetupRepo) EquipmentExists(_ context.Context, _, _ string) (bool, error) {
	return m.equipExistsResult, m.equipExistsErr
}

func (m *mockSetupRepo) EquipmentLinked(_ context.Context, _, _ string) (bool, error) {
	return m.equipLinkedResult, m.equipLinkedErr
}

func (m *mockSetupRepo) AddEquipment(_ context.Context, _, _ string) error {
	return m.addEquipErr
}

func (m *mockSetupRepo) RemoveEquipment(_ context.Context, _, _ string) (bool, error) {
	return m.removeEquipResult, m.removeEquipErr
}

// ── Helpers ───────────────────────────────────────────────────────────

func setupRequest(method, path, userID, setupID string) *http.Request {
	req := authedRequest(method, path, userID)
	if setupID != "" {
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", setupID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	}
	return req
}

func setupEquipRequest(method, path, userID, setupID, equipID string) *http.Request {
	req := authedRequest(method, path, userID)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", setupID)
	rctx.URLParams.Add("equipment_id", equipID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	return req
}

func sampleSetupProfile(id string) *repository.SetupProfileOut {
	return &repository.SetupProfileOut{
		ID:        id,
		Name:      "My Setup",
		Equipment: []repository.EquipmentOut{},
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

// ── List ──────────────────────────────────────────────────────────────

func TestSetups_List_Success(t *testing.T) {
	mock := &mockSetupRepo{
		listResult: []repository.SetupProfileSummary{
			{ID: "s-1", Name: "Outdoor", EquipmentCount: 3, CreatedAt: time.Now()},
		},
	}
	h := &SetupsHandler{Setups: mock}

	rr := httptest.NewRecorder()
	h.List(rr, authedRequest(http.MethodGet, "/", "user-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result []repository.SetupProfileSummary
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result))
	}
	if result[0].Name != "Outdoor" {
		t.Errorf("expected 'Outdoor', got '%s'", result[0].Name)
	}
}

func TestSetups_List_DBError(t *testing.T) {
	mock := &mockSetupRepo{listErr: errors.New("db error")}
	h := &SetupsHandler{Setups: mock}

	rr := httptest.NewRecorder()
	h.List(rr, authedRequest(http.MethodGet, "/", "user-1"))

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

// ── Create ────────────────────────────────────────────────────────────

func TestSetups_Create_Success(t *testing.T) {
	id := uuid.New().String()
	mock := &mockSetupRepo{
		createID:   id,
		loadResult: sampleSetupProfile(id),
	}
	h := &SetupsHandler{Setups: mock}

	body := strings.NewReader(`{"name":"My Setup"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-1")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}

	var result repository.SetupProfileOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.ID != id {
		t.Errorf("expected id '%s', got '%s'", id, result.ID)
	}
}

func TestSetups_Create_MissingName(t *testing.T) {
	h := &SetupsHandler{Setups: &mockSetupRepo{}}

	body := strings.NewReader(`{"description":"no name"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-1")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rr.Code)
	}
}

// ── Get ───────────────────────────────────────────────────────────────

func TestSetups_Get_Success(t *testing.T) {
	id := uuid.New().String()
	mock := &mockSetupRepo{
		loadResult: sampleSetupProfile(id),
	}
	h := &SetupsHandler{Setups: mock}

	rr := httptest.NewRecorder()
	h.Get(rr, setupRequest(http.MethodGet, "/"+id, "user-1", id))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result repository.SetupProfileOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.ID != id {
		t.Errorf("expected id '%s', got '%s'", id, result.ID)
	}
}

func TestSetups_Get_InvalidUUID(t *testing.T) {
	h := &SetupsHandler{Setups: &mockSetupRepo{}}

	rr := httptest.NewRecorder()
	h.Get(rr, setupRequest(http.MethodGet, "/not-a-uuid", "user-1", "not-a-uuid"))

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestSetups_Get_NotFound(t *testing.T) {
	id := uuid.New().String()
	mock := &mockSetupRepo{loadErr: errors.New("not found")}
	h := &SetupsHandler{Setups: mock}

	rr := httptest.NewRecorder()
	h.Get(rr, setupRequest(http.MethodGet, "/"+id, "user-1", id))

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

// ── Update ────────────────────────────────────────────────────────────

func TestSetups_Update_Success(t *testing.T) {
	id := uuid.New().String()
	updated := sampleSetupProfile(id)
	updated.Name = "Updated Setup"
	mock := &mockSetupRepo{
		existsResult: true,
		updateResult: true,
		loadResult:   updated,
	}
	h := &SetupsHandler{Setups: mock}

	body := strings.NewReader(`{"name":"Updated Setup"}`)
	req := httptest.NewRequest(http.MethodPut, "/"+id, body)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.Update(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result repository.SetupProfileOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.Name != "Updated Setup" {
		t.Errorf("expected 'Updated Setup', got '%s'", result.Name)
	}
}

func TestSetups_Update_NotFound(t *testing.T) {
	id := uuid.New().String()
	mock := &mockSetupRepo{existsResult: false}
	h := &SetupsHandler{Setups: mock}

	body := strings.NewReader(`{"name":"Updated"}`)
	req := httptest.NewRequest(http.MethodPut, "/"+id, body)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.Update(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

// ── Delete ────────────────────────────────────────────────────────────

func TestSetups_Delete_Success(t *testing.T) {
	id := uuid.New().String()
	mock := &mockSetupRepo{deleteResult: true}
	h := &SetupsHandler{Setups: mock}

	rr := httptest.NewRecorder()
	h.Delete(rr, setupRequest(http.MethodDelete, "/"+id, "user-1", id))

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

func TestSetups_Delete_NotFound(t *testing.T) {
	id := uuid.New().String()
	mock := &mockSetupRepo{deleteResult: false}
	h := &SetupsHandler{Setups: mock}

	rr := httptest.NewRecorder()
	h.Delete(rr, setupRequest(http.MethodDelete, "/"+id, "user-1", id))

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

// ── AddEquipment ──────────────────────────────────────────────────────

func TestSetups_AddEquipment_Success(t *testing.T) {
	setupID := uuid.New().String()
	equipID := uuid.New().String()
	mock := &mockSetupRepo{
		existsResult:      true,
		equipExistsResult: true,
		equipLinkedResult: false,
		loadResult:        sampleSetupProfile(setupID),
	}
	h := &SetupsHandler{Setups: mock}

	rr := httptest.NewRecorder()
	h.AddEquipment(rr, setupEquipRequest(http.MethodPost, "/"+setupID+"/equipment/"+equipID, "user-1", setupID, equipID))

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}

	var result repository.SetupProfileOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.ID != setupID {
		t.Errorf("expected id '%s', got '%s'", setupID, result.ID)
	}
}

func TestSetups_AddEquipment_AlreadyLinked(t *testing.T) {
	setupID := uuid.New().String()
	equipID := uuid.New().String()
	mock := &mockSetupRepo{
		existsResult:      true,
		equipExistsResult: true,
		equipLinkedResult: true,
	}
	h := &SetupsHandler{Setups: mock}

	rr := httptest.NewRecorder()
	h.AddEquipment(rr, setupEquipRequest(http.MethodPost, "/"+setupID+"/equipment/"+equipID, "user-1", setupID, equipID))

	if rr.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", rr.Code)
	}
}

// ── RemoveEquipment ───────────────────────────────────────────────────

func TestSetups_RemoveEquipment_Success(t *testing.T) {
	setupID := uuid.New().String()
	equipID := uuid.New().String()
	mock := &mockSetupRepo{
		existsResult:    true,
		removeEquipResult: true,
	}
	h := &SetupsHandler{Setups: mock}

	rr := httptest.NewRecorder()
	h.RemoveEquipment(rr, setupEquipRequest(http.MethodDelete, "/"+setupID+"/equipment/"+equipID, "user-1", setupID, equipID))

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

func TestSetups_RemoveEquipment_NotLinked(t *testing.T) {
	setupID := uuid.New().String()
	equipID := uuid.New().String()
	mock := &mockSetupRepo{
		existsResult:    true,
		removeEquipResult: false,
	}
	h := &SetupsHandler{Setups: mock}

	rr := httptest.NewRecorder()
	h.RemoveEquipment(rr, setupEquipRequest(http.MethodDelete, "/"+setupID+"/equipment/"+equipID, "user-1", setupID, equipID))

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}
