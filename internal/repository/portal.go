package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/openclaw/relay-server-go/internal/model"
)

type PortalUserRepository interface {
	FindByID(ctx context.Context, id string) (*model.PortalUser, error)
	FindByEmail(ctx context.Context, email string) (*model.PortalUser, error)
	Create(ctx context.Context, params model.CreatePortalUserParams) (*model.PortalUser, error)
	UpdateLastLogin(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
}

type portalUserRepo struct {
	db *sqlx.DB
}

func NewPortalUserRepository(db *sqlx.DB) PortalUserRepository {
	return &portalUserRepo{db: db}
}

func (r *portalUserRepo) FindByID(ctx context.Context, id string) (*model.PortalUser, error) {
	var user model.PortalUser
	err := r.db.GetContext(ctx, &user, `SELECT * FROM portal_users WHERE id = $1`, id)
	return HandleNotFound(&user, err)
}

func (r *portalUserRepo) FindByEmail(ctx context.Context, email string) (*model.PortalUser, error) {
	var user model.PortalUser
	err := r.db.GetContext(ctx, &user, `SELECT * FROM portal_users WHERE email = $1`, email)
	return HandleNotFound(&user, err)
}

func (r *portalUserRepo) Create(ctx context.Context, params model.CreatePortalUserParams) (*model.PortalUser, error) {
	var user model.PortalUser
	err := r.db.GetContext(ctx, &user, `
		INSERT INTO portal_users (email, account_id)
		VALUES ($1, $2)
		RETURNING *
	`, params.Email, params.AccountID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *portalUserRepo) UpdateLastLogin(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE portal_users SET last_login_at = $2 WHERE id = $1
	`, id, time.Now())
	return err
}

func (r *portalUserRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM portal_users WHERE id = $1`, id)
	return err
}

// Portal Session Repository

type PortalSessionRepository interface {
	FindByTokenHash(ctx context.Context, tokenHash string) (*model.PortalSession, error)
	Create(ctx context.Context, params model.CreatePortalSessionParams) (*model.PortalSession, error)
	Delete(ctx context.Context, id string) error
	DeleteByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) (int64, error)
}

type portalSessionRepo struct {
	db *sqlx.DB
}

func NewPortalSessionRepository(db *sqlx.DB) PortalSessionRepository {
	return &portalSessionRepo{db: db}
}

func (r *portalSessionRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*model.PortalSession, error) {
	var session model.PortalSession
	err := r.db.GetContext(ctx, &session, `
		SELECT * FROM portal_sessions
		WHERE token_hash = $1 AND expires_at > NOW()
	`, tokenHash)
	return HandleNotFound(&session, err)
}

func (r *portalSessionRepo) Create(ctx context.Context, params model.CreatePortalSessionParams) (*model.PortalSession, error) {
	var session model.PortalSession
	err := r.db.GetContext(ctx, &session, `
		INSERT INTO portal_sessions (token_hash, user_id, expires_at)
		VALUES ($1, $2, $3)
		RETURNING *
	`, params.TokenHash, params.UserID, params.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *portalSessionRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM portal_sessions WHERE id = $1`, id)
	return err
}

func (r *portalSessionRepo) DeleteByUserID(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM portal_sessions WHERE user_id = $1`, userID)
	return err
}

func (r *portalSessionRepo) DeleteExpired(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx, `DELETE FROM portal_sessions WHERE expires_at < NOW()`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
