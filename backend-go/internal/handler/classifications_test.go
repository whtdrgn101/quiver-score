package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
)

type mockClassificationRepo struct {
	listResult    []repository.ClassificationRecordOut
	listErr       error
	currentResult []repository.CurrentClassificationOut
	currentErr    error
}

func (m *mockClassificationRepo) List(_ context.Context, _ string) ([]repository.ClassificationRecordOut, error) {
	return m.listResult, m.listErr
}

func (m *mockClassificationRepo) Current(_ context.Context, _ string) ([]repository.CurrentClassificationOut, error) {
	return m.currentResult, m.currentErr
}

func authedRequest(method, path, userID string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	return req.WithContext(ctx)
}

func TestClassifications_List_Success(t *testing.T) {
	mock := &mockClassificationRepo{
		listResult: []repository.ClassificationRecordOut{
			{
				ID:             "c-1",
				System:         "AGB",
				Classification: "Bowman",
				RoundType:      "indoor",
				Score:          250,
				AchievedAt:     time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
				SessionID:      "s-1",
			},
		},
	}
	h := &ClassificationsHandler{Classifications: mock}

	rr := httptest.NewRecorder()
	h.List(rr, authedRequest(http.MethodGet, "/", "user-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result []repository.ClassificationRecordOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result))
	}
	if result[0].Classification != "Bowman" {
		t.Errorf("expected 'Bowman', got '%s'", result[0].Classification)
	}
}

func TestClassifications_List_Empty(t *testing.T) {
	mock := &mockClassificationRepo{
		listResult: []repository.ClassificationRecordOut{},
	}
	h := &ClassificationsHandler{Classifications: mock}

	rr := httptest.NewRecorder()
	h.List(rr, authedRequest(http.MethodGet, "/", "user-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result []repository.ClassificationRecordOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 items, got %d", len(result))
	}
}

func TestClassifications_List_DBError(t *testing.T) {
	mock := &mockClassificationRepo{
		listErr: errors.New("connection refused"),
	}
	h := &ClassificationsHandler{Classifications: mock}

	rr := httptest.NewRecorder()
	h.List(rr, authedRequest(http.MethodGet, "/", "user-1"))

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestClassifications_Current_Success(t *testing.T) {
	mock := &mockClassificationRepo{
		currentResult: []repository.CurrentClassificationOut{
			{
				System:         "AGB",
				Classification: "Archer",
				RoundType:      "outdoor",
				Score:          300,
				AchievedAt:     time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	h := &ClassificationsHandler{Classifications: mock}

	rr := httptest.NewRecorder()
	h.Current(rr, authedRequest(http.MethodGet, "/current", "user-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result []repository.CurrentClassificationOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result))
	}
	if result[0].System != "AGB" {
		t.Errorf("expected 'AGB', got '%s'", result[0].System)
	}
}

func TestClassifications_Current_DBError(t *testing.T) {
	mock := &mockClassificationRepo{
		currentErr: errors.New("timeout"),
	}
	h := &ClassificationsHandler{Classifications: mock}

	rr := httptest.NewRecorder()
	h.Current(rr, authedRequest(http.MethodGet, "/current", "user-1"))

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}
