package service

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRedisClient(t *testing.T) *redis.Client {
	t.Helper()
	redisURL := "redis://localhost:6379/15"
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		t.Skip("Redis URL not parseable, skipping")
	}
	client := redis.NewClient(opts)
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		t.Skip("Redis not available for testing")
	}
	client.FlushDB(ctx)
	return client
}

func TestRateLimiter_Basic(t *testing.T) {
	redisClient := newTestRedisClient(t)
	defer redisClient.Close()

	ctx := context.Background()

	limiter := NewRateLimiter(redisClient)

	t.Run("allows requests within limit", func(t *testing.T) {
		key := "test:user1"
		limit := 3
		window := 10 * time.Second

		// First 3 requests should be allowed
		for i := 0; i < limit; i++ {
			allowed, _ := limiter.CheckLimit(ctx, key, limit, window)
			assert.True(t, allowed, "Request %d should be allowed", i+1)
		}

		// 4th request should be denied
		allowed, resetAt := limiter.CheckLimit(ctx, key, limit, window)
		assert.False(t, allowed, "Request should be rate limited")
		assert.True(t, resetAt.After(time.Now()), "Reset time should be in future")
	})

	t.Run("sliding window behavior", func(t *testing.T) {
		key := "test:user2"
		limit := 2
		window := 2 * time.Second

		// Use up limit
		allowed, _ := limiter.CheckLimit(ctx, key, limit, window)
		assert.True(t, allowed)
		allowed, _ = limiter.CheckLimit(ctx, key, limit, window)
		assert.True(t, allowed)

		// Should be limited now
		allowed, _ = limiter.CheckLimit(ctx, key, limit, window)
		assert.False(t, allowed)

		// Wait for window to pass
		time.Sleep(2100 * time.Millisecond)

		// Should be allowed again
		allowed, _ = limiter.CheckLimit(ctx, key, limit, window)
		assert.True(t, allowed)
	})

	t.Run("different keys are independent", func(t *testing.T) {
		limit := 1
		window := 10 * time.Second

		key1 := "test:independent1"
		key2 := "test:independent2"

		// Use up key1 limit
		allowed, _ := limiter.CheckLimit(ctx, key1, limit, window)
		assert.True(t, allowed)
		allowed, _ = limiter.CheckLimit(ctx, key1, limit, window)
		assert.False(t, allowed)

		// key2 should still be allowed
		allowed, _ = limiter.CheckLimit(ctx, key2, limit, window)
		assert.True(t, allowed)
	})
}

func TestRateLimiter_GracefulFailure(t *testing.T) {
	// Test with invalid Redis client (should deny requests for safety)
	invalidClient := redis.NewClient(&redis.Options{
		Addr: "localhost:9999", // Invalid port
	})
	defer invalidClient.Close()

	limiter := NewRateLimiter(invalidClient)
	ctx := context.Background()

	// Should deny request when Redis fails (fail-closed for safety)
	allowed, resetAt := limiter.CheckLimit(ctx, "test:key", 1, 1*time.Minute)
	require.False(t, allowed, "Should deny request on Redis failure for safety")
	require.True(t, resetAt.After(time.Now()), "Should return valid reset time")
}

