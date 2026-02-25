package repository

import (
	"context"
	"testing"
	"time"

	"github.com/openclaw/relay-server-go/internal/database"
	"github.com/openclaw/relay-server-go/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPortalAccessCodeRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPortalAccessCodeRepository(db.DB)
	ctx := context.Background()

	code, err := repo.Create(ctx, model.CreatePortalAccessCodeParams{
		Code:            "ABCD-1234",
		ConversationKey: "test-conversation",
		ExpiresAt:       time.Now().Add(30 * time.Minute),
	})

	require.NoError(t, err)
	assert.Equal(t, "ABCD-1234", code.Code)
	assert.Equal(t, "test-conversation", code.ConversationKey)
	assert.Nil(t, code.UsedAt)
}

func TestPortalAccessCodeRepository_FindActiveByCode(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPortalAccessCodeRepository(db.DB)
	ctx := context.Background()

	expiresAt := time.Now().Add(30 * time.Minute)
	_, err := repo.Create(ctx, model.CreatePortalAccessCodeParams{
		Code:            "ABCD-1234",
		ConversationKey: "test-conversation",
		ExpiresAt:       expiresAt,
	})
	require.NoError(t, err)

	t.Run("finds active code", func(t *testing.T) {
		code, err := repo.FindActiveByCode(ctx, "ABCD-1234")
		require.NoError(t, err)
		assert.Equal(t, "ABCD-1234", code.Code)
		assert.Equal(t, "test-conversation", code.ConversationKey)
	})

	t.Run("returns nil for non-existent code", func(t *testing.T) {
		code, err := repo.FindActiveByCode(ctx, "XXXX-9999")
		require.NoError(t, err)
		assert.Nil(t, code)
	})

	t.Run("does not find used code", func(t *testing.T) {
		err := repo.MarkUsed(ctx, "ABCD-1234")
		require.NoError(t, err)

		code, err := repo.FindActiveByCode(ctx, "ABCD-1234")
		require.NoError(t, err)
		assert.Nil(t, code)
	})
}

func TestPortalAccessCodeRepository_FindActiveByConversationKey(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPortalAccessCodeRepository(db.DB)
	ctx := context.Background()

	// Create multiple codes for same conversation
	_, err := repo.Create(ctx, model.CreatePortalAccessCodeParams{
		Code:            "CODE-0001",
		ConversationKey: "test-conv",
		ExpiresAt:       time.Now().Add(10 * time.Minute),
	})
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond) // Ensure different timestamps

	_, err = repo.Create(ctx, model.CreatePortalAccessCodeParams{
		Code:            "CODE-0002",
		ConversationKey: "test-conv",
		ExpiresAt:       time.Now().Add(20 * time.Minute),
	})
	require.NoError(t, err)

	t.Run("finds most recent active code", func(t *testing.T) {
		code, err := repo.FindActiveByConversationKey(ctx, "test-conv")
		require.NoError(t, err)
		assert.Equal(t, "CODE-0002", code.Code) // Most recent
	})

	t.Run("returns nil for non-existent conversation", func(t *testing.T) {
		code, err := repo.FindActiveByConversationKey(ctx, "non-existent")
		require.NoError(t, err)
		assert.Nil(t, code)
	})
}

func TestPortalAccessCodeRepository_MarkUsed(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPortalAccessCodeRepository(db.DB)
	ctx := context.Background()

	_, err := repo.Create(ctx, model.CreatePortalAccessCodeParams{
		Code:            "TEST-CODE",
		ConversationKey: "test-conv",
		ExpiresAt:       time.Now().Add(30 * time.Minute),
	})
	require.NoError(t, err)

	err = repo.MarkUsed(ctx, "TEST-CODE")
	require.NoError(t, err)

	// Verify code is no longer active
	code, err := repo.FindActiveByCode(ctx, "TEST-CODE")
	require.NoError(t, err)
	assert.Nil(t, code)
}

func TestPortalAccessCodeRepository_DeleteExpired(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPortalAccessCodeRepository(db.DB)
	ctx := context.Background()

	// Create expired code
	_, err := repo.Create(ctx, model.CreatePortalAccessCodeParams{
		Code:            "EXPIRED-1",
		ConversationKey: "test-conv",
		ExpiresAt:       time.Now().Add(-1 * time.Hour), // Already expired
	})
	require.NoError(t, err)

	// Create valid code
	_, err = repo.Create(ctx, model.CreatePortalAccessCodeParams{
		Code:            "VALID-001",
		ConversationKey: "test-conv",
		ExpiresAt:       time.Now().Add(30 * time.Minute),
	})
	require.NoError(t, err)

	count, err := repo.DeleteExpired(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Verify expired code is deleted
	code, err := repo.FindActiveByCode(ctx, "EXPIRED-1")
	require.NoError(t, err)
	assert.Nil(t, code)

	// Verify valid code still exists
	code, err = repo.FindActiveByCode(ctx, "VALID-001")
	require.NoError(t, err)
	assert.NotNil(t, code)
}

func setupTestDB(t *testing.T) *database.DB {
	t.Helper()
	db, err := database.Connect("postgres://postgres:postgres@localhost:5432/relay_test?sslmode=disable")
	require.NoError(t, err)
	return db
}
