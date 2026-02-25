package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/openclaw/relay-server-go/internal/model"
)

// Mock inbound repository
type mockInboundRepo struct {
	mock.Mock
}

func (m *mockInboundRepo) FindByID(ctx context.Context, id string) (*model.InboundMessage, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.InboundMessage), args.Error(1)
}

func (m *mockInboundRepo) FindQueuedByAccountID(ctx context.Context, accountID string) ([]model.InboundMessage, error) {
	args := m.Called(ctx, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.InboundMessage), args.Error(1)
}

func (m *mockInboundRepo) FindByAccountID(ctx context.Context, accountID string, limit, offset int) ([]model.InboundMessage, error) {
	args := m.Called(ctx, accountID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.InboundMessage), args.Error(1)
}

func (m *mockInboundRepo) CountByAccountID(ctx context.Context, accountID string) (int, error) {
	args := m.Called(ctx, accountID)
	return args.Int(0), args.Error(1)
}

func (m *mockInboundRepo) Create(ctx context.Context, params model.CreateInboundMessageParams) (*model.InboundMessage, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.InboundMessage), args.Error(1)
}

func (m *mockInboundRepo) MarkDelivered(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockInboundRepo) MarkAcked(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockInboundRepo) MarkExpired(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockInboundRepo) CountByStatus(ctx context.Context, status model.InboundMessageStatus) (int, error) {
	args := m.Called(ctx, status)
	return args.Int(0), args.Error(1)
}

func (m *mockInboundRepo) CountByAccountIDAndStatus(ctx context.Context, accountID string, status model.InboundMessageStatus) (int, error) {
	args := m.Called(ctx, accountID, status)
	return args.Int(0), args.Error(1)
}

func (m *mockInboundRepo) CountByAccountIDSince(ctx context.Context, accountID string, since time.Time) (int, error) {
	args := m.Called(ctx, accountID, since)
	return args.Int(0), args.Error(1)
}

func (m *mockInboundRepo) FindByConversationKey(ctx context.Context, conversationKey string, limit, offset int) ([]model.InboundMessage, error) {
	args := m.Called(ctx, conversationKey, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.InboundMessage), args.Error(1)
}

func (m *mockInboundRepo) CountByConversationKey(ctx context.Context, conversationKey string) (int, error) {
	args := m.Called(ctx, conversationKey)
	return args.Int(0), args.Error(1)
}

func (m *mockInboundRepo) CountByConversationKeySince(ctx context.Context, conversationKey string, since time.Time) (int, error) {
	args := m.Called(ctx, conversationKey, since)
	return args.Int(0), args.Error(1)
}

type mockOutboundRepo struct {
	mock.Mock
}

func (m *mockOutboundRepo) FindByID(ctx context.Context, id string) (*model.OutboundMessage, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.OutboundMessage), args.Error(1)
}

func (m *mockOutboundRepo) FindPendingByAccountID(ctx context.Context, accountID string) ([]model.OutboundMessage, error) {
	args := m.Called(ctx, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.OutboundMessage), args.Error(1)
}

func (m *mockOutboundRepo) FindByAccountID(ctx context.Context, accountID string, limit, offset int) ([]model.OutboundMessage, error) {
	args := m.Called(ctx, accountID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.OutboundMessage), args.Error(1)
}

func (m *mockOutboundRepo) CountByAccountID(ctx context.Context, accountID string) (int, error) {
	args := m.Called(ctx, accountID)
	return args.Int(0), args.Error(1)
}

func (m *mockOutboundRepo) Create(ctx context.Context, params model.CreateOutboundMessageParams) (*model.OutboundMessage, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.OutboundMessage), args.Error(1)
}

func (m *mockOutboundRepo) MarkSent(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockOutboundRepo) MarkFailed(ctx context.Context, id string, errorMsg string) error {
	args := m.Called(ctx, id, errorMsg)
	return args.Error(0)
}

func (m *mockOutboundRepo) CountByAccountIDAndStatus(ctx context.Context, accountID string, status model.OutboundMessageStatus) (int, error) {
	args := m.Called(ctx, accountID, status)
	return args.Int(0), args.Error(1)
}

func (m *mockOutboundRepo) CountByAccountIDSince(ctx context.Context, accountID string, since time.Time) (int, error) {
	args := m.Called(ctx, accountID, since)
	return args.Int(0), args.Error(1)
}

