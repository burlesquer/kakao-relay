package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/openclaw/relay-server-go/internal/model"
	"github.com/openclaw/relay-server-go/internal/repository"
	"github.com/openclaw/relay-server-go/internal/util"
)

type mockAccountRepo struct {
	findByTokenHashFunc func(ctx context.Context, tokenHash string) (*model.Account, error)
	findByIDFunc        func(ctx context.Context, id string) (*model.Account, error)
}

type mockSessionRepo struct {
	findByTokenHashFunc func(ctx context.Context, tokenHash string) (*model.Session, error)
}

func (m *mockAccountRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*model.Account, error) {
	if m.findByTokenHashFunc != nil {
		return m.findByTokenHashFunc(ctx, tokenHash)
	}
	return nil, nil
}

func (m *mockAccountRepo) FindByID(ctx context.Context, id string) (*model.Account, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockSessionRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*model.Session, error) {
	if m.findByTokenHashFunc != nil {
		return m.findByTokenHashFunc(ctx, tokenHash)
	}
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
	return 0, nil
}

func (m *mockSessionRepo) CountPendingByIP(ctx context.Context, ip string, since time.Time) (int, error) {
	return 0, nil
}

func (m *mockSessionRepo) MarkDisconnected(ctx context.Context, id string) error {
	return nil
}

func (m *mockSessionRepo) WithTx(tx *sqlx.Tx) repository.SessionRepository {
	return m
}

func (m *mockAccountRepo) FindAll(ctx context.Context, limit, offset int) ([]model.Account, error) {
	return nil, nil
}

func (m *mockAccountRepo) Create(ctx context.Context, params model.CreateAccountParams) (*model.Account, error) {
	return nil, nil
}

func (m *mockAccountRepo) Update(ctx context.Context, id string, params model.UpdateAccountParams) (*model.Account, error) {
	return nil, nil
}

func (m *mockAccountRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockAccountRepo) Count(ctx context.Context) (int, error) {
	return 0, nil
}

func (m *mockAccountRepo) UpdateToken(ctx context.Context, id, tokenHash string) (*model.Account, error) {
	return nil, nil
}

func (m *mockAccountRepo) WithTx(tx *sqlx.Tx) repository.AccountRepository {
	return m
}

func TestAuthMiddleware(t *testing.T) {
	testAccount := &model.Account{
		ID:              "acc-123",
		Mode:            model.AccountModeRelay,
		RateLimitPerMin: 60,
	}
	accountID := "acc-123"
	testSession := &model.Session{
		ID:        "sess-123",
		Status:    model.SessionStatusPaired,
		AccountID: &accountID,
	}
	validToken := "valid-token"
	validTokenHash := util.HashToken(validToken)

	t.Run("allows request with valid paired session token", func(t *testing.T) {
		accountRepo := &mockAccountRepo{
			findByIDFunc: func(ctx context.Context, id string) (*model.Account, error) {
				if id == accountID {
					return testAccount, nil
				}
				return nil, nil
			},
		}
		sessionRepo := &mockSessionRepo{
			findByTokenHashFunc: func(ctx context.Context, tokenHash string) (*model.Session, error) {
				if tokenHash == validTokenHash {
					return testSession, nil
				}
				return nil, nil
			},
		}

		middleware := NewAuthMiddleware(accountRepo, sessionRepo)
		handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			account := GetAccount(r.Context())
			require.NotNil(t, account)
			assert.Equal(t, "acc-123", account.ID)
			session := GetSession(r.Context())
			require.NotNil(t, session)
			assert.Equal(t, "sess-123", session.ID)
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("allows request with query token", func(t *testing.T) {
		accountRepo := &mockAccountRepo{
			findByIDFunc: func(ctx context.Context, id string) (*model.Account, error) {
				return testAccount, nil
			},
		}
		sessionRepo := &mockSessionRepo{
			findByTokenHashFunc: func(ctx context.Context, tokenHash string) (*model.Session, error) {
				if tokenHash == validTokenHash {
					return testSession, nil
				}
				return nil, nil
			},
		}

		middleware := NewAuthMiddleware(accountRepo, sessionRepo)
		handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test?token="+validToken, nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("rejects request without token", func(t *testing.T) {
		accountRepo := &mockAccountRepo{}
		sessionRepo := &mockSessionRepo{}
		middleware := NewAuthMiddleware(accountRepo, sessionRepo)
		handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called")
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("rejects request with invalid token", func(t *testing.T) {
		accountRepo := &mockAccountRepo{}
		sessionRepo := &mockSessionRepo{
			findByTokenHashFunc: func(ctx context.Context, tokenHash string) (*model.Session, error) {
				return nil, nil
			},
		}

		middleware := NewAuthMiddleware(accountRepo, sessionRepo)
		handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called")
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("returns 500 on database error", func(t *testing.T) {
		accountRepo := &mockAccountRepo{}
		sessionRepo := &mockSessionRepo{
			findByTokenHashFunc: func(ctx context.Context, tokenHash string) (*model.Session, error) {
				return nil, errors.New("database error")
			},
		}

		middleware := NewAuthMiddleware(accountRepo, sessionRepo)
		handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called")
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("allows pending session without account", func(t *testing.T) {
		pendingSession := &model.Session{
			ID:     "sess-pending",
			Status: model.SessionStatusPendingPairing,
		}
		accountRepo := &mockAccountRepo{}
		sessionRepo := &mockSessionRepo{
			findByTokenHashFunc: func(ctx context.Context, tokenHash string) (*model.Session, error) {
				return pendingSession, nil
			},
		}

		middleware := NewAuthMiddleware(accountRepo, sessionRepo)
		handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := GetSession(r.Context())
			require.NotNil(t, session)
			assert.Equal(t, "sess-pending", session.ID)
			account := GetAccount(r.Context())
			assert.Nil(t, account)
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestGetAccount(t *testing.T) {
	t.Run("returns account from context", func(t *testing.T) {
		account := &model.Account{ID: "test-id"}
		ctx := context.WithValue(context.Background(), AccountContextKey, account)

		result := GetAccount(ctx)

		assert.NotNil(t, result)
		assert.Equal(t, "test-id", result.ID)
	})

	t.Run("returns nil when no account in context", func(t *testing.T) {
		result := GetAccount(context.Background())
		assert.Nil(t, result)
	})
}
