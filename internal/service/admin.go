package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"github.com/openclaw/relay-server-go/internal/model"
	"github.com/openclaw/relay-server-go/internal/repository"
	"github.com/openclaw/relay-server-go/internal/util"
)

type queryBuilder struct {
	conditions []string
	args       []interface{}
}

func newQueryBuilder() *queryBuilder {
	return &queryBuilder{
		conditions: make([]string, 0),
		args:       make([]interface{}, 0),
	}
}

func (qb *queryBuilder) addCondition(column string, value interface{}) {
	if value == nil {
		return
	}
	if s, ok := value.(string); ok && s == "" {
		return
	}
	qb.args = append(qb.args, value)
	qb.conditions = append(qb.conditions, fmt.Sprintf("%s = $%d", column, len(qb.args)))
}

func (qb *queryBuilder) buildSelect(table string, limit, offset int) (selectQuery, countQuery string, args []interface{}) {
	whereClause := ""
	if len(qb.conditions) > 0 {
		whereClause = " WHERE " + strings.Join(qb.conditions, " AND ")
	}

	countQuery = fmt.Sprintf("SELECT COUNT(*) FROM %s%s", table, whereClause)

	limitIdx := len(qb.args) + 1
	offsetIdx := len(qb.args) + 2
	selectQuery = fmt.Sprintf(
		"SELECT * FROM %s%s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		table, whereClause, limitIdx, offsetIdx,
	)

	args = append(qb.args, limit, offset)
	return selectQuery, countQuery, args
}

type AdminService struct {
	db                *sqlx.DB
	sessionRepo       repository.AdminSessionRepository
	accountRepo       repository.AccountRepository
	convRepo          repository.ConversationRepository
	inboundRepo       repository.InboundMessageRepository
	outboundRepo      repository.OutboundMessageRepository
	portalUserRepo    repository.PortalUserRepository
	pluginSessionRepo repository.SessionRepository
	adminPasswordHash string
	sessionSecret     string
}

func NewAdminService(
	db *sqlx.DB,
	sessionRepo repository.AdminSessionRepository,
	accountRepo repository.AccountRepository,
	convRepo repository.ConversationRepository,
	inboundRepo repository.InboundMessageRepository,
	outboundRepo repository.OutboundMessageRepository,
	portalUserRepo repository.PortalUserRepository,
	pluginSessionRepo repository.SessionRepository,
	adminPasswordHash, sessionSecret string,
) *AdminService {
	return &AdminService{
		db:                db,
		sessionRepo:       sessionRepo,
		accountRepo:       accountRepo,
		convRepo:          convRepo,
		inboundRepo:       inboundRepo,
		outboundRepo:      outboundRepo,
		portalUserRepo:    portalUserRepo,
		pluginSessionRepo: pluginSessionRepo,
		adminPasswordHash: adminPasswordHash,
		sessionSecret:     sessionSecret,
	}
}

