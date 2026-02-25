package model

import (
	"encoding/json"
	"time"
)

type PairingCode struct {
	Code      string           `db:"code" json:"code"`
	AccountID string           `db:"account_id" json:"accountId"`
	ExpiresAt time.Time        `db:"expires_at" json:"expiresAt"`
	UsedAt    *time.Time       `db:"used_at" json:"usedAt,omitempty"`
	UsedBy    *string          `db:"used_by" json:"usedBy,omitempty"`
	Metadata  *json.RawMessage `db:"metadata" json:"metadata,omitempty"`
	CreatedAt time.Time        `db:"created_at" json:"createdAt"`
}

type CreatePairingCodeParams struct {
	Code      string
	AccountID string
	ExpiresAt time.Time
	Metadata  *json.RawMessage
}