func (m *mockOutboundRepo) FindRecentFailedByAccountID(ctx context.Context, accountID string, limit int) ([]model.OutboundMessage, error) {
	args := m.Called(ctx, accountID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.OutboundMessage), args.Error(1)
}

func (m *mockOutboundRepo) FindByConversationKey(ctx context.Context, conversationKey string, limit, offset int) ([]model.OutboundMessage, error) {
	args := m.Called(ctx, conversationKey, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.OutboundMessage), args.Error(1)
}

func (m *mockOutboundRepo) CountByConversationKey(ctx context.Context, conversationKey string) (int, error) {
	args := m.Called(ctx, conversationKey)
	return args.Int(0), args.Error(1)
}

func (m *mockOutboundRepo) CountByConversationKeySince(ctx context.Context, conversationKey string, since time.Time) (int, error) {
	args := m.Called(ctx, conversationKey, since)
	return args.Int(0), args.Error(1)
}

func (m *mockOutboundRepo) CountByConversationKeyAndStatus(ctx context.Context, conversationKey string, status model.OutboundMessageStatus) (int, error) {
	args := m.Called(ctx, conversationKey, status)
	return args.Int(0), args.Error(1)
}

func TestMessageService_CreateInbound(t *testing.T) {
	t.Run("creates inbound message successfully", func(t *testing.T) {
		inboundRepo := new(mockInboundRepo)
		outboundRepo := new(mockOutboundRepo)
		svc := NewMessageService(inboundRepo, outboundRepo)

		ctx := context.Background()
		params := CreateInboundParams{
			AccountID:         "acc-1",
			ConversationKey:   "conv-1",
			KakaoPayload:      json.RawMessage(`{"type": "text"}`),
			NormalizedMessage: json.RawMessage(`{"text": "Hello"}`),
		}

		expectedMsg := &model.InboundMessage{
			ID:              "msg-1",
			AccountID:       "acc-1",
			ConversationKey: "conv-1",
			Status:          model.InboundStatusQueued,
		}

		inboundRepo.On("Create", ctx, mock.MatchedBy(func(p model.CreateInboundMessageParams) bool {
			return p.AccountID == "acc-1" && p.ConversationKey == "conv-1"
		})).Return(expectedMsg, nil)

		msg, err := svc.CreateInbound(ctx, params)

		assert.NoError(t, err)
		assert.NotNil(t, msg)
		assert.Equal(t, "msg-1", msg.ID)
		assert.Equal(t, "acc-1", msg.AccountID)
		inboundRepo.AssertExpectations(t)
	})

	t.Run("returns error when repository fails", func(t *testing.T) {
		inboundRepo := new(mockInboundRepo)
		outboundRepo := new(mockOutboundRepo)
		svc := NewMessageService(inboundRepo, outboundRepo)

		ctx := context.Background()
		params := CreateInboundParams{
			AccountID:       "acc-1",
			ConversationKey: "conv-1",
		}

		inboundRepo.On("Create", ctx, mock.Anything).Return(nil, assert.AnError)

		msg, err := svc.CreateInbound(ctx, params)

		assert.Error(t, err)
		assert.Nil(t, msg)
		assert.Contains(t, err.Error(), "create inbound message")
		inboundRepo.AssertExpectations(t)
	})
}

func TestMessageService_FindInboundByID(t *testing.T) {
	t.Run("finds message by ID", func(t *testing.T) {
		inboundRepo := new(mockInboundRepo)
		outboundRepo := new(mockOutboundRepo)
		svc := NewMessageService(inboundRepo, outboundRepo)

		ctx := context.Background()
		expectedMsg := &model.InboundMessage{
			ID:              "msg-1",
			AccountID:       "acc-1",
			ConversationKey: "conv-1",
		}

		inboundRepo.On("FindByID", ctx, "msg-1").Return(expectedMsg, nil)

		msg, err := svc.FindInboundByID(ctx, "msg-1")

		assert.NoError(t, err)
		assert.NotNil(t, msg)
		assert.Equal(t, "msg-1", msg.ID)
		inboundRepo.AssertExpectations(t)
	})

	t.Run("returns nil when not found", func(t *testing.T) {
		inboundRepo := new(mockInboundRepo)
		outboundRepo := new(mockOutboundRepo)
		svc := NewMessageService(inboundRepo, outboundRepo)

		ctx := context.Background()
		inboundRepo.On("FindByID", ctx, "msg-unknown").Return(nil, nil)

		msg, err := svc.FindInboundByID(ctx, "msg-unknown")

		assert.NoError(t, err)
		assert.Nil(t, msg)
		inboundRepo.AssertExpectations(t)
	})
}

