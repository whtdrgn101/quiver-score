package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/quiverscore/backend-go/internal/repository"
)

type mockUserRepo struct {
	getMeResult         *repository.UserOut
	getMeErr            error
	updateProfileResult *repository.UserOut
	updateProfileErr    error
	updateAvatarResult  *repository.UserOut
	updateAvatarErr     error
	deleteAvatarResult  *repository.UserOut
	deleteAvatarErr     error
	publicProfileResult *repository.PublicProfileOut
	publicProfileErr    error
}

func (m *mockUserRepo) GetMe(_ context.Context, _ string) (*repository.UserOut, error) {
	return m.getMeResult, m.getMeErr
}

func (m *mockUserRepo) UpdateProfile(_ context.Context, _ string,
	_, _, _, _ *string, _, _, _, _ bool, _ *bool,
) (*repository.UserOut, error) {
	return m.updateProfileResult, m.updateProfileErr
}

func (m *mockUserRepo) UpdateAvatar(_ context.Context, _, _ string) (*repository.UserOut, error) {
	return m.updateAvatarResult, m.updateAvatarErr
}

func (m *mockUserRepo) DeleteAvatar(_ context.Context, _ string) (*repository.UserOut, error) {
	return m.deleteAvatarResult, m.deleteAvatarErr
}

func (m *mockUserRepo) GetPublicProfile(_ context.Context, _ string) (*repository.PublicProfileOut, error) {
	return m.publicProfileResult, m.publicProfileErr
}

func sampleUser() *repository.UserOut {
	dn := "Test Archer"
	return &repository.UserOut{
		ID:            "user-1",
		Email:         "test@example.com",
		Username:      "testuser",
		DisplayName:   &dn,
		EmailVerified: true,
		ProfilePublic: false,
		CreatedAt:     time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

// ── GetMe ────────────────────────────────────────────────────────────

func TestGetMe_Success(t *testing.T) {
	mock := &mockUserRepo{getMeResult: sampleUser()}
	h := &UsersHandler{Users: mock}

	rr := httptest.NewRecorder()
	h.GetMe(rr, authedRequest(http.MethodGet, "/users/me", "user-1"))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var u repository.UserOut
	if err := json.NewDecoder(rr.Body).Decode(&u); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if u.ID != "user-1" {
		t.Errorf("expected id 'user-1', got '%s'", u.ID)
	}
	if u.Username != "testuser" {
		t.Errorf("expected username 'testuser', got '%s'", u.Username)
	}
}

func TestGetMe_NotFound(t *testing.T) {
	mock := &mockUserRepo{getMeErr: errors.New("no rows")}
	h := &UsersHandler{Users: mock}

	rr := httptest.NewRecorder()
	h.GetMe(rr, authedRequest(http.MethodGet, "/users/me", "user-1"))

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

// ── UpdateMe ─────────────────────────────────────────────────────────

func TestUpdateMe_Success(t *testing.T) {
	updated := sampleUser()
	dn := "New Name"
	updated.DisplayName = &dn
	mock := &mockUserRepo{updateProfileResult: updated}
	h := &UsersHandler{Users: mock}

	body := strings.NewReader(`{"display_name":"New Name"}`)
	req := authedRequest(http.MethodPut, "/users/me", "user-1")
	req.Body = io.NopCloser(body)

	rr := httptest.NewRecorder()
	h.UpdateMe(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var u repository.UserOut
	if err := json.NewDecoder(rr.Body).Decode(&u); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if *u.DisplayName != "New Name" {
		t.Errorf("expected display_name 'New Name', got '%s'", *u.DisplayName)
	}
}

func TestUpdateMe_InvalidBody(t *testing.T) {
	mock := &mockUserRepo{}
	h := &UsersHandler{Users: mock}

	body := strings.NewReader(`not json`)
	req := authedRequest(http.MethodPut, "/users/me", "user-1")
	req.Body = io.NopCloser(body)

	rr := httptest.NewRecorder()
	h.UpdateMe(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rr.Code)
	}
}

// ── DeleteAvatar ─────────────────────────────────────────────────────

func TestDeleteAvatar_Success(t *testing.T) {
	mock := &mockUserRepo{deleteAvatarResult: sampleUser()}
	h := &UsersHandler{Users: mock}

	rr := httptest.NewRecorder()
	h.DeleteAvatar(rr, authedRequest(http.MethodDelete, "/users/me/avatar", "user-1"))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var u repository.UserOut
	if err := json.NewDecoder(rr.Body).Decode(&u); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if u.ID != "user-1" {
		t.Errorf("expected id 'user-1', got '%s'", u.ID)
	}
}

func TestDeleteAvatar_Error(t *testing.T) {
	mock := &mockUserRepo{deleteAvatarErr: errors.New("db error")}
	h := &UsersHandler{Users: mock}

	rr := httptest.NewRecorder()
	h.DeleteAvatar(rr, authedRequest(http.MethodDelete, "/users/me/avatar", "user-1"))

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}

// ── GetPublicProfile ─────────────────────────────────────────────────

func publicProfileRequest(method, path, username string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("username", username)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestGetPublicProfile_Success(t *testing.T) {
	dn := "Public Archer"
	mock := &mockUserRepo{
		publicProfileResult: &repository.PublicProfileOut{
			ID:             "user-2",
			Username:       "publicuser",
			DisplayName:    &dn,
			CreatedAt:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			RecentSessions: []repository.PublicSessionSummary{},
			Clubs:          []repository.ProfileClubOut{},
		},
	}
	h := &UsersHandler{Users: mock}

	rr := httptest.NewRecorder()
	h.GetPublicProfile(rr, publicProfileRequest(http.MethodGet, "/users/publicuser", "publicuser"))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var p repository.PublicProfileOut
	if err := json.NewDecoder(rr.Body).Decode(&p); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if p.Username != "publicuser" {
		t.Errorf("expected username 'publicuser', got '%s'", p.Username)
	}
}

func TestGetPublicProfile_NotFound(t *testing.T) {
	mock := &mockUserRepo{publicProfileErr: errors.New("no rows")}
	h := &UsersHandler{Users: mock}

	rr := httptest.NewRecorder()
	h.GetPublicProfile(rr, publicProfileRequest(http.MethodGet, "/users/unknown", "unknown"))

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
