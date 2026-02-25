package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/openclaw/relay-server-go/internal/model"
	"github.com/openclaw/relay-server-go/internal/repository"
)

type CreateInboundParams struct {
	AccountID         string
	ConversationKey   string
	KakaoPayload      json.RawMessage
	NormalizedMessage json.RawMessage
	CallbackURL       *string
	CallbackExpiresAt *time.Time
	SourceEventID     *string
}

type MessageService struct {
	inboundRepo  repository.InboundMessageRepository
	outboundRepo repository.OutboundMessageRepository
}

func NewMessageService(
	inboundRepo repository.InboundMessageRepository,
	outboundRepo repository.OutboundMessageRepository,
) *MessageService {
	return &MessageService{
		inboundRepo:  inboundRepo,
		outboundRepo: outboundRepo,
	}
}

func (s *MessageService) CreateInbound(ctx context.Context, params CreateInboundParams) (*model.InboundMessage, error) {
	msg, err := s.inboundRepo.Create(ctx, model.CreateInboundMessageParams{
		AccountID:         params.AccountID,
		ConversationKey:   params.ConversationKey,
		KakaoPayload:      params.KakaoPayload,
		NormalizedMessage: params.NormalizedMessage,
		CallbackURL:       params.CallbackURL,
		CallbackExpiresAt: params.CallbackExpiresAt,
		SourceEventID:     params.SourceEventID,
	})
	if err != nil {
		return nil, fmt.Errorf("create inbound message: %w", err)
	}

	log.Info().
		Str("messageId", msg.ID).
		Str("accountId", params.AccountID).
		Str("conversationKey", params.ConversationKey).
		Msg("inbound message created")

	return msg, nil
}

func (s *MessageService) FindInboundByID(ctx context.Context, id string) (*model.InboundMessage, error) {
	return s.inboundRepo.FindByID(ctx, id)
}

func (s *MessageService) FindQueuedByAccountID(ctx context.Context, accountID string) ([]model.InboundMessage, error) {
	return s.inboundRepo.FindQueuedByAccountID(ctx, accountID)
}

func (s *MessageService) MarkDelivered(ctx context.Context, id string) error {
	if err := s.inboundRepo.MarkDelivered(ctx, id); err != nil {
		return fmt.Errorf("mark delivered: %w", err)
	}
	log.Debug().Str("messageId", id).Msg("message marked as delivered")
	return nil
}

func (s *MessageService) MarkAcked(ctx context.Context, id string) error {
	if err := s.inboundRepo.MarkAcked(ctx, id); err != nil {
		return fmt.Errorf("mark acked: %w", err)
	}
	log.Debug().Str("messageId", id).Msg("message marked as acked")
	return nil
}

func (s *MessageService) CreateOutbound(ctx context.Context, params model.CreateOutboundMessageParams) (*model.OutboundMessage, error) {
	msg, err := s.outboundRepo.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create outbound message: %w", err)
	}

	log.Info().
		Str("messageId", msg.ID).
		Str("accountId", params.AccountID).
		Str("conversationKey", params.ConversationKey).
		Msg("outbound message created")

	return msg, nil
}

func (s *MessageService) MarkOutboundSent(ctx context.Context, id string) error {
	return s.outboundRepo.MarkSent(ctx, id)
}

func (s *MessageService) MarkOutboundFailed(ctx context.Context, id, errorMsg string) error {
	return s.outboundRepo.MarkFailed(ctx, id, errorMsg)
}

type MessageHistoryParams struct {
	AccountID string
	Type      string // "inbound", "outbound", or "" for all
	Limit     int
	Offset    int
}

type MessageHistoryResult struct {
	Messages []MessageHistoryItem
	Total    int
	HasMore  bool
}

type MessageHistoryItem struct {
	ID              string           `json:"id"`
	ConversationKey string           `json:"conversationKey"`
	Direction       string           `json:"direction"`
	Content         *json.RawMessage `json:"content,omitempty"`
	CreatedAt       time.Time        `json:"createdAt"`
}

