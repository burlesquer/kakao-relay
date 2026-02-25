package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/openclaw/relay-server-go/internal/audit"
	"github.com/openclaw/relay-server-go/internal/middleware"
	"github.com/openclaw/relay-server-go/internal/model"
	"github.com/openclaw/relay-server-go/internal/service"
	"github.com/openclaw/relay-server-go/internal/util"
)

type AdminHandler struct {
	adminService      *service.AdminService
	sessionMiddleware func(http.Handler) http.Handler
	loginRateLimiter  *middleware.LoginRateLimiter
	isProduction      bool
}

func NewAdminHandler(
	adminService *service.AdminService,
	sessionMiddleware func(http.Handler) http.Handler,
	isProduction bool,
) *AdminHandler {
	return &AdminHandler{
		adminService:      adminService,
		sessionMiddleware: sessionMiddleware,
		loginRateLimiter:  middleware.NewLoginRateLimiter(),
		isProduction:      isProduction,
	}
}

func (h *AdminHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.With(h.loginRateLimiter.Handler).Post("/api/login", h.Login)
	r.Post("/api/logout", h.Logout)

	r.Group(func(r chi.Router) {
		r.Use(h.sessionMiddleware)
		r.Get("/api/stats", h.Stats)

		// Accounts
		r.Get("/api/accounts", h.ListAccounts)
		r.Post("/api/accounts", h.CreateAccount)
		r.Get("/api/accounts/{id}", h.GetAccount)
		r.Delete("/api/accounts/{id}", h.DeleteAccount)
		r.Post("/api/accounts/{id}/regenerate-token", h.RegenerateToken)

		// Mappings
		r.Get("/api/mappings", h.ListMappings)
		r.Delete("/api/mappings/{id}", h.DeleteMapping)

		// Messages
		r.Get("/api/messages/inbound", h.ListInboundMessages)
		r.Get("/api/messages/outbound", h.ListOutboundMessages)

		// Users
		r.Get("/api/users", h.ListUsers)
		r.Get("/api/users/{id}", h.GetUser)
		r.Patch("/api/users/{id}", h.UpdateUser)
		r.Delete("/api/users/{id}", h.DeleteUser)

		// Sessions (Plugin Sessions)
		r.Get("/api/sessions", h.ListSessions)
		r.Delete("/api/sessions/{id}", h.DeleteSession)
		r.Post("/api/sessions/{id}/disconnect", h.DisconnectSession)
	})

	return r
}

func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password is required"})
		return
	}

	token, err := h.adminService.Login(r.Context(), req.Password)
	if err != nil {
		log.Error().Err(err).Msg("admin login error")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Login failed"})
		return
	}

	if token == "" {
		audit.LogFromRequest(r, audit.Event{
			Type: audit.EventLoginFailure,
			Details: map[string]interface{}{
				"reason": "invalid_password",
				"target": "admin",
			},
		})
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid password"})
		return
	}

	audit.LogFromRequest(r, audit.Event{
		Type: audit.EventLoginSuccess,
		Details: map[string]interface{}{
			"target": "admin",
		},
	})

	middleware.SetSessionCookie(w, middleware.AdminSessionCookie, token, "/admin", h.isProduction)
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *AdminHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(middleware.AdminSessionCookie)
	if err == nil && cookie.Value != "" {
		h.adminService.Logout(r.Context(), cookie.Value)
	}

	audit.LogFromRequest(r, audit.Event{
		Type: audit.EventLogout,
		Details: map[string]interface{}{
			"target": "admin",
		},
	})

	middleware.ClearSessionCookie(w, middleware.AdminSessionCookie, "/admin")
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *AdminHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.adminService.GetStats(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to get stats")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

func (h *AdminHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	p := ParsePagination(r)

	accounts, err := h.adminService.GetAccounts(r.Context(), p.Limit, p.Offset)
	if err != nil {
		log.Error().Err(err).Msg("failed to list accounts")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": accounts,
		"total": len(accounts),
	})
}

