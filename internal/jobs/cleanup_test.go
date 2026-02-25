package jobs

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"gitlab.tepseg.com/ai/kakao-relay/internal/model"
	"gitlab.tepseg.com/ai/kakao-relay/internal/repository"
)

type mockInboundMsgRepo struct {
	markExpiredCount int64
}

func (m *mockInboundMsgRepo) FindByID(ctx context.Context, id string) (*model.InboundMessage, error) {
	return nil, nil
}

func (m *mockInboundMsgRepo) FindQueuedByAccountID(ctx context.Context, accountID string) ([]model.InboundMessage, error) {
	return nil, nil
}

func (m *mockInboundMsgRepo) FindByAccountID(ctx context.Context, accountID string, limit, offset int) ([]model.InboundMessage, error) {
	return nil, nil
}

func (m *mockInboundMsgRepo) Create(ctx context.Context, params model.CreateInboundMessageParams) (*model.InboundMessage, error) {
	return nil, nil
}

func (m *mockInboundMsgRepo) MarkDelivered(ctx context.Context, id string) error {
	return nil
}

func (m *mockInboundMsgRepo) MarkAcked(ctx context.Context, id string) error {
	return nil
}

func (m *mockInboundMsgRepo) MarkExpired(ctx context.Context) (int64, error) {
	return m.markExpiredCount, nil
}

func (m *mockInboundMsgRepo) CountByStatus(ctx context.Context, status model.InboundMessageStatus) (int, error) {
	return 0, nil
}

func (m *mockInboundMsgRepo) CountByAccountID(ctx context.Context, accountID string) (int, error) {
	return 0, nil
}

func (m *mockInboundMsgRepo) CountByAccountIDAndStatus(ctx context.Context, accountID string, status model.InboundMessageStatus) (int, error) {
	return 0, nil
}

func (m *mockInboundMsgRepo) CountByAccountIDSince(ctx context.Context, accountID string, since time.Time) (int, error) {
	return 0, nil
}

func (m *mockInboundMsgRepo) FindByConversationKey(ctx context.Context, conversationKey string, limit, offset int) ([]model.InboundMessage, error) {
	return nil, nil
}

func (m *mockInboundMsgRepo) CountByConversationKey(ctx context.Context, conversationKey string) (int, error) {
	return 0, nil
}

func (m *mockInboundMsgRepo) CountByConversationKeySince(ctx context.Context, conversationKey string, since time.Time) (int, error) {
	return 0, nil
}

type mockSessionRepo struct {
	deleteExpiredCount int64
}

func (m *mockSessionRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*model.Session, error) {
	return nil, nil
}

func (m *mockSessionRepo) FindByPairingCode(ctx context.Context, code string) (*model.Session, error) {
	return nil, nil
}

func (m *mockSessionRepo) FindByID(ctx context.Context, id string) (*model.Session, error) {
	return nil, nil
}

func (m *mockSessionRepo) Create(ctx context.Context, params model.CreateSessionParams) (*model.Session, error) {
	return nil, nil
}

func (m *mockSessionRepo) MarkPaired(ctx context.Context, id, accountID, conversationKey string) error {
	return nil
}

func (m *mockSessionRepo) MarkExpired(ctx context.Context, id string) error {
	return nil
}

func (m *mockSessionRepo) DeleteExpired(ctx context.Context) (int64, error) {
	return m.deleteExpiredCount, nil
}

func (m *mockSessionRepo) CountPendingByIP(ctx context.Context, ip string, since time.Time) (int, error) {
	return 0, nil
}

func (m *mockSessionRepo) MarkDisconnected(ctx context.Context, id string) error {
	return nil
}

func (m *mockSessionRepo) FindRecent(ctx context.Context, limit int) ([]model.Session, error) {
	return nil, nil
}

func (m *mockSessionRepo) CountByStatus(ctx context.Context, status model.SessionStatus) (int, error) {
	return 0, nil
}

func (m *mockSessionRepo) UpdateMetadata(ctx context.Context, id string, metadata json.RawMessage) error {
	return nil
}

func (m *mockSessionRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockSessionRepo) WithTx(tx *sqlx.Tx) repository.SessionRepository {
	return m
}

func TestCleanupJob(t *testing.T) {
	t.Run("creates job with correct interval", func(t *testing.T) {
		job := NewCleanupJob(nil, nil, 5*time.Minute)

		assert.NotNil(t, job)
		assert.Equal(t, 5*time.Minute, job.interval)
	})

	t.Run("starts and stops without panic", func(t *testing.T) {
		msgRepo := &mockInboundMsgRepo{}
		sessionRepo := &mockSessionRepo{}

		job := NewCleanupJob(msgRepo, sessionRepo, 100*time.Millisecond)

		job.Start()
		time.Sleep(50 * time.Millisecond)
		job.Stop()
	})

	t.Run("runs cleanup on start", func(t *testing.T) {
		msgRepo := &mockInboundMsgRepo{markExpiredCount: 5}
		sessionRepo := &mockSessionRepo{deleteExpiredCount: 6}

		job := NewCleanupJob(msgRepo, sessionRepo, 1*time.Hour)

		job.Start()
		time.Sleep(10 * time.Millisecond)
		job.Stop()
	})
}
