package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/openclaw/relay-server-go/internal/model"
)

type SessionRepository interface {
	FindByID(ctx context.Context, id string) (*model.Session, error)
	FindByTokenHash(ctx context.Context, tokenHash string) (*model.Session, error)
	FindByPairingCode(ctx context.Context, code string) (*model.Session, error)
	Create(ctx context.Context, params model.CreateSessionParams) (*model.Session, error)
	MarkPaired(ctx context.Context, id string, accountID string, conversationKey string) error
	MarkExpired(ctx context.Context, id string) error
	MarkDisconnected(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context) (int64, error)
	CountPendingByIP(ctx context.Context, ip string, since time.Time) (int, error)
	// WithTx returns a new repository that uses the given transaction
	WithTx(tx *sqlx.Tx) SessionRepository
}

// sessionDB is an interface satisfied by both *sqlx.DB and *sqlx.Tx
type sessionDB interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type sessionRepo struct {
	db sessionDB
}

func NewSessionRepository(db *sqlx.DB) SessionRepository {
	return &sessionRepo{db: db}
}

func (r *sessionRepo) WithTx(tx *sqlx.Tx) SessionRepository {
	return &sessionRepo{db: tx}
}

func (r *sessionRepo) FindByID(ctx context.Context, id string) (*model.Session, error) {
	var session model.Session
	err := r.db.GetContext(ctx, &session, `
		SELECT * FROM sessions WHERE id = $1
	`, id)
	return HandleNotFound(&session, err)
}

func (r *sessionRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*model.Session, error) {
	var session model.Session
	err := r.db.GetContext(ctx, &session, `
		SELECT * FROM sessions
		WHERE session_token_hash = $1
		AND status IN ('pending_pairing', 'paired')
	`, tokenHash)
	return HandleNotFound(&session, err)
}

func (r *sessionRepo) FindByPairingCode(ctx context.Context, code string) (*model.Session, error) {
	var session model.Session
	err := r.db.GetContext(ctx, &session, `
		SELECT * FROM sessions
		WHERE pairing_code = $1
		AND status = 'pending_pairing'
		AND expires_at > NOW()
	`, code)
	return HandleNotFound(&session, err)
}

func (r *sessionRepo) Create(ctx context.Context, params model.CreateSessionParams) (*model.Session, error) {
	var session model.Session
	err := r.db.GetContext(ctx, &session, `
		INSERT INTO sessions (session_token_hash, pairing_code, expires_at, metadata)
		VALUES ($1, $2, $3, $4)
		RETURNING *
	`, params.SessionTokenHash, params.PairingCode, params.ExpiresAt, params.Metadata)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepo) MarkPaired(ctx context.Context, id string, accountID string, conversationKey string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE sessions SET
			status = 'paired',
			account_id = $2,
			paired_conversation_key = $3,
			paired_at = $4,
			expires_at = '9999-12-31 23:59:59+00',
			updated_at = $4
		WHERE id = $1 AND status = 'pending_pairing'
	`, id, accountID, conversationKey, time.Now())
	return err
}

func (r *sessionRepo) MarkExpired(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE sessions SET
			status = 'expired',
			updated_at = $2
		WHERE id = $1
	`, id, time.Now())
	return err
}

func (r *sessionRepo) MarkDisconnected(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE sessions SET
			status = 'disconnected',
			updated_at = $2
		WHERE id = $1
	`, id, time.Now())
	return err
}

func (r *sessionRepo) DeleteExpired(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM sessions
		WHERE (status = 'pending_pairing' AND expires_at < NOW())
		OR status IN ('expired', 'disconnected')
	`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *sessionRepo) CountPendingByIP(ctx context.Context, ip string, since time.Time) (int, error) {
	// Note: This requires storing IP in metadata when creating session
	// For now, return 0 to not block (rate limiting can be added later)
	return 0, nil
}
