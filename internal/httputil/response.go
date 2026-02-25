package httputil

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/openclaw/relay-server-go/internal/errors"
)

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// ErrorResponse is the standard error response format
type ErrorResponse struct {
	Error   string                  `json:"error"`
	Code    apperrors.ErrorCode     `json:"code"`
	Details any                     `json:"details,omitempty"`
}

// WriteError writes an AppError as an HTTP response with appropriate status code
func WriteError(w http.ResponseWriter, err error) {
	appErr, ok := apperrors.AsAppError(err)
	if !ok {
		// Wrap unknown errors as internal errors
		appErr = apperrors.Internal("An unexpected error occurred")
	}

	status := statusFromCode(appErr.Code)
	response := ErrorResponse{
		Error:   appErr.Message,
		Code:    appErr.Code,
		Details: appErr.Details,
	}

	WriteJSON(w, status, response)
}

// WriteErrorWithStatus writes an error with a specific HTTP status code
func WriteErrorWithStatus(w http.ResponseWriter, status int, err *apperrors.AppError) {
	response := ErrorResponse{
		Error:   err.Message,
		Code:    err.Code,
		Details: err.Details,
	}
	WriteJSON(w, status, response)
}

// statusFromCode maps ErrorCode to HTTP status code
func statusFromCode(code apperrors.ErrorCode) int {
	switch code {
	// 400 Bad Request
	case apperrors.ErrCodeValidation,
		apperrors.ErrCodeInvalidInput,
		apperrors.ErrCodeMissingRequired,
		apperrors.ErrCodeInvalidPairingCode,
		apperrors.ErrCodePairingExpired,
		apperrors.ErrCodeCallbackExpired:
		return http.StatusBadRequest

	// 401 Unauthorized
	case apperrors.ErrCodeUnauthorized,
		apperrors.ErrCodeInvalidToken,
		apperrors.ErrCodeTokenExpired,
		apperrors.ErrCodeSessionNotPaired:
		return http.StatusUnauthorized

	// 403 Forbidden
	case apperrors.ErrCodeForbidden:
		return http.StatusForbidden

	// 404 Not Found
	case apperrors.ErrCodeNotFound:
		return http.StatusNotFound

	// 409 Conflict
	case apperrors.ErrCodeAlreadyExists,
		apperrors.ErrCodeConflict,
		apperrors.ErrCodeAlreadyPaired:
		return http.StatusConflict

	// 429 Too Many Requests
	case apperrors.ErrCodeRateLimitExceeded:
		return http.StatusTooManyRequests

	// 502 Bad Gateway
	case apperrors.ErrCodeCallbackFailed,
		apperrors.ErrCodeExternal:
		return http.StatusBadGateway

	// 500 Internal Server Error
	case apperrors.ErrCodeInternal,
		apperrors.ErrCodeDatabase:
		return http.StatusInternalServerError

	default:
		return http.StatusInternalServerError
	}
}