// UserStats represents statistics for a user's account
type UserStats struct {
	Connections struct {
		Total   int `json:"total"`
		Paired  int `json:"paired"`
		Blocked int `json:"blocked"`
	} `json:"connections"`
	Messages struct {
		Inbound struct {
			Today   int `json:"today"`
			Total   int `json:"total"`
			Queued  int `json:"queued"`
			Expired int `json:"expired"`
		} `json:"inbound"`
		Outbound struct {
			Today  int `json:"today"`
			Total  int `json:"total"`
			Sent   int `json:"sent"`
			Failed int `json:"failed"`
		} `json:"outbound"`
	} `json:"messages"`
	RecentErrors []RecentError `json:"recentErrors"`
	LastActivity *time.Time    `json:"lastActivity"`
}

// RecentError represents a recent failed message
type RecentError struct {
	ID              string    `json:"id"`
	ConversationKey string    `json:"conversationKey"`
	ErrorMessage    string    `json:"errorMessage"`
	CreatedAt       time.Time `json:"createdAt"`
}

// GetUserStats returns statistics for a specific account
func (s *MessageService) GetUserStats(ctx context.Context, accountID string, connections []ConnectionStat) (*UserStats, error) {
	stats := &UserStats{}

	for _, conn := range connections {
		stats.Connections.Total++
		switch conn.State {
		case "paired":
			stats.Connections.Paired++
		case "blocked":
			stats.Connections.Blocked++
		}
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	inboundTotal, err := s.inboundRepo.CountByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("count inbound messages: %w", err)
	}
	stats.Messages.Inbound.Total = inboundTotal

	inboundToday, err := s.inboundRepo.CountByAccountIDSince(ctx, accountID, todayStart)
	if err != nil {
		return nil, fmt.Errorf("count inbound messages today: %w", err)
	}
	stats.Messages.Inbound.Today = inboundToday

	queuedCount, err := s.inboundRepo.CountByAccountIDAndStatus(ctx, accountID, model.InboundStatusQueued)
	if err != nil {
		return nil, fmt.Errorf("count queued messages: %w", err)
	}
	stats.Messages.Inbound.Queued = queuedCount

	expiredCount, err := s.inboundRepo.CountByAccountIDAndStatus(ctx, accountID, model.InboundStatusExpired)
	if err != nil {
		return nil, fmt.Errorf("count expired messages: %w", err)
	}
	stats.Messages.Inbound.Expired = expiredCount

	outboundTotal, err := s.outboundRepo.CountByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("count outbound messages: %w", err)
	}
	stats.Messages.Outbound.Total = outboundTotal

	outboundToday, err := s.outboundRepo.CountByAccountIDSince(ctx, accountID, todayStart)
	if err != nil {
		return nil, fmt.Errorf("count outbound messages today: %w", err)
	}
	stats.Messages.Outbound.Today = outboundToday

	sentCount, err := s.outboundRepo.CountByAccountIDAndStatus(ctx, accountID, model.OutboundStatusSent)
	if err != nil {
		return nil, fmt.Errorf("count sent messages: %w", err)
	}
	stats.Messages.Outbound.Sent = sentCount

	failedCount, err := s.outboundRepo.CountByAccountIDAndStatus(ctx, accountID, model.OutboundStatusFailed)
	if err != nil {
		return nil, fmt.Errorf("count failed messages: %w", err)
	}
	stats.Messages.Outbound.Failed = failedCount

	failedMsgs, err := s.outboundRepo.FindRecentFailedByAccountID(ctx, accountID, 5)
	if err != nil {
		return nil, fmt.Errorf("find recent errors: %w", err)
	}
	stats.RecentErrors = make([]RecentError, len(failedMsgs))
	for i, msg := range failedMsgs {
		errorMsg := ""
		if msg.ErrorMessage != nil {
			errorMsg = *msg.ErrorMessage
		}
		stats.RecentErrors[i] = RecentError{
			ID:              msg.ID,
			ConversationKey: msg.ConversationKey,
			ErrorMessage:    errorMsg,
			CreatedAt:       msg.CreatedAt,
		}
	}

	for _, conn := range connections {
		if conn.LastSeenAt != nil {
			if stats.LastActivity == nil || conn.LastSeenAt.After(*stats.LastActivity) {
				stats.LastActivity = conn.LastSeenAt
			}
		}
	}

	return stats, nil
}

