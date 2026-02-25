package service

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/openclaw/relay-server-go/internal/model"
	"github.com/openclaw/relay-server-go/internal/repository"
)

type ConversationService struct {
	repo repository.ConversationRepository
}

func NewConversationService(repo repository.ConversationRepository) *ConversationService {
	return &ConversationService{repo: repo}
}

func BuildConversationKey(channelID, userKey string) string {
	return fmt.Sprintf("%s:%s", channelID, userKey)
}

func (s *ConversationService) FindByKey(ctx context.Context, key string) (*model.ConversationMapping, error) {
	return s.repo.FindByKey(ctx, key)
}

func (s *ConversationService) FindOrCreate(
	ctx context.Context,
	channelID, userKey string,
	callbackURL *string,
	callbackExpiresAt *time.Time,
) (*model.ConversationMapping, error) {
	key := BuildConversationKey(channelID, userKey)

	conv, err := s.repo.Upsert(ctx, model.UpsertConversationParams{
		ConversationKey:   key,
		KakaoChannelID:    channelID,
		PlusfriendUserKey: userKey,
		CallbackURL:       callbackURL,
		CallbackExpiresAt: callbackExpiresAt,
	})
	if err != nil {
		return nil, fmt.Errorf("find or create conversation: %w", err)
	}

	return conv, nil
}

func (s *ConversationService) UpdateState(
	ctx context.Context,
	key string,
	state model.PairingState,
	accountID *string,
) error {
	if err := s.repo.UpdateState(ctx, key, state, accountID); err != nil {
		return fmt.Errorf("update state: %w", err)
	}

	log.Info().
		Str("conversationKey", key).
		Str("state", string(state)).
		Msg("conversation state updated")

	return nil
}

func (s *ConversationService) Unpair(ctx context.Context, key string) error {
	return s.repo.UpdateState(ctx, key, model.PairingStateUnpaired, nil)
}

func (s *ConversationService) ListByAccountID(ctx context.Context, accountID string) ([]model.ConversationMapping, error) {
	return s.repo.FindPairedByAccountID(ctx, accountID)
}
