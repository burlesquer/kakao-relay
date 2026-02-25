package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/openclaw/relay-server-go/internal/model"
	"github.com/openclaw/relay-server-go/internal/repository"
	"github.com/openclaw/relay-server-go/internal/util"
)

type contextKey string

const AccountContextKey contextKey = "account"
const SessionContextKey contextKey = "session"

func GetAccount(ctx context.Context) *model.Account {
	if account, ok := ctx.Value(AccountContextKey).(*model.Account); ok {
		return account
	}
	return nil
}

func GetSession(ctx context.Context) *model.Session {
	if session, ok := ctx.Value(SessionContextKey).(*model.Session); ok {
		return session
	}
	return nil
}

type AuthMiddleware struct {
	accountRepo repository.AccountRepository
	sessionRepo repository.SessionRepository
}

func NewAuthMiddleware(
	accountRepo repository.AccountRepository,
	sessionRepo repository.SessionRepository,
) *AuthMiddleware {
	return &AuthMiddleware{
		accountRepo: accountRepo,
		sessionRepo: sessionRepo,
	}
}

func (m *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Missing authentication token",
			})
			return
		}

		tokenHash := util.HashToken(token)
		ctx := r.Context()

		session, err := m.sessionRepo.FindByTokenHash(ctx, tokenHash)
		if err != nil {
			log.Error().Err(err).Msg("auth middleware: session lookup error")
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Authentication failed",
			})
			return
		}

		if session == nil {
			log.Warn().Msg("auth middleware: invalid token attempt")
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Invalid token",
			})
			return
		}

		ctx = context.WithValue(ctx, SessionContextKey, session)

		// If session is paired, also add the linked account
		if session.Status == model.SessionStatusPaired && session.AccountID != nil {
			linkedAccount, err := m.accountRepo.FindByID(ctx, *session.AccountID)
			if err == nil && linkedAccount != nil {
				ctx = context.WithValue(ctx, AccountContextKey, linkedAccount)
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractToken(r *http.Request) string {
	if token := r.URL.Query().Get("token"); token != "" {
		return token
	}

	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}
