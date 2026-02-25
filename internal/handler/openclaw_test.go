package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/openclaw/relay-server-go/internal/middleware"
	"github.com/openclaw/relay-server-go/internal/model"
	"github.com/openclaw/relay-server-go/internal/service"
)

// Mock repositories
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
	return args.Get(0).([]model.InboundMessage), args.Error(1)
}

func (m *mockInboundRepo) FindByAccountID(ctx context.Context, accountID string, limit, offset int) ([]model.InboundMessage, error) {
	args := m.Called(ctx, accountID, limit, offset)
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

type mockOutboundRepo struct{
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
	return args.Get(0).([]model.OutboundMessage), args.Error(1)
}

func (m *mockOutboundRepo) FindByAccountID(ctx context.Context, accountID string, limit, offset int) ([]model.OutboundMessage, error) {
	args := m.Called(ctx, accountID, limit, offset)
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

type mockKakaoService struct {
	mock.Mock
}

func (m *mockKakaoService) SendCallback(ctx context.Context, callbackURL string, payload any) error {
	args := m.Called(ctx, callbackURL, payload)
	return args.Error(0)
}

// Helper to add account to context
func withAccount(ctx context.Context, account *model.Account) context.Context {
	return context.WithValue(ctx, middleware.AccountContextKey, account)
}

func TestOpenClawHandler_Reply(t *testing.T) {
	t.Run("returns 401 when no account in context", func(t *testing.T) {
		inboundRepo := new(mockInboundRepo)
		outboundRepo := new(mockOutboundRepo)
		msgService := service.NewMessageService(inboundRepo, outboundRepo)
		kakaoService := service.NewKakaoService()

		handler := NewOpenClawHandler(msgService, kakaoService)

		body := bytes.NewBufferString(`{"messageId": "msg-1", "response": {"text": "Hello"}}`)
		req := httptest.NewRequest(http.MethodPost, "/openclaw/reply", body)
		rec := httptest.NewRecorder()

		handler.Reply(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "SESSION_NOT_PAIRED")
	})

	t.Run("returns 400 when messageId is missing", func(t *testing.T) {
		inboundRepo := new(mockInboundRepo)
		outboundRepo := new(mockOutboundRepo)
		msgService := service.NewMessageService(inboundRepo, outboundRepo)
		kakaoService := service.NewKakaoService()

		handler := NewOpenClawHandler(msgService, kakaoService)

		account := &model.Account{ID: "acc-1"}
		body := bytes.NewBufferString(`{"response": {"text": "Hello"}}`)
		req := httptest.NewRequest(http.MethodPost, "/openclaw/reply", body)
		req = req.WithContext(withAccount(req.Context(), account))
		rec := httptest.NewRecorder()

		handler.Reply(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "MISSING_REQUIRED")
	})

	t.Run("returns 400 when request body is invalid", func(t *testing.T) {
		inboundRepo := new(mockInboundRepo)
		outboundRepo := new(mockOutboundRepo)
		msgService := service.NewMessageService(inboundRepo, outboundRepo)
		kakaoService := service.NewKakaoService()

		handler := NewOpenClawHandler(msgService, kakaoService)

		account := &model.Account{ID: "acc-1"}
		body := bytes.NewBufferString(`{invalid json}`)
		req := httptest.NewRequest(http.MethodPost, "/openclaw/reply", body)
		req = req.WithContext(withAccount(req.Context(), account))
		rec := httptest.NewRecorder()

		handler.Reply(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "VALIDATION_ERROR")
	})

	t.Run("returns 404 when message not found", func(t *testing.T) {
		inboundRepo := new(mockInboundRepo)
		outboundRepo := new(mockOutboundRepo)
		msgService := service.NewMessageService(inboundRepo, outboundRepo)
		kakaoService := service.NewKakaoService()

		inboundRepo.On("FindByID", mock.Anything, "msg-1").Return(nil, nil)

		handler := NewOpenClawHandler(msgService, kakaoService)

		account := &model.Account{ID: "acc-1"}
		body := bytes.NewBufferString(`{"messageId": "msg-1", "response": {"text": "Hello"}}`)
		req := httptest.NewRequest(http.MethodPost, "/openclaw/reply", body)
		req = req.WithContext(withAccount(req.Context(), account))
		rec := httptest.NewRecorder()

		handler.Reply(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), "NOT_FOUND")
		inboundRepo.AssertExpectations(t)
	})

	t.Run("returns 404 when message belongs to different account", func(t *testing.T) {
		inboundRepo := new(mockInboundRepo)
		outboundRepo := new(mockOutboundRepo)
		msgService := service.NewMessageService(inboundRepo, outboundRepo)
		kakaoService := service.NewKakaoService()

		callbackURL := "https://callback.kakao.com/v1"
		expiresAt := time.Now().Add(1 * time.Hour)
		inboundMsg := &model.InboundMessage{
			ID:                "msg-1",
			AccountID:         "acc-other", // Different account
			ConversationKey:   "conv-1",
			CallbackURL:       &callbackURL,
			CallbackExpiresAt: &expiresAt,
		}
		inboundRepo.On("FindByID", mock.Anything, "msg-1").Return(inboundMsg, nil)

		handler := NewOpenClawHandler(msgService, kakaoService)

		account := &model.Account{ID: "acc-1"}
		body := bytes.NewBufferString(`{"messageId": "msg-1", "response": {"text": "Hello"}}`)
		req := httptest.NewRequest(http.MethodPost, "/openclaw/reply", body)
		req = req.WithContext(withAccount(req.Context(), account))
		rec := httptest.NewRecorder()

		handler.Reply(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), "NOT_FOUND")
		inboundRepo.AssertExpectations(t)
	})

	t.Run("returns 400 when callback URL is nil", func(t *testing.T) {
		inboundRepo := new(mockInboundRepo)
		outboundRepo := new(mockOutboundRepo)
		msgService := service.NewMessageService(inboundRepo, outboundRepo)
		kakaoService := service.NewKakaoService()

		inboundMsg := &model.InboundMessage{
			ID:              "msg-1",
			AccountID:       "acc-1",
			ConversationKey: "conv-1",
			CallbackURL:     nil, // No callback URL
		}
		inboundRepo.On("FindByID", mock.Anything, "msg-1").Return(inboundMsg, nil)

		handler := NewOpenClawHandler(msgService, kakaoService)

		account := &model.Account{ID: "acc-1"}
		body := bytes.NewBufferString(`{"messageId": "msg-1", "response": {"text": "Hello"}}`)
		req := httptest.NewRequest(http.MethodPost, "/openclaw/reply", body)
		req = req.WithContext(withAccount(req.Context(), account))
		rec := httptest.NewRecorder()

		handler.Reply(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "CALLBACK_EXPIRED")
		inboundRepo.AssertExpectations(t)
	})

	t.Run("returns 400 when callback URL is expired", func(t *testing.T) {
		inboundRepo := new(mockInboundRepo)
		outboundRepo := new(mockOutboundRepo)
		msgService := service.NewMessageService(inboundRepo, outboundRepo)
		kakaoService := service.NewKakaoService()

		callbackURL := "https://callback.kakao.com/v1"
		expiresAt := time.Now().Add(-1 * time.Hour) // Expired
		inboundMsg := &model.InboundMessage{
			ID:                "msg-1",
			AccountID:         "acc-1",
			ConversationKey:   "conv-1",
			CallbackURL:       &callbackURL,
			CallbackExpiresAt: &expiresAt,
		}
		inboundRepo.On("FindByID", mock.Anything, "msg-1").Return(inboundMsg, nil)

		handler := NewOpenClawHandler(msgService, kakaoService)

		account := &model.Account{ID: "acc-1"}
		body := bytes.NewBufferString(`{"messageId": "msg-1", "response": {"text": "Hello"}}`)
		req := httptest.NewRequest(http.MethodPost, "/openclaw/reply", body)
		req = req.WithContext(withAccount(req.Context(), account))
		rec := httptest.NewRecorder()

		handler.Reply(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "CALLBACK_EXPIRED")
		inboundRepo.AssertExpectations(t)
	})
}

