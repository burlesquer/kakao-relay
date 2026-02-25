package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"github.com/openclaw/relay-server-go/internal/database"
	"github.com/openclaw/relay-server-go/internal/model"
	"github.com/openclaw/relay-server-go/internal/repository"
	"github.com/openclaw/relay-server-go/internal/sse"
	"github.com/openclaw/relay-server-go/internal/util"
)

const (
	sessionPairingCodeChars  = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	sessionPairingExpiryMins = 5
)

type CreateSessionResult struct {
	SessionToken string    `json:"sessionToken"`
	PairingCode  string    `json:"pairingCode"`
	ExpiresIn    int       `json:"expiresIn"`
	Status       string    `json:"status"`
	ExpiresAt    time.Time `json:"-"`
}

type SessionStatusResult struct {
	Status      model.SessionStatus `json:"status"`
	PairedAt    *time.Time          `json:"pairedAt,omitempty"`
	KakaoUserID *string             `json:"kakaoUserId,omitempty"`
	AccountID   *string             `json:"accountId,omitempty"`
}

type SessionPairResult struct {
	Success   bool
	SessionID string
	AccountID string
	Error     string
}

type SessionService struct {
	db          *database.DB
	sessionRepo repository.SessionRepository
	accountRepo repository.AccountRepository
	broker      *sse.Broker
}

func NewSessionService(
	db *database.DB,
	sessionRepo repository.SessionRepository,
	accountRepo repository.AccountRepository,
	broker *sse.Broker,
) *SessionService {
	return &SessionService{
		db:          db,
		sessionRepo: sessionRepo,
		accountRepo: accountRepo,
		broker:      broker,
	}
}

func (s *SessionService) CreateSession(ctx context.Context) (*CreateSessionResult, error) {
	token, err := util.GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	tokenHash := util.HashToken(token)
	pairingCode := generateSessionPairingCode()
	expiresAt := time.Now().Add(sessionPairingExpiryMins * time.Minute)

	session, err := s.sessionRepo.Create(ctx, model.CreateSessionParams{
		SessionTokenHash: tokenHash,
		PairingCode:      pairingCode,
		ExpiresAt:        expiresAt,
	})
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	log.Info().
		Str("sessionId", session.ID).
		Str("pairingCode", util.MaskCode(pairingCode)).
		Time("expiresAt", expiresAt).
		Msg("session created")

	return &CreateSessionResult{
		SessionToken: token,
		PairingCode:  pairingCode,
		ExpiresIn:    sessionPairingExpiryMins * 60,
		Status:       string(model.SessionStatusPendingPairing),
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *SessionService) GetStatus(ctx context.Context, tokenHash string) (*SessionStatusResult, error) {
	session, err := s.sessionRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("find session: %w", err)
	}

	if session == nil {
		return nil, nil
	}

	// Check if pending session has expired
	if session.Status == model.SessionStatusPendingPairing && time.Now().After(session.ExpiresAt) {
		s.sessionRepo.MarkExpired(ctx, session.ID)
		return &SessionStatusResult{
			Status: model.SessionStatusExpired,
		}, nil
	}

	result := &SessionStatusResult{
		Status:    session.Status,
		PairedAt:  session.PairedAt,
		AccountID: session.AccountID,
	}

	// Extract kakaoUserId from conversation key if available
	if session.PairedConversationKey != nil {
		parts := strings.Split(*session.PairedConversationKey, ":")
		if len(parts) >= 2 {
			result.KakaoUserID = &parts[1]
		}
	}

	return result, nil
}

func (s *SessionService) FindByTokenHash(ctx context.Context, tokenHash string) (*model.Session, error) {
	return s.sessionRepo.FindByTokenHash(ctx, tokenHash)
}

func (s *SessionService) FindByID(ctx context.Context, id string) (*model.Session, error) {
	return s.sessionRepo.FindByID(ctx, id)
}

