package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"gitlab.tepseg.com/ai/kakao-relay/internal/model"
	"gitlab.tepseg.com/ai/kakao-relay/internal/repository"
	"gitlab.tepseg.com/ai/kakao-relay/internal/service"
	"gitlab.tepseg.com/ai/kakao-relay/internal/sse"
	"gitlab.tepseg.com/ai/kakao-relay/internal/util"
)

type DashboardHandler struct {
	dashboardRepo  repository.DashboardRepository
	accountRepo    repository.AccountRepository
	convRepo       repository.ConversationRepository
	inboundRepo    repository.InboundMessageRepository
	outboundRepo   repository.OutboundMessageRepository
	sessionService *service.SessionService
	messageService *service.MessageService
	broker         *sse.Broker
	indexHTML       []byte
}

func NewDashboardHandler(
	dashboardRepo repository.DashboardRepository,
	accountRepo repository.AccountRepository,
	convRepo repository.ConversationRepository,
	inboundRepo repository.InboundMessageRepository,
	outboundRepo repository.OutboundMessageRepository,
	sessionService *service.SessionService,
	messageService *service.MessageService,
	broker *sse.Broker,
	indexHTML []byte,
) *DashboardHandler {
	return &DashboardHandler{
		dashboardRepo:  dashboardRepo,
		accountRepo:    accountRepo,
		convRepo:       convRepo,
		inboundRepo:    inboundRepo,
		outboundRepo:   outboundRepo,
		sessionService: sessionService,
		messageService: messageService,
		broker:         broker,
		indexHTML:       indexHTML,
	}
}

func (h *DashboardHandler) ServeIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(h.indexHTML)
}

func (h *DashboardHandler) Overview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := h.dashboardRepo.GetOverviewStats(ctx)
	if err != nil {
		log.Error().Err(err).Msg("dashboard: failed to get overview stats")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to get stats"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"accounts":             stats.AccountCount,
		"sessionPending":       stats.SessionPending,
		"sessionPaired":        stats.SessionPaired,
		"sessionTotal":         stats.SessionTotal,
		"conversationPaired":   stats.ConversationPaired,
		"conversationUnpaired": stats.ConversationUnpaired,
		"inboundTotal":         stats.InboundTotal,
		"outboundTotal":        stats.OutboundTotal,
		"outboundFailed":       stats.OutboundFailed,
		"sseClients":           h.broker.TotalClients(),
		"timestamp":            time.Now().UnixMilli(),
	})
}

func (h *DashboardHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	accounts, err := h.accountRepo.FindAll(ctx, 100, 0)
	if err != nil {
		log.Error().Err(err).Msg("dashboard: failed to list accounts")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to list accounts"})
		return
	}

	result := make([]map[string]any, 0, len(accounts))
	for _, acc := range accounts {
		entry := map[string]any{
			"id":                 acc.ID,
			"rateLimitPerMinute": acc.RateLimitPerMin,
			"createdAt":          acc.CreatedAt.Format(time.RFC3339),
			"updatedAt":          acc.UpdatedAt.Format(time.RFC3339),
		}
		if stats, err := h.messageService.GetQuickStats(ctx, acc.ID); err == nil {
			entry["inboundToday"] = stats.InboundToday
			entry["inboundTotal"] = stats.InboundTotal
			entry["outboundToday"] = stats.OutboundToday
			entry["outboundTotal"] = stats.OutboundTotal
			entry["outboundFailed"] = stats.OutboundFailed
		}
		result = append(result, entry)
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *DashboardHandler) AccountConversations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accountID := chi.URLParam(r, "id")

	convs, err := h.convRepo.FindByAccountID(ctx, accountID)
	if err != nil {
		log.Error().Err(err).Msg("dashboard: failed to get conversations")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to get conversations"})
		return
	}

	result := make([]map[string]any, 0, len(convs))
	for _, conv := range convs {
		result = append(result, formatConversation(conv))
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *DashboardHandler) AccountMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accountID := chi.URLParam(r, "id")
	msgType := r.URL.Query().Get("type")
	limit, offset := parsePagination(r)

	if msgType == "outbound" {
		msgs, err := h.outboundRepo.FindByAccountID(ctx, accountID, limit, offset)
		if err != nil {
			log.Error().Err(err).Msg("dashboard: failed to get outbound messages")
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to get messages"})
			return
		}
		if msgs == nil {
			msgs = []model.OutboundMessage{}
		}
		writeJSON(w, http.StatusOK, msgs)
		return
	}

	msgs, err := h.inboundRepo.FindByAccountID(ctx, accountID, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("dashboard: failed to get inbound messages")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to get messages"})
		return
	}
	if msgs == nil {
		msgs = []model.InboundMessage{}
	}
	writeJSON(w, http.StatusOK, msgs)
}

