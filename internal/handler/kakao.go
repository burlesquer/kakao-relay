package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/openclaw/relay-server-go/internal/audit"
	"github.com/openclaw/relay-server-go/internal/model"
	"github.com/openclaw/relay-server-go/internal/service"
	"github.com/openclaw/relay-server-go/internal/sse"
)

type Command struct {
	Type string // PAIR, UNPAIR, STATUS, HELP, CODE
	Code string
}

func parseCommand(utterance string) *Command {
	trimmed := strings.TrimSpace(utterance)

	if strings.HasPrefix(trimmed, "/pair ") {
		code := strings.ToUpper(strings.TrimSpace(trimmed[6:]))
		if code != "" {
			return &Command{Type: "PAIR", Code: code}
		}
	}

	if trimmed == "/unpair" {
		return &Command{Type: "UNPAIR"}
	}

	if trimmed == "/status" {
		return &Command{Type: "STATUS"}
	}

	if trimmed == "/help" {
		return &Command{Type: "HELP"}
	}

	if trimmed == "/code" {
		return &Command{Type: "CODE"}
	}

	return nil
}

type KakaoHandler struct {
	convService         *service.ConversationService
	sessionService      *service.SessionService
	messageService      *service.MessageService
	portalAccessService *service.PortalAccessService
	broker              *sse.Broker
	callbackTTL         time.Duration
	portalBaseURL       string
}

func NewKakaoHandler(
	convService *service.ConversationService,
	sessionService *service.SessionService,
	messageService *service.MessageService,
	portalAccessService *service.PortalAccessService,
	broker *sse.Broker,
	callbackTTL time.Duration,
	portalBaseURL string,
) *KakaoHandler {
	return &KakaoHandler{
		convService:         convService,
		sessionService:      sessionService,
		messageService:      messageService,
		portalAccessService: portalAccessService,
		broker:              broker,
		callbackTTL:         callbackTTL,
		portalBaseURL:       portalBaseURL,
	}
}

