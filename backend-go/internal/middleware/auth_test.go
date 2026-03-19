package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/quiverscore/backend-go/internal/auth"
)

const testSecret = "test-secret-key-for-middleware"

func TestRequireAuth_ValidToken(t *testing.T) {
	token, err := auth.CreateAccessToken("user-123", 30, testSecret)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	var capturedUserID string
	handler := RequireAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if capturedUserID != "user-123" {
		t.Errorf("expected user ID 'user-123', got '%s'", capturedUserID)
	}
}

func TestRequireAuth_MissingHeader(t *testing.T) {
	handler := RequireAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	handler := RequireAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestRequireAuth_WrongSecret(t *testing.T) {
	token, err := auth.CreateAccessToken("user-123", 30, "other-secret")
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	handler := RequireAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestRequireAuth_RefreshTokenRejected(t *testing.T) {
	token, err := auth.CreateRefreshToken("user-123", 7, testSecret)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	handler := RequireAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with refresh token")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestRequireAuth_MalformedHeader(t *testing.T) {
	handler := RequireAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "NotBearer some-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestOptionalAuth_ValidToken(t *testing.T) {
	token, err := auth.CreateAccessToken("user-456", 30, testSecret)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	var capturedUserID string
	handler := OptionalAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if capturedUserID != "user-456" {
		t.Errorf("expected user ID 'user-456', got '%s'", capturedUserID)
	}
}

func TestOptionalAuth_NoToken(t *testing.T) {
	var capturedUserID string
	handler := OptionalAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if capturedUserID != "" {
		t.Errorf("expected empty user ID, got '%s'", capturedUserID)
	}
}

func TestOptionalAuth_InvalidToken(t *testing.T) {
	var capturedUserID string
	handler := OptionalAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 even with invalid token, got %d", rr.Code)
	}
	if capturedUserID != "" {
		t.Errorf("expected empty user ID, got '%s'", capturedUserID)
	}
}

func TestGetUserID_NoValue(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	userID := GetUserID(req.Context())
	if userID != "" {
		t.Errorf("expected empty string, got '%s'", userID)
	}
}
