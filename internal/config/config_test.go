package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigMethods(t *testing.T) {
	t.Run("Addr returns formatted port", func(t *testing.T) {
		cfg := &Config{Port: 3000}
		assert.Equal(t, ":3000", cfg.Addr())
	})

	t.Run("QueueTTL converts seconds to duration", func(t *testing.T) {
		cfg := &Config{QueueTTLSeconds: 900}
		assert.Equal(t, 900*time.Second, cfg.QueueTTL())
	})

	t.Run("CallbackTTL converts seconds to duration", func(t *testing.T) {
		cfg := &Config{CallbackTTLSeconds: 55}
		assert.Equal(t, 55*time.Second, cfg.CallbackTTL())
	})
}

func TestLoad(t *testing.T) {
	originalEnv := map[string]string{
		"PORT":                  os.Getenv("PORT"),
		"DATABASE_URL":          os.Getenv("DATABASE_URL"),
		"REDIS_URL":             os.Getenv("REDIS_URL"),
		"KAKAO_SIGNATURE_SECRET": os.Getenv("KAKAO_SIGNATURE_SECRET"),
		"ADMIN_PASSWORD_HASH":   os.Getenv("ADMIN_PASSWORD_HASH"),
		"QUEUE_TTL_SECONDS":     os.Getenv("QUEUE_TTL_SECONDS"),
		"CALLBACK_TTL_SECONDS":  os.Getenv("CALLBACK_TTL_SECONDS"),
		"LOG_LEVEL":             os.Getenv("LOG_LEVEL"),
	}

	defer func() {
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	t.Run("loads config with defaults", func(t *testing.T) {
		os.Setenv("DATABASE_URL", "postgres://localhost/test")
		os.Setenv("REDIS_URL", "redis://localhost:6379")
		os.Unsetenv("PORT")
		os.Unsetenv("QUEUE_TTL_SECONDS")
		os.Unsetenv("CALLBACK_TTL_SECONDS")
		os.Unsetenv("LOG_LEVEL")

		cfg, err := Load()
		require.NoError(t, err)

		assert.Equal(t, 8080, cfg.Port)
		assert.Equal(t, "postgres://localhost/test", cfg.DatabaseURL)
		assert.Equal(t, "redis://localhost:6379", cfg.RedisURL)
		assert.Equal(t, 900, cfg.QueueTTLSeconds)
		assert.Equal(t, 55, cfg.CallbackTTLSeconds)
		assert.Equal(t, "info", cfg.LogLevel)
	})

	t.Run("loads custom values", func(t *testing.T) {
		os.Setenv("DATABASE_URL", "postgres://localhost/test")
		os.Setenv("REDIS_URL", "redis://localhost:6379")
		os.Setenv("PORT", "3000")
		os.Setenv("QUEUE_TTL_SECONDS", "600")
		os.Setenv("LOG_LEVEL", "debug")

		cfg, err := Load()
		require.NoError(t, err)

		assert.Equal(t, 3000, cfg.Port)
		assert.Equal(t, 600, cfg.QueueTTLSeconds)
		assert.Equal(t, "debug", cfg.LogLevel)
	})

	t.Run("fails without required DATABASE_URL", func(t *testing.T) {
		os.Unsetenv("DATABASE_URL")
		os.Setenv("REDIS_URL", "redis://localhost:6379")

		_, err := Load()
		assert.Error(t, err)
	})

	t.Run("fails without required REDIS_URL", func(t *testing.T) {
		os.Setenv("DATABASE_URL", "postgres://localhost/test")
		os.Unsetenv("REDIS_URL")

		_, err := Load()
		assert.Error(t, err)
	})
}
