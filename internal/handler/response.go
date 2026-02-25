package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/openclaw/relay-server-go/internal/httputil"
	"github.com/openclaw/relay-server-go/internal/model"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	httputil.WriteJSON(w, status, data)
}

func formatTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.Format(time.RFC3339)
}

func formatConversation(conv model.ConversationMapping) map[string]any {
	return map[string]any{
		"conversationKey": conv.ConversationKey,
		"state":           conv.State,
		"pairedAt":        formatTime(conv.PairedAt),
		"lastSeenAt":      conv.LastSeenAt.Format(time.RFC3339),
	}
}

func parseIntParam(s string) (int, error) {
	return strconv.Atoi(s)
}