func (h *DashboardHandler) AccountStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accountID := chi.URLParam(r, "id")

	stats, err := h.messageService.GetQuickStats(ctx, accountID)
	if err != nil {
		log.Error().Err(err).Msg("dashboard: failed to get account stats")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to get stats"})
		return
	}

	convs, err := h.convRepo.FindByAccountID(ctx, accountID)
	if err != nil {
		log.Error().Err(err).Msg("dashboard: failed to get conversations for stats")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to get stats"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"inboundToday":   stats.InboundToday,
		"inboundTotal":   stats.InboundTotal,
		"outboundToday":  stats.OutboundToday,
		"outboundTotal":  stats.OutboundTotal,
		"outboundFailed": stats.OutboundFailed,
		"conversations":  len(convs),
		"sseClients":     h.broker.ClientCount(accountID),
	})
}

func (h *DashboardHandler) AccountFailedMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accountID := chi.URLParam(r, "id")

	msgs, err := h.outboundRepo.FindRecentFailedByAccountID(ctx, accountID, 50)
	if err != nil {
		log.Error().Err(err).Msg("dashboard: failed to get failed messages")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to get messages"})
		return
	}
	if msgs == nil {
		msgs = []model.OutboundMessage{}
	}

	writeJSON(w, http.StatusOK, msgs)
}

func (h *DashboardHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	sessions, err := h.sessionService.FindRecent(ctx, limit)
	if err != nil {
		log.Error().Err(err).Msg("dashboard: failed to list sessions")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to list sessions"})
		return
	}
	if sessions == nil {
		sessions = []model.Session{}
	}

	writeJSON(w, http.StatusOK, sessions)
}

func (h *DashboardHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result, err := h.sessionService.CreateSession(ctx)
	if err != nil {
		log.Error().Err(err).Msg("dashboard: failed to create session")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create session"})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *DashboardHandler) RegenerateToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accountID := chi.URLParam(r, "id")

	token, err := util.GenerateToken()
	if err != nil {
		log.Error().Err(err).Msg("dashboard: failed to generate token")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
		return
	}

	tokenHash := util.HashToken(token)
	if _, err := h.accountRepo.UpdateToken(ctx, accountID, tokenHash); err != nil {
		log.Error().Err(err).Msg("dashboard: failed to update token")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to regenerate token"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"relayToken": token})
}

func (h *DashboardHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accountID := chi.URLParam(r, "id")

	if err := h.accountRepo.Delete(ctx, accountID); err != nil {
		log.Error().Err(err).Msg("dashboard: failed to delete account")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to delete account"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *DashboardHandler) DisconnectSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessionID := chi.URLParam(r, "id")

	if err := h.sessionService.Disconnect(ctx, sessionID); err != nil {
		log.Error().Err(err).Msg("dashboard: failed to disconnect session")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to disconnect session"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *DashboardHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessionID := chi.URLParam(r, "id")

	if err := h.sessionService.Delete(ctx, sessionID); err != nil {
		log.Error().Err(err).Msg("dashboard: failed to delete session")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to delete session"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *DashboardHandler) DeleteConversation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	convID := chi.URLParam(r, "convId")

	if err := h.convRepo.Delete(ctx, convID); err != nil {
		log.Error().Err(err).Msg("dashboard: failed to delete conversation")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to delete conversation"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func parsePagination(r *http.Request) (limit, offset int) {
	limit = 20
	offset = 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return limit, offset
}
