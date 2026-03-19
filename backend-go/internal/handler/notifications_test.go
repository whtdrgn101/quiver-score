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

	"github.com/quiverscore/backend-go/internal/repository"
)

type mockNotificationRepo struct {
	listResult     []repository.NotificationOut
	listErr        error
	unreadResult   int
	unreadErr      error
	markReadResult *repository.NotificationOut
	markReadErr    error
	markAllErr     error
}

func (m *mockNotificationRepo) List(_ context.Context, _ string) ([]repository.NotificationOut, error) {
	return m.listResult, m.listErr
}

func (m *mockNotificationRepo) UnreadCount(_ context.Context, _ string) (int, error) {
	return m.unreadResult, m.unreadErr
}

func (m *mockNotificationRepo) MarkRead(_ context.Context, _, _ string) (*repository.NotificationOut, error) {
	return m.markReadResult, m.markReadErr
}

func (m *mockNotificationRepo) MarkAllRead(_ context.Context, _ string) error {
	return m.markAllErr
}

func TestNotifications_List_Success(t *testing.T) {
	mock := &mockNotificationRepo{
		listResult: []repository.NotificationOut{
			{ID: "n-1", Type: "score", Title: "New PR!", Message: "You set a new personal record", Read: false, CreatedAt: time.Now()},
		},
	}
	h := &NotificationsHandler{Notifications: mock}

	rr := httptest.NewRecorder()
	h.List(rr, authedRequest(http.MethodGet, "/", "user-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	var result []repository.NotificationOut
	json.NewDecoder(rr.Body).Decode(&result)
	if len(result) != 1 {
		t.Errorf("expected 1 item, got %d", len(result))
	}
}

func TestNotifications_List_DBError(t *testing.T) {
	mock := &mockNotificationRepo{listErr: errors.New("db error")}
	h := &NotificationsHandler{Notifications: mock}

	rr := httptest.NewRecorder()
	h.List(rr, authedRequest(http.MethodGet, "/", "user-1"))

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestNotifications_UnreadCount_Success(t *testing.T) {
	mock := &mockNotificationRepo{unreadResult: 3}
	h := &NotificationsHandler{Notifications: mock}

	rr := httptest.NewRecorder()
	h.UnreadCount(rr, authedRequest(http.MethodGet, "/unread-count", "user-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	var result map[string]int
	json.NewDecoder(rr.Body).Decode(&result)
	if result["count"] != 3 {
		t.Errorf("expected count 3, got %d", result["count"])
	}
}

func TestNotifications_UnreadCount_DBError(t *testing.T) {
	mock := &mockNotificationRepo{unreadErr: errors.New("db error")}
	h := &NotificationsHandler{Notifications: mock}

	rr := httptest.NewRecorder()
	h.UnreadCount(rr, authedRequest(http.MethodGet, "/unread-count", "user-1"))

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestNotifications_MarkRead_Success(t *testing.T) {
	mock := &mockNotificationRepo{
		markReadResult: &repository.NotificationOut{ID: "n-1", Read: true},
	}
	h := &NotificationsHandler{Notifications: mock}

	rr := httptest.NewRecorder()
	req := authedRequest(http.MethodPatch, "/n-1/read", "user-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("notificationID", "n-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.MarkRead(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestNotifications_MarkRead_NotFound(t *testing.T) {
	mock := &mockNotificationRepo{markReadErr: repository.ErrNotFound}
	h := &NotificationsHandler{Notifications: mock}

	rr := httptest.NewRecorder()
	req := authedRequest(http.MethodPatch, "/n-1/read", "user-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("notificationID", "n-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.MarkRead(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestNotifications_MarkAllRead_Success(t *testing.T) {
	mock := &mockNotificationRepo{}
	h := &NotificationsHandler{Notifications: mock}

	rr := httptest.NewRecorder()
	h.MarkAllRead(rr, authedRequest(http.MethodPost, "/read-all", "user-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestNotifications_MarkAllRead_DBError(t *testing.T) {
	mock := &mockNotificationRepo{markAllErr: errors.New("db error")}
	h := &NotificationsHandler{Notifications: mock}

	rr := httptest.NewRecorder()
	h.MarkAllRead(rr, authedRequest(http.MethodPost, "/read-all", "user-1"))

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}
