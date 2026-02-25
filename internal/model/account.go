package model

import (
	"time"
)

type Account struct {
	ID              string  `db:"id" json:"id"`
	RelayTokenHash  *string `db:"relay_token_hash" json:"-"`
	RateLimitPerMin int     `db:"rate_limit_per_minute" json:"rateLimitPerMinute"`
	CreatedAt       time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt       time.Time `db:"updated_at" json:"updatedAt"`
}

type CreateAccountParams struct {
	RelayTokenHash  string
	RateLimitPerMin int
}

type UpdateAccountParams struct {
	RateLimitPerMin *int
}
