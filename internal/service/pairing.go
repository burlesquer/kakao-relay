package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/openclaw/relay-server-go/internal/model"
	"github.com/openclaw/relay-server-go/internal/repository"
)

const (
	pairingCodeChars       = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	maxActiveCodesPerAcct  = 5
	defaultExpirySeconds   = 600
	maxExpirySeconds       = 1800
)

type VerifyResult struct {
	Success   bool
	AccountID string
	Error     string // INVALID_CODE, EXPIRED_CODE, ALREADY_USED
}

type PairingService struct {
	codeRepo repository.PairingCodeRepository
	convRepo repository.ConversationRepository
}

func NewPairingService(
	codeRepo repository.PairingCodeRepository,
	convRepo repository.ConversationRepository,
) *PairingService {
	return &PairingService{
		codeRepo: codeRepo,
		convRepo: convRepo,
	}
}

func (s *PairingService) GenerateCode(
	ctx context.Context,
	accountID string,
	expiresInSeconds int,
	metadata map[string]any,
) (*model.PairingCode, error) {
	if expiresInSeconds <= 0 {
		expiresInSeconds = defaultExpirySeconds
	}
	if expiresInSeconds > maxExpirySeconds {
		expiresInSeconds = maxExpirySeconds
	}

	activeCount, err := s.codeRepo.CountActiveByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("count active codes: %w", err)
	}

	if activeCount >= maxActiveCodesPerAcct {
		return nil, fmt.Errorf("maximum active codes (%d) reached", maxActiveCodesPerAcct)
	}

	var code string
	for attempts := 0; attempts < 10; attempts++ {
		code = generateRandomCode()
		existing, _ := s.codeRepo.FindByCode(ctx, code)
		if existing == nil {
			break
		}
	}

	var metadataJSON *json.RawMessage
	if metadata != nil {
		data, _ := json.Marshal(metadata)
		raw := json.RawMessage(data)
		metadataJSON = &raw
	}

	pc, err := s.codeRepo.Create(ctx, model.CreatePairingCodeParams{
		Code:      code,
		AccountID: accountID,
		ExpiresAt: time.Now().Add(time.Duration(expiresInSeconds) * time.Second),
		Metadata:  metadataJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("create pairing code: %w", err)
	}

	log.Info().
		Str("code", code).
		Str("accountId", accountID).
		Time("expiresAt", pc.ExpiresAt).
		Msg("pairing code created")

	return pc, nil
}

func (s *PairingService) VerifyCode(ctx context.Context, code, conversationKey string) VerifyResult {
	normalizedCode := strings.ToUpper(strings.TrimSpace(code))

	pc, err := s.codeRepo.FindByCode(ctx, normalizedCode)
	if err != nil {
		log.Error().Err(err).Msg("verify code: database error")
		return VerifyResult{Success: false, Error: "INVALID_CODE"}
	}

	if pc == nil {
		log.Warn().Str("code", normalizedCode).Msg("invalid pairing code")
		return VerifyResult{Success: false, Error: "INVALID_CODE"}
	}

	if err := s.codeRepo.MarkUsed(ctx, normalizedCode, conversationKey); err != nil {
		log.Error().Err(err).Msg("verify code: mark used")
		return VerifyResult{Success: false, Error: "INVALID_CODE"}
	}

	if err := s.convRepo.UpdateState(ctx, conversationKey, model.PairingStatePaired, &pc.AccountID); err != nil {
		log.Error().Err(err).Msg("verify code: update conversation state")
		return VerifyResult{Success: false, Error: "INVALID_CODE"}
	}

	log.Info().
		Str("code", normalizedCode).
		Str("accountId", pc.AccountID).
		Str("conversationKey", conversationKey).
		Msg("pairing successful")

	return VerifyResult{Success: true, AccountID: pc.AccountID}
}

func (s *PairingService) Unpair(ctx context.Context, conversationKey string) error {
	if err := s.convRepo.UpdateState(ctx, conversationKey, model.PairingStateUnpaired, nil); err != nil {
		return fmt.Errorf("unpair: %w", err)
	}

	log.Info().Str("conversationKey", conversationKey).Msg("conversation unpaired")
	return nil
}

func (s *PairingService) ListActiveCodes(ctx context.Context, accountID string) ([]model.PairingCode, error) {
	return s.codeRepo.FindActiveByAccountID(ctx, accountID)
}

func generateRandomCode() string {
	chars := []byte(pairingCodeChars)
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
