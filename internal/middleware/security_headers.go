package middleware

import (
	"net/http"
)

type SecurityHeadersMiddleware struct {
	isProduction bool
}

func NewSecurityHeadersMiddleware(isProduction bool) *SecurityHeadersMiddleware {
	return &SecurityHeadersMiddleware{isProduction: isProduction}
}

func (m *SecurityHeadersMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		if m.isProduction {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		csp := "default-src 'self'; " +
			"script-src 'self'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self'; " +
			"connect-src 'self'; " +
			"frame-ancestors 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'"

		w.Header().Set("Content-Security-Policy", csp)

		next.ServeHTTP(w, r)
	})
}