func (s *AdminService) Login(ctx context.Context, password string) (string, error) {
	if !util.CheckPasswordHash(password, s.adminPasswordHash) {
		return "", nil
	}

	token, err := util.GenerateToken()
	if err != nil {
		return "", err
	}

	tokenHash := util.HmacSHA256(s.sessionSecret, token)
	expiresAt := time.Now().Add(24 * time.Hour)

	_, err = s.sessionRepo.Create(ctx, model.CreateAdminSessionParams{
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AdminService) Logout(ctx context.Context, token string) error {
	tokenHash := util.HmacSHA256(s.sessionSecret, token)
	return s.sessionRepo.DeleteByTokenHash(ctx, tokenHash)
}

func (s *AdminService) ValidateSession(ctx context.Context, token string) bool {
	tokenHash := util.HmacSHA256(s.sessionSecret, token)
	session, err := s.sessionRepo.FindByTokenHash(ctx, tokenHash)
	return err == nil && session != nil
}

type Stats struct {
	Accounts int `json:"accounts"`
	Mappings int `json:"mappings"`
	Sessions struct {
		Pending int `json:"pending"`
		Paired  int `json:"paired"`
		Total   int `json:"total"`
	} `json:"sessions"`
	Messages struct {
		Inbound struct {
			Today  int `json:"today"`
			Week   int `json:"week"`
			Queued int `json:"queued"`
		} `json:"inbound"`
		Outbound struct {
			Today  int `json:"today"`
			Week   int `json:"week"`
			Failed int `json:"failed"`
		} `json:"outbound"`
	} `json:"messages"`
}

func (s *AdminService) GetStats(ctx context.Context) (*Stats, error) {
	stats := &Stats{}

	accounts, err := s.accountRepo.Count(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("failed to get account count for stats")
	}
	stats.Accounts = accounts

	pairedCount, err := s.convRepo.CountByState(ctx, model.PairingStatePaired)
	if err != nil {
		log.Warn().Err(err).Msg("failed to get paired count for stats")
	}
	stats.Mappings = pairedCount

	queuedCount, err := s.inboundRepo.CountByStatus(ctx, model.InboundStatusQueued)
	if err != nil {
		log.Warn().Err(err).Msg("failed to get queued count for stats")
	}
	stats.Messages.Inbound.Queued = queuedCount

	// Session stats
	var sessionStats struct {
		Pending int `db:"pending"`
		Paired  int `db:"paired"`
		Total   int `db:"total"`
	}
	err = s.db.GetContext(ctx, &sessionStats, `
		SELECT
			COUNT(*) FILTER (WHERE status = 'pending_pairing') as pending,
			COUNT(*) FILTER (WHERE status = 'paired') as paired,
			COUNT(*) as total
		FROM sessions
	`)
	if err != nil {
		log.Warn().Err(err).Msg("failed to get session stats")
	}
	stats.Sessions.Pending = sessionStats.Pending
	stats.Sessions.Paired = sessionStats.Paired
	stats.Sessions.Total = sessionStats.Total

	return stats, nil
}

func (s *AdminService) CreateAccount(ctx context.Context, openclawUserID *string, mode model.AccountMode, rateLimit int) (*model.Account, string, error) {
	token, err := util.GenerateToken()
	if err != nil {
		return nil, "", err
	}

	tokenHash := util.HashToken(token)

	account, err := s.accountRepo.Create(ctx, model.CreateAccountParams{
		OpenclawUserID:  openclawUserID,
		RelayTokenHash:  tokenHash,
		Mode:            mode,
		RateLimitPerMin: rateLimit,
	})
	if err != nil {
		return nil, "", err
	}

	return account, token, nil
}

func (s *AdminService) RegenerateToken(ctx context.Context, accountID string) (string, error) {
	token, err := util.GenerateToken()
	if err != nil {
		return "", err
	}

	tokenHash := util.HashToken(token)

	_, err = s.db.ExecContext(ctx, `
		UPDATE accounts SET relay_token_hash = $2, updated_at = NOW()
		WHERE id = $1
	`, accountID, tokenHash)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AdminService) GetAccounts(ctx context.Context, limit, offset int) ([]model.Account, error) {
	return s.accountRepo.FindAll(ctx, limit, offset)
}

func (s *AdminService) GetAccountByID(ctx context.Context, id string) (*model.Account, error) {
	return s.accountRepo.FindByID(ctx, id)
}

func (s *AdminService) DeleteAccount(ctx context.Context, id string) error {
	return s.accountRepo.Delete(ctx, id)
}

// Mappings

func (s *AdminService) GetMappings(ctx context.Context, limit, offset int, accountID string) ([]model.ConversationMapping, int, error) {
	var mappings []model.ConversationMapping
	var total int
	var err error

	if accountID != "" {
		mappings, err = s.convRepo.FindByAccountID(ctx, accountID)
		total = len(mappings)
	} else {
		err = s.db.SelectContext(ctx, &mappings, `
			SELECT * FROM conversation_mappings
			ORDER BY first_seen_at DESC
			LIMIT $1 OFFSET $2
		`, limit, offset)
		if err == nil {
			if countErr := s.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM conversation_mappings`); countErr != nil {
				log.Warn().Err(countErr).Msg("failed to get mappings count")
			}
		}
	}

	if err != nil {
		return nil, 0, err
	}

	// Apply pagination for accountID case
	if accountID != "" && len(mappings) > 0 {
		start := offset
		end := offset + limit
		if start > len(mappings) {
			mappings = []model.ConversationMapping{}
		} else {
			if end > len(mappings) {
				end = len(mappings)
			}
			mappings = mappings[start:end]
		}
	}

	return mappings, total, nil
}

func (s *AdminService) DeleteMapping(ctx context.Context, id string) error {
	return s.convRepo.Delete(ctx, id)
}

// Messages

func (s *AdminService) GetInboundMessages(ctx context.Context, limit, offset int, accountID, status string) ([]model.InboundMessage, int, error) {
	var messages []model.InboundMessage
	var total int

	qb := newQueryBuilder()
	qb.addCondition("account_id", accountID)
	qb.addCondition("status", status)

	selectQuery, countQuery, args := qb.buildSelect("inbound_messages", limit, offset)

	if err := s.db.SelectContext(ctx, &messages, selectQuery, args...); err != nil {
		return nil, 0, err
	}

	countArgs := args[:len(args)-2]
	if len(countArgs) > 0 {
		if countErr := s.db.GetContext(ctx, &total, countQuery, countArgs...); countErr != nil {
			log.Warn().Err(countErr).Msg("failed to get inbound messages count")
		}
	} else {
		if countErr := s.db.GetContext(ctx, &total, countQuery); countErr != nil {
			log.Warn().Err(countErr).Msg("failed to get inbound messages count")
		}
	}

	return messages, total, nil
}

func (s *AdminService) GetOutboundMessages(ctx context.Context, limit, offset int, accountID, status string) ([]model.OutboundMessage, int, error) {
	var messages []model.OutboundMessage
	var total int

	qb := newQueryBuilder()
	qb.addCondition("account_id", accountID)
	qb.addCondition("status", status)

	selectQuery, countQuery, args := qb.buildSelect("outbound_messages", limit, offset)

	if err := s.db.SelectContext(ctx, &messages, selectQuery, args...); err != nil {
		return nil, 0, err
	}

	countArgs := args[:len(args)-2]
	if len(countArgs) > 0 {
		if countErr := s.db.GetContext(ctx, &total, countQuery, countArgs...); countErr != nil {
			log.Warn().Err(countErr).Msg("failed to get outbound messages count")
		}
	} else {
		if countErr := s.db.GetContext(ctx, &total, countQuery); countErr != nil {
			log.Warn().Err(countErr).Msg("failed to get outbound messages count")
		}
	}

	return messages, total, nil
}

// Users (Portal Users)

func (s *AdminService) GetUsers(ctx context.Context, limit, offset int) ([]model.PortalUser, int, error) {
	var users []model.PortalUser
	var total int

	err := s.db.SelectContext(ctx, &users, `
		SELECT * FROM portal_users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	if countErr := s.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM portal_users`); countErr != nil {
		log.Warn().Err(countErr).Msg("failed to get portal users count")
	}

	return users, total, nil
}

func (s *AdminService) GetUserByID(ctx context.Context, id string) (*model.PortalUser, error) {
	return s.portalUserRepo.FindByID(ctx, id)
}

func (s *AdminService) UpdateUser(ctx context.Context, id string, isActive *bool) (*model.PortalUser, error) {
	if isActive != nil {
		_, err := s.db.ExecContext(ctx, `
			UPDATE portal_users SET is_active = $2 WHERE id = $1
		`, id, *isActive)
		if err != nil {
			return nil, err
		}
	}
	return s.portalUserRepo.FindByID(ctx, id)
}

func (s *AdminService) DeleteUser(ctx context.Context, id string) error {
	return s.portalUserRepo.Delete(ctx, id)
}

// Sessions (Plugin Sessions)

func (s *AdminService) GetSessions(ctx context.Context, limit, offset int, status string) ([]model.Session, int, error) {
	var sessions []model.Session
	var total int

	qb := newQueryBuilder()
	qb.addCondition("status", status)

	selectQuery, countQuery, args := qb.buildSelect("sessions", limit, offset)

	if err := s.db.SelectContext(ctx, &sessions, selectQuery, args...); err != nil {
		return nil, 0, err
	}

	countArgs := args[:len(args)-2]
	if len(countArgs) > 0 {
		if countErr := s.db.GetContext(ctx, &total, countQuery, countArgs...); countErr != nil {
			log.Warn().Err(countErr).Msg("failed to get sessions count")
		}
	} else {
		if countErr := s.db.GetContext(ctx, &total, countQuery); countErr != nil {
			log.Warn().Err(countErr).Msg("failed to get sessions count")
		}
	}

	return sessions, total, nil
}

func (s *AdminService) DeleteSession(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	return err
}

func (s *AdminService) DisconnectSession(ctx context.Context, id string) error {
	return s.pluginSessionRepo.MarkDisconnected(ctx, id)
}
