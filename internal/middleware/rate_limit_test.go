package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"concert-go/internal/middleware"
	"concert-go/internal/util"

	"github.com/gin-gonic/gin"
)

type mockRateLimitStore struct {
	count int
}

func (m *mockRateLimitStore) Hit(ctx context.Context, key string, window time.Duration) (util.RateLimitState, error) {
	m.count++
	return util.RateLimitState{
		Count:      m.count,
		RetryAfter: window,
	}, nil
}

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &mockRateLimitStore{}

	r := gin.New()
	r.Use(middleware.RateLimit(store, "test", 2, time.Minute))
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Request 1: allowed (count = 1)
	req1 := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200 on request 1, got %d", w1.Code)
	}

	// Request 2: allowed (count = 2)
	req2 := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 on request 2, got %d", w2.Code)
	}

	// Request 3: blocked (count = 3 > limit 2)
	req3 := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 on request 3, got %d", w3.Code)
	}
}
