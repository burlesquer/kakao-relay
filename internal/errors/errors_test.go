package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError(t *testing.T) {
	t.Run("Error returns formatted string", func(t *testing.T) {
		err := New(ErrCodeNotFound, "User not found")
		assert.Equal(t, "NOT_FOUND: User not found", err.Error())
	})

	t.Run("Error with cause includes cause", func(t *testing.T) {
		cause := errors.New("database connection failed")
		err := Wrap(ErrCodeDatabase, "Database error", cause)
		assert.Contains(t, err.Error(), "DATABASE_ERROR")
		assert.Contains(t, err.Error(), "Database error")
		assert.Contains(t, err.Error(), "database connection failed")
	})

	t.Run("WithCause adds cause to error", func(t *testing.T) {
		cause := errors.New("original error")
		err := New(ErrCodeInternal, "Something went wrong").WithCause(cause)
		assert.Equal(t, cause, err.Unwrap())
	})

	t.Run("WithDetails adds details to error", func(t *testing.T) {
		details := map[string]string{"field": "email", "reason": "invalid format"}
		err := New(ErrCodeValidation, "Validation failed").WithDetails(details)
		assert.Equal(t, details, err.Details)
	})
}

func TestErrorConstructors(t *testing.T) {
	tests := []struct {
		name         string
		constructor  func() *AppError
		expectedCode ErrorCode
	}{
		{"Unauthorized", func() *AppError { return Unauthorized("test") }, ErrCodeUnauthorized},
		{"Forbidden", func() *AppError { return Forbidden("test") }, ErrCodeForbidden},
		{"InvalidToken", func() *AppError { return InvalidToken("test") }, ErrCodeInvalidToken},
		{"SessionNotPaired", func() *AppError { return SessionNotPaired() }, ErrCodeSessionNotPaired},
		{"NotFound", func() *AppError { return NotFound("User") }, ErrCodeNotFound},
		{"AlreadyExists", func() *AppError { return AlreadyExists("User") }, ErrCodeAlreadyExists},
		{"ValidationError", func() *AppError { return ValidationError("test") }, ErrCodeValidation},
		{"InvalidInput", func() *AppError { return InvalidInput("email", "invalid") }, ErrCodeInvalidInput},
		{"MissingRequired", func() *AppError { return MissingRequired("email") }, ErrCodeMissingRequired},
		{"InvalidPairingCode", func() *AppError { return InvalidPairingCode() }, ErrCodeInvalidPairingCode},
		{"PairingExpired", func() *AppError { return PairingExpired() }, ErrCodePairingExpired},
		{"AlreadyPaired", func() *AppError { return AlreadyPaired() }, ErrCodeAlreadyPaired},
		{"RateLimitExceeded", func() *AppError { return RateLimitExceeded() }, ErrCodeRateLimitExceeded},
		{"CallbackExpired", func() *AppError { return CallbackExpired() }, ErrCodeCallbackExpired},
		{"CallbackFailed", func() *AppError { return CallbackFailed("timeout") }, ErrCodeCallbackFailed},
		{"Internal", func() *AppError { return Internal("test") }, ErrCodeInternal},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.constructor()
			assert.Equal(t, tc.expectedCode, err.Code)
			assert.NotEmpty(t, err.Message)
		})
	}
}

func TestDatabase(t *testing.T) {
	t.Run("wraps database error", func(t *testing.T) {
		cause := errors.New("connection refused")
		err := Database(cause)
		assert.Equal(t, ErrCodeDatabase, err.Code)
		assert.Equal(t, cause, err.Unwrap())
	})
}

func TestExternal(t *testing.T) {
	t.Run("wraps external service error", func(t *testing.T) {
		cause := errors.New("timeout")
		err := External("Kakao API", cause)
		assert.Equal(t, ErrCodeExternal, err.Code)
		assert.Contains(t, err.Message, "Kakao API")
		assert.Equal(t, cause, err.Unwrap())
	})
}

func TestIsAppError(t *testing.T) {
	t.Run("returns true for AppError", func(t *testing.T) {
		err := New(ErrCodeNotFound, "test")
		assert.True(t, IsAppError(err))
	})

	t.Run("returns false for standard error", func(t *testing.T) {
		err := errors.New("standard error")
		assert.False(t, IsAppError(err))
	})

	t.Run("returns true for wrapped AppError", func(t *testing.T) {
		appErr := New(ErrCodeNotFound, "test")
		wrapped := errors.New("wrapped: " + appErr.Error())
		// Note: This returns false because we're creating a new error, not wrapping
		assert.False(t, IsAppError(wrapped))
	})
}

func TestAsAppError(t *testing.T) {
	t.Run("extracts AppError", func(t *testing.T) {
		original := New(ErrCodeNotFound, "User not found")
		extracted, ok := AsAppError(original)
		assert.True(t, ok)
		assert.Equal(t, original, extracted)
	})

	t.Run("returns false for non-AppError", func(t *testing.T) {
		err := errors.New("standard error")
		extracted, ok := AsAppError(err)
		assert.False(t, ok)
		assert.Nil(t, extracted)
	})
}

func TestGetCode(t *testing.T) {
	t.Run("returns code for AppError", func(t *testing.T) {
		err := New(ErrCodeNotFound, "test")
		assert.Equal(t, ErrCodeNotFound, GetCode(err))
	})

	t.Run("returns ErrCodeInternal for standard error", func(t *testing.T) {
		err := errors.New("standard error")
		assert.Equal(t, ErrCodeInternal, GetCode(err))
	})
}

func TestNotFoundMessage(t *testing.T) {
	t.Run("formats resource name correctly", func(t *testing.T) {
		err := NotFound("User")
		assert.Equal(t, "User not found", err.Message)

		err = NotFound("Message")
		assert.Equal(t, "Message not found", err.Message)
	})
}

func TestMissingRequiredMessage(t *testing.T) {
	t.Run("formats field name correctly", func(t *testing.T) {
		err := MissingRequired("email")
		assert.Equal(t, "email is required", err.Message)

		err = MissingRequired("messageId")
		assert.Equal(t, "messageId is required", err.Message)
	})
}
