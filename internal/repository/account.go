package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/openclaw/relay-server-go/internal/model"
)

type AccountRepository interface {
	FindByID(ctx context.Context, id string) (*model.Account, error)
	FindByTokenHash(ctx context.Context, tokenHash string) (*model.Account, error)
	FindAll(ctx context.Context, limit, offset int) ([]model.Account, error)
	Create(ctx context.Context, params model.CreateAccountParams) (*model.Account, error)
	Update(ctx context.Context, id string, params model.UpdateAccountParams) (*model.Account, error)
	UpdateToken(ctx context.Context, id, tokenHash string) (*model.Account, error)
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int, error)
	// WithTx returns a new repository that uses the given transaction
	WithTx(tx *sqlx.Tx) AccountRepository
}

type accountRepo struct {
	db sqlxDB
}

// sqlxDB is an interface satisfied by both *sqlx.DB and *sqlx.Tx
type sqlxDB interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func NewAccountRepository(db *sqlx.DB) AccountRepository {
	return &accountRepo{db: db}
}

func (r *accountRepo) WithTx(tx *sqlx.Tx) AccountRepository {
	return &accountRepo{db: tx}
}

func (r *accountRepo) FindByID(ctx context.Context, id string) (*model.Account, error) {
	var account model.Account
	err := r.db.GetContext(ctx, &account, `
		SELECT * FROM accounts WHERE id = $1
	`, id)
	return HandleNotFound(&account, err)
}

func (r *accountRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*model.Account, error) {
	var account model.Account
	err := r.db.GetContext(ctx, &account, `
		SELECT * FROM accounts
		WHERE relay_token_hash = $1 AND disabled_at IS NULL
	`, tokenHash)
	return HandleNotFound(&account, err)
}

func (r *accountRepo) FindAll(ctx context.Context, limit, offset int) ([]model.Account, error) {
	var accounts []model.Account
	err := r.db.SelectContext(ctx, &accounts, `
		SELECT * FROM accounts
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *accountRepo) Create(ctx context.Context, params model.CreateAccountParams) (*model.Account, error) {
	var account model.Account
	err := r.db.GetContext(ctx, &account, `
		INSERT INTO accounts (openclaw_user_id, relay_token_hash, mode, rate_limit_per_minute)
		VALUES ($1, $2, $3, $4)
		RETURNING *
	`, params.OpenclawUserID, params.RelayTokenHash, params.Mode, params.RateLimitPerMin)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *accountRepo) Update(ctx context.Context, id string, params model.UpdateAccountParams) (*model.Account, error) {
	var account model.Account
	err := r.db.GetContext(ctx, &account, `
		UPDATE accounts SET
			openclaw_user_id = COALESCE($2, openclaw_user_id),
			mode = COALESCE($3, mode),
			rate_limit_per_minute = COALESCE($4, rate_limit_per_minute),
			disabled_at = $5,
			updated_at = $6
		WHERE id = $1
		RETURNING *
	`, id, params.OpenclawUserID, params.Mode, params.RateLimitPerMin, params.DisabledAt, time.Now())
	return HandleNotFound(&account, err)
}

func (r *accountRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM accounts WHERE id = $1`, id)
	return err
}

func (r *accountRepo) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM accounts`)
	return count, err
}

func (r *accountRepo) UpdateToken(ctx context.Context, id, tokenHash string) (*model.Account, error) {
	var account model.Account
	err := r.db.GetContext(ctx, &account, `
		UPDATE accounts SET
			relay_token_hash = $2,
			updated_at = $3
		WHERE id = $1
		RETURNING *
	`, id, tokenHash, time.Now())
	return HandleNotFound(&account, err)
}
