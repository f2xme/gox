package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/f2xme/gox/httpx"
	ginadapter "github.com/f2xme/gox/httpx/adapter/gin"
	"github.com/gin-gonic/gin"
)

func TestRateLimit_TokenBucket_Success(t *testing.T) {
	engine := ginadapter.New()
	engine.Use(New(
		WithStrategy(StrategyTokenBucket),
		WithRate(10),
		WithBurst(10),
	))

	engine.GET("/test", func(c httpx.Context) error {
		return c.JSON(200, map[string]any{"status": "ok"})
	})

	// First 10 requests should succeed
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		engine.Raw().(*gin.Engine).ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("request %d: expected status 200, got %d", i, w.Code)
		}
	}
}

func TestRateLimit_TokenBucket_Exceeded(t *testing.T) {
	engine := ginadapter.New()
	engine.Use(New(
		WithStrategy(StrategyTokenBucket),
		WithRate(5),
		WithBurst(5),
	))

	engine.GET("/test", func(c httpx.Context) error {
		return c.JSON(200, map[string]any{"status": "ok"})
	})

	// First 5 requests should succeed
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		engine.Raw().(*gin.Engine).ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("request %d: expected status 200, got %d", i, w.Code)
		}
	}

	// 6th request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	engine.Raw().(*gin.Engine).ServeHTTP(w, req)

	if w.Code != 429 {
		t.Errorf("expected status 429, got %d", w.Code)
	}
}

func TestRateLimit_TokenBucket_Refill(t *testing.T) {
	engine := ginadapter.New()
	engine.Use(New(
		WithStrategy(StrategyTokenBucket),
		WithRate(10), // 10 tokens per second
		WithBurst(5),
	))

	engine.GET("/test", func(c httpx.Context) error {
		return c.JSON(200, map[string]any{"status": "ok"})
	})

	// Exhaust tokens
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		engine.Raw().(*gin.Engine).ServeHTTP(w, req)
	}

	// Wait for refill (0.5 seconds = 5 tokens)
	time.Sleep(500 * time.Millisecond)

	// Should succeed after refill
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	engine.Raw().(*gin.Engine).ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200 after refill, got %d", w.Code)
	}
}

func TestRateLimit_FixedWindow(t *testing.T) {
	engine := ginadapter.New()
	engine.Use(New(
		WithStrategy(StrategyFixedWindow),
		WithRate(5),
		WithWindow(time.Second),
	))

	engine.GET("/test", func(c httpx.Context) error {
		return c.JSON(200, map[string]any{"status": "ok"})
	})

	// First 5 requests should succeed
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		engine.Raw().(*gin.Engine).ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("request %d: expected status 200, got %d", i, w.Code)
		}
	}

	// 6th request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	engine.Raw().(*gin.Engine).ServeHTTP(w, req)

	if w.Code != 429 {
		t.Errorf("expected status 429, got %d", w.Code)
	}

	// Wait for window to reset
	time.Sleep(1100 * time.Millisecond)

	// Should succeed in new window
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	w = httptest.NewRecorder()
	engine.Raw().(*gin.Engine).ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200 in new window, got %d", w.Code)
	}
}

func TestRateLimit_SlidingWindow(t *testing.T) {
	engine := ginadapter.New()
	engine.Use(New(
		WithStrategy(StrategySlidingWindow),
		WithRate(3),
		WithWindow(time.Second),
	))

	engine.GET("/test", func(c httpx.Context) error {
		return c.JSON(200, map[string]any{"status": "ok"})
	})

	// First 3 requests should succeed
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		engine.Raw().(*gin.Engine).ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("request %d: expected status 200, got %d", i, w.Code)
		}
	}

	// 4th request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	engine.Raw().(*gin.Engine).ServeHTTP(w, req)

	if w.Code != 429 {
		t.Errorf("expected status 429, got %d", w.Code)
	}

	// Wait for oldest request to expire
	time.Sleep(1100 * time.Millisecond)

	// Should succeed after window slides
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	w = httptest.NewRecorder()
	engine.Raw().(*gin.Engine).ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200 after window slide, got %d", w.Code)
	}
}

func TestRateLimit_CustomHandler(t *testing.T) {
	customHandlerCalled := false

	engine := ginadapter.New()
	engine.Use(New(
		WithRate(1),
		WithBurst(1),
		WithHandler(func(c httpx.Context) {
			customHandlerCalled = true
			c.JSON(503, map[string]any{"error": "custom rate limit"})
		}),
	))

	engine.GET("/test", func(c httpx.Context) error {
		return c.JSON(200, map[string]any{"status": "ok"})
	})

	// First request succeeds
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	engine.Raw().(*gin.Engine).ServeHTTP(w, req)

	// Second request triggers custom handler
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	w = httptest.NewRecorder()
	engine.Raw().(*gin.Engine).ServeHTTP(w, req)

	if !customHandlerCalled {
		t.Error("expected custom handler to be called")
	}

	if w.Code != 503 {
		t.Errorf("expected status 503, got %d", w.Code)
	}
}

func TestRateLimit_ByHeader(t *testing.T) {
	engine := ginadapter.New()
	engine.Use(New(
		WithRate(2),
		WithBurst(2),
		WithKeyFunc(ByHeader("X-API-Key")),
	))

	engine.GET("/test", func(c httpx.Context) error {
		return c.JSON(200, map[string]any{"status": "ok"})
	})

	// User 1: first 2 requests succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-API-Key", "user1")
		w := httptest.NewRecorder()
		engine.Raw().(*gin.Engine).ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("user1 request %d: expected status 200, got %d", i, w.Code)
		}
	}

	// User 1: 3rd request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "user1")
	w := httptest.NewRecorder()
	engine.Raw().(*gin.Engine).ServeHTTP(w, req)

	if w.Code != 429 {
		t.Errorf("user1: expected status 429, got %d", w.Code)
	}

	// User 2: should still have quota
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "user2")
	w = httptest.NewRecorder()
	engine.Raw().(*gin.Engine).ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("user2: expected status 200, got %d", w.Code)
	}
}

func TestRateLimit_ByIPAndPath(t *testing.T) {
	engine := ginadapter.New()
	engine.Use(New(
		WithRate(1),
		WithBurst(1),
		WithKeyFunc(ByIPAndPath),
	))

	engine.GET("/api/users", func(c httpx.Context) error {
		return c.JSON(200, map[string]any{"status": "ok"})
	})

	engine.GET("/api/posts", func(c httpx.Context) error {
		return c.JSON(200, map[string]any{"status": "ok"})
	})

	// First request to /api/users succeeds
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()
	engine.Raw().(*gin.Engine).ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Second request to /api/users should be rate limited
	req = httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w = httptest.NewRecorder()
	engine.Raw().(*gin.Engine).ServeHTTP(w, req)

	if w.Code != 429 {
		t.Errorf("expected status 429, got %d", w.Code)
	}

	// Request to /api/posts should succeed (different path)
	req = httptest.NewRequest(http.MethodGet, "/api/posts", nil)
	w = httptest.NewRecorder()
	engine.Raw().(*gin.Engine).ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200 for different path, got %d", w.Code)
	}
}
