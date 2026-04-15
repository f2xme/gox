package timeout

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/f2xme/gox/httpx"
	ginadapter "github.com/f2xme/gox/httpx/adapter/gin"
	"github.com/gin-gonic/gin"
)

func TestTimeout_Success(t *testing.T) {
	engine := ginadapter.New()
	engine.Use(New(WithTimeout(100 * time.Millisecond)))

	engine.GET("/fast", func(c httpx.Context) error {
		time.Sleep(10 * time.Millisecond)
		return c.JSON(200, map[string]any{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/fast", nil)
	w := httptest.NewRecorder()

	engine.Raw().(*gin.Engine).ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestTimeout_Exceeded(t *testing.T) {
	engine := ginadapter.New()
	engine.Use(New(WithTimeout(50 * time.Millisecond)))

	engine.GET("/slow", func(c httpx.Context) error {
		time.Sleep(200 * time.Millisecond)
		return c.JSON(200, map[string]any{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	w := httptest.NewRecorder()

	engine.Raw().(*gin.Engine).ServeHTTP(w, req)

	if w.Code != 503 {
		t.Errorf("expected status 503, got %d", w.Code)
	}
}

func TestTimeout_CustomHandler(t *testing.T) {
	customHandlerCalled := false

	engine := ginadapter.New()
	engine.Use(New(
		WithTimeout(50*time.Millisecond),
		WithHandler(func(c httpx.Context) {
			customHandlerCalled = true
			c.JSON(408, map[string]any{"error": "custom timeout"})
		}),
	))

	engine.GET("/slow", func(c httpx.Context) error {
		time.Sleep(200 * time.Millisecond)
		return c.JSON(200, map[string]any{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	w := httptest.NewRecorder()

	engine.Raw().(*gin.Engine).ServeHTTP(w, req)

	if !customHandlerCalled {
		t.Error("expected custom handler to be called")
	}

	if w.Code != 408 {
		t.Errorf("expected status 408, got %d", w.Code)
	}
}

func TestTimeout_ContextCancellation(t *testing.T) {
	engine := ginadapter.New()
	engine.Use(New(WithTimeout(50 * time.Millisecond)))

	contextCancelled := false

	engine.GET("/check-context", func(c httpx.Context) error {
		time.Sleep(100 * time.Millisecond)

		// Check if context was cancelled
		if c.Request().Context().Err() == context.DeadlineExceeded {
			contextCancelled = true
		}

		return c.JSON(200, map[string]any{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/check-context", nil)
	w := httptest.NewRecorder()

	engine.Raw().(*gin.Engine).ServeHTTP(w, req)

	// Give goroutine time to complete
	time.Sleep(150 * time.Millisecond)

	if !contextCancelled {
		t.Error("expected context to be cancelled")
	}
}

func TestTimeout_HandlerError(t *testing.T) {
	engine := ginadapter.New()
	engine.Use(New(WithTimeout(100 * time.Millisecond)))

	expectedErr := errors.New("handler error")

	engine.GET("/error", func(c httpx.Context) error {
		return expectedErr
	})

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	w := httptest.NewRecorder()

	engine.Raw().(*gin.Engine).ServeHTTP(w, req)

	// Error should be propagated normally (not timeout)
	if w.Code == 503 {
		t.Error("should not return timeout status for handler errors")
	}
}
