package handler

import (
	"context"
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

type mockSightMarkRepo struct {
	listResult   []repository.SightMarkOut
	listErr      error
	createResult *repository.SightMarkOut
	createErr    error
	existsResult bool
	existsErr    error
	updateResult *repository.SightMarkOut
	updateErr    error
	deleteResult bool
	deleteErr    error
}

func (m *mockSightMarkRepo) List(_ context.Context, _ string, _, _ *string) ([]repository.SightMarkOut, error) {
	return m.listResult, m.listErr
}

func (m *mockSightMarkRepo) Create(_ context.Context, _, _ string, _, _ *string, _, _ string, _ *string, _ time.Time) (*repository.SightMarkOut, error) {
	return m.createResult, m.createErr
}

func (m *mockSightMarkRepo) Exists(_ context.Context, _, _ string) (bool, error) {
	return m.existsResult, m.existsErr
}

func (m *mockSightMarkRepo) Update(_ context.Context, _, _ string, _, _, _ *string, _ bool, _ *time.Time, _ bool, _ *string, _ bool, _ *string, _ bool) (*repository.SightMarkOut, error) {
	return m.updateResult, m.updateErr
}

func (m *mockSightMarkRepo) Delete(_ context.Context, _, _ string) (bool, error) {
	return m.deleteResult, m.deleteErr
}

func sightMarkRequest(method, path, userID, smID string) *http.Request {
	req := authedRequest(method, path, userID)
	if smID != "" {
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", smID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	}
	return req
}

func TestSightMarks_List_Success(t *testing.T) {
	mock := &mockSightMarkRepo{
		listResult: []repository.SightMarkOut{
			{ID: "sm-1", Distance: "20yd", Setting: "4.5"},
		},
	}
	h := &SightMarksHandler{SightMarks: mock}

	rr := httptest.NewRecorder()
	h.List(rr, authedRequest(http.MethodGet, "/", "user-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestSightMarks_List_DBError(t *testing.T) {
	mock := &mockSightMarkRepo{listErr: errors.New("db error")}
	h := &SightMarksHandler{SightMarks: mock}

	rr := httptest.NewRecorder()
	h.List(rr, authedRequest(http.MethodGet, "/", "user-1"))

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestSightMarks_List_InvalidFilterUUID(t *testing.T) {
	mock := &mockSightMarkRepo{}
	h := &SightMarksHandler{SightMarks: mock}

	rr := httptest.NewRecorder()
	h.List(rr, authedRequest(http.MethodGet, "/?equipment_id=not-valid", "user-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 with empty list, got %d", rr.Code)
	}
}

func TestSightMarks_Create_Success(t *testing.T) {
	mock := &mockSightMarkRepo{
		createResult: &repository.SightMarkOut{ID: "sm-1", Distance: "20yd", Setting: "4.5"},
	}
	h := &SightMarksHandler{SightMarks: mock}

	body := strings.NewReader(`{"distance":"20yd","setting":"4.5","date_recorded":"2026-01-15T00:00:00Z"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-1")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}
}

func TestSightMarks_Create_MissingRequired(t *testing.T) {
	h := &SightMarksHandler{SightMarks: &mockSightMarkRepo{}}

	body := strings.NewReader(`{"distance":"20yd"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-1")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rr.Code)
	}
}

func TestSightMarks_Create_MissingDateRecorded(t *testing.T) {
	h := &SightMarksHandler{SightMarks: &mockSightMarkRepo{}}

	body := strings.NewReader(`{"distance":"20yd","setting":"4.5"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-1")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rr.Code)
	}
}

func TestSightMarks_Create_InvalidDate(t *testing.T) {
	h := &SightMarksHandler{SightMarks: &mockSightMarkRepo{}}

	body := strings.NewReader(`{"distance":"20yd","setting":"4.5","date_recorded":"not-a-date"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-1")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rr.Code)
	}
}

func TestSightMarks_Delete_Success(t *testing.T) {
	id := uuid.New().String()
	mock := &mockSightMarkRepo{deleteResult: true}
	h := &SightMarksHandler{SightMarks: mock}

	rr := httptest.NewRecorder()
	h.Delete(rr, sightMarkRequest(http.MethodDelete, "/"+id, "user-1", id))

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

func TestSightMarks_Delete_NotFound(t *testing.T) {
	id := uuid.New().String()
	mock := &mockSightMarkRepo{deleteResult: false}
	h := &SightMarksHandler{SightMarks: mock}

	rr := httptest.NewRecorder()
	h.Delete(rr, sightMarkRequest(http.MethodDelete, "/"+id, "user-1", id))

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestSightMarks_Delete_InvalidUUID(t *testing.T) {
	h := &SightMarksHandler{SightMarks: &mockSightMarkRepo{}}

	rr := httptest.NewRecorder()
	h.Delete(rr, sightMarkRequest(http.MethodDelete, "/bad", "user-1", "bad"))

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestSightMarks_Update_Success(t *testing.T) {
	id := uuid.New().String()
	mock := &mockSightMarkRepo{
		existsResult: true,
		updateResult: &repository.SightMarkOut{ID: id, Distance: "30yd", Setting: "5.0"},
	}
	h := &SightMarksHandler{SightMarks: mock}

	body := strings.NewReader(`{"distance":"30yd"}`)
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

func TestSightMarks_Update_NotFound(t *testing.T) {
	id := uuid.New().String()
	mock := &mockSightMarkRepo{existsResult: false}
	h := &SightMarksHandler{SightMarks: mock}

	body := strings.NewReader(`{"distance":"30yd"}`)
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
