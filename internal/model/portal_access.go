package model

import (
	"time"
)

// PortalAccessCode represents a temporary access code for portal login
type PortalAccessCode struct {
	Code            string     `db:"code" json:"code"`
	ConversationKey string     `db:"conversation_key" json:"conversationKey"`
	ExpiresAt       time.Time  `db:"expires_at" json:"expiresAt"`
	CreatedAt       time.Time  `db:"created_at" json:"createdAt"`
	UsedAt          *time.Time `db:"used_at" json:"usedAt,omitempty"`
	LastAccessedAt  *time.Time `db:"last_accessed_at" json:"lastAccessedAt,omitempty"`
}

// CreatePortalAccessCodeParams contains parameters for creating a portal access code
type CreatePortalAccessCodeParams struct {
	Code            string
	ConversationKey string
	ExpiresAt       time.Time
}

// IsExpired checks if the code has expired
func (c *PortalAccessCode) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// IsValid checks if the code is valid (not expired and not used)
func (c *PortalAccessCode) IsValid() bool {
	return !c.IsExpired() && c.UsedAt == nil
}
