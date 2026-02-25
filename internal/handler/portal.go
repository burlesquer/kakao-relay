package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/openclaw/relay-server-go/internal/audit"
	"github.com/openclaw/relay-server-go/internal/middleware"
	"github.com/openclaw/relay-server-go/internal/model"
	"github.com/openclaw/relay-server-go/internal/service"
	"github.com/openclaw/relay-server-go/internal/util"
)

type PortalHandler struct {
	portalService       *service.PortalService
	pairingService      *service.PairingService
	portalAccessService *service.PortalAccessService
	convService         *service.ConversationService
	msgService          *service.MessageService
	adminService        *service.AdminService
	isProduction        bool
}

func NewPortalHandler(
	portalService *service.PortalService,
	pairingService *service.PairingService,
	portalAccessService *service.PortalAccessService,
	convService *service.ConversationService,
	msgService *service.MessageService,
	adminService *service.AdminService,
	isProduction bool,
) *PortalHandler {
	return &PortalHandler{
		portalService:       portalService,
		pairingService:      pairingService,
		portalAccessService: portalAccessService,
		convService:         convService,
		msgService:          msgService,
		adminService:        adminService,
		isProduction:        isProduction,
	}
}

func (h *PortalHandler) PublicRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/api/stats/public", h.GetPublicStats)
	r.Post("/api/auth/code", h.LoginWithCode)
	r.Get("/api/code/stats", h.GetCodeStats)
	r.Get("/api/code/messages", h.GetCodeMessages)

	return r
}

func (h *PortalHandler) AuthenticatedRoutes() chi.Router {
	r := chi.NewRouter()

	r.Post("/api/logout", h.Logout)
	r.Get("/api/me", h.Me)
	r.Get("/api/stats", h.GetStats)
	r.Post("/api/pairing/generate", h.GeneratePairingCode)
	r.Get("/api/connections", h.ListConnections)
	r.Post("/api/connections/{conversationKey}/unpair", h.UnpairConnection)
	r.Patch("/api/connections/{conversationKey}/block", h.BlockConnection)
	r.Get("/api/token", h.GetToken)
	r.Post("/api/token/regenerate", h.RegenerateToken)
	r.Delete("/api/account", h.DeleteAccount)
	r.Get("/api/messages", h.GetMessages)

	return r
}

func (h *PortalHandler) requireUser(w http.ResponseWriter, r *http.Request) *model.PortalUser {
	user := middleware.GetPortalUser(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Not authenticated"})
		return nil
	}
	return user
}

func (h *PortalHandler) Logout(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	cookie, err := r.Cookie(middleware.PortalSessionCookie)
	if err == nil && cookie.Value != "" {
		h.portalService.Logout(r.Context(), cookie.Value)
	}

	audit.LogFromRequest(r, audit.Event{
		Type:      audit.EventLogout,
		UserID:    user.ID,
		AccountID: user.AccountID,
		Details: map[string]interface{}{
			"target": "portal",
		},
	})

	middleware.ClearSessionCookie(w, middleware.PortalSessionCookie, "/portal")
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *PortalHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user": map[string]any{
			"id":        user.ID,
			"email":     user.Email,
			"accountId": user.AccountID,
			"createdAt": user.CreatedAt.Format(time.RFC3339),
		},
	})
}

func (h *PortalHandler) GetPublicStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.adminService.GetStats(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to get public stats")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	publicStats := map[string]any{
		"system": map[string]any{
			"accounts":    stats.Accounts,
			"connections": stats.Mappings,
			"sessions": map[string]int{
				"pending": stats.Sessions.Pending,
				"paired":  stats.Sessions.Paired,
				"total":   stats.Sessions.Total,
			},
		},
		"messages": map[string]any{
			"inbound": map[string]int{
				"queued": stats.Messages.Inbound.Queued,
			},
		},
		"isPublic": true,
	}

	writeJSON(w, http.StatusOK, publicStats)
}

