package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/rs/zerolog/log"
)

var knownWeakSecrets = []string{
	"change-me", "dev-secret-change-me", "secret", "admin", "password",
}

type Config struct {
	Port                 int    `env:"PORT" envDefault:"8080"`
	DatabaseURL          string `env:"DATABASE_URL,required"`
	RedisURL             string `env:"REDIS_URL,required"`
	KakaoSignatureSecret string `env:"KAKAO_SIGNATURE_SECRET"`
	AdminPasswordHash    string `env:"ADMIN_PASSWORD_HASH"`
	AdminSessionSecret   string `env:"ADMIN_SESSION_SECRET"`
	PortalSessionSecret  string `env:"PORTAL_SESSION_SECRET"`
	EncryptionKey        string `env:"ENCRYPTION_KEY"`
	QueueTTLSeconds      int    `env:"QUEUE_TTL_SECONDS" envDefault:"900"`
	CallbackTTLSeconds   int    `env:"CALLBACK_TTL_SECONDS" envDefault:"55"`
	LogLevel             string `env:"LOG_LEVEL" envDefault:"info"`
	PortalBaseURL        string `env:"PORTAL_BASE_URL" envDefault:""`
}

func (c *Config) QueueTTL() time.Duration {
	return time.Duration(c.QueueTTLSeconds) * time.Second
}

func (c *Config) CallbackTTL() time.Duration {
	return time.Duration(c.CallbackTTLSeconds) * time.Second
}

func (c *Config) Addr() string {
	return fmt.Sprintf(":%d", c.Port)
}

func (c *Config) Validate(isProduction bool) error {
	if c.AdminPasswordHash != "" {
		if !strings.HasPrefix(c.AdminPasswordHash, "$2a$") &&
			!strings.HasPrefix(c.AdminPasswordHash, "$2b$") &&
			!strings.HasPrefix(c.AdminPasswordHash, "$2y$") {
			return fmt.Errorf("ADMIN_PASSWORD_HASH must be a bcrypt hash (generate with: go run scripts/hash-password.go <password>)")
		}
	}

	if isProduction {
		if err := validateSecret("ADMIN_SESSION_SECRET", c.AdminSessionSecret); err != nil {
			return err
		}
		if err := validateSecret("PORTAL_SESSION_SECRET", c.PortalSessionSecret); err != nil {
			return err
		}

		if c.KakaoSignatureSecret == "" {
			log.Warn().Msg("KAKAO_SIGNATURE_SECRET is empty in production: webhook signature verification disabled")
		}
		if strings.HasPrefix(c.RedisURL, "redis://") {
			log.Warn().Msg("REDIS_URL uses redis:// (not TLS) in production: consider using rediss://")
		}
		if c.EncryptionKey == "" {
			log.Warn().Msg("ENCRYPTION_KEY is empty in production: sensitive data will not be encrypted at rest")
		}
	}

	return nil
}

func validateSecret(name, value string) error {
	if len(value) < 32 {
		return fmt.Errorf("%s must be at least 32 characters in production (generate with: openssl rand -base64 32)", name)
	}
	for _, weak := range knownWeakSecrets {
		if value == weak {
			return fmt.Errorf("%s is a known weak default; set a strong secret in production", name)
		}
	}
	return nil
}

func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &cfg, nil
}
