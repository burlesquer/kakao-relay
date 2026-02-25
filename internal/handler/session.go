package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/openclaw/relay-server-go/internal/service"
	"github.com/openclaw/relay-server-go/internal/util"
)

type SessionHandler struct {
	sessionService *service.SessionService
}

func NewSessionHandler(sessionService *service.SessionService) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
	}
}

func (h *SessionHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/create", h.CreateSession)
	r.Get("/{sessionToken}/status", h.GetSessionStatus)

	return r
}

// POST /v1/sessions/create
func (h *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result, err := h.sessionService.CreateSession(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to create session")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create session"})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GET /v1/sessions/{sessionToken}/status
func (h *SessionHandler) GetSessionStatus(w http.ResponseWriter, r *http.Request) {
	sessionToken := chi.URLParam(r, "sessionToken")
	if sessionToken == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Session token is required"})
		return
	}

	ctx := r.Context()
	tokenHash := util.HashToken(sessionToken)

	result, err := h.sessionService.GetStatus(ctx, tokenHash)
	if err != nil {
		log.Error().Err(err).Msg("failed to get session status")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	if result == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Session not found"})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

