package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/openclaw/relay-server-go/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock repositories
type mockPortalAccessCodeRepo struct {
	mock.Mock
}

func (m *mockPortalAccessCodeRepo) Create(ctx context.Context, params model.CreatePortalAccessCodeParams) (*model.PortalAccessCode, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.PortalAccessCode), args.Error(1)
}

func (m *mockPortalAccessCodeRepo) FindActiveByCode(ctx context.Context, code string) (*model.PortalAccessCode, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.PortalAccessCode), args.Error(1)
}

func (m *mockPortalAccessCodeRepo) FindActiveByConversationKey(ctx context.Context, conversationKey string) (*model.PortalAccessCode, error) {
	args := m.Called(ctx, conversationKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.PortalAccessCode), args.Error(1)
}

func (m *mockPortalAccessCodeRepo) MarkUsed(ctx context.Context, code string) error {
	args := m.Called(ctx, code)
	return args.Error(0)
}

func (m *mockPortalAccessCodeRepo) UpdateLastAccessed(ctx context.Context, code string) error {
	args := m.Called(ctx, code)
	return args.Error(0)
}

func (m *mockPortalAccessCodeRepo) DeleteExpired(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

type mockConversationRepo struct {
	mock.Mock
}

func (m *mockConversationRepo) FindByKey(ctx context.Context, key string) (*model.ConversationMapping, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ConversationMapping), args.Error(1)
}

func (m *mockConversationRepo) FindByAccountID(ctx context.Context, accountID string) ([]model.ConversationMapping, error) {
	args := m.Called(ctx, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.ConversationMapping), args.Error(1)
}

func (m *mockConversationRepo) FindPairedByAccountID(ctx context.Context, accountID string) ([]model.ConversationMapping, error) {
	args := m.Called(ctx, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.ConversationMapping), args.Error(1)
}

func (m *mockConversationRepo) Upsert(ctx context.Context, params model.UpsertConversationParams) (*model.ConversationMapping, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ConversationMapping), args.Error(1)
}

func (m *mockConversationRepo) UpdateState(ctx context.Context, key string, state model.PairingState, accountID *string) error {
	args := m.Called(ctx, key, state, accountID)
	return args.Error(0)
}

func (m *mockConversationRepo) UpdateCallback(ctx context.Context, key string, callbackURL string, expiresAt time.Time) error {
	args := m.Called(ctx, key, callbackURL, expiresAt)
	return args.Error(0)
}

func (m *mockConversationRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockConversationRepo) CountByState(ctx context.Context, state model.PairingState) (int, error) {
	args := m.Called(ctx, state)
	return args.Int(0), args.Error(1)
}

func TestGenerateCode_ReusePolicy(t *testing.T) {
	mockCodeRepo := new(mockPortalAccessCodeRepo)
	mockConvRepo := new(mockConversationRepo)

	// Mock existing active code
	existingCode := &model.PortalAccessCode{
		Code:            "ABCD-1234",
		ConversationKey: "test-conv",
		ExpiresAt:       time.Now().Add(15 * time.Minute),
		CreatedAt:       time.Now().Add(-15 * time.Minute),
	}

	mockCodeRepo.On("FindActiveByConversationKey", mock.Anything, "test-conv").
		Return(existingCode, nil)

	service := &PortalAccessService{
		codeRepo: mockCodeRepo,
		convRepo: mockConvRepo,
	}

	ctx := context.Background()
	code, err := service.GenerateCode(ctx, "test-conv")

	require.NoError(t, err)
	assert.Equal(t, "ABCD-1234", code.Code)
	assert.Equal(t, "test-conv", code.ConversationKey)

	// Verify Create was NOT called (code reused)
	mockCodeRepo.AssertNotCalled(t, "Create")
}

