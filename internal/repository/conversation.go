package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/openclaw/relay-server-go/internal/model"
)

type ConversationRepository interface {
	FindByKey(ctx context.Context, key string) (*model.ConversationMapping, error)
	FindByAccountID(ctx context.Context, accountID string) ([]model.ConversationMapping, error)
	FindPairedByAccountID(ctx context.Context, accountID string) ([]model.ConversationMapping, error)
	Upsert(ctx context.Context, params model.UpsertConversationParams) (*model.ConversationMapping, error)
	UpdateState(ctx context.Context, key string, state model.PairingState, accountID *string) error
	UpdateCallback(ctx context.Context, key string, callbackURL string, expiresAt time.Time) error
	Delete(ctx context.Context, id string) error
	CountByState(ctx context.Context, state model.PairingState) (int, error)
}

type conversationRepo struct {
	db *sqlx.DB
}

func NewConversationRepository(db *sqlx.DB) ConversationRepository {
	return &conversationRepo{db: db}
}

func (r *conversationRepo) FindByKey(ctx context.Context, key string) (*model.ConversationMapping, error) {
	var conv model.ConversationMapping
	err := r.db.GetContext(ctx, &conv, `
		SELECT * FROM conversation_mappings WHERE conversation_key = $1
	`, key)
	return HandleNotFound(&conv, err)
}

func (r *conversationRepo) FindByAccountID(ctx context.Context, accountID string) ([]model.ConversationMapping, error) {
	var convs []model.ConversationMapping
	err := r.db.SelectContext(ctx, &convs, `
		SELECT * FROM conversation_mappings
		WHERE account_id = $1
		ORDER BY last_seen_at DESC
	`, accountID)
	return convs, err
}

func (r *conversationRepo) FindPairedByAccountID(ctx context.Context, accountID string) ([]model.ConversationMapping, error) {
	var convs []model.ConversationMapping
	err := r.db.SelectContext(ctx, &convs, `
		SELECT * FROM conversation_mappings
		WHERE account_id = $1 AND state = 'paired'
		ORDER BY paired_at DESC
	`, accountID)
	return convs, err
}

func (r *conversationRepo) Upsert(ctx context.Context, params model.UpsertConversationParams) (*model.ConversationMapping, error) {
	var conv model.ConversationMapping
	err := r.db.GetContext(ctx, &conv, `
		INSERT INTO conversation_mappings
			(conversation_key, kakao_channel_id, plusfriend_user_key, last_callback_url, last_callback_expires_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (conversation_key) DO UPDATE SET
			last_callback_url = EXCLUDED.last_callback_url,
			last_callback_expires_at = EXCLUDED.last_callback_expires_at,
			last_seen_at = NOW()
		RETURNING *
	`, params.ConversationKey, params.KakaoChannelID, params.PlusfriendUserKey,
		params.CallbackURL, params.CallbackExpiresAt)
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *conversationRepo) UpdateState(ctx context.Context, key string, state model.PairingState, accountID *string) error {
	var pairedAt interface{}
	if state == model.PairingStatePaired {
		pairedAt = time.Now()
	}

	_, err := r.db.ExecContext(ctx, `
		UPDATE conversation_mappings SET
			state = $2,
			account_id = $3,
			paired_at = COALESCE($4, paired_at)
		WHERE conversation_key = $1
	`, key, state, accountID, pairedAt)
	return err
}

func (r *conversationRepo) UpdateCallback(ctx context.Context, key string, callbackURL string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE conversation_mappings SET
			last_callback_url = $2,
			last_callback_expires_at = $3,
			last_seen_at = NOW()
		WHERE conversation_key = $1
	`, key, callbackURL, expiresAt)
	return err
}

func (r *conversationRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM conversation_mappings WHERE id = $1`, id)
	return err
}

func (r *conversationRepo) CountByState(ctx context.Context, state model.PairingState) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM conversation_mappings WHERE state = $1
	`, state)
	return count, err
}
