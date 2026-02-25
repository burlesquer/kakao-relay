package config

import "time"

// Database connection pool settings
const (
	DBMaxOpenConns    = 25
	DBMaxIdleConns    = 5
	DBConnMaxLifetime = 5 * time.Minute
)

// HTTP server timeouts
const (
	ServerRequestTimeout  = 60 * time.Second
	ServerReadTimeout     = 15 * time.Second
	ServerIdleTimeout     = 120 * time.Second
	ServerShutdownTimeout = 30 * time.Second
)

// Database ping timeout for health checks
const DBPingTimeout = 5 * time.Second

// Background job intervals
const CleanupJobInterval = 5 * time.Minute

// Default rate limiting
const DefaultRateLimitPerMin = 60
