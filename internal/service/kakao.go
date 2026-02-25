package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	callbackTimeout = 5 * time.Second
)

var allowedCallbackHosts = []string{
	".kakao.com",
	".kakaocdn.net",
	".kakaoenterprise.com",
}

type KakaoService struct {
	client *http.Client
}

func NewKakaoService() *KakaoService {
	return &KakaoService{
		client: &http.Client{
			Timeout: callbackTimeout,
		},
	}
}

func (s *KakaoService) SendCallback(ctx context.Context, callbackURL string, payload any) error {
	if !isValidCallbackURL(callbackURL) {
		log.Warn().Str("url", callbackURL).Msg("invalid callback URL rejected")
		return fmt.Errorf("invalid callback URL")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, callbackURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	log.Info().
		Str("url", callbackURL).
		Msg("sending callback to Kakao")

	resp, err := s.client.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		log.Error().
			Err(err).
			Str("url", callbackURL).
			Dur("elapsed", elapsed).
			Msg("kakao callback error")
		return fmt.Errorf("callback request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Error().
			Str("url", callbackURL).
			Int("status", resp.StatusCode).
			Dur("elapsed", elapsed).
			Msg("kakao callback failed")
		return fmt.Errorf("callback failed with status %d", resp.StatusCode)
	}

	log.Info().
		Str("url", callbackURL).
		Int("status", resp.StatusCode).
		Dur("elapsed", elapsed).
		Msg("kakao callback successful")

	return nil
}

func isValidCallbackURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	if parsed.Scheme != "https" {
		return false
	}

	hostname := strings.ToLower(parsed.Hostname())
	for _, suffix := range allowedCallbackHosts {
		if strings.HasSuffix(hostname, suffix) {
			return true
		}
	}

	return false
}