// ConnectionStat is used for passing connection info to GetUserStats
type ConnectionStat struct {
	State      string
	LastSeenAt *time.Time
}

// QuickStats represents basic message statistics for display in chat
type QuickStats struct {
	InboundToday   int
	InboundTotal   int
	OutboundToday  int
	OutboundTotal  int
	OutboundFailed int
}

// GetQuickStats returns simple message counts for an account (used by /status command)
func (s *MessageService) GetQuickStats(ctx context.Context, accountID string) (*QuickStats, error) {
	stats := &QuickStats{}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	inboundTotal, err := s.inboundRepo.CountByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("count inbound messages: %w", err)
	}
	stats.InboundTotal = inboundTotal

	inboundToday, err := s.inboundRepo.CountByAccountIDSince(ctx, accountID, todayStart)
	if err != nil {
		return nil, fmt.Errorf("count inbound today: %w", err)
	}
	stats.InboundToday = inboundToday

	outboundTotal, err := s.outboundRepo.CountByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("count outbound messages: %w", err)
	}
	stats.OutboundTotal = outboundTotal

	outboundToday, err := s.outboundRepo.CountByAccountIDSince(ctx, accountID, todayStart)
	if err != nil {
		return nil, fmt.Errorf("count outbound today: %w", err)
	}
	stats.OutboundToday = outboundToday

	failedCount, err := s.outboundRepo.CountByAccountIDAndStatus(ctx, accountID, model.OutboundStatusFailed)
	if err != nil {
		return nil, fmt.Errorf("count failed messages: %w", err)
	}
	stats.OutboundFailed = failedCount

	return stats, nil
}

func (s *MessageService) GetMessageHistory(ctx context.Context, params MessageHistoryParams) (*MessageHistoryResult, error) {
	var messages []MessageHistoryItem
	var total int

	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	switch params.Type {
	case "inbound":
		inboundMsgs, err := s.inboundRepo.FindByAccountID(ctx, params.AccountID, params.Limit, params.Offset)
		if err != nil {
			return nil, fmt.Errorf("find inbound messages: %w", err)
		}
		total, err = s.inboundRepo.CountByAccountID(ctx, params.AccountID)
		if err != nil {
			return nil, fmt.Errorf("count inbound messages: %w", err)
		}
		for _, msg := range inboundMsgs {
			messages = append(messages, MessageHistoryItem{
				ID:              msg.ID,
				ConversationKey: msg.ConversationKey,
				Direction:       "inbound",
				Content:         msg.NormalizedMessage,
				CreatedAt:       msg.CreatedAt,
			})
		}

	case "outbound":
		outboundMsgs, err := s.outboundRepo.FindByAccountID(ctx, params.AccountID, params.Limit, params.Offset)
		if err != nil {
			return nil, fmt.Errorf("find outbound messages: %w", err)
		}
		total, err = s.outboundRepo.CountByAccountID(ctx, params.AccountID)
		if err != nil {
			return nil, fmt.Errorf("count outbound messages: %w", err)
		}
		for _, msg := range outboundMsgs {
			payload := msg.ResponsePayload
			messages = append(messages, MessageHistoryItem{
				ID:              msg.ID,
				ConversationKey: msg.ConversationKey,
				Direction:       "outbound",
				Content:         &payload,
				CreatedAt:       msg.CreatedAt,
			})
		}

	default:
		// Fetch both and merge by created_at
		inboundMsgs, err := s.inboundRepo.FindByAccountID(ctx, params.AccountID, params.Limit, params.Offset)
		if err != nil {
			return nil, fmt.Errorf("find inbound messages: %w", err)
		}
		outboundMsgs, err := s.outboundRepo.FindByAccountID(ctx, params.AccountID, params.Limit, params.Offset)
		if err != nil {
			return nil, fmt.Errorf("find outbound messages: %w", err)
		}

		inboundCount, err := s.inboundRepo.CountByAccountID(ctx, params.AccountID)
		if err != nil {
			return nil, fmt.Errorf("count inbound messages: %w", err)
		}
		outboundCount, err := s.outboundRepo.CountByAccountID(ctx, params.AccountID)
		if err != nil {
			return nil, fmt.Errorf("count outbound messages: %w", err)
		}
		total = inboundCount + outboundCount

		for _, msg := range inboundMsgs {
			messages = append(messages, MessageHistoryItem{
				ID:              msg.ID,
				ConversationKey: msg.ConversationKey,
				Direction:       "inbound",
				Content:         msg.NormalizedMessage,
				CreatedAt:       msg.CreatedAt,
			})
		}
		for _, msg := range outboundMsgs {
			payload := msg.ResponsePayload
			messages = append(messages, MessageHistoryItem{
				ID:              msg.ID,
				ConversationKey: msg.ConversationKey,
				Direction:       "outbound",
				Content:         &payload,
				CreatedAt:       msg.CreatedAt,
			})
		}

		// Sort by created_at descending
		sort.Slice(messages, func(i, j int) bool {
			return messages[i].CreatedAt.After(messages[j].CreatedAt)
		})

		// Limit results
		if len(messages) > params.Limit {
			messages = messages[:params.Limit]
		}
	}

	return &MessageHistoryResult{
		Messages: messages,
		Total:    total,
		HasMore:  params.Offset+len(messages) < total,
	}, nil
}

