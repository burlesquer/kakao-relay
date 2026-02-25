package middleware

import (
	"net/http"

	"github.com/openclaw/relay-server-go/internal/util"
)

const (
	CSRFCookieName = "csrf_token"
	CSRFHeaderName = "X-CSRF-Token"
	CSRFTokenBytes = 32
)

// CSRFMiddleware provides CSRF protection for state-changing requests.
// It uses the double-submit cookie pattern:
// 1. A CSRF token is set in a cookie (readable by JavaScript)
// 2. The same token must be sent in the X-CSRF-Token header
// 3. For state-changing methods (POST, PUT, PATCH, DELETE), both must match
type CSRFMiddleware struct {
	isProduction bool
}

func NewCSRFMiddleware(isProduction bool) *CSRFMiddleware {
	return &CSRFMiddleware{isProduction: isProduction}
}

func (m *CSRFMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ensure CSRF cookie exists
		cookie, err := r.Cookie(CSRFCookieName)
		if err != nil || cookie.Value == "" {
			// Generate new CSRF token
			token, err := util.GenerateToken()
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{
					"error": "Failed to generate security token",
				})
				return
			}
			m.setCSRFCookie(w, token)
			cookie = &http.Cookie{Value: token}
		}

		// For safe methods (GET, HEAD, OPTIONS), just ensure cookie exists
		if isSafeMethod(r.Method) {
			next.ServeHTTP(w, r)
			return
		}

		// For state-changing methods, validate the token
		headerToken := r.Header.Get(CSRFHeaderName)
		if headerToken == "" {
			writeJSON(w, http.StatusForbidden, map[string]string{
				"error": "Missing CSRF token",
			})
			return
		}

		if !util.ConstantTimeEqual(cookie.Value, headerToken) {
			writeJSON(w, http.StatusForbidden, map[string]string{
				"error": "Invalid CSRF token",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *CSRFMiddleware) setCSRFCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CSRFCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(SessionMaxAge.Seconds()),
		HttpOnly: false, // Must be readable by JavaScript to send in header
		Secure:   m.isProduction,
		SameSite: http.SameSiteLaxMode,
	})
}

func isSafeMethod(method string) bool {
	return method == http.MethodGet ||
		method == http.MethodHead ||
		method == http.MethodOptions
}
