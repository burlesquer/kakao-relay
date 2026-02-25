package service

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// rateLimitScript is a Lua script for sliding window rate limiting
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
    return {0, resetAt}
end

redis.call('ZADD', key, now, now .. '-' .. math.random())
redis.call('EXPIRE', key, window + 10)

local resetAt = now + window
return {1, resetAt}
`)

// RateLimiter provides generic rate limiting functionality
type RateLimiter struct {
	client *redis.Client
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{client: client}
}

// CheckLimit checks if a request is allowed under the rate limit
func (rl *RateLimiter) CheckLimit(
	ctx context.Context,
	key string,
	limit int,
	window time.Duration,
) (allowed bool, resetAt time.Time) {
	now := time.Now().Unix()
	fullKey := fmt.Sprintf("ratelimit:%s", key)

	result, err := rateLimitScript.Run(
		ctx,
		rl.client,
		[]string{fullKey},
		now,
		int64(window.Seconds()),
		limit,
	).Int64Slice()

	if err != nil {
		log.Warn().
			Err(err).
			Str("key", key).
			Msg("rate limit check failed, denying request for safety")
		return false, time.Now().Add(window)
	}

	if len(result) != 2 {
		log.Warn().Str("key", key).Msg("unexpected rate limit result, denying request for safety")
		return false, time.Now().Add(window)
	}

	return result[0] == 1, time.Unix(result[1], 0)
}
