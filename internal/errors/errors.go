package errors

import (
	"errors"
	"fmt"
)

// ErrorCode represents a unique error identifier
type ErrorCode string

const (
	// Authentication & Authorization
	ErrCodeUnauthorized     ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden        ErrorCode = "FORBIDDEN"
	ErrCodeInvalidToken     ErrorCode = "INVALID_TOKEN"
	ErrCodeTokenExpired     ErrorCode = "TOKEN_EXPIRED"
	ErrCodeSessionNotPaired ErrorCode = "SESSION_NOT_PAIRED"

	// Validation
	ErrCodeValidation      ErrorCode = "VALIDATION_ERROR"
	ErrCodeInvalidInput    ErrorCode = "INVALID_INPUT"
	ErrCodeMissingRequired ErrorCode = "MISSING_REQUIRED"

	// Resource
	ErrCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrCodeAlreadyExists ErrorCode = "ALREADY_EXISTS"
	ErrCodeConflict      ErrorCode = "CONFLICT"

	// Pairing
	ErrCodeInvalidPairingCode ErrorCode = "INVALID_PAIRING_CODE"
	ErrCodePairingExpired     ErrorCode = "PAIRING_EXPIRED"
	ErrCodeAlreadyPaired      ErrorCode = "ALREADY_PAIRED"

	// Rate Limiting
	ErrCodeRateLimitExceeded ErrorCode = "RATE_LIMIT_EXCEEDED"

	// Callback
	ErrCodeCallbackExpired ErrorCode = "CALLBACK_EXPIRED"
	ErrCodeCallbackFailed  ErrorCode = "CALLBACK_FAILED"

	// Internal
	ErrCodeInternal ErrorCode = "INTERNAL_ERROR"
	ErrCodeDatabase ErrorCode = "DATABASE_ERROR"
	ErrCodeExternal ErrorCode = "EXTERNAL_SERVICE_ERROR"
)

// AppError is a structured error that can be returned to clients
type AppError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details any       `json:"details,omitempty"`
	cause   error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s (cause: %v)", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.cause
}

// WithCause adds a cause to the error
func (e *AppError) WithCause(err error) *AppError {
	e.cause = err
	return e
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details any) *AppError {
	e.Details = details
	return e
}

// New creates a new AppError
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an existing error with an AppError
func Wrap(code ErrorCode, message string, cause error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		cause:   cause,
	}
}

// Common error constructors

func Unauthorized(message string) *AppError {
	return New(ErrCodeUnauthorized, message)
}

func Forbidden(message string) *AppError {
	return New(ErrCodeForbidden, message)
}

func InvalidToken(message string) *AppError {
	return New(ErrCodeInvalidToken, message)
}

func SessionNotPaired() *AppError {
	return New(ErrCodeSessionNotPaired, "Session not paired")
}

func NotFound(resource string) *AppError {
	return New(ErrCodeNotFound, fmt.Sprintf("%s not found", resource))
}

func AlreadyExists(resource string) *AppError {
	return New(ErrCodeAlreadyExists, fmt.Sprintf("%s already exists", resource))
}

func ValidationError(message string) *AppError {
	return New(ErrCodeValidation, message)
}

func InvalidInput(field string, reason string) *AppError {
	return New(ErrCodeInvalidInput, fmt.Sprintf("Invalid %s: %s", field, reason))
}

func MissingRequired(field string) *AppError {
	return New(ErrCodeMissingRequired, fmt.Sprintf("%s is required", field))
}

func InvalidPairingCode() *AppError {
	return New(ErrCodeInvalidPairingCode, "Invalid or expired pairing code")
}

func PairingExpired() *AppError {
	return New(ErrCodePairingExpired, "Pairing code has expired")
}

func AlreadyPaired() *AppError {
	return New(ErrCodeAlreadyPaired, "Session is already paired")
}

func RateLimitExceeded() *AppError {
	return New(ErrCodeRateLimitExceeded, "Rate limit exceeded")
}

func CallbackExpired() *AppError {
	return New(ErrCodeCallbackExpired, "Callback URL expired or not available")
}

func CallbackFailed(reason string) *AppError {
	return New(ErrCodeCallbackFailed, fmt.Sprintf("Failed to send callback: %s", reason))
}

func Internal(message string) *AppError {
	return New(ErrCodeInternal, message)
}

func Database(cause error) *AppError {
	return Wrap(ErrCodeDatabase, "Database error", cause)
}

func External(service string, cause error) *AppError {
	return Wrap(ErrCodeExternal, fmt.Sprintf("External service error: %s", service), cause)
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// AsAppError converts an error to an AppError if possible
func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// GetCode returns the error code if the error is an AppError, otherwise returns ErrCodeInternal
func GetCode(err error) ErrorCode {
	if appErr, ok := AsAppError(err); ok {
		return appErr.Code
	}
	return ErrCodeInternal
}
