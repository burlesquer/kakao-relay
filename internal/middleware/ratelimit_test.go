package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/openclaw/relay-server-go/internal/model"
)

func TestRateLimiter(t *testing.T) {
	t.Run("allows requests under limit", func(t *testing.T) {
		limiter := NewRateLimiter()

		for i := 0; i < 5; i++ {
			allowed, remaining, _ := limiter.Check("account-1", 10)
			assert.True(t, allowed)
			assert.Equal(t, 10-i-1, remaining)
		}
	})

	t.Run("blocks requests over limit", func(t *testing.T) {
		limiter := NewRateLimiter()

		for i := 0; i < 5; i++ {
			limiter.Check("account-2", 5)
		}

		allowed, remaining, _ := limiter.Check("account-2", 5)
		assert.False(t, allowed)
		assert.Equal(t, 0, remaining)
	})

	t.Run("tracks accounts separately", func(t *testing.T) {
		limiter := NewRateLimiter()

		for i := 0; i < 5; i++ {
			limiter.Check("account-a", 5)
		}

		allowed, _, _ := limiter.Check("account-b", 5)
		assert.True(t, allowed)
	})

	t.Run("returns reset time", func(t *testing.T) {
		limiter := NewRateLimiter()

		_, _, resetAt := limiter.Check("account-3", 10)
		assert.Greater(t, resetAt, int64(0))
	})
}

func TestRateLimitMiddleware(t *testing.T) {
	t.Run("allows request without account", func(t *testing.T) {
		middleware := NewRateLimitMiddleware()
		handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("sets rate limit headers", func(t *testing.T) {
		middleware := NewRateLimitMiddleware()
		handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		account := &model.Account{ID: "acc-1", RateLimitPerMin: 100}
		ctx := context.WithValue(context.Background(), AccountContextKey, account)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "100", rec.Header().Get("X-RateLimit-Limit"))
		assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Remaining"))
		assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Reset"))
	})

	t.Run("returns 429 when rate limited", func(t *testing.T) {
		middleware := NewRateLimitMiddleware()
		handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		account := &model.Account{ID: "acc-2", RateLimitPerMin: 2}
		ctx := context.WithValue(context.Background(), AccountContextKey, account)

		for i := 0; i < 2; i++ {
			req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
		}

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusTooManyRequests, rec.Code)
		assert.Equal(t, "60", rec.Header().Get("Retry-After"))
	})

	t.Run("uses default limit when account limit is zero", func(t *testing.T) {
		middleware := NewRateLimitMiddleware()
		handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		account := &model.Account{ID: "acc-3", RateLimitPerMin: 0}
		ctx := context.WithValue(context.Background(), AccountContextKey, account)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "60", rec.Header().Get("X-RateLimit-Limit"))
	})
}
