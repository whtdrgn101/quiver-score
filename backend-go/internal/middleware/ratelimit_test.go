package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimiter_AllowsInitialBurst(t *testing.T) {
	rl := NewRateLimiter(10, 5)

	for i := 0; i < 5; i++ {
		if !rl.Allow("192.168.1.1") {
			t.Errorf("request %d should be allowed within burst", i+1)
		}
	}
}

func TestRateLimiter_BlocksAfterBurst(t *testing.T) {
	rl := NewRateLimiter(10, 3)

	for i := 0; i < 3; i++ {
		rl.Allow("192.168.1.1")
	}

	if rl.Allow("192.168.1.1") {
		t.Error("request after burst should be blocked")
	}
}

func TestRateLimiter_SeparateIPs(t *testing.T) {
	rl := NewRateLimiter(10, 2)

	rl.Allow("192.168.1.1")
	rl.Allow("192.168.1.1")

	if rl.Allow("192.168.1.1") {
		t.Error("third request from same IP should be blocked")
	}
	if !rl.Allow("192.168.1.2") {
		t.Error("first request from different IP should be allowed")
	}
}

func TestRateLimit_Middleware_Returns429(t *testing.T) {
	rl := NewRateLimiter(10, 1)

	handler := RateLimit(rl)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request allowed
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("first request: expected 200, got %d", rr.Code)
	}

	// Second request blocked
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("second request: expected 429, got %d", rr.Code)
	}
	if got := rr.Header().Get("Retry-After"); got != "60" {
		t.Errorf("expected Retry-After '60', got '%s'", got)
	}
}

func TestClientIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.50, 70.41.3.18")

	ip := clientIP(req)
	if ip != "203.0.113.50" {
		t.Errorf("expected '203.0.113.50', got '%s'", ip)
	}
}

func TestClientIP_XForwardedFor_Single(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.50")

	ip := clientIP(req)
	if ip != "203.0.113.50" {
		t.Errorf("expected '203.0.113.50', got '%s'", ip)
	}
}

func TestClientIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:54321"

	ip := clientIP(req)
	if ip != "192.168.1.1" {
		t.Errorf("expected '192.168.1.1', got '%s'", ip)
	}
}

func TestClientIP_RemoteAddrNoPort(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1"
	req.Header.Del("X-Forwarded-For")

	ip := clientIP(req)
	if ip != "192.168.1.1" {
		t.Errorf("expected '192.168.1.1', got '%s'", ip)
	}
}
