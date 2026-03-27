package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterAllowsUnderLimit(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(5, 15*time.Minute)

	for i := 0; i < 5; i++ {
		if !rl.Allow("192.168.1.1") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
}

func TestRateLimiterBlocksOverLimit(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(5, 15*time.Minute)

	// Use up all 5 allowed requests
	for i := 0; i < 5; i++ {
		rl.Allow("192.168.1.1")
	}

	// 6th request should be blocked
	if rl.Allow("192.168.1.1") {
		t.Fatal("6th request should be blocked")
	}
}

func TestRateLimiterResetsAfterWindow(t *testing.T) {
	t.Parallel()
	// Use a very short window for testing
	rl := NewRateLimiter(2, 50*time.Millisecond)

	// Use up limit
	rl.Allow("10.0.0.1")
	rl.Allow("10.0.0.1")
	if rl.Allow("10.0.0.1") {
		t.Fatal("3rd request should be blocked")
	}

	// Wait for window to expire
	time.Sleep(60 * time.Millisecond)

	// Should be allowed again
	if !rl.Allow("10.0.0.1") {
		t.Fatal("request after window reset should be allowed")
	}
}

func TestRateLimiterTracksIPsSeparately(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(2, 15*time.Minute)

	// IP1 uses its limit
	rl.Allow("192.168.1.1")
	rl.Allow("192.168.1.1")
	if rl.Allow("192.168.1.1") {
		t.Fatal("IP1 should be blocked")
	}

	// IP2 should still be allowed
	if !rl.Allow("192.168.1.2") {
		t.Fatal("IP2 should be allowed")
	}
}

func TestRateLimiterHTTPMiddleware5RequestsOK(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(5, 15*time.Minute)
	handler := rl.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, w.Code)
		}
	}
}

func TestRateLimiterHTTPMiddleware6thRequestBlocked(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(5, 15*time.Minute)
	handler := rl.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// 5 allowed requests
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	// 6th request → 429
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}
}

func TestRateLimiterHTTPMiddlewareResetsAfterWindow(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(2, 50*time.Millisecond)
	handler := rl.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Use up limit
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	// 3rd request blocked
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}

	// Wait for window to expire
	time.Sleep(60 * time.Millisecond)

	// Should be allowed again
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 after window reset, got %d", w.Code)
	}
}

func TestRateLimiterHTTPMiddlewareDifferentPortsSameIP(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(2, 15*time.Minute)
	handler := rl.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Two requests from same IP but different ports share the same counter
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	req.RemoteAddr = "10.0.0.1:11111"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	req.RemoteAddr = "10.0.0.1:22222"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// 3rd request from yet another port should be blocked (limit=2)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	req.RemoteAddr = "10.0.0.1:33333"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}
}

func TestRateLimiterHTTPMiddlewareIPv6(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(1, 15*time.Minute)
	handler := rl.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	req.RemoteAddr = "[::1]:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// 2nd request from same IPv6 should be blocked (limit=1)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	req.RemoteAddr = "[::1]:54321"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}
}

func TestRateLimiterReset(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(1, 15*time.Minute)

	rl.Allow("10.0.0.1")
	if rl.Allow("10.0.0.1") {
		t.Fatal("should be blocked")
	}

	rl.Reset("10.0.0.1")

	if !rl.Allow("10.0.0.1") {
		t.Fatal("should be allowed after reset")
	}
}
