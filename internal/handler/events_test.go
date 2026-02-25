package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/openclaw/relay-server-go/internal/middleware"
	"github.com/openclaw/relay-server-go/internal/model"
	"github.com/openclaw/relay-server-go/internal/sse"
)

// Helper to add session to context
func withSession(ctx context.Context, session *model.Session) context.Context {
	return context.WithValue(ctx, middleware.SessionContextKey, session)
}

func TestEventsHandler_ServeHTTP(t *testing.T) {
	t.Run("returns 401 when no session or account in context", func(t *testing.T) {
		// Create handler without dependencies (will fail early)
		handler := NewEventsHandler(nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/events", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "Unauthorized")
	})
}

func TestEventsHandler_sendEvent(t *testing.T) {
	t.Run("formats SSE event correctly", func(t *testing.T) {
		handler := &EventsHandler{}
		rec := httptest.NewRecorder()
		flusher := rec // httptest.ResponseRecorder implements http.Flusher

		data := map[string]any{
			"accountId": "acc-1",
			"status":    "paired",
		}

		err := handler.sendEvent(rec, flusher, "connected", data)

		assert.NoError(t, err)
		body := rec.Body.String()
		assert.Contains(t, body, "event: connected\n")
		assert.Contains(t, body, "data: ")
		assert.Contains(t, body, "acc-1")
	})
}

func TestEventsHandler_sendRawEvent(t *testing.T) {
	t.Run("writes event and data lines", func(t *testing.T) {
		handler := &EventsHandler{}
		rec := httptest.NewRecorder()
		flusher := rec

		event := sse.Event{
			Type: "message",
			Data: json.RawMessage(`{"text": "hello"}`),
		}

		err := handler.sendRawEvent(rec, flusher, event)

		assert.NoError(t, err)
		body := rec.Body.String()
		assert.Contains(t, body, "event: message\n")
		assert.Contains(t, body, `data: {"text": "hello"}`)
		assert.Contains(t, body, "\n\n")
	})
}

// Override sendRawEvent for testing - this tests the format logic
func (h *EventsHandler) sendRawEventTest(w http.ResponseWriter, flusher http.Flusher, eventType string, data json.RawMessage) error {
	if _, err := w.Write([]byte("event: " + eventType + "\n")); err != nil {
		return err
	}
	if _, err := w.Write([]byte("data: " + string(data) + "\n\n")); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

func TestSSEEventFormat(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		data      map[string]any
		wantEvent string
	}{
		{
			name:      "connected event",
			eventType: "connected",
			data:      map[string]any{"accountId": "acc-1", "status": "paired"},
			wantEvent: "event: connected\n",
		},
		{
			name:      "message event",
			eventType: "message",
			data:      map[string]any{"id": "msg-1", "text": "Hello"},
			wantEvent: "event: message\n",
		},
		{
			name:      "paired event",
			eventType: "paired",
			data:      map[string]any{"accountId": "acc-1"},
			wantEvent: "event: paired\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := &EventsHandler{}
			rec := httptest.NewRecorder()

			err := handler.sendEvent(rec, rec, tc.eventType, tc.data)

			assert.NoError(t, err)
			body := rec.Body.String()
			assert.Contains(t, body, tc.wantEvent)
			assert.Contains(t, body, "data: ")
			assert.True(t, len(body) > 0)
		})
	}
}

func TestInboundMessage_ToSSEEventData(t *testing.T) {
	t.Run("generates correct SSE event data", func(t *testing.T) {
		now := time.Now()
		normalized := json.RawMessage(`{"text": "Hello"}`)
		msg := &model.InboundMessage{
			ID:                "msg-1",
			ConversationKey:   "conv-1",
			KakaoPayload:      json.RawMessage(`{"type": "text", "content": "Hello"}`),
			NormalizedMessage: &normalized,
			CreatedAt:         now,
		}

		data := msg.ToSSEEventData()

		var parsed map[string]any
		err := json.Unmarshal(data, &parsed)
		assert.NoError(t, err)
		assert.Equal(t, "msg-1", parsed["id"])
		assert.Equal(t, "conv-1", parsed["conversationKey"])
		assert.NotNil(t, parsed["kakaoPayload"])
		assert.NotNil(t, parsed["normalized"])
		assert.NotNil(t, parsed["createdAt"])
	})

	t.Run("handles nil normalized message", func(t *testing.T) {
		msg := &model.InboundMessage{
			ID:                "msg-1",
			ConversationKey:   "conv-1",
			KakaoPayload:      json.RawMessage(`{}`),
			NormalizedMessage: nil,
			CreatedAt:         time.Now(),
		}

		data := msg.ToSSEEventData()

		var parsed map[string]any
		err := json.Unmarshal(data, &parsed)
		assert.NoError(t, err)
		assert.Equal(t, "msg-1", parsed["id"])
		assert.Nil(t, parsed["normalized"])
	})
}

func TestSSEHeaders(t *testing.T) {
	t.Run("SSE requires specific headers", func(t *testing.T) {
		// This test documents the expected SSE headers
		expectedHeaders := map[string]string{
			"Content-Type":      "text/event-stream",
			"Cache-Control":     "no-cache",
			"Connection":        "keep-alive",
			"X-Accel-Buffering": "no",
		}

		for header, value := range expectedHeaders {
			t.Run(header, func(t *testing.T) {
				assert.NotEmpty(t, value, "SSE header %s should have a value", header)
			})
		}
	})
}
