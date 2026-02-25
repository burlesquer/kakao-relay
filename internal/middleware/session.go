package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/openclaw/relay-server-go/internal/model"
	"github.com/openclaw/relay-server-go/internal/repository"
	"github.com/openclaw/relay-server-go/internal/util"
)

const (
	AdminSessionCookie  = "admin_session"
	PortalSessionCookie = "portal_session"
	SessionMaxAge       = 24 * time.Hour
)

const (
	AdminSessionContextKey  contextKey = "adminSession"
	PortalUserContextKey    contextKey = "portalUser"
	PortalSessionContextKey contextKey = "portalSession"
)

func GetAdminSession(ctx context.Context) *model.AdminSession {
	if session, ok := ctx.Value(AdminSessionContextKey).(*model.AdminSession); ok {
		return session
	}
	return nil
}

func GetPortalUser(ctx context.Context) *model.PortalUser {
	if user, ok := ctx.Value(PortalUserContextKey).(*model.PortalUser); ok {
		return user
	}
	return nil
}

// Admin Session Middleware

type AdminSessionMiddleware struct {
	sessionRepo       repository.AdminSessionRepository
	adminPasswordHash string
	sessionSecret     string
}

func NewAdminSessionMiddleware(
	sessionRepo repository.AdminSessionRepository,
	adminPasswordHash, sessionSecret string,
) *AdminSessionMiddleware {
	return &AdminSessionMiddleware{
		sessionRepo:       sessionRepo,
		adminPasswordHash: adminPasswordHash,
		sessionSecret:     sessionSecret,
	}
}

func (m *AdminSessionMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.adminPasswordHash == "" {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"error": "Admin not configured",
			})
			return
		}

		cookie, err := r.Cookie(AdminSessionCookie)
		if err != nil || cookie.Value == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Unauthorized",
			})
			return
		}

		tokenHash := hashSessionToken(cookie.Value, m.sessionSecret)
		session, err := m.sessionRepo.FindByTokenHash(r.Context(), tokenHash)
		if err != nil {
			log.Error().Err(err).Msg("admin session middleware: database error")
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Session validation failed",
			})
			return
		}

		if session == nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Unauthorized",
			})
			return
		}

		ctx := context.WithValue(r.Context(), AdminSessionContextKey, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AdminSessionMiddleware) ValidatePassword(password string) bool {
	return util.CheckPasswordHash(password, m.adminPasswordHash)
}

// Portal Session Middleware

type PortalSessionMiddleware struct {
	sessionRepo repository.PortalSessionRepository
	userRepo    repository.PortalUserRepository
	sessionSecret string
}

func NewPortalSessionMiddleware(
	sessionRepo repository.PortalSessionRepository,
	userRepo repository.PortalUserRepository,
	sessionSecret string,
) *PortalSessionMiddleware {
	return &PortalSessionMiddleware{
		sessionRepo:   sessionRepo,
		userRepo:      userRepo,
		sessionSecret: sessionSecret,
	}
}

func (m *PortalSessionMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(PortalSessionCookie)
		if err != nil || cookie.Value == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Unauthorized",
			})
			return
		}

		tokenHash := hashSessionToken(cookie.Value, m.sessionSecret)
		session, err := m.sessionRepo.FindByTokenHash(r.Context(), tokenHash)
		if err != nil {
			log.Error().Err(err).Msg("portal session middleware: database error")
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Session validation failed",
			})
			return
		}

		if session == nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Unauthorized",
			})
			return
		}

		user, err := m.userRepo.FindByID(r.Context(), session.UserID)
		if err != nil || user == nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Unauthorized",
			})
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, PortalSessionContextKey, session)
		ctx = context.WithValue(ctx, PortalUserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func hashSessionToken(token, secret string) string {
	return util.HmacSHA256(secret, token)
}

func SetSessionCookie(w http.ResponseWriter, name, token string, path string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    token,
		Path:     path,
		MaxAge:   int(SessionMaxAge.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func ClearSessionCookie(w http.ResponseWriter, name, path string) {
	http.SetCookie(w, &http.Cookie{
		Name:   name,
		Value:  "",
		Path:   path,
		MaxAge: -1,
	})
}
