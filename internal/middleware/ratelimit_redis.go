package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/openclaw/relay-server-go/internal/config"
)

const (
	rateLimitKeyPrefix = "ratelimit:"
	rateLimitWindow    = 60 * time.Second
)

var rateLimitScript = redis.NewScript(`
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])

local windowStart = now - window

redis.call('ZREMRANGEBYSCORE', key, '-inf', windowStart)

local count = redis.call('ZCARD', key)

if count >= limit then
    local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
    local resetAt = 0
    if #oldest >= 2 then
        resetAt = tonumber(oldest[2]) + window
    else
        resetAt = now + window
    end
    return {0, 0, resetAt}
end

redis.call('ZADD', key, now, now .. '-' .. math.random())
redis.call('EXPIRE', key, window + 10)

local remaining = limit - count - 1
local resetAt = now + window

return {1, remaining, resetAt}
`)

type RedisRateLimiter struct {
	client *redis.Client
}

func NewRedisRateLimiter(client *redis.Client) *RedisRateLimiter {
	return &RedisRateLimiter{client: client}
}

func (rl *RedisRateLimiter) Check(ctx context.Context, accountID string, limit int) (allowed bool, remaining int, resetAt int64) {
	now := time.Now().Unix()
	key := rateLimitKeyPrefix + accountID

	result, err := rateLimitScript.Run(ctx, rl.client, []string{key}, now, int64(rateLimitWindow.Seconds()), limit).Int64Slice()
	if err != nil {
		log.Warn().Err(err).Str("accountId", accountID).Msg("redis rate limit check failed, denying request for safety")
		return false, 0, now + int64(rateLimitWindow.Seconds())
	}

	if len(result) != 3 {
		log.Warn().Str("accountId", accountID).Msg("unexpected redis rate limit result, denying request for safety")
		return false, 0, now + int64(rateLimitWindow.Seconds())
	}

	return result[0] == 1, int(result[1]), result[2]
}

type RedisRateLimitMiddleware struct {
	limiter *RedisRateLimiter
}

func NewRedisRateLimitMiddleware(redisClient *redis.Client) *RedisRateLimitMiddleware {
	return &RedisRateLimitMiddleware{
		limiter: NewRedisRateLimiter(redisClient),
	}
}

func (m *RedisRateLimitMiddleware) Handler(next http.Handler) http.Handler {
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

		allowed, remaining, resetAt := m.limiter.Check(r.Context(), account.ID, limit)

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
