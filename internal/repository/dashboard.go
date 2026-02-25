package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type DashboardOverview struct {
	AccountCount         int `db:"account_count"`
	SessionPending       int `db:"session_pending"`
	SessionPaired        int `db:"session_paired"`
	SessionTotal         int `db:"session_total"`
	ConversationPaired   int `db:"conversation_paired"`
	ConversationUnpaired int `db:"conversation_unpaired"`
	InboundTotal         int `db:"inbound_total"`
	OutboundTotal        int `db:"outbound_total"`
	OutboundFailed       int `db:"outbound_failed"`
}

type DashboardRepository interface {
	GetOverviewStats(ctx context.Context) (*DashboardOverview, error)
}

type dashboardRepo struct {
	db *sqlx.DB
}

func NewDashboardRepository(db *sqlx.DB) DashboardRepository {
	return &dashboardRepo{db: db}
}

func (r *dashboardRepo) GetOverviewStats(ctx context.Context) (*DashboardOverview, error) {
	var stats DashboardOverview
	err := r.db.GetContext(ctx, &stats, `
		WITH
			acc AS (SELECT COUNT(*) AS cnt FROM accounts),
			sess_pending AS (SELECT COUNT(*) AS cnt FROM sessions WHERE status = 'pending_pairing' AND expires_at > NOW()),
			sess_paired AS (SELECT COUNT(*) AS cnt FROM sessions WHERE status = 'paired'),
			sess_total AS (SELECT COUNT(*) AS cnt FROM sessions),
			conv_paired AS (SELECT COUNT(*) AS cnt FROM conversation_mappings WHERE state = 'paired'),
			conv_unpaired AS (SELECT COUNT(*) AS cnt FROM conversation_mappings WHERE state != 'paired'),
			inbound AS (SELECT COUNT(*) AS cnt FROM inbound_messages),
			outbound AS (SELECT COUNT(*) AS cnt FROM outbound_messages),
			outbound_failed AS (SELECT COUNT(*) AS cnt FROM outbound_messages WHERE status = 'failed')
		SELECT
			acc.cnt AS account_count,
			sess_pending.cnt AS session_pending,
			sess_paired.cnt AS session_paired,
			sess_total.cnt AS session_total,
			conv_paired.cnt AS conversation_paired,
			conv_unpaired.cnt AS conversation_unpaired,
			inbound.cnt AS inbound_total,
			outbound.cnt AS outbound_total,
			outbound_failed.cnt AS outbound_failed
		FROM acc, sess_pending, sess_paired, sess_total,
			conv_paired, conv_unpaired, inbound, outbound, outbound_failed
	`)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}
