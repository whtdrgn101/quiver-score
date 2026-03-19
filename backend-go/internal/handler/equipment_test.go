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

type mockEquipmentRepo struct {
	listResult   []repository.EquipmentOut
	listErr      error
	createResult *repository.EquipmentOut
	createErr    error
	getResult    *repository.EquipmentOut
	getErr       error
	updateResult *repository.EquipmentOut
	updateErr    error
	deleteResult bool
	deleteErr    error
	statsResult  []repository.EquipmentUsageOut
	statsErr     error
}

func (m *mockEquipmentRepo) List(_ context.Context, _ string) ([]repository.EquipmentOut, error) {
	return m.listResult, m.listErr
}

func (m *mockEquipmentRepo) Create(_ context.Context, _, _, _, _ string, _, _ *string, _ json.RawMessage, _ *string) (*repository.EquipmentOut, error) {
	return m.createResult, m.createErr
}

func (m *mockEquipmentRepo) Get(_ context.Context, _, _ string) (*repository.EquipmentOut, error) {
	return m.getResult, m.getErr
}

func (m *mockEquipmentRepo) Update(_ context.Context, _, _ string, _, _, _, _ *string, _, _ bool, _ json.RawMessage, _ bool, _ *string, _ bool) (*repository.EquipmentOut, error) {
	return m.updateResult, m.updateErr
}

func (m *mockEquipmentRepo) Delete(_ context.Context, _, _ string) (bool, error) {
	return m.deleteResult, m.deleteErr
}

func (m *mockEquipmentRepo) Stats(_ context.Context, _ string) ([]repository.EquipmentUsageOut, error) {
	return m.statsResult, m.statsErr
}

func equipmentRequest(method, path, userID, equipID string) *http.Request {
	req := authedRequest(method, path, userID)
	if equipID != "" {
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", equipID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	}
	return req
}

func TestEquipment_List_Success(t *testing.T) {
	mock := &mockEquipmentRepo{
		listResult: []repository.EquipmentOut{
			{ID: "e-1", Category: "riser", Name: "Hoyt Formula Xi", CreatedAt: time.Now()},
		},
	}
	h := &EquipmentHandler{Equipment: mock}

	rr := httptest.NewRecorder()
	h.List(rr, authedRequest(http.MethodGet, "/", "user-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestEquipment_List_DBError(t *testing.T) {
	mock := &mockEquipmentRepo{listErr: errors.New("db error")}
	h := &EquipmentHandler{Equipment: mock}

	rr := httptest.NewRecorder()
	h.List(rr, authedRequest(http.MethodGet, "/", "user-1"))

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestEquipment_Create_Success(t *testing.T) {
	mock := &mockEquipmentRepo{
		createResult: &repository.EquipmentOut{ID: "e-1", Category: "riser", Name: "Hoyt Formula Xi"},
	}
	h := &EquipmentHandler{Equipment: mock}

	body := strings.NewReader(`{"category":"riser","name":"Hoyt Formula Xi"}`)
	req := authedRequest(http.MethodPost, "/", "user-1")
	req.Body = http.NoBody
	req = httptest.NewRequest(http.MethodPost, "/", body)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-1")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}
}

func TestEquipment_Create_MissingName(t *testing.T) {
	h := &EquipmentHandler{Equipment: &mockEquipmentRepo{}}

	body := strings.NewReader(`{"category":"riser"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-1")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rr.Code)
	}
}

func TestEquipment_Create_InvalidCategory(t *testing.T) {
	h := &EquipmentHandler{Equipment: &mockEquipmentRepo{}}

	body := strings.NewReader(`{"category":"invalid","name":"Test"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-1")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rr.Code)
	}
}

func TestEquipment_Get_Success(t *testing.T) {
	id := uuid.New().String()
	mock := &mockEquipmentRepo{
		getResult: &repository.EquipmentOut{ID: id, Category: "riser", Name: "Test"},
	}
	h := &EquipmentHandler{Equipment: mock}

	rr := httptest.NewRecorder()
	h.Get(rr, equipmentRequest(http.MethodGet, "/"+id, "user-1", id))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestEquipment_Get_InvalidUUID(t *testing.T) {
	h := &EquipmentHandler{Equipment: &mockEquipmentRepo{}}

	rr := httptest.NewRecorder()
	h.Get(rr, equipmentRequest(http.MethodGet, "/not-a-uuid", "user-1", "not-a-uuid"))

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestEquipment_Get_NotFound(t *testing.T) {
	id := uuid.New().String()
	mock := &mockEquipmentRepo{getErr: errors.New("not found")}
	h := &EquipmentHandler{Equipment: mock}

	rr := httptest.NewRecorder()
	h.Get(rr, equipmentRequest(http.MethodGet, "/"+id, "user-1", id))

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestEquipment_Delete_Success(t *testing.T) {
	id := uuid.New().String()
	mock := &mockEquipmentRepo{deleteResult: true}
	h := &EquipmentHandler{Equipment: mock}

	rr := httptest.NewRecorder()
	h.Delete(rr, equipmentRequest(http.MethodDelete, "/"+id, "user-1", id))

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

func TestEquipment_Delete_NotFound(t *testing.T) {
	id := uuid.New().String()
	mock := &mockEquipmentRepo{deleteResult: false}
	h := &EquipmentHandler{Equipment: mock}

	rr := httptest.NewRecorder()
	h.Delete(rr, equipmentRequest(http.MethodDelete, "/"+id, "user-1", id))

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestEquipment_Stats_Success(t *testing.T) {
	mock := &mockEquipmentRepo{
		statsResult: []repository.EquipmentUsageOut{
			{ItemID: "e-1", ItemName: "Riser", Category: "riser", SessionsCount: 5, TotalArrows: 300},
		},
	}
	h := &EquipmentHandler{Equipment: mock}

	rr := httptest.NewRecorder()
	h.Stats(rr, authedRequest(http.MethodGet, "/stats", "user-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestEquipment_Stats_DBError(t *testing.T) {
	mock := &mockEquipmentRepo{statsErr: errors.New("db error")}
	h := &EquipmentHandler{Equipment: mock}

	rr := httptest.NewRecorder()
	h.Stats(rr, authedRequest(http.MethodGet, "/stats", "user-1"))

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestEquipment_Update_Success(t *testing.T) {
	id := uuid.New().String()
	mock := &mockEquipmentRepo{
		updateResult: &repository.EquipmentOut{ID: id, Category: "riser", Name: "Updated"},
	}
	h := &EquipmentHandler{Equipment: mock}

	body := strings.NewReader(`{"name":"Updated"}`)
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
}

func TestEquipment_Update_InvalidCategory(t *testing.T) {
	id := uuid.New().String()
	h := &EquipmentHandler{Equipment: &mockEquipmentRepo{}}

	body := strings.NewReader(`{"category":"invalid"}`)
	req := httptest.NewRequest(http.MethodPut, "/"+id, body)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.Update(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rr.Code)
	}
}
