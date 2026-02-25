package middleware

import (
	"net/http"
)

const (
	DefaultMaxBodySize = 1 << 20 // 1MB
)

type BodyLimitMiddleware struct {
	maxSize int64
}

func NewBodyLimitMiddleware(maxSize int64) *BodyLimitMiddleware {
	if maxSize <= 0 {
		maxSize = DefaultMaxBodySize
	}
	return &BodyLimitMiddleware{maxSize: maxSize}
}

func (m *BodyLimitMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil && r.ContentLength > m.maxSize {
			writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{
				"error": "Request body too large",
			})
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, m.maxSize)
		next.ServeHTTP(w, r)
	})
}