func (h *KakaoHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	var req KakaoWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("invalid kakao webhook request")
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	channelID := req.GetChannelID()
	userKey := req.GetPlusfriendUserKey()
	utterance := req.UserRequest.Utterance
	callbackURL := req.UserRequest.CallbackURL

	conversationKey := service.BuildConversationKey(channelID, userKey)

	log.Info().
		Str("conversationKey", conversationKey).
		Str("utterance", truncate(utterance, 50)).
		Bool("hasCallback", callbackURL != "").
		Msg("received kakao webhook")

	var callbackURLPtr *string
	var callbackExpiresAt *time.Time
	if callbackURL != "" {
		callbackURLPtr = &callbackURL
		expires := time.Now().Add(h.callbackTTL)
		callbackExpiresAt = &expires
	}

	ctx := r.Context()

	conv, err := h.convService.FindOrCreate(ctx, channelID, userKey, callbackURLPtr, callbackExpiresAt)
	if err != nil {
		log.Error().Err(err).Msg("failed to find or create conversation")
		writeJSON(w, http.StatusOK, NewCallbackResponse())
		return
	}

	cmd := parseCommand(utterance)
	if cmd != nil {
		response := h.handleCommand(r, cmd, conv, conversationKey)
		writeJSON(w, http.StatusOK, response)
		return
	}

	if conv.State != model.PairingStatePaired || conv.AccountID == nil {
		writeJSON(w, http.StatusOK, NewTextResponse(
			"OpenClaw에 연결되지 않았습니다.\n\n"+
				"연결하려면 페어링 코드를 받은 후:\n"+
				"/pair <코드>\n\n"+
				"를 입력해주세요.\n\n"+
				"도움말: /help",
		))
		return
	}

	normalizedMsg, _ := json.Marshal(map[string]string{
		"userId":    userKey,
		"text":      utterance,
		"channelId": channelID,
	})

	msg, err := h.messageService.CreateInbound(ctx, service.CreateInboundParams{
		AccountID:         *conv.AccountID,
		ConversationKey:   conversationKey,
		KakaoPayload:      req.ToJSON(),
		NormalizedMessage: normalizedMsg,
		CallbackURL:       callbackURLPtr,
		CallbackExpiresAt: callbackExpiresAt,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create inbound message")
		writeJSON(w, http.StatusOK, NewCallbackResponse())
		return
	}

	sseData := msg.ToSSEEventData()
	log.Debug().
		Str("messageId", msg.ID).
		RawJSON("sseEventData", sseData).
		Msg("publishing sse message event")

	if err := h.broker.Publish(ctx, *conv.AccountID, sse.Event{
		Type: "message",
		Data: sseData,
	}); err != nil {
		log.Warn().Err(err).Msg("failed to publish message event")
	}

	writeJSON(w, http.StatusOK, NewCallbackResponse())
}

func (h *KakaoHandler) handleCommand(r *http.Request, cmd *Command, conv *model.ConversationMapping, conversationKey string) *KakaoResponse {
	ctx := r.Context()

	switch cmd.Type {
	case "PAIR":
		if cmd.Code == "" {
			return NewTextResponse("페어링 코드를 입력해주세요.\n\n예: /pair ABCD-1234")
		}

		if conv.State == model.PairingStatePaired {
			return NewTextResponse(
				"이미 OpenClaw에 연결되어 있습니다.\n\n" +
					"다른 봇에 연결하려면 먼저 /unpair 로 연결을 해제하세요.",
			)
		}

		result := h.sessionService.VerifyPairingCode(ctx, cmd.Code, conversationKey)
		if !result.Success {
			errorMessages := map[string]string{
				"INVALID_CODE":   "❌ 유효하지 않은 코드입니다.\n\n코드를 다시 확인해주세요.",
				"INTERNAL_ERROR": "❌ 오류가 발생했습니다. 다시 시도해주세요.",
			}
			msg := errorMessages[result.Error]
			if msg == "" {
				msg = "페어링에 실패했습니다."
			}
			return NewTextResponse(msg)
		}

		// Update conversation state
		if err := h.convService.UpdateState(ctx, conversationKey, model.PairingStatePaired, &result.AccountID); err != nil {
			log.Error().Err(err).Msg("failed to update conversation state after session pairing")
		}

		// Publish pairing_complete event
		session, err := h.sessionService.FindByID(ctx, result.SessionID)
		if err == nil && session != nil {
			if err := h.sessionService.PublishPairingComplete(ctx, session, conversationKey); err != nil {
				log.Warn().Err(err).Msg("failed to publish pairing_complete event")
			}
		}

		return NewTextResponse("✅ OpenClaw에 연결되었습니다!\n\n이제 자유롭게 대화를 시작하세요.")

	case "UNPAIR":
		if conv.State != model.PairingStatePaired {
			return NewTextResponse("연결된 OpenClaw가 없습니다.")
		}

		if err := h.convService.UpdateState(ctx, conversationKey, model.PairingStateUnpaired, nil); err != nil {
			log.Error().Err(err).Msg("failed to unpair")
			return NewTextResponse("연결 해제에 실패했습니다. 다시 시도해주세요.")
		}

		return NewTextResponse("연결이 해제되었습니다.\n\n다시 연결하려면 /pair <코드>를 사용하세요.")

	case "STATUS":
		if conv.State == model.PairingStatePaired && conv.AccountID != nil {
			pairedAt := "알 수 없음"
			if conv.PairedAt != nil {
				pairedAt = conv.PairedAt.Format("2006-01-02 15:04:05")
			}

			stats, err := h.messageService.GetQuickStats(ctx, *conv.AccountID)
			if err != nil {
				log.Error().Err(err).Msg("failed to get quick stats for status command")
				return NewTextResponse("✅ 연결됨\n\n연결 시간: " + pairedAt)
			}

			return NewTextResponse(fmt.Sprintf(
				"✅ 연결됨\n\n"+
					"📊 오늘 통계\n"+
					"• 수신: %d건\n"+
					"• 발신: %d건 (실패 %d)\n\n"+
					"📈 전체 통계\n"+
					"• 총 수신: %d건\n"+
					"• 총 발신: %d건\n\n"+
					"연결 시간: %s",
				stats.InboundToday,
				stats.OutboundToday,
				stats.OutboundFailed,
				stats.InboundTotal,
				stats.OutboundTotal,
				pairedAt,
			))
		}
		return NewTextResponse("❌ 연결되지 않음\n\n/pair <코드>로 연결하세요.")

	case "CODE":
		if conv.State != model.PairingStatePaired {
			return NewTextResponse(
				"포털 접속 코드는 연결된 대화에서만 발급할 수 있습니다.\n\n" +
					"먼저 /pair <코드>로 연결하세요.",
			)
		}

		// Rate limit check: 3 times per 5 minutes
		allowed, resetAt := h.portalAccessService.CheckCodeGenerationLimit(ctx, conversationKey)
		if !allowed {
			minutesLeft := int(time.Until(resetAt).Minutes()) + 1
			return NewTextResponse(fmt.Sprintf(
				"⏱️ 코드 생성 한도 초과\n\n"+
					"5분에 최대 3회까지 코드를 생성할 수 있습니다.\n\n"+
					"%d분 후 다시 시도해주세요.",
				minutesLeft,
			))
		}

		code, err := h.portalAccessService.GenerateCode(ctx, conversationKey)
		if err != nil {
			log.Error().Err(err).Msg("failed to generate portal access code")
			return NewTextResponse("코드 생성에 실패했습니다. 잠시 후 다시 시도해주세요.")
		}

		// Audit log
		auditEvent := audit.Event{
			Type: audit.EventCodeGenerate,
			Details: map[string]interface{}{
				"conversationKey": conversationKey,
				"code":            code.Code,
				"expiresAt":       code.ExpiresAt,
			},
		}
		if conv.AccountID != nil {
			auditEvent.AccountID = *conv.AccountID
		}
		audit.LogFromRequest(r, auditEvent)

		expiresIn := int(time.Until(code.ExpiresAt).Minutes())
		msg := fmt.Sprintf(
			"🔑 포털 접속 코드\n\n"+
				"코드: %s\n"+
				"유효시간: %d분\n\n"+
				"이 코드로 포털에서 대화 내역과 통계를 확인할 수 있습니다.",
			code.Code,
			expiresIn,
		)
		if h.portalBaseURL != "" {
			msg += fmt.Sprintf("\n\n포털 주소:\n%s/portal/code", h.portalBaseURL)
		}
		return NewTextResponse(msg)

	case "HELP":
		return NewTextResponse(
			"📖 도움말\n\n" +
				"이 봇은 OpenClaw AI 에이전트와 연결하는 중계 서비스입니다.\n\n" +
				"명령어:\n" +
				"• /pair <코드> - OpenClaw에 연결\n" +
				"• /unpair - 연결 해제\n" +
				"• /status - 연결 상태 확인\n" +
				"• /code - 포털 접속 코드 발급\n" +
				"• /help - 이 도움말",
		)

	default:
		return NewTextResponse("알 수 없는 명령어입니다. /help를 입력해 도움말을 확인하세요.")
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
