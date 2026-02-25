package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateToken(t *testing.T) {
	t.Run("generates 64 character hex string", func(t *testing.T) {
		token, err := GenerateToken()
		require.NoError(t, err)
		assert.Len(t, token, 64)
	})

	t.Run("generates unique tokens", func(t *testing.T) {
		token1, _ := GenerateToken()
		token2, _ := GenerateToken()
		assert.NotEqual(t, token1, token2)
	})

	t.Run("generates valid hex", func(t *testing.T) {
		token, _ := GenerateToken()
		for _, c := range token {
			assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'))
		}
	})
}

func TestHashToken(t *testing.T) {
	t.Run("returns 64 character hex string", func(t *testing.T) {
		hash := HashToken("test-token")
		assert.Len(t, hash, 64)
	})

	t.Run("same input produces same hash", func(t *testing.T) {
		hash1 := HashToken("test-token")
		hash2 := HashToken("test-token")
		assert.Equal(t, hash1, hash2)
	})

	t.Run("different input produces different hash", func(t *testing.T) {
		hash1 := HashToken("token-1")
		hash2 := HashToken("token-2")
		assert.NotEqual(t, hash1, hash2)
	})
}

func TestHmacSHA256(t *testing.T) {
	t.Run("returns 64 character hex string", func(t *testing.T) {
		result := HmacSHA256("secret", "data")
		assert.Len(t, result, 64)
	})

	t.Run("same inputs produce same result", func(t *testing.T) {
		result1 := HmacSHA256("secret", "data")
		result2 := HmacSHA256("secret", "data")
		assert.Equal(t, result1, result2)
	})

	t.Run("different secret produces different result", func(t *testing.T) {
		result1 := HmacSHA256("secret1", "data")
		result2 := HmacSHA256("secret2", "data")
		assert.NotEqual(t, result1, result2)
	})

	t.Run("different data produces different result", func(t *testing.T) {
		result1 := HmacSHA256("secret", "data1")
		result2 := HmacSHA256("secret", "data2")
		assert.NotEqual(t, result1, result2)
	})

	t.Run("produces expected HMAC", func(t *testing.T) {
		// Known test vector
		result := HmacSHA256("key", "The quick brown fox jumps over the lazy dog")
		assert.Equal(t, "f7bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8", result)
	})
}

func TestConstantTimeEqual(t *testing.T) {
	t.Run("returns true for equal strings", func(t *testing.T) {
		assert.True(t, ConstantTimeEqual("abc", "abc"))
	})

	t.Run("returns false for different strings", func(t *testing.T) {
		assert.False(t, ConstantTimeEqual("abc", "def"))
	})

	t.Run("returns false for different lengths", func(t *testing.T) {
		assert.False(t, ConstantTimeEqual("abc", "abcd"))
	})

	t.Run("returns true for empty strings", func(t *testing.T) {
		assert.True(t, ConstantTimeEqual("", ""))
	})
}
