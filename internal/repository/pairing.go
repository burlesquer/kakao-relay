package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/openclaw/relay-server-go/internal/model"
)

type PairingCodeRepository interface {
	FindByCode(ctx context.Context, code string) (*model.PairingCode, error)
	FindActiveByAccountID(ctx context.Context, accountID string) ([]model.PairingCode, error)
	CountActiveByAccountID(ctx context.Context, accountID string) (int, error)
	Create(ctx context.Context, params model.CreatePairingCodeParams) (*model.PairingCode, error)
	MarkUsed(ctx context.Context, code string, usedBy string) error
	DeleteExpired(ctx context.Context) (int64, error)
}

type pairingCodeRepo struct {
	db *sqlx.DB
}

func NewPairingCodeRepository(db *sqlx.DB) PairingCodeRepository {
	return &pairingCodeRepo{db: db}
}

func (r *pairingCodeRepo) FindByCode(ctx context.Context, code string) (*model.PairingCode, error) {
	var pc model.PairingCode
	err := r.db.GetContext(ctx, &pc, `
		SELECT * FROM pairing_codes
		WHERE code = $1 AND used_at IS NULL AND expires_at > NOW()
	`, code)
	return HandleNotFound(&pc, err)
}

func (r *pairingCodeRepo) FindActiveByAccountID(ctx context.Context, accountID string) ([]model.PairingCode, error) {
	var codes []model.PairingCode
	err := r.db.SelectContext(ctx, &codes, `
		SELECT * FROM pairing_codes
		WHERE account_id = $1 AND used_at IS NULL AND expires_at > NOW()
		ORDER BY created_at DESC
	`, accountID)
	return codes, err
}

func (r *pairingCodeRepo) CountActiveByAccountID(ctx context.Context, accountID string) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM pairing_codes
		WHERE account_id = $1 AND used_at IS NULL AND expires_at > NOW()
	`, accountID)
	return count, err
}

func (r *pairingCodeRepo) Create(ctx context.Context, params model.CreatePairingCodeParams) (*model.PairingCode, error) {
	var pc model.PairingCode
	err := r.db.GetContext(ctx, &pc, `
		INSERT INTO pairing_codes (code, account_id, expires_at, metadata)
		VALUES ($1, $2, $3, $4)
		RETURNING *
	`, params.Code, params.AccountID, params.ExpiresAt, params.Metadata)
	if err != nil {
		return nil, err
	}
	return &pc, nil
}

func (r *pairingCodeRepo) MarkUsed(ctx context.Context, code string, usedBy string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE pairing_codes SET
			used_at = $2,
			used_by = $3
		WHERE code = $1
	`, code, time.Now(), usedBy)
	return err
}

func (r *pairingCodeRepo) DeleteExpired(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM pairing_codes
		WHERE expires_at < NOW() OR used_at IS NOT NULL
	`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
