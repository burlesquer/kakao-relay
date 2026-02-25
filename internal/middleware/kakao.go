package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/openclaw/relay-server-go/internal/util"
)

const KakaoBodyContextKey contextKey = "kakaoBody"

func GetKakaoBody(ctx context.Context) any {
	return ctx.Value(KakaoBodyContextKey)
}

type KakaoSignatureMiddleware struct {
	secret string
}

func NewKakaoSignatureMiddleware(secret string) *KakaoSignatureMiddleware {
	return &KakaoSignatureMiddleware{secret: secret}
}

func (m *KakaoSignatureMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.secret == "" {
			log.Warn().Msg("kakao signature verification bypassed: KAKAO_SIGNATURE_SECRET is not configured")
			next.ServeHTTP(w, r)
			return
		}

		signature := r.Header.Get("X-Kakao-Signature")
		if signature == "" {
			log.Warn().Msg("kakao signature middleware: missing signature header")
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Missing signature",
			})
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).Msg("kakao signature middleware: failed to read body")
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Failed to read request body",
			})
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))

		computed := util.HmacSHA256(m.secret, string(body))
		if !util.ConstantTimeEqual(computed, signature) {
			log.Warn().Msg("kakao signature middleware: invalid signature")
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Invalid signature",
			})
			return
		}

		var parsed any
		if err := json.Unmarshal(body, &parsed); err != nil {
			log.Error().Err(err).Msg("kakao signature middleware: failed to parse body")
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid JSON body",
			})
			return
		}

		ctx := context.WithValue(r.Context(), KakaoBodyContextKey, parsed)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