func TestMessageService_FindQueuedByAccountID(t *testing.T) {
	t.Run("finds queued messages", func(t *testing.T) {
		inboundRepo := new(mockInboundRepo)
		outboundRepo := new(mockOutboundRepo)
		svc := NewMessageService(inboundRepo, outboundRepo)

		ctx := context.Background()
		expectedMsgs := []model.InboundMessage{
			{ID: "msg-1", AccountID: "acc-1", Status: model.InboundStatusQueued},
			{ID: "msg-2", AccountID: "acc-1", Status: model.InboundStatusQueued},
		}

		inboundRepo.On("FindQueuedByAccountID", ctx, "acc-1").Return(expectedMsgs, nil)

		msgs, err := svc.FindQueuedByAccountID(ctx, "acc-1")

		assert.NoError(t, err)
		assert.Len(t, msgs, 2)
		inboundRepo.AssertExpectations(t)
	})
}

func TestMessageService_MarkDelivered(t *testing.T) {
	t.Run("marks message as delivered", func(t *testing.T) {
		inboundRepo := new(mockInboundRepo)
		outboundRepo := new(mockOutboundRepo)
		svc := NewMessageService(inboundRepo, outboundRepo)

		ctx := context.Background()
		inboundRepo.On("MarkDelivered", ctx, "msg-1").Return(nil)

		err := svc.MarkDelivered(ctx, "msg-1")

		assert.NoError(t, err)
		inboundRepo.AssertExpectations(t)
	})

	t.Run("returns error when repository fails", func(t *testing.T) {
		inboundRepo := new(mockInboundRepo)
		outboundRepo := new(mockOutboundRepo)
		svc := NewMessageService(inboundRepo, outboundRepo)

		ctx := context.Background()
		inboundRepo.On("MarkDelivered", ctx, "msg-1").Return(assert.AnError)

		err := svc.MarkDelivered(ctx, "msg-1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mark delivered")
		inboundRepo.AssertExpectations(t)
	})
}

func TestMessageService_MarkAcked(t *testing.T) {
	t.Run("marks message as acked", func(t *testing.T) {
		inboundRepo := new(mockInboundRepo)
		outboundRepo := new(mockOutboundRepo)
		svc := NewMessageService(inboundRepo, outboundRepo)

		ctx := context.Background()
		inboundRepo.On("MarkAcked", ctx, "msg-1").Return(nil)

		err := svc.MarkAcked(ctx, "msg-1")

		assert.NoError(t, err)
		inboundRepo.AssertExpectations(t)
	})
}

func TestMessageService_CreateOutbound(t *testing.T) {
	t.Run("creates outbound message successfully", func(t *testing.T) {
		inboundRepo := new(mockInboundRepo)
		outboundRepo := new(mockOutboundRepo)
		svc := NewMessageService(inboundRepo, outboundRepo)

		ctx := context.Background()
		inboundID := "msg-in-1"
		params := model.CreateOutboundMessageParams{
			AccountID:        "acc-1",
			InboundMessageID: &inboundID,
			ConversationKey:  "conv-1",
			KakaoTarget:      json.RawMessage(`{}`),
			ResponsePayload:  json.RawMessage(`{"text": "Response"}`),
		}

		expectedMsg := &model.OutboundMessage{
			ID:              "msg-out-1",
			AccountID:       "acc-1",
			ConversationKey: "conv-1",
			Status:          model.OutboundStatusPending,
		}

		outboundRepo.On("Create", ctx, params).Return(expectedMsg, nil)

		msg, err := svc.CreateOutbound(ctx, params)

		assert.NoError(t, err)
		assert.NotNil(t, msg)
		assert.Equal(t, "msg-out-1", msg.ID)
		outboundRepo.AssertExpectations(t)
	})
}

