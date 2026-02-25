package middleware

import (
	"net/http"

	"github.com/openclaw/relay-server-go/internal/httputil"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	httputil.WriteJSON(w, status, data)
}
