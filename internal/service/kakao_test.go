package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidCallbackURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "valid kakao.com URL",
			url:      "https://api.kakao.com/v1/callback",
			expected: true,
		},
		{
			name:     "valid kakaocdn.net URL",
			url:      "https://t1.kakaocdn.net/callback",
			expected: true,
		},
		{
			name:     "valid kakaoenterprise.com URL",
			url:      "https://bot.kakaoenterprise.com/webhook",
			expected: true,
		},
		{
			name:     "invalid - http scheme",
			url:      "http://api.kakao.com/callback",
			expected: false,
		},
		{
			name:     "invalid - non-kakao domain",
			url:      "https://evil.com/callback",
			expected: false,
		},
		{
			name:     "invalid - kakao in subdomain but wrong TLD",
			url:      "https://kakao.evil.com/callback",
			expected: false,
		},
		{
			name:     "invalid - empty URL",
			url:      "",
			expected: false,
		},
		{
			name:     "invalid - malformed URL",
			url:      "not-a-url",
			expected: false,
		},
		{
			name:     "valid - nested subdomain",
			url:      "https://deep.nested.kakao.com/path",
			expected: true,
		},
		{
			name:     "invalid - suffix match attempt",
			url:      "https://faketalkakao.com/callback",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isValidCallbackURL(tc.url)
			assert.Equal(t, tc.expected, result)
		})
	}
}