func TestMessageService_MarkOutboundSent(t *testing.T) {
	t.Run("marks outbound as sent", func(t *testing.T) {
		inboundRepo := new(mockInboundRepo)
		outboundRepo := new(mockOutboundRepo)
		svc := NewMessageService(inboundRepo, outboundRepo)

		ctx := context.Background()
		outboundRepo.On("MarkSent", ctx, "msg-out-1").Return(nil)

		err := svc.MarkOutboundSent(ctx, "msg-out-1")

		assert.NoError(t, err)
		outboundRepo.AssertExpectations(t)
	})
}

func TestMessageService_MarkOutboundFailed(t *testing.T) {
	t.Run("marks outbound as failed", func(t *testing.T) {
		inboundRepo := new(mockInboundRepo)
		outboundRepo := new(mockOutboundRepo)
		svc := NewMessageService(inboundRepo, outboundRepo)

		ctx := context.Background()
		outboundRepo.On("MarkFailed", ctx, "msg-out-1", "connection timeout").Return(nil)

		err := svc.MarkOutboundFailed(ctx, "msg-out-1", "connection timeout")

		assert.NoError(t, err)
		outboundRepo.AssertExpectations(t)
	})
}

func TestMessageService_GetMessageHistory(t *testing.T) {
	t.Run("returns inbound messages only", func(t *testing.T) {
		inboundRepo := new(mockInboundRepo)
		outboundRepo := new(mockOutboundRepo)
		svc := NewMessageService(inboundRepo, outboundRepo)

		ctx := context.Background()
		now := time.Now()
		normalized := json.RawMessage(`{"text": "Hello"}`)
		inboundMsgs := []model.InboundMessage{
			{ID: "msg-1", ConversationKey: "conv-1", NormalizedMessage: &normalized, CreatedAt: now},
		}

		inboundRepo.On("FindByAccountID", ctx, "acc-1", 20, 0).Return(inboundMsgs, nil)
		inboundRepo.On("CountByAccountID", ctx, "acc-1").Return(1, nil)

		result, err := svc.GetMessageHistory(ctx, MessageHistoryParams{
			AccountID: "acc-1",
			Type:      "inbound",
		})

		assert.NoError(t, err)
		assert.Len(t, result.Messages, 1)
		assert.Equal(t, "inbound", result.Messages[0].Direction)
		assert.Equal(t, 1, result.Total)
		assert.False(t, result.HasMore)
		inboundRepo.AssertExpectations(t)
	})

	t.Run("returns outbound messages only", func(t *testing.T) {
		inboundRepo := new(mockInboundRepo)
		outboundRepo := new(mockOutboundRepo)
		svc := NewMessageService(inboundRepo, outboundRepo)

		ctx := context.Background()
		now := time.Now()
		outboundMsgs := []model.OutboundMessage{
			{ID: "msg-out-1", ConversationKey: "conv-1", ResponsePayload: json.RawMessage(`{"text": "Reply"}`), CreatedAt: now},
		}

		outboundRepo.On("FindByAccountID", ctx, "acc-1", 20, 0).Return(outboundMsgs, nil)
		outboundRepo.On("CountByAccountID", ctx, "acc-1").Return(1, nil)

		result, err := svc.GetMessageHistory(ctx, MessageHistoryParams{
			AccountID: "acc-1",
			Type:      "outbound",
		})

		assert.NoError(t, err)
		assert.Len(t, result.Messages, 1)
		assert.Equal(t, "outbound", result.Messages[0].Direction)
		outboundRepo.AssertExpectations(t)
	})

	t.Run("returns all messages sorted by created_at", func(t *testing.T) {
		inboundRepo := new(mockInboundRepo)
		outboundRepo := new(mockOutboundRepo)
		svc := NewMessageService(inboundRepo, outboundRepo)

		ctx := context.Background()
		now := time.Now()
		normalized := json.RawMessage(`{"text": "Hello"}`)
		inboundMsgs := []model.InboundMessage{
			{ID: "msg-1", ConversationKey: "conv-1", NormalizedMessage: &normalized, CreatedAt: now.Add(-2 * time.Minute)},
		}
		outboundMsgs := []model.OutboundMessage{
			{ID: "msg-out-1", ConversationKey: "conv-1", ResponsePayload: json.RawMessage(`{"text": "Reply"}`), CreatedAt: now.Add(-1 * time.Minute)},
		}

		inboundRepo.On("FindByAccountID", ctx, "acc-1", 20, 0).Return(inboundMsgs, nil)
		outboundRepo.On("FindByAccountID", ctx, "acc-1", 20, 0).Return(outboundMsgs, nil)
		inboundRepo.On("CountByAccountID", ctx, "acc-1").Return(1, nil)
		outboundRepo.On("CountByAccountID", ctx, "acc-1").Return(1, nil)

		result, err := svc.GetMessageHistory(ctx, MessageHistoryParams{
			AccountID: "acc-1",
			Type:      "", // All
		})

		assert.NoError(t, err)
		assert.Len(t, result.Messages, 2)
		// Should be sorted by created_at descending (newest first)
		assert.Equal(t, "msg-out-1", result.Messages[0].ID)
		assert.Equal(t, "msg-1", result.Messages[1].ID)
		inboundRepo.AssertExpectations(t)
		outboundRepo.AssertExpectations(t)
	})

	t.Run("limits results to specified limit", func(t *testing.T) {
		inboundRepo := new(mockInboundRepo)
		outboundRepo := new(mockOutboundRepo)
		svc := NewMessageService(inboundRepo, outboundRepo)

		ctx := context.Background()
		inboundRepo.On("FindByAccountID", ctx, "acc-1", 5, 0).Return([]model.InboundMessage{}, nil)
		inboundRepo.On("CountByAccountID", ctx, "acc-1").Return(0, nil)

		_, err := svc.GetMessageHistory(ctx, MessageHistoryParams{
			AccountID: "acc-1",
			Type:      "inbound",
			Limit:     5,
		})

		assert.NoError(t, err)
		inboundRepo.AssertExpectations(t)
	})

	t.Run("enforces max limit of 100", func(t *testing.T) {
		inboundRepo := new(mockInboundRepo)
		outboundRepo := new(mockOutboundRepo)
		svc := NewMessageService(inboundRepo, outboundRepo)

		ctx := context.Background()
		inboundRepo.On("FindByAccountID", ctx, "acc-1", 100, 0).Return([]model.InboundMessage{}, nil)
		inboundRepo.On("CountByAccountID", ctx, "acc-1").Return(0, nil)

		_, err := svc.GetMessageHistory(ctx, MessageHistoryParams{
			AccountID: "acc-1",
			Type:      "inbound",
			Limit:     200, // Exceeds max
		})

		assert.NoError(t, err)
		inboundRepo.AssertExpectations(t)
	})

	t.Run("defaults limit to 20 when not specified", func(t *testing.T) {
		inboundRepo := new(mockInboundRepo)
		outboundRepo := new(mockOutboundRepo)
		svc := NewMessageService(inboundRepo, outboundRepo)

		ctx := context.Background()
		inboundRepo.On("FindByAccountID", ctx, "acc-1", 20, 0).Return([]model.InboundMessage{}, nil)
		inboundRepo.On("CountByAccountID", ctx, "acc-1").Return(0, nil)

		_, err := svc.GetMessageHistory(ctx, MessageHistoryParams{
			AccountID: "acc-1",
			Type:      "inbound",
			Limit:     0, // Not specified
		})

		assert.NoError(t, err)
		inboundRepo.AssertExpectations(t)
	})

	t.Run("calculates HasMore correctly", func(t *testing.T) {
		inboundRepo := new(mockInboundRepo)
		outboundRepo := new(mockOutboundRepo)
		svc := NewMessageService(inboundRepo, outboundRepo)

		ctx := context.Background()
		now := time.Now()
		normalized := json.RawMessage(`{"text": "Hello"}`)
		inboundMsgs := []model.InboundMessage{
			{ID: "msg-1", NormalizedMessage: &normalized, CreatedAt: now},
			{ID: "msg-2", NormalizedMessage: &normalized, CreatedAt: now},
		}

		inboundRepo.On("FindByAccountID", ctx, "acc-1", 2, 0).Return(inboundMsgs, nil)
		inboundRepo.On("CountByAccountID", ctx, "acc-1").Return(5, nil) // Total is 5

		result, err := svc.GetMessageHistory(ctx, MessageHistoryParams{
			AccountID: "acc-1",
			Type:      "inbound",
			Limit:     2,
			Offset:    0,
		})

		assert.NoError(t, err)
		assert.True(t, result.HasMore) // 0 + 2 < 5
		inboundRepo.AssertExpectations(t)
	})
}
