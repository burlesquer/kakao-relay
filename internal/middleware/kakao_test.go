package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/openclaw/relay-server-go/internal/util"
)

func TestKakaoSignatureMiddleware(t *testing.T) {
	secret := "test-secret"
	body := `{"key":"value"}`
	validSignature := util.HmacSHA256(secret, body)

	t.Run("passes through when secret is empty", func(t *testing.T) {
		middleware := NewKakaoSignatureMiddleware("")
		handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("POST", "/webhook", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("rejects request without signature header", func(t *testing.T) {
		middleware := NewKakaoSignatureMiddleware(secret)
		handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called")
		}))

		req := httptest.NewRequest("POST", "/webhook", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("rejects request with invalid signature", func(t *testing.T) {
		middleware := NewKakaoSignatureMiddleware(secret)
		handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called")
		}))

		req := httptest.NewRequest("POST", "/webhook", bytes.NewBufferString(body))
		req.Header.Set("X-Kakao-Signature", "invalid-signature")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("allows request with valid signature", func(t *testing.T) {
		middleware := NewKakaoSignatureMiddleware(secret)
		handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("POST", "/webhook", bytes.NewBufferString(body))
		req.Header.Set("X-Kakao-Signature", validSignature)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("stores parsed body in context", func(t *testing.T) {
		middleware := NewKakaoSignatureMiddleware(secret)
		handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			parsed := GetKakaoBody(r.Context())
			assert.NotNil(t, parsed)

			m, ok := parsed.(map[string]any)
			assert.True(t, ok)
			assert.Equal(t, "value", m["key"])
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("POST", "/webhook", bytes.NewBufferString(body))
		req.Header.Set("X-Kakao-Signature", validSignature)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {
		invalidBody := `{invalid json}`
		invalidSignature := util.HmacSHA256(secret, invalidBody)

		middleware := NewKakaoSignatureMiddleware(secret)
		handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called")
		}))

		req := httptest.NewRequest("POST", "/webhook", bytes.NewBufferString(invalidBody))
		req.Header.Set("X-Kakao-Signature", invalidSignature)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
