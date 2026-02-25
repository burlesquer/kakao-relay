package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/openclaw/relay-server-go/internal/model"
	"github.com/openclaw/relay-server-go/internal/repository"
	redisclient "github.com/openclaw/relay-server-go/internal/redis"
	"github.com/openclaw/relay-server-go/internal/util"
)

const (
	portalCodeChars      = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	portalCodeTTLMinutes = 30
	sessionTTLMinutes    = 30
)

// PortalCodeSession represents a temporary portal session
type PortalCodeSession struct {
	Token           string
	ConversationKey string
	ExpiresAt       time.Time
}

// PortalAccessService handles portal access code operations
type PortalAccessService struct {
	codeRepo    repository.PortalAccessCodeRepository
	convRepo    repository.ConversationRepository
	redisClient *redisclient.Client
	rateLimiter *RateLimiter
}

// NewPortalAccessService creates a new portal access service
func NewPortalAccessService(
	codeRepo repository.PortalAccessCodeRepository,
	convRepo repository.ConversationRepository,
	redisClient *redisclient.Client,
) *PortalAccessService {
	return &PortalAccessService{
		codeRepo:    codeRepo,
		convRepo:    convRepo,
		redisClient: redisClient,
		rateLimiter: NewRateLimiter(redisClient.Client),
	}
}

// GenerateCode generates a new portal access code or returns existing valid one
func (s *PortalAccessService) GenerateCode(
	ctx context.Context,
	conversationKey string,
) (*model.PortalAccessCode, error) {
	// Check if there's already an active code for this conversation
	existing, err := s.codeRepo.FindActiveByConversationKey(ctx, conversationKey)
	if err == nil && existing != nil {
		log.Info().
			Str("code", util.MaskCode(existing.Code)).
			Str("conversationKey", conversationKey).
			Time("expiresAt", existing.ExpiresAt).
			Msg("reusing existing portal access code")
		return existing, nil
	}

	// Generate new code
	var code string
	for attempts := 0; attempts < 10; attempts++ {
		code = generatePortalCode()
		existingCode, _ := s.codeRepo.FindActiveByCode(ctx, code)
		if existingCode == nil {
			break
		}
	}

	expiresAt := time.Now().Add(portalCodeTTLMinutes * time.Minute)
	pac, err := s.codeRepo.Create(ctx, model.CreatePortalAccessCodeParams{
		Code:            code,
		ConversationKey: conversationKey,
		ExpiresAt:       expiresAt,
	})
	if err != nil {
		return nil, fmt.Errorf("create portal access code: %w", err)
	}

	log.Info().
		Str("code", util.MaskCode(code)).
		Str("conversationKey", conversationKey).
		Time("expiresAt", pac.ExpiresAt).
		Msg("portal access code created")

	return pac, nil
}

// VerifyCode verifies a portal access code and returns the conversation key
func (s *PortalAccessService) VerifyCode(
	ctx context.Context,
	code string,
) (string, error) {
	normalizedCode := strings.ToUpper(strings.TrimSpace(code))

	pac, err := s.codeRepo.FindActiveByCode(ctx, normalizedCode)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warn().Str("code", util.MaskCode(normalizedCode)).Msg("invalid or expired portal code")
			return "", fmt.Errorf("invalid or expired code")
		}
		return "", fmt.Errorf("verify portal code: %w", err)
	}

	// Mark code as used
	if err := s.codeRepo.MarkUsed(ctx, normalizedCode); err != nil {
		log.Error().Err(err).Msg("mark portal code as used")
		return "", fmt.Errorf("verify portal code: %w", err)
	}

	log.Info().
		Str("code", util.MaskCode(normalizedCode)).
		Str("conversationKey", pac.ConversationKey).
		Msg("portal code verified")

	return pac.ConversationKey, nil
}

// CreateCodeSession creates a temporary session for portal access
func (s *PortalAccessService) CreateCodeSession(
	conversationKey string,
) (*PortalCodeSession, error) {
	// Generate secure random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("generate session token: %w", err)
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	session := &PortalCodeSession{
		Token:           token,
		ConversationKey: conversationKey,
		ExpiresAt:       time.Now().Add(sessionTTLMinutes * time.Minute),
	}

	log.Info().
		Str("conversationKey", conversationKey).
		Time("expiresAt", session.ExpiresAt).
		Msg("portal code session created")

	return session, nil
}

// ValidateCodeSession validates a portal session token using Redis
func (s *PortalAccessService) ValidateCodeSession(
	ctx context.Context,
	token string,
) (string, error) {
	key := fmt.Sprintf("portal_session:%s", token)
	data, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("session expired or invalid")
		}
		return "", fmt.Errorf("validate session: %w", err)
	}

	var session PortalCodeSession
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return "", fmt.Errorf("unmarshal session: %w", err)
	}

	return session.ConversationKey, nil
}

// StoreSession stores a session in Redis with automatic expiry
func (s *PortalAccessService) StoreSession(ctx context.Context, session *PortalCodeSession) error {
	key := fmt.Sprintf("portal_session:%s", session.Token)
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return fmt.Errorf("session already expired")
	}

	if err := s.redisClient.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("store session: %w", err)
	}

	log.Debug().
		Str("token", session.Token[:16]+"...").
		Str("conversationKey", session.ConversationKey).
		Dur("ttl", ttl).
		Msg("session stored in redis")

	return nil
}

// CheckCodeGenerationLimit checks if code generation is allowed for a conversation
// Limit: 3 times per 5 minutes per conversationKey
func (s *PortalAccessService) CheckCodeGenerationLimit(
	ctx context.Context,
	conversationKey string,
) (allowed bool, resetAt time.Time) {
	key := fmt.Sprintf("code_gen:%s", conversationKey)
	return s.rateLimiter.CheckLimit(ctx, key, 3, 5*time.Minute)
}

// CheckLoginLimit checks if login attempts are allowed for an IP
// Limit: 5 times per 1 minute per IP
func (s *PortalAccessService) CheckLoginLimit(
	ctx context.Context,
	ip string,
) (allowed bool, resetAt time.Time) {
	key := fmt.Sprintf("code_login:%s", ip)
	return s.rateLimiter.CheckLimit(ctx, key, 5, 1*time.Minute)
}

// generatePortalCode generates an 8-character code in XXXX-XXXX format
func generatePortalCode() string {
	chars := []byte(portalCodeChars)
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
