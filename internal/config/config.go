package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Port                 int    `env:"PORT" envDefault:"8080"`
	DatabaseURL          string `env:"DATABASE_URL,required"`
	RedisURL             string `env:"REDIS_URL,required"`
	KakaoSignatureSecret string `env:"KAKAO_SIGNATURE_SECRET"`
	EncryptionKey        string `env:"ENCRYPTION_KEY"`
	QueueTTLSeconds      int    `env:"QUEUE_TTL_SECONDS" envDefault:"900"`
	CallbackTTLSeconds   int    `env:"CALLBACK_TTL_SECONDS" envDefault:"55"`
	LogLevel             string `env:"LOG_LEVEL" envDefault:"info"`
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
	if isProduction {
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

func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &cfg, nil
}
