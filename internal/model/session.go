package model

import (
	"encoding/json"
	"time"
)

type Session struct {
	ID                    string           `db:"id" json:"id"`
	SessionTokenHash      string           `db:"session_token_hash" json:"-"`
	PairingCode           string           `db:"pairing_code" json:"pairingCode"`
	Status                SessionStatus    `db:"status" json:"status"`
	AccountID             *string          `db:"account_id" json:"accountId,omitempty"`
	PairedConversationKey *string          `db:"paired_conversation_key" json:"pairedConversationKey,omitempty"`
	Metadata              *json.RawMessage `db:"metadata" json:"metadata,omitempty"`
	ExpiresAt             time.Time        `db:"expires_at" json:"expiresAt"`
	PairedAt              *time.Time       `db:"paired_at" json:"pairedAt,omitempty"`
	CreatedAt             time.Time        `db:"created_at" json:"createdAt"`
	UpdatedAt             time.Time        `db:"updated_at" json:"updatedAt"`
}

type CreateSessionParams struct {
	SessionTokenHash string
	PairingCode      string
	ExpiresAt        time.Time
	Metadata         *json.RawMessage
}
