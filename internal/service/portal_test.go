package service

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/openclaw/relay-server-go/internal/model"
	"github.com/openclaw/relay-server-go/internal/repository"
)

type mockPortalUserRepo struct {
	users map[string]*model.PortalUser
}

func newMockPortalUserRepo() *mockPortalUserRepo {
	return &mockPortalUserRepo{users: make(map[string]*model.PortalUser)}
}

func (m *mockPortalUserRepo) FindByID(ctx context.Context, id string) (*model.PortalUser, error) {
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, nil
}

func (m *mockPortalUserRepo) FindByEmail(ctx context.Context, email string) (*model.PortalUser, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, nil
}

func (m *mockPortalUserRepo) Create(ctx context.Context, params model.CreatePortalUserParams) (*model.PortalUser, error) {
	user := &model.PortalUser{
		ID:        "user-123",
		Email:     params.Email,
		AccountID: params.AccountID,
		CreatedAt: time.Now(),
	}
	m.users[user.ID] = user
	return user, nil
}

func (m *mockPortalUserRepo) UpdateLastLogin(ctx context.Context, id string) error {
	if user, ok := m.users[id]; ok {
		now := time.Now()
		user.LastLoginAt = &now
	}
	return nil
}

func (m *mockPortalUserRepo) Delete(ctx context.Context, id string) error {
	delete(m.users, id)
	return nil
}

type mockPortalSessionRepo struct {
	sessions map[string]*model.PortalSession
}

func newMockPortalSessionRepo() *mockPortalSessionRepo {
	return &mockPortalSessionRepo{sessions: make(map[string]*model.PortalSession)}
}

func (m *mockPortalSessionRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*model.PortalSession, error) {
	for _, session := range m.sessions {
		if session.TokenHash == tokenHash {
			if session.ExpiresAt.After(time.Now()) {
				return session, nil
			}
		}
	}
	return nil, nil
}

func (m *mockPortalSessionRepo) Create(ctx context.Context, params model.CreatePortalSessionParams) (*model.PortalSession, error) {
	session := &model.PortalSession{
		ID:        "session-123",
		TokenHash: params.TokenHash,
		UserID:    params.UserID,
		ExpiresAt: params.ExpiresAt,
		CreatedAt: time.Now(),
	}
	m.sessions[session.ID] = session
	return session, nil
}

func (m *mockPortalSessionRepo) Delete(ctx context.Context, id string) error {
	delete(m.sessions, id)
	return nil
}

func (m *mockPortalSessionRepo) DeleteByUserID(ctx context.Context, userID string) error {
	for id, session := range m.sessions {
		if session.UserID == userID {
			delete(m.sessions, id)
		}
	}
	return nil
}

func (m *mockPortalSessionRepo) DeleteExpired(ctx context.Context) (int64, error) {
	count := int64(0)
	for id, session := range m.sessions {
		if session.ExpiresAt.Before(time.Now()) {
			delete(m.sessions, id)
			count++
		}
	}
	return count, nil
}

type mockAccountRepo struct {
	accounts map[string]*model.Account
}

func newMockAccountRepo() *mockAccountRepo {
	return &mockAccountRepo{accounts: make(map[string]*model.Account)}
}

func (m *mockAccountRepo) FindByID(ctx context.Context, id string) (*model.Account, error) {
	if acc, ok := m.accounts[id]; ok {
		return acc, nil
	}
	return nil, nil
}

func (m *mockAccountRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*model.Account, error) {
	return nil, nil
}

func (m *mockAccountRepo) Create(ctx context.Context, params model.CreateAccountParams) (*model.Account, error) {
	return nil, nil
}

func (m *mockAccountRepo) Update(ctx context.Context, id string, params model.UpdateAccountParams) (*model.Account, error) {
	return nil, nil
}

func (m *mockAccountRepo) UpdateToken(ctx context.Context, id, tokenHash string) (*model.Account, error) {
	return nil, nil
}

func (m *mockAccountRepo) Delete(ctx context.Context, id string) error {
	delete(m.accounts, id)
	return nil
}

func (m *mockAccountRepo) FindAll(ctx context.Context, limit, offset int) ([]model.Account, error) {
	return nil, nil
}

func (m *mockAccountRepo) Count(ctx context.Context) (int, error) {
	return len(m.accounts), nil
}

func (m *mockAccountRepo) WithTx(tx *sqlx.Tx) repository.AccountRepository {
	return m
}

func TestPortalService(t *testing.T) {
	t.Run("CreateSession creates valid session", func(t *testing.T) {
		userRepo := newMockPortalUserRepo()
		sessionRepo := newMockPortalSessionRepo()
		accountRepo := newMockAccountRepo()

		svc := NewPortalService(userRepo, sessionRepo, accountRepo, "test-secret")

		token, err := svc.CreateSession(context.Background(), "user-123")

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Len(t, sessionRepo.sessions, 1)
	})

	t.Run("ValidateSession returns user for valid session", func(t *testing.T) {
		userRepo := newMockPortalUserRepo()
		sessionRepo := newMockPortalSessionRepo()
		accountRepo := newMockAccountRepo()

		// Add a user
		userRepo.users["user-123"] = &model.PortalUser{
			ID:        "user-123",
			Email:     "test@example.com",
			AccountID: "account-123",
		}

		svc := NewPortalService(userRepo, sessionRepo, accountRepo, "test-secret")

		// Create a session
		token, _ := svc.CreateSession(context.Background(), "user-123")

		// Validate the session
		user, err := svc.ValidateSession(context.Background(), token)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "user-123", user.ID)
	})

	t.Run("ValidateSession returns nil for invalid token", func(t *testing.T) {
		userRepo := newMockPortalUserRepo()
		sessionRepo := newMockPortalSessionRepo()
		accountRepo := newMockAccountRepo()

		svc := NewPortalService(userRepo, sessionRepo, accountRepo, "test-secret")

		user, err := svc.ValidateSession(context.Background(), "invalid-token")

		assert.NoError(t, err)
		assert.Nil(t, user)
	})

	t.Run("Logout deletes session", func(t *testing.T) {
		userRepo := newMockPortalUserRepo()
		sessionRepo := newMockPortalSessionRepo()
		accountRepo := newMockAccountRepo()

		svc := NewPortalService(userRepo, sessionRepo, accountRepo, "test-secret")

		// Create a session
		token, _ := svc.CreateSession(context.Background(), "user-123")
		assert.Len(t, sessionRepo.sessions, 1)

		// Logout
		err := svc.Logout(context.Background(), token)

		assert.NoError(t, err)
		assert.Len(t, sessionRepo.sessions, 0)
	})
}
