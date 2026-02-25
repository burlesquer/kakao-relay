package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/openclaw/relay-server-go/internal/model"
)

// PortalAccessCodeRepository handles portal access code data operations
type PortalAccessCodeRepository interface {
	Create(ctx context.Context, params model.CreatePortalAccessCodeParams) (*model.PortalAccessCode, error)
	FindActiveByCode(ctx context.Context, code string) (*model.PortalAccessCode, error)
	FindActiveByConversationKey(ctx context.Context, conversationKey string) (*model.PortalAccessCode, error)
	MarkUsed(ctx context.Context, code string) error
	UpdateLastAccessed(ctx context.Context, code string) error
	DeleteExpired(ctx context.Context) (int64, error)
}

type portalAccessCodeRepo struct {
	db *sqlx.DB
}

// NewPortalAccessCodeRepository creates a new portal access code repository
func NewPortalAccessCodeRepository(db *sqlx.DB) PortalAccessCodeRepository {
	return &portalAccessCodeRepo{db: db}
}

// Create creates a new portal access code
func (r *portalAccessCodeRepo) Create(ctx context.Context, params model.CreatePortalAccessCodeParams) (*model.PortalAccessCode, error) {
	var code model.PortalAccessCode
	err := r.db.GetContext(ctx, &code, `
		INSERT INTO portal_access_codes (code, conversation_key, expires_at)
		VALUES ($1, $2, $3)
		RETURNING *
	`, params.Code, params.ConversationKey, params.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &code, nil
}

// FindActiveByCode finds an active (not expired, not used) code by its code string
func (r *portalAccessCodeRepo) FindActiveByCode(ctx context.Context, code string) (*model.PortalAccessCode, error) {
	var pac model.PortalAccessCode
	err := r.db.GetContext(ctx, &pac, `
		SELECT * FROM portal_access_codes
		WHERE code = $1 AND used_at IS NULL AND expires_at > NOW()
	`, code)
	return HandleNotFound(&pac, err)
}

// FindActiveByConversationKey finds an active code for a conversation key
func (r *portalAccessCodeRepo) FindActiveByConversationKey(ctx context.Context, conversationKey string) (*model.PortalAccessCode, error) {
	var pac model.PortalAccessCode
	err := r.db.GetContext(ctx, &pac, `
		SELECT * FROM portal_access_codes
		WHERE conversation_key = $1 AND used_at IS NULL AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`, conversationKey)
	return HandleNotFound(&pac, err)
}

// MarkUsed marks a code as used
func (r *portalAccessCodeRepo) MarkUsed(ctx context.Context, code string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE portal_access_codes
		SET used_at = $1
		WHERE code = $2
	`, time.Now(), code)
	return err
}

// UpdateLastAccessed updates the last accessed timestamp
func (r *portalAccessCodeRepo) UpdateLastAccessed(ctx context.Context, code string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE portal_access_codes
		SET last_accessed_at = $1
		WHERE code = $2
	`, time.Now(), code)
	return err
}

// DeleteExpired deletes expired codes
func (r *portalAccessCodeRepo) DeleteExpired(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM portal_access_codes
		WHERE expires_at < NOW()
	`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
