package model

import (
	"time"
)

type AdminSession struct {
	ID        string    `db:"id" json:"id"`
	TokenHash string    `db:"token_hash" json:"-"`
	ExpiresAt time.Time `db:"expires_at" json:"expiresAt"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type CreateAdminSessionParams struct {
	TokenHash string
	ExpiresAt time.Time
}
