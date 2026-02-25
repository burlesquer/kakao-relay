package model

import (
	"time"
)

type Account struct {
	ID              string      `db:"id" json:"id"`
	OpenclawUserID  *string     `db:"openclaw_user_id" json:"openclawUserId,omitempty"`
	RelayTokenHash  *string     `db:"relay_token_hash" json:"-"`
	Mode            AccountMode `db:"mode" json:"mode"`
	RateLimitPerMin int         `db:"rate_limit_per_minute" json:"rateLimitPerMinute"`
	CreatedAt       time.Time   `db:"created_at" json:"createdAt"`
	UpdatedAt       time.Time   `db:"updated_at" json:"updatedAt"`
	DisabledAt      *time.Time  `db:"disabled_at" json:"disabledAt,omitempty"`
}

type CreateAccountParams struct {
	OpenclawUserID  *string
	RelayTokenHash  string
	Mode            AccountMode
	RateLimitPerMin int
}

type UpdateAccountParams struct {
	OpenclawUserID  *string
	Mode            *AccountMode
	RateLimitPerMin *int
	DisabledAt      *time.Time
}
