package model

import (
	"encoding/json"
	"time"
)

type InboundMessage struct {
	ID                string               `db:"id" json:"id"`
	AccountID         string               `db:"account_id" json:"accountId"`
	ConversationKey   string               `db:"conversation_key" json:"conversationKey"`
	KakaoPayload      json.RawMessage      `db:"kakao_payload" json:"kakaoPayload"`
	NormalizedMessage *json.RawMessage     `db:"normalized_message" json:"normalizedMessage,omitempty"`
	CallbackURL       *string              `db:"callback_url" json:"-"`
	CallbackExpiresAt *time.Time           `db:"callback_expires_at" json:"-"`
	Status            InboundMessageStatus `db:"status" json:"status"`
	SourceEventID     *string              `db:"source_event_id" json:"sourceEventId,omitempty"`
	CreatedAt         time.Time            `db:"created_at" json:"createdAt"`
	DeliveredAt       *time.Time           `db:"delivered_at" json:"deliveredAt,omitempty"`
	AckedAt           *time.Time           `db:"acked_at" json:"ackedAt,omitempty"`
}

// ToSSEEventData returns JSON data for SSE message events
func (m *InboundMessage) ToSSEEventData() json.RawMessage {
	data, _ := json.Marshal(map[string]any{
		"id":              m.ID,
		"conversationKey": m.ConversationKey,
		"kakaoPayload":    m.KakaoPayload,
		"normalized":      m.NormalizedMessage,
		"createdAt":       m.CreatedAt,
	})
	return data
}

type CreateInboundMessageParams struct {
	AccountID         string
	ConversationKey   string
	KakaoPayload      json.RawMessage
	NormalizedMessage json.RawMessage
	CallbackURL       *string
	CallbackExpiresAt *time.Time
	SourceEventID     *string
}

type OutboundMessage struct {
	ID               string                `db:"id" json:"id"`
	AccountID        string                `db:"account_id" json:"accountId"`
	InboundMessageID *string               `db:"inbound_message_id" json:"inboundMessageId,omitempty"`
	ConversationKey  string                `db:"conversation_key" json:"conversationKey"`
	KakaoTarget      json.RawMessage       `db:"kakao_target" json:"kakaoTarget"`
	ResponsePayload  json.RawMessage       `db:"response_payload" json:"responsePayload"`
	Status           OutboundMessageStatus `db:"status" json:"status"`
	ErrorMessage     *string               `db:"error_message" json:"errorMessage,omitempty"`
	CreatedAt        time.Time             `db:"created_at" json:"createdAt"`
	SentAt           *time.Time            `db:"sent_at" json:"sentAt,omitempty"`
}

type CreateOutboundMessageParams struct {
	AccountID        string
	InboundMessageID *string
	ConversationKey  string
	KakaoTarget      json.RawMessage
	ResponsePayload  json.RawMessage
}
