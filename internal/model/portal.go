package model

import (
	"time"
)

type PortalUser struct {
	ID          string     `db:"id" json:"id"`
	Email       string     `db:"email" json:"email"`
	AccountID   string     `db:"account_id" json:"accountId"`
	CreatedAt   time.Time  `db:"created_at" json:"createdAt"`
	LastLoginAt *time.Time `db:"last_login_at" json:"lastLoginAt,omitempty"`
}

type CreatePortalUserParams struct {
	Email     string
	AccountID string
}

type PortalSession struct {
	ID        string    `db:"id" json:"id"`
	TokenHash string    `db:"token_hash" json:"-"`
	UserID    string    `db:"user_id" json:"userId"`
	ExpiresAt time.Time `db:"expires_at" json:"expiresAt"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type CreatePortalSessionParams struct {
	TokenHash string
	UserID    string
	ExpiresAt time.Time
}
