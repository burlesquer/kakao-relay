package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/openclaw/relay-server-go/internal/config"
)

const (
	maxEntries      = 10000
	cleanupInterval = time.Minute
	entryTTL        = 5 * time.Minute
	windowDuration  = time.Minute
)

type rateLimitEntry struct {
	timestamps []time.Time
	lastAccess time.Time
}

type RateLimiter struct {
	mu          sync.RWMutex
	store       map[string]*rateLimitEntry
	lastCleanup time.Time
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		store:       make(map[string]*rateLimitEntry),
		lastCleanup: time.Now(),
	}
}

func (rl *RateLimiter) cleanup() {
	now := time.Now()
	if now.Sub(rl.lastCleanup) < cleanupInterval {
		return
	}
	rl.lastCleanup = now

	for key, entry := range rl.store {
		if now.Sub(entry.lastAccess) > entryTTL {
			delete(rl.store, key)
		}
	}

	if len(rl.store) > maxEntries {
		oldest := make([]string, 0, len(rl.store)/5)
		for key := range rl.store {
			oldest = append(oldest, key)
			if len(oldest) >= len(rl.store)/5 {
				break
			}
		}
		for _, key := range oldest {
			delete(rl.store, key)
		}
	}
}

func (rl *RateLimiter) Check(accountID string, limit int) (allowed bool, remaining int, resetAt int64) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.cleanup()

	now := time.Now()
	windowStart := now.Add(-windowDuration)

	entry, exists := rl.store[accountID]
	if !exists {
		entry = &rateLimitEntry{
			timestamps: make([]time.Time, 0),
			lastAccess: now,
		}
		rl.store[accountID] = entry
	}

	entry.lastAccess = now

	filtered := entry.timestamps[:0]
	for _, ts := range entry.timestamps {
		if ts.After(windowStart) {
			filtered = append(filtered, ts)
		}
	}
	entry.timestamps = filtered

	remaining = limit - len(entry.timestamps)
	if remaining < 0 {
		remaining = 0
	}

	if len(entry.timestamps) > 0 {
		resetAt = entry.timestamps[0].Add(windowDuration).Unix()
	} else {
		resetAt = now.Add(windowDuration).Unix()
	}

	if len(entry.timestamps) >= limit {
		return false, 0, resetAt
	}

	entry.timestamps = append(entry.timestamps, now)
	return true, remaining - 1, resetAt
}

type RateLimitMiddleware struct {
	limiter *RateLimiter
}

func NewRateLimitMiddleware() *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter: NewRateLimiter(),
	}
}

func (m *RateLimitMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		account := GetAccount(r.Context())
		if account == nil {
			next.ServeHTTP(w, r)
			return
		}

		limit := account.RateLimitPerMin
		if limit <= 0 {
			limit = config.DefaultRateLimitPerMin
		}

		allowed, remaining, resetAt := m.limiter.Check(account.ID, limit)

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))

		if !allowed {
			log.Warn().Str("accountId", account.ID).Msg("rate limit exceeded")
			w.Header().Set("Retry-After", "60")
			writeJSON(w, http.StatusTooManyRequests, map[string]string{
				"error": "Rate limit exceeded",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
