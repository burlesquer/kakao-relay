package repository

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/openclaw/relay-server-go/internal/model"
)

type AdminSessionRepository interface {
	FindByTokenHash(ctx context.Context, tokenHash string) (*model.AdminSession, error)
	Create(ctx context.Context, params model.CreateAdminSessionParams) (*model.AdminSession, error)
	Delete(ctx context.Context, id string) error
	DeleteByTokenHash(ctx context.Context, tokenHash string) error
	DeleteExpired(ctx context.Context) (int64, error)
}

type adminSessionRepo struct {
	db *sqlx.DB
}

func NewAdminSessionRepository(db *sqlx.DB) AdminSessionRepository {
	return &adminSessionRepo{db: db}
}

func (r *adminSessionRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*model.AdminSession, error) {
	var session model.AdminSession
	err := r.db.GetContext(ctx, &session, `
		SELECT * FROM admin_sessions
		WHERE token_hash = $1 AND expires_at > NOW()
	`, tokenHash)
	return HandleNotFound(&session, err)
}

func (r *adminSessionRepo) Create(ctx context.Context, params model.CreateAdminSessionParams) (*model.AdminSession, error) {
	var session model.AdminSession
	err := r.db.GetContext(ctx, &session, `
		INSERT INTO admin_sessions (token_hash, expires_at)
		VALUES ($1, $2)
		RETURNING *
	`, params.TokenHash, params.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *adminSessionRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM admin_sessions WHERE id = $1`, id)
	return err
}

func (r *adminSessionRepo) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM admin_sessions WHERE token_hash = $1`, tokenHash)
	return err
}

func (r *adminSessionRepo) DeleteExpired(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx, `DELETE FROM admin_sessions WHERE expires_at < NOW()`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