func TestOpenClawHandler_Routes(t *testing.T) {
	t.Run("registers /reply route", func(t *testing.T) {
		inboundRepo := new(mockInboundRepo)
		outboundRepo := new(mockOutboundRepo)
		msgService := service.NewMessageService(inboundRepo, outboundRepo)
		kakaoService := service.NewKakaoService()

		handler := NewOpenClawHandler(msgService, kakaoService)
		router := handler.Routes()

		// Verify the route is registered by making a request
		req := httptest.NewRequest(http.MethodPost, "/reply", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		// Should get 401 (no account) not 404 (not found)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestReplyRequest_Parsing(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantErr     bool
		wantMsgID   string
		wantHasResp bool
	}{
		{
			name:        "valid request with text response",
			body:        `{"messageId": "msg-123", "response": {"text": "Hello"}}`,
			wantErr:     false,
			wantMsgID:   "msg-123",
			wantHasResp: true,
		},
		{
			name:        "valid request with complex response",
			body:        `{"messageId": "msg-456", "response": {"template": {"type": "basic", "outputs": []}}}`,
			wantErr:     false,
			wantMsgID:   "msg-456",
			wantHasResp: true,
		},
		{
			name:    "invalid json",
			body:    `{invalid}`,
			wantErr: true,
		},
		{
			name:        "missing messageId",
			body:        `{"response": {"text": "Hello"}}`,
			wantErr:     false,
			wantMsgID:   "",
			wantHasResp: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var req struct {
				MessageID string          `json:"messageId"`
				Response  json.RawMessage `json:"response"`
			}
			err := json.NewDecoder(bytes.NewBufferString(tc.body)).Decode(&req)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantMsgID, req.MessageID)
				if tc.wantHasResp {
					assert.NotEmpty(t, req.Response)
				}
			}
		})
	}
}