func (h *PortalHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	conversations, err := h.convService.ListByAccountID(r.Context(), user.AccountID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list connections for stats")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	connStats := make([]service.ConnectionStat, len(conversations))
	for i, conv := range conversations {
		lastSeenAt := conv.LastSeenAt
		connStats[i] = service.ConnectionStat{
			State:      string(conv.State),
			LastSeenAt: &lastSeenAt,
		}
	}

	stats, err := h.msgService.GetUserStats(r.Context(), user.AccountID, connStats)
	if err != nil {
		log.Error().Err(err).Msg("failed to get user stats")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

func (h *PortalHandler) GeneratePairingCode(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	var req struct {
		ExpirySeconds int `json:"expirySeconds"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	code, err := h.pairingService.GenerateCode(r.Context(), user.AccountID, req.ExpirySeconds, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate pairing code")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"code":      code.Code,
		"expiresAt": code.ExpiresAt.Format(time.RFC3339),
	})
}

func (h *PortalHandler) ListConnections(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	conversations, err := h.convService.ListByAccountID(r.Context(), user.AccountID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list connections")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	formatted := make([]map[string]any, len(conversations))
	for i, conv := range conversations {
		formatted[i] = formatConversation(conv)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"connections": formatted,
		"total":       len(conversations),
	})
}

func (h *PortalHandler) UnpairConnection(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	conversationKey := chi.URLParam(r, "conversationKey")
	if conversationKey == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Conversation key is required"})
		return
	}

	conv, err := h.convService.FindByKey(r.Context(), conversationKey)
	if err != nil {
		log.Error().Err(err).Msg("failed to find conversation")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}
	if conv == nil || conv.AccountID == nil || *conv.AccountID != user.AccountID {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Connection not found"})
		return
	}

	if err := h.convService.Unpair(r.Context(), conversationKey); err != nil {
		log.Error().Err(err).Msg("failed to unpair connection")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to unpair connection"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *PortalHandler) BlockConnection(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	conversationKey := chi.URLParam(r, "conversationKey")
	if conversationKey == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Conversation key is required"})
		return
	}

	conv, err := h.convService.FindByKey(r.Context(), conversationKey)
	if err != nil {
		log.Error().Err(err).Msg("failed to find conversation")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}
	if conv == nil || conv.AccountID == nil || *conv.AccountID != user.AccountID {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Connection not found"})
		return
	}

	var newState model.PairingState
	if conv.State == model.PairingStateBlocked {
		newState = model.PairingStatePaired
	} else {
		newState = model.PairingStateBlocked
	}

	if err := h.convService.UpdateState(r.Context(), conversationKey, newState, &user.AccountID); err != nil {
		log.Error().Err(err).Msg("failed to update connection state")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update connection state"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"state":   string(newState),
	})
}

func (h *PortalHandler) GetToken(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	account, err := h.portalService.GetAccountByID(r.Context(), user.AccountID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get account")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}
	if account == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Account not found"})
		return
	}

	hasToken := account.RelayTokenHash != nil && *account.RelayTokenHash != ""

	writeJSON(w, http.StatusOK, map[string]any{
		"hasToken":  hasToken,
		"message":   "Plaintext tokens are no longer stored. Use /api/token/regenerate to get a new token.",
		"createdAt": account.CreatedAt.Format(time.RFC3339),
	})
}

func (h *PortalHandler) RegenerateToken(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	account, newToken, err := h.portalService.RegenerateToken(r.Context(), user.AccountID)
	if err != nil {
		log.Error().Err(err).Msg("failed to regenerate token")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to regenerate token"})
		return
	}

	audit.LogFromRequest(r, audit.Event{
		Type:      audit.EventTokenRegenerate,
		UserID:    user.ID,
		AccountID: user.AccountID,
		Details: map[string]interface{}{
			"regenerated_by": "portal_user",
		},
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"token":     newToken,
		"createdAt": account.UpdatedAt.Format(time.RFC3339),
	})
}

func (h *PortalHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	var req struct {
		Confirm string `json:"confirm"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Confirm != "DELETE" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Please confirm deletion by sending {\"confirm\": \"DELETE\"}"})
		return
	}

	audit.LogFromRequest(r, audit.Event{
		Type:      audit.EventAccountDelete,
		UserID:    user.ID,
		AccountID: user.AccountID,
		Details: map[string]interface{}{
			"deleted_by": "self",
		},
	})

	if err := h.portalService.DeleteAccount(r.Context(), user.ID); err != nil {
		log.Error().Err(err).Msg("failed to delete account")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to delete account"})
		return
	}

	middleware.ClearSessionCookie(w, middleware.PortalSessionCookie, "/portal")
	w.WriteHeader(http.StatusNoContent)
}

func (h *PortalHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	msgType := r.URL.Query().Get("type")
	if msgType != "" && msgType != "inbound" && msgType != "outbound" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid type parameter"})
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := parseIntParam(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := parseIntParam(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	result, err := h.msgService.GetMessageHistory(r.Context(), service.MessageHistoryParams{
		AccountID: user.AccountID,
		Type:      msgType,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to get message history")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	messages := result.Messages
	if messages == nil {
		messages = []service.MessageHistoryItem{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"messages": messages,
		"total":    result.Total,
		"hasMore":  result.HasMore,
	})
}

// Code-based authentication handlers

func (h *PortalHandler) LoginWithCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Code is required"})
		return
	}

	// Rate limit check: 5 times per 1 minute per IP
	clientIP := getClientIP(r)
	allowed, resetAt := h.portalAccessService.CheckLoginLimit(r.Context(), clientIP)
	if !allowed {
		secondsLeft := int(time.Until(resetAt).Seconds()) + 1
		log.Warn().
			Str("ip", clientIP).
			Str("code", util.MaskCode(req.Code)).
			Msg("code login rate limit exceeded")

		w.Header().Set("Retry-After", fmt.Sprintf("%d", secondsLeft))
		writeJSON(w, http.StatusTooManyRequests, map[string]string{
			"error": "Too many login attempts. Please try again later.",
		})
		return
	}

	conversationKey, err := h.portalAccessService.VerifyCode(r.Context(), req.Code)
	if err != nil {
		log.Warn().Err(err).Str("code", util.MaskCode(req.Code)).Msg("invalid portal code")
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid or expired code"})
		return
	}

	session, err := h.portalAccessService.CreateCodeSession(conversationKey)
	if err != nil {
		log.Error().Err(err).Msg("failed to create code session")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create session"})
		return
	}

	// Store session in Redis
	if err := h.portalAccessService.StoreSession(r.Context(), session); err != nil {
		log.Error().Err(err).Msg("failed to store session")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create session"})
		return
	}

	// Set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "portal_code_session",
		Value:    session.Token,
		Path:     "/portal",
		MaxAge:   1800, // 30 minutes
		HttpOnly: true,
		Secure:   h.isProduction,
		SameSite: http.SameSiteLaxMode,
	})

	audit.LogFromRequest(r, audit.Event{
		Type: "code_login",
		Details: map[string]interface{}{
			"conversationKey": conversationKey,
			"readOnly":        true,
		},
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"success":         true,
		"conversationKey": conversationKey,
	})
}

func (h *PortalHandler) GetCodeStats(w http.ResponseWriter, r *http.Request) {
	conversationKey := h.getCodeSessionConversationKey(r)
	if conversationKey == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Not authenticated"})
		return
	}

	stats, err := h.msgService.GetConversationStats(r.Context(), conversationKey)
	if err != nil {
		log.Error().Err(err).Msg("failed to get conversation stats")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to fetch stats"})
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

func (h *PortalHandler) GetCodeMessages(w http.ResponseWriter, r *http.Request) {
	conversationKey := h.getCodeSessionConversationKey(r)
	if conversationKey == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Not authenticated"})
		return
	}

	msgType := r.URL.Query().Get("type")
	limit := 20
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscan(l, &limit)
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		fmt.Sscan(o, &offset)
	}

	result, err := h.msgService.GetConversationMessages(r.Context(), service.ConversationMessagesParams{
		ConversationKey: conversationKey,
		Type:            msgType,
		Limit:           limit,
		Offset:          offset,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to get conversation messages")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to fetch messages"})
		return
	}

	messages := result.Messages
	if messages == nil {
		messages = []service.MessageHistoryItem{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"messages": messages,
		"total":    result.Total,
		"hasMore":  result.HasMore,
	})
}

func (h *PortalHandler) getCodeSessionConversationKey(r *http.Request) string {
	cookie, err := r.Cookie("portal_code_session")
	if err != nil {
		return ""
	}

	conversationKey, err := h.portalAccessService.ValidateCodeSession(r.Context(), cookie.Value)
	if err != nil {
		return ""
	}

	return conversationKey
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
