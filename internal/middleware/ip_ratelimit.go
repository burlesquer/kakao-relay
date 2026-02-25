package middleware

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/openclaw/relay-server-go/internal/service"
)

type IPRateLimitMiddleware struct {
	limiter *service.RateLimiter
	limit   int
	window  time.Duration
	prefix  string
}

func NewIPRateLimitMiddleware(limiter *service.RateLimiter, limit int, window time.Duration, prefix string) *IPRateLimitMiddleware {
	return &IPRateLimitMiddleware{
		limiter: limiter,
		limit:   limit,
		window:  window,
		prefix:  prefix,
	}
}

func (m *IPRateLimitMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if host, _, err := net.SplitHostPort(ip); err == nil {
			ip = host
		}

		key := fmt.Sprintf("ip:%s:%s", m.prefix, ip)
		allowed, resetAt := m.limiter.CheckLimit(r.Context(), key, m.limit, m.window)

		if !allowed {
			secondsLeft := int(time.Until(resetAt).Seconds()) + 1
			w.Header().Set("Retry-After", fmt.Sprintf("%d", secondsLeft))
			writeJSON(w, http.StatusTooManyRequests, map[string]string{
				"error": "Too many requests. Please try again later.",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