func (h *AdminHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OpenclawUserID     *string `json:"openclawUserId"`
		Mode               string  `json:"mode"`
		RateLimitPerMinute int     `json:"rateLimitPerMinute"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	mode := model.AccountModeRelay
	if req.Mode == "direct" {
		mode = model.AccountModeDirect
	}

	rateLimit := req.RateLimitPerMinute
	if rateLimit <= 0 {
		rateLimit = 60
	}

	account, token, err := h.adminService.CreateAccount(r.Context(), req.OpenclawUserID, mode, rateLimit)
	if err != nil {
		log.Error().Err(err).Msg("failed to create account")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":                 account.ID,
		"openclawUserId":     account.OpenclawUserID,
		"mode":               account.Mode,
		"rateLimitPerMinute": account.RateLimitPerMin,
		"createdAt":          account.CreatedAt,
		"relayToken":         token,
	})
}

func (h *AdminHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	account, err := h.adminService.GetAccountByID(r.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("failed to get account")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	if account == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Account not found"})
		return
	}

	writeJSON(w, http.StatusOK, account)
}

func (h *AdminHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.adminService.DeleteAccount(r.Context(), id); err != nil {
		log.Error().Err(err).Msg("failed to delete account")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	audit.LogFromRequest(r, audit.Event{
		Type:      audit.EventAccountDelete,
		AccountID: id,
		Details: map[string]interface{}{
			"deleted_by": "admin",
		},
	})

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *AdminHandler) RegenerateToken(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	token, err := h.adminService.RegenerateToken(r.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("failed to regenerate token")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	audit.LogFromRequest(r, audit.Event{
		Type:      audit.EventTokenRegenerate,
		AccountID: id,
		Details: map[string]interface{}{
			"regenerated_by": "admin",
		},
	})

	writeJSON(w, http.StatusOK, map[string]string{"relayToken": token})
}

// Mappings

func (h *AdminHandler) ListMappings(w http.ResponseWriter, r *http.Request) {
	p := ParsePagination(r)
	accountID := r.URL.Query().Get("accountId")

	mappings, total, err := h.adminService.GetMappings(r.Context(), p.Limit, p.Offset, accountID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list mappings")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": mappings,
		"total": total,
	})
}

func (h *AdminHandler) DeleteMapping(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.adminService.DeleteMapping(r.Context(), id); err != nil {
		log.Error().Err(err).Msg("failed to delete mapping")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// Messages

var validInboundStatuses = []string{"queued", "delivered", "expired", "failed"}

func (h *AdminHandler) ListInboundMessages(w http.ResponseWriter, r *http.Request) {
	p := ParsePagination(r)
	accountID := r.URL.Query().Get("accountId")
	status := r.URL.Query().Get("status")

	if accountID != "" && !util.IsValidUUID(accountID) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid accountId format"})
		return
	}
	if !util.IsValidEnum(status, validInboundStatuses) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid status value"})
		return
	}

	messages, total, err := h.adminService.GetInboundMessages(r.Context(), p.Limit, p.Offset, accountID, status)
	if err != nil {
		log.Error().Err(err).Msg("failed to list inbound messages")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": messages,
		"total": total,
	})
}

var validOutboundStatuses = []string{"pending", "sent", "failed"}

func (h *AdminHandler) ListOutboundMessages(w http.ResponseWriter, r *http.Request) {
	p := ParsePagination(r)
	accountID := r.URL.Query().Get("accountId")
	status := r.URL.Query().Get("status")

	if accountID != "" && !util.IsValidUUID(accountID) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid accountId format"})
		return
	}
	if !util.IsValidEnum(status, validOutboundStatuses) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid status value"})
		return
	}

	messages, total, err := h.adminService.GetOutboundMessages(r.Context(), p.Limit, p.Offset, accountID, status)
	if err != nil {
		log.Error().Err(err).Msg("failed to list outbound messages")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": messages,
		"total": total,
	})
}

// Users

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	p := ParsePagination(r)

	users, total, err := h.adminService.GetUsers(r.Context(), p.Limit, p.Offset)
	if err != nil {
		log.Error().Err(err).Msg("failed to list users")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": users,
		"total": total,
	})
}

func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	user, err := h.adminService.GetUserByID(r.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("failed to get user")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	if user == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req struct {
		IsActive *bool `json:"isActive"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	user, err := h.adminService.UpdateUser(r.Context(), id, req.IsActive)
	if err != nil {
		log.Error().Err(err).Msg("failed to update user")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	if user == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.adminService.DeleteUser(r.Context(), id); err != nil {
		log.Error().Err(err).Msg("failed to delete user")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	audit.LogFromRequest(r, audit.Event{
		Type:   audit.EventUserDelete,
		UserID: id,
		Details: map[string]interface{}{
			"deleted_by": "admin",
		},
	})

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// Sessions (Plugin Sessions)

var validSessionStatuses = []string{"pending_pairing", "paired", "expired", "disconnected"}

func (h *AdminHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	p := ParsePagination(r)
	status := r.URL.Query().Get("status")

	if !util.IsValidEnum(status, validSessionStatuses) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid status value"})
		return
	}

	sessions, total, err := h.adminService.GetSessions(r.Context(), p.Limit, p.Offset, status)
	if err != nil {
		log.Error().Err(err).Msg("failed to list sessions")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": sessions,
		"total": total,
	})
}

func (h *AdminHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.adminService.DeleteSession(r.Context(), id); err != nil {
		log.Error().Err(err).Msg("failed to delete session")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *AdminHandler) DisconnectSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.adminService.DisconnectSession(r.Context(), id); err != nil {
		log.Error().Err(err).Msg("failed to disconnect session")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}