// ConversationStats represents statistics for a specific conversation
type ConversationStats struct {
	ConversationKey string `json:"conversationKey"`
	Messages        struct {
		Inbound struct {
			Today int `json:"today"`
			Total int `json:"total"`
		} `json:"inbound"`
		Outbound struct {
			Today  int `json:"today"`
			Total  int `json:"total"`
			Failed int `json:"failed"`
		} `json:"outbound"`
	} `json:"messages"`
}

// GetConversationStats returns statistics for a specific conversation
func (s *MessageService) GetConversationStats(ctx context.Context, conversationKey string) (*ConversationStats, error) {
	stats := &ConversationStats{
		ConversationKey: conversationKey,
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	inboundTotal, err := s.inboundRepo.CountByConversationKey(ctx, conversationKey)
	if err != nil {
		return nil, fmt.Errorf("count inbound messages: %w", err)
	}
	stats.Messages.Inbound.Total = inboundTotal

	inboundToday, err := s.inboundRepo.CountByConversationKeySince(ctx, conversationKey, todayStart)
	if err != nil {
		return nil, fmt.Errorf("count inbound messages today: %w", err)
	}
	stats.Messages.Inbound.Today = inboundToday

	outboundTotal, err := s.outboundRepo.CountByConversationKey(ctx, conversationKey)
	if err != nil {
		return nil, fmt.Errorf("count outbound messages: %w", err)
	}
	stats.Messages.Outbound.Total = outboundTotal

	outboundToday, err := s.outboundRepo.CountByConversationKeySince(ctx, conversationKey, todayStart)
	if err != nil {
		return nil, fmt.Errorf("count outbound messages today: %w", err)
	}
	stats.Messages.Outbound.Today = outboundToday

	failedCount, err := s.outboundRepo.CountByConversationKeyAndStatus(ctx, conversationKey, model.OutboundStatusFailed)
	if err != nil {
		return nil, fmt.Errorf("count failed messages: %w", err)
	}
	stats.Messages.Outbound.Failed = failedCount

	return stats, nil
}

// ConversationMessagesParams contains parameters for fetching conversation messages
type ConversationMessagesParams struct {
	ConversationKey string
	Type            string // "inbound", "outbound", or "" for all
	Limit           int
	Offset          int
}

// GetConversationMessages returns message history for a specific conversation
func (s *MessageService) GetConversationMessages(ctx context.Context, params ConversationMessagesParams) (*MessageHistoryResult, error) {
	var messages []MessageHistoryItem
	var total int

	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	switch params.Type {
	case "inbound":
		inboundMsgs, err := s.inboundRepo.FindByConversationKey(ctx, params.ConversationKey, params.Limit, params.Offset)
		if err != nil {
			return nil, fmt.Errorf("find inbound messages: %w", err)
		}
		total, err = s.inboundRepo.CountByConversationKey(ctx, params.ConversationKey)
		if err != nil {
			return nil, fmt.Errorf("count inbound messages: %w", err)
		}
		for _, msg := range inboundMsgs {
			messages = append(messages, MessageHistoryItem{
				ID:              msg.ID,
				ConversationKey: msg.ConversationKey,
				Direction:       "inbound",
				Content:         msg.NormalizedMessage,
				CreatedAt:       msg.CreatedAt,
			})
		}

	case "outbound":
		outboundMsgs, err := s.outboundRepo.FindByConversationKey(ctx, params.ConversationKey, params.Limit, params.Offset)
		if err != nil {
			return nil, fmt.Errorf("find outbound messages: %w", err)
		}
		total, err = s.outboundRepo.CountByConversationKey(ctx, params.ConversationKey)
		if err != nil {
			return nil, fmt.Errorf("count outbound messages: %w", err)
		}
		for _, msg := range outboundMsgs {
			payload := msg.ResponsePayload
			messages = append(messages, MessageHistoryItem{
				ID:              msg.ID,
				ConversationKey: msg.ConversationKey,
				Direction:       "outbound",
				Content:         &payload,
				CreatedAt:       msg.CreatedAt,
			})
		}

	default:
		// Fetch both and merge
		inboundMsgs, err := s.inboundRepo.FindByConversationKey(ctx, params.ConversationKey, params.Limit, params.Offset)
		if err != nil {
			return nil, fmt.Errorf("find inbound messages: %w", err)
		}
		outboundMsgs, err := s.outboundRepo.FindByConversationKey(ctx, params.ConversationKey, params.Limit, params.Offset)
		if err != nil {
			return nil, fmt.Errorf("find outbound messages: %w", err)
		}

		inboundCount, err := s.inboundRepo.CountByConversationKey(ctx, params.ConversationKey)
		if err != nil {
			return nil, fmt.Errorf("count inbound messages: %w", err)
		}
		outboundCount, err := s.outboundRepo.CountByConversationKey(ctx, params.ConversationKey)
		if err != nil {
			return nil, fmt.Errorf("count outbound messages: %w", err)
		}
		total = inboundCount + outboundCount

		for _, msg := range inboundMsgs {
			messages = append(messages, MessageHistoryItem{
				ID:              msg.ID,
				ConversationKey: msg.ConversationKey,
				Direction:       "inbound",
				Content:         msg.NormalizedMessage,
				CreatedAt:       msg.CreatedAt,
			})
		}
		for _, msg := range outboundMsgs {
			payload := msg.ResponsePayload
			messages = append(messages, MessageHistoryItem{
				ID:              msg.ID,
				ConversationKey: msg.ConversationKey,
				Direction:       "outbound",
				Content:         &payload,
				CreatedAt:       msg.CreatedAt,
			})
		}

		// Sort by created_at descending
		sort.Slice(messages, func(i, j int) bool {
			return messages[i].CreatedAt.After(messages[j].CreatedAt)
		})

		// Limit results
		if len(messages) > params.Limit {
			messages = messages[:params.Limit]
		}
	}

	return &MessageHistoryResult{
		Messages: messages,
		Total:    total,
		HasMore:  params.Offset+len(messages) < total,
	}, nil
}