func (s *SessionService) VerifyPairingCode(ctx context.Context, code, conversationKey string) SessionPairResult {
	normalizedCode := strings.ToUpper(strings.TrimSpace(code))

	// First, find the session (outside transaction for quick validation)
	session, err := s.sessionRepo.FindByPairingCode(ctx, normalizedCode)
	if err != nil {
		log.Error().Err(err).Msg("verify pairing code: database error")
		return SessionPairResult{Success: false, Error: "INVALID_CODE"}
	}

	if session == nil {
		log.Warn().Str("code", util.MaskCode(normalizedCode)).Msg("invalid session pairing code")
		return SessionPairResult{Success: false, Error: "INVALID_CODE"}
	}

	var account *model.Account

	// Use transaction to ensure atomicity of account creation + session pairing
	err = s.db.WithTx(ctx, func(tx *sqlx.Tx) error {
		txAccountRepo := s.accountRepo.WithTx(tx)
		txSessionRepo := s.sessionRepo.WithTx(tx)

		// Create account for this session within transaction
		var createErr error
		account, createErr = s.createAccountForSessionTx(ctx, txAccountRepo, session.ID)
		if createErr != nil {
			return fmt.Errorf("create account: %w", createErr)
		}

		// Mark session as paired within the same transaction
		if markErr := txSessionRepo.MarkPaired(ctx, session.ID, account.ID, conversationKey); markErr != nil {
			return fmt.Errorf("mark paired: %w", markErr)
		}

		return nil
	})

	if err != nil {
		log.Error().Err(err).Msg("verify pairing code: transaction failed")
		return SessionPairResult{Success: false, Error: "INTERNAL_ERROR"}
	}

	log.Info().
		Str("sessionId", session.ID).
		Str("accountId", account.ID).
		Str("conversationKey", conversationKey).
		Msg("session paired successfully")

	return SessionPairResult{
		Success:   true,
		SessionID: session.ID,
		AccountID: account.ID,
	}
}

func (s *SessionService) createAccountForSession(ctx context.Context, sessionID string) (*model.Account, error) {
	return s.createAccountForSessionTx(ctx, s.accountRepo, sessionID)
}

func (s *SessionService) createAccountForSessionTx(ctx context.Context, accountRepo repository.AccountRepository, sessionID string) (*model.Account, error) {
	token, err := util.GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	tokenHash := util.HashToken(token)

	account, err := accountRepo.Create(ctx, model.CreateAccountParams{
		RelayTokenHash:  tokenHash,
		Mode:            model.AccountModeRelay,
		RateLimitPerMin: 60,
	})
	if err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}

	return account, nil
}

func (s *SessionService) PublishPairingComplete(ctx context.Context, session *model.Session, conversationKey string) error {
	if session.AccountID == nil {
		return fmt.Errorf("session not paired")
	}

	// Extract kakaoUserId from conversation key
	var kakaoUserID string
	parts := strings.Split(conversationKey, ":")
	if len(parts) >= 2 {
		kakaoUserID = parts[1]
	}

	eventDataBytes, err := json.Marshal(map[string]string{
		"kakaoUserId": kakaoUserID,
		"pairedAt":    time.Now().Format(time.RFC3339),
		"accountId":   *session.AccountID,
	})
	if err != nil {
		return fmt.Errorf("marshal pairing complete event: %w", err)
	}

	event := sse.Event{
		Type: "pairing_complete",
		Data: eventDataBytes,
	}

	// Publish to session channel (for pending SSE connections)
	sessionChannel := "session:" + session.ID
	if err := s.broker.Publish(ctx, sessionChannel, event); err != nil {
		log.Warn().Err(err).Str("sessionId", session.ID).Msg("failed to publish to session channel")
	}

	// Also publish to account channel (for any existing account connections)
	return s.broker.Publish(ctx, *session.AccountID, event)
}

func (s *SessionService) PublishPairingExpired(ctx context.Context, sessionID string, reason string) error {
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil || session == nil {
		return fmt.Errorf("session not found")
	}

	if session.AccountID == nil {
		return nil // No account to notify
	}

	eventDataBytes, err := json.Marshal(map[string]string{
		"reason": reason,
	})
	if err != nil {
		return fmt.Errorf("marshal pairing expired event: %w", err)
	}

	return s.broker.Publish(ctx, *session.AccountID, sse.Event{
		Type: "pairing_expired",
		Data: eventDataBytes,
	})
}

func generateSessionPairingCode() string {
	chars := []byte(sessionPairingCodeChars)
	part1 := make([]byte, 4)
	part2 := make([]byte, 4)

	for i := 0; i < 4; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		part1[i] = chars[n.Int64()]
	}
	for i := 0; i < 4; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		part2[i] = chars[n.Int64()]
	}

	return fmt.Sprintf("%s-%s", string(part1), string(part2))
}