func TestGenerateCode_NewCode(t *testing.T) {
	mockCodeRepo := new(mockPortalAccessCodeRepo)
	mockConvRepo := new(mockConversationRepo)

	// No existing active code
	mockCodeRepo.On("FindActiveByConversationKey", mock.Anything, "new-conv").
		Return(nil, sql.ErrNoRows)

	// No collision on first attempt
	mockCodeRepo.On("FindActiveByCode", mock.Anything, mock.AnythingOfType("string")).
		Return(nil, sql.ErrNoRows).Once()

	newCode := &model.PortalAccessCode{
		Code:            "WXYZ-5678",
		ConversationKey: "new-conv",
		ExpiresAt:       time.Now().Add(30 * time.Minute),
		CreatedAt:       time.Now(),
	}

	mockCodeRepo.On("Create", mock.Anything, mock.AnythingOfType("model.CreatePortalAccessCodeParams")).
		Return(newCode, nil)

	service := &PortalAccessService{
		codeRepo: mockCodeRepo,
		convRepo: mockConvRepo,
	}

	ctx := context.Background()
	code, err := service.GenerateCode(ctx, "new-conv")

	require.NoError(t, err)
	assert.Len(t, code.Code, 9) // XXXX-XXXX format
	assert.Equal(t, "new-conv", code.ConversationKey)

	mockCodeRepo.AssertCalled(t, "Create", mock.Anything, mock.AnythingOfType("model.CreatePortalAccessCodeParams"))
}

func TestVerifyCode_Success(t *testing.T) {
	mockCodeRepo := new(mockPortalAccessCodeRepo)
	mockConvRepo := new(mockConversationRepo)

	activeCode := &model.PortalAccessCode{
		Code:            "ABCD-1234",
		ConversationKey: "test-conv",
		ExpiresAt:       time.Now().Add(15 * time.Minute),
	}

	mockCodeRepo.On("FindActiveByCode", mock.Anything, "ABCD-1234").
		Return(activeCode, nil)
	mockCodeRepo.On("MarkUsed", mock.Anything, "ABCD-1234").
		Return(nil)

	service := &PortalAccessService{
		codeRepo: mockCodeRepo,
		convRepo: mockConvRepo,
	}

	ctx := context.Background()
	conversationKey, err := service.VerifyCode(ctx, "abcd-1234") // Test case insensitive

	require.NoError(t, err)
	assert.Equal(t, "test-conv", conversationKey)

	mockCodeRepo.AssertCalled(t, "MarkUsed", mock.Anything, "ABCD-1234")
}

func TestVerifyCode_InvalidCode(t *testing.T) {
	mockCodeRepo := new(mockPortalAccessCodeRepo)
	mockConvRepo := new(mockConversationRepo)

	mockCodeRepo.On("FindActiveByCode", mock.Anything, "INVALID").
		Return(nil, sql.ErrNoRows)

	service := &PortalAccessService{
		codeRepo: mockCodeRepo,
		convRepo: mockConvRepo,
	}

	ctx := context.Background()
	_, err := service.VerifyCode(ctx, "INVALID")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid or expired code")

	mockCodeRepo.AssertNotCalled(t, "MarkUsed")
}

func TestCreateCodeSession(t *testing.T) {
	service := &PortalAccessService{}

	session, err := service.CreateCodeSession("test-conv")

	require.NoError(t, err)
	assert.NotEmpty(t, session.Token)
	assert.Equal(t, "test-conv", session.ConversationKey)
	assert.True(t, session.ExpiresAt.After(time.Now()))
	assert.True(t, session.ExpiresAt.Before(time.Now().Add(31*time.Minute)))
}

func TestGeneratePortalCode_Format(t *testing.T) {
	for i := 0; i < 100; i++ {
		code := generatePortalCode()
		assert.Len(t, code, 9, "Code should be 9 characters (XXXX-XXXX)")
		assert.Equal(t, "-", string(code[4]), "5th character should be hyphen")

		// Verify only allowed characters
		for j, ch := range code {
			if j == 4 {
				continue // Skip hyphen
			}
			assert.Contains(t, portalCodeChars, string(ch),
				"Code should only contain characters from portalCodeChars")
		}
	}
}

func TestGeneratePortalCode_Uniqueness(t *testing.T) {
	// Generate 1000 codes and ensure no duplicates
	codes := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		code := generatePortalCode()
		assert.False(t, codes[code], "Generated duplicate code: %s", code)
		codes[code] = true
	}
}
