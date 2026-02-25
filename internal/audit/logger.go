package audit

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type EventType string

const (
	EventLoginSuccess    EventType = "login_success"
	EventLoginFailure    EventType = "login_failure"
	EventLogout          EventType = "logout"
	EventTokenRegenerate EventType = "token_regenerate"
	EventAccountCreate   EventType = "account_create"
	EventAccountDelete   EventType = "account_delete"
	EventUserDelete      EventType = "user_delete"
	EventRateLimitExceed EventType = "rate_limit_exceeded"
	EventCSRFFailure     EventType = "csrf_failure"
	EventAuthFailure     EventType = "auth_failure"
	EventSessionCreate   EventType = "session_create"
	EventSessionDelete   EventType = "session_delete"
	EventCodeGenerate    EventType = "code_generate"
	EventCodeLogin       EventType = "code_login"
)

type Event struct {
	Type      EventType
	UserID    string
	AccountID string
	IP        string
	UserAgent string
	Details   map[string]interface{}
}

func Log(ctx context.Context, event Event) {
	logger := log.With().
		Str("audit", "security").
		Str("event_type", string(event.Type)).
		Time("timestamp", time.Now()).
		Logger()

	if event.UserID != "" {
		logger = logger.With().Str("user_id", event.UserID).Logger()
	}
	if event.AccountID != "" {
		logger = logger.With().Str("account_id", event.AccountID).Logger()
	}
	if event.IP != "" {
		logger = logger.With().Str("ip", event.IP).Logger()
	}
	if event.UserAgent != "" {
		logger = logger.With().Str("user_agent", event.UserAgent).Logger()
	}

	logEvent := logger.Info()
	for k, v := range event.Details {
		logEvent = addField(logEvent, k, v)
	}
	logEvent.Msg("security audit event")
}

func addField(e *zerolog.Event, key string, value interface{}) *zerolog.Event {
	switch v := value.(type) {
	case string:
		return e.Str(key, v)
	case int:
		return e.Int(key, v)
	case int64:
		return e.Int64(key, v)
	case bool:
		return e.Bool(key, v)
	default:
		return e.Interface(key, v)
	}
}

func LogFromRequest(r *http.Request, event Event) {
	event.IP = getClientIP(r)
	event.UserAgent = r.UserAgent()
	Log(r.Context(), event)
}

func getClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	return r.RemoteAddr
}
