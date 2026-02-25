package middleware

import (
	"net/http"
	"sync"
	"time"
)

const (
	loginMaxAttempts    = 5
	loginWindowDuration = time.Minute
	loginCleanupPeriod  = 5 * time.Minute
)

type loginAttempt struct {
	count       int
	windowStart time.Time
}

type LoginRateLimiter struct {
	mu          sync.RWMutex
	attempts    map[string]*loginAttempt
	lastCleanup time.Time
}

func NewLoginRateLimiter() *LoginRateLimiter {
	return &LoginRateLimiter{
		attempts:    make(map[string]*loginAttempt),
		lastCleanup: time.Now(),
	}
}

func (l *LoginRateLimiter) cleanup() {
	now := time.Now()
	if now.Sub(l.lastCleanup) < loginCleanupPeriod {
		return
	}
	l.lastCleanup = now

	for ip, attempt := range l.attempts {
		if now.Sub(attempt.windowStart) > loginWindowDuration {
			delete(l.attempts, ip)
		}
	}
}

func (l *LoginRateLimiter) isAllowed(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.cleanup()

	now := time.Now()
	attempt, exists := l.attempts[ip]

	if !exists {
		l.attempts[ip] = &loginAttempt{
			count:       1,
			windowStart: now,
		}
		return true
	}

	if now.Sub(attempt.windowStart) > loginWindowDuration {
		attempt.count = 1
		attempt.windowStart = now
		return true
	}

	if attempt.count >= loginMaxAttempts {
		return false
	}

	attempt.count++
	return true
}

func (l *LoginRateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = forwarded
		}

		if !l.isAllowed(ip) {
			w.Header().Set("Retry-After", "60")
			writeJSON(w, http.StatusTooManyRequests, map[string]string{
				"error": "Too many login attempts. Please try again later.",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
