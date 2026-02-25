package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name      string
		utterance string
		expected  *Command
	}{
		{
			name:      "parse /pair command with code",
			utterance: "/pair ABCD-1234",
			expected:  &Command{Type: "PAIR", Code: "ABCD-1234"},
		},
		{
			name:      "parse /pair command with lowercase code",
			utterance: "/pair abcd-1234",
			expected:  &Command{Type: "PAIR", Code: "ABCD-1234"},
		},
		{
			name:      "parse /pair command with extra spaces",
			utterance: "/pair   ABCD-1234  ",
			expected:  &Command{Type: "PAIR", Code: "ABCD-1234"},
		},
		{
			name:      "reject /pair without code",
			utterance: "/pair ",
			expected:  nil,
		},
		{
			name:      "reject /pair without space",
			utterance: "/pairABCD",
			expected:  nil,
		},
		{
			name:      "parse /unpair command",
			utterance: "/unpair",
			expected:  &Command{Type: "UNPAIR"},
		},
		{
			name:      "parse /status command",
			utterance: "/status",
			expected:  &Command{Type: "STATUS"},
		},
		{
			name:      "parse /help command",
			utterance: "/help",
			expected:  &Command{Type: "HELP"},
		},
		{
			name:      "return nil for regular message",
			utterance: "Hello, how are you?",
			expected:  nil,
		},
		{
			name:      "return nil for unknown command",
			utterance: "/unknown",
			expected:  nil,
		},
		{
			name:      "trim whitespace from utterance",
			utterance: "  /help  ",
			expected:  &Command{Type: "HELP"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parseCommand(tc.utterance)
			if tc.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tc.expected.Type, result.Type)
				assert.Equal(t, tc.expected.Code, result.Code)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "string shorter than max",
			input:    "Hello",
			maxLen:   10,
			expected: "Hello",
		},
		{
			name:     "string equal to max",
			input:    "Hello",
			maxLen:   5,
			expected: "Hello",
		},
		{
			name:     "string longer than max",
			input:    "Hello World",
			maxLen:   5,
			expected: "Hello...",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   5,
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := truncate(tc.input, tc.maxLen)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestWriteJSON(t *testing.T) {
	t.Run("sets correct content type and status", func(t *testing.T) {
		rec := httptest.NewRecorder()

		writeJSON(rec, http.StatusOK, map[string]string{"message": "hello"})

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Contains(t, rec.Body.String(), "hello")
	})

	t.Run("handles error status codes", func(t *testing.T) {
		rec := httptest.NewRecorder()

		writeJSON(rec, http.StatusBadRequest, map[string]string{"error": "bad request"})

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "bad request")
	})
}
