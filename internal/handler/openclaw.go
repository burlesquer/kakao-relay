package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	apperrors "github.com/openclaw/relay-server-go/internal/errors"
	"github.com/openclaw/relay-server-go/internal/httputil"
	"github.com/openclaw/relay-server-go/internal/middleware"
	"github.com/openclaw/relay-server-go/internal/model"
	"github.com/openclaw/relay-server-go/internal/service"
)

type OpenClawHandler struct {
	messageService *service.MessageService
	kakaoService   *service.KakaoService
}

func NewOpenClawHandler(
	messageService *service.MessageService,
	kakaoService *service.KakaoService,
) *OpenClawHandler {
	return &OpenClawHandler{
		messageService: messageService,
		kakaoService:   kakaoService,
	}
}

func (h *OpenClawHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/reply", h.Reply)
	return r
}

// POST /openclaw/reply
// Core API: Send reply to Kakao user.
func (h *OpenClawHandler) Reply(w http.ResponseWriter, r *http.Request) {
	account := middleware.GetAccount(r.Context())
	if account == nil {
		httputil.WriteError(w, apperrors.SessionNotPaired())
		return
	}

	var req struct {
		MessageID string          `json:"messageId"`
		Response  json.RawMessage `json:"response"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, apperrors.ValidationError("Invalid request body"))
		return
	}

	if req.MessageID == "" {
		httputil.WriteError(w, apperrors.MissingRequired("messageId"))
		return
	}

	ctx := r.Context()

	inbound, err := h.messageService.FindInboundByID(ctx, req.MessageID)
	if err != nil {
		log.Error().Err(err).Msg("failed to find inbound message")
		httputil.WriteError(w, apperrors.Database(err))
		return
	}

	if inbound == nil || inbound.AccountID != account.ID {
		httputil.WriteError(w, apperrors.NotFound("Message"))
		return
	}

	hasValidCallback := inbound.CallbackURL != nil &&
		(inbound.CallbackExpiresAt == nil || inbound.CallbackExpiresAt.After(time.Now()))

	if !hasValidCallback {
		log.Warn().
			Str("messageId", req.MessageID).
			Bool("hasCallbackUrl", inbound.CallbackURL != nil).
			Msg("no valid callback URL for reply")
		httputil.WriteError(w, apperrors.CallbackExpired())
		return
	}

	outbound, err := h.messageService.CreateOutbound(ctx, model.CreateOutboundMessageParams{
		AccountID:        account.ID,
		InboundMessageID: &req.MessageID,
		ConversationKey:  inbound.ConversationKey,
		KakaoTarget:      json.RawMessage("{}"),
		ResponsePayload:  req.Response,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create outbound message")
		httputil.WriteError(w, apperrors.Database(err))
		return
	}

	var responsePayload any
	json.Unmarshal(req.Response, &responsePayload)

	if err := h.kakaoService.SendCallback(ctx, *inbound.CallbackURL, responsePayload); err != nil {
		h.messageService.MarkOutboundFailed(ctx, outbound.ID, err.Error())
		log.Error().
			Err(err).
			Str("outboundId", outbound.ID).
			Str("messageId", req.MessageID).
			Msg("failed to send callback to Kakao")
		httputil.WriteError(w, apperrors.CallbackFailed("Kakao callback failed"))
		return
	}

	h.messageService.MarkOutboundSent(ctx, outbound.ID)

	deliveredAt := time.Now().UnixMilli()

	log.Info().
		Str("outboundId", outbound.ID).
		Str("messageId", req.MessageID).
		Str("accountId", account.ID).
		Msg("reply sent to Kakao")

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"success":     true,
		"deliveredAt": deliveredAt,
	})
}
