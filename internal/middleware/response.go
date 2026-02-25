package middleware

import (
	"net/http"

	"gitlab.tepseg.com/ai/kakao-relay/internal/httputil"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	httputil.WriteJSON(w, status, data)
}
