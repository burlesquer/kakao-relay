package service

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/openclaw/relay-server-go/internal/model"
	"github.com/openclaw/relay-server-go/internal/repository"
	"github.com/openclaw/relay-server-go/internal/util"
)

type PortalService struct {
	userRepo      repository.PortalUserRepository
	sessionRepo   repository.PortalSessionRepository
	accountRepo   repository.AccountRepository
	sessionSecret string
}

func NewPortalService(
	userRepo repository.PortalUserRepository,
	sessionRepo repository.PortalSessionRepository,
	accountRepo repository.AccountRepository,
	sessionSecret string,
) *PortalService {
	return &PortalService{
		userRepo:      userRepo,
		sessionRepo:   sessionRepo,
		accountRepo:   accountRepo,
		sessionSecret: sessionSecret,
	}
}

func (s *PortalService) Logout(ctx context.Context, token string) error {
	tokenHash := util.HmacSHA256(s.sessionSecret, token)
	session, _ := s.sessionRepo.FindByTokenHash(ctx, tokenHash)
	if session != nil {
		return s.sessionRepo.Delete(ctx, session.ID)
	}
	return nil
}

func (s *PortalService) ValidateSession(ctx context.Context, token string) (*model.PortalUser, error) {
	tokenHash := util.HmacSHA256(s.sessionSecret, token)
	session, err := s.sessionRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil || session == nil {
		return nil, nil
	}

	return s.userRepo.FindByID(ctx, session.UserID)
}

func (s *PortalService) GetAccountByID(ctx context.Context, accountID string) (*model.Account, error) {
	return s.accountRepo.FindByID(ctx, accountID)
}

func (s *PortalService) RegenerateToken(ctx context.Context, accountID string) (*model.Account, string, error) {
	newToken, err := util.GenerateToken()
	if err != nil {
		return nil, "", err
	}

	tokenHash := util.HashToken(newToken)
	account, err := s.accountRepo.UpdateToken(ctx, accountID, tokenHash)
	if err != nil {
		return nil, "", err
	}

	log.Info().Str("accountId", accountID).Msg("relay token regenerated")

	return account, newToken, nil
}

func (s *PortalService) DeleteAccount(ctx context.Context, userID string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return err
	}

	if err := s.accountRepo.Delete(ctx, user.AccountID); err != nil {
		return err
	}

	log.Info().Str("userId", userID).Str("accountId", user.AccountID).Msg("portal user account deleted")

	return nil
}

func (s *PortalService) CreateSession(ctx context.Context, userID string) (string, error) {
	token, err := util.GenerateToken()
	if err != nil {
		return "", err
	}

	tokenHash := util.HmacSHA256(s.sessionSecret, token)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	_, err = s.sessionRepo.Create(ctx, model.CreatePortalSessionParams{
		TokenHash: tokenHash,
		UserID:    userID,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return "", err
	}

	return token, nil
}
