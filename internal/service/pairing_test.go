package service

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRandomCode(t *testing.T) {
	t.Run("generates code in correct format XXXX-XXXX", func(t *testing.T) {
		code := generateRandomCode()

		pattern := regexp.MustCompile(`^[A-Z0-9]{4}-[A-Z0-9]{4}$`)
		assert.True(t, pattern.MatchString(code), "code should match XXXX-XXXX format, got: %s", code)
	})

	t.Run("uses only allowed characters", func(t *testing.T) {
		code := generateRandomCode()

		// Remove the dash and check each character
		chars := code[:4] + code[5:]
		for _, c := range chars {
			found := false
			for _, allowed := range pairingCodeChars {
				if c == allowed {
					found = true
					break
				}
			}
			assert.True(t, found, "character '%c' should be in allowed set", c)
		}
	})

	t.Run("generates unique codes", func(t *testing.T) {
		codes := make(map[string]bool)
		for i := 0; i < 100; i++ {
			code := generateRandomCode()
			assert.False(t, codes[code], "duplicate code generated: %s", code)
			codes[code] = true
		}
	})

	t.Run("excludes ambiguous characters", func(t *testing.T) {
		// O, I, 0, 1 are excluded from pairingCodeChars
		for i := 0; i < 100; i++ {
			code := generateRandomCode()
			assert.NotContains(t, code, "O")
			assert.NotContains(t, code, "I")
			assert.NotContains(t, code, "0")
			assert.NotContains(t, code, "1")
		}
	})
}

func TestPairingCodeChars(t *testing.T) {
	t.Run("contains no ambiguous characters", func(t *testing.T) {
		assert.NotContains(t, pairingCodeChars, "O")
		assert.NotContains(t, pairingCodeChars, "I")
		assert.NotContains(t, pairingCodeChars, "0")
		assert.NotContains(t, pairingCodeChars, "1")
	})

	t.Run("contains expected character count", func(t *testing.T) {
		// 26 letters - O, I = 24 letters
		// 10 digits - 0, 1 = 8 digits
		// Total = 32 characters
		assert.Len(t, pairingCodeChars, 32)
	})
}
