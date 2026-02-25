package model

import (
	"time"
)

type ConversationMapping struct {
	ID                    string       `db:"id" json:"id"`
	ConversationKey       string       `db:"conversation_key" json:"conversationKey"`
	KakaoChannelID        string       `db:"kakao_channel_id" json:"kakaoChannelId"`
	PlusfriendUserKey     string       `db:"plusfriend_user_key" json:"plusfriendUserKey"`
	AccountID             *string      `db:"account_id" json:"accountId,omitempty"`
	State                 PairingState `db:"state" json:"state"`
	LastCallbackURL       *string      `db:"last_callback_url" json:"-"`
	LastCallbackExpiresAt *time.Time   `db:"last_callback_expires_at" json:"-"`
	FirstSeenAt           time.Time    `db:"first_seen_at" json:"firstSeenAt"`
	LastSeenAt            time.Time    `db:"last_seen_at" json:"lastSeenAt"`
	PairedAt              *time.Time   `db:"paired_at" json:"pairedAt,omitempty"`
}

type UpsertConversationParams struct {
	ConversationKey   string
	KakaoChannelID    string
	PlusfriendUserKey string
	CallbackURL       *string
	CallbackExpiresAt *time.Time
}
