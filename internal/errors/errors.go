package errors

import (
	"encoding/json"
	"net/http"
)

// ErrorCode represents a unique error code for client handling
type ErrorCode string

const (
	// Client errors (4xx)
	ErrCodeInvalidRequestBody ErrorCode = "INVALID_REQUEST_BODY"
	ErrCodeMissingFields      ErrorCode = "MISSING_REQUIRED_FIELDS"
	ErrCodeValidationFailed   ErrorCode = "VALIDATION_FAILED"
	ErrCodePlayerNotFound     ErrorCode = "PLAYER_NOT_FOUND"
	ErrCodeMatchesNotFound    ErrorCode = "MATCHES_NOT_FOUND"
	ErrCodeInvalidRegion      ErrorCode = "INVALID_REGION"
	ErrCodeMissingAPIKey      ErrorCode = "MISSING_API_KEY"
	ErrCodeInvalidAPIKey      ErrorCode = "INVALID_API_KEY"
	ErrCodeRateLimitExceeded  ErrorCode = "RATE_LIMIT_EXCEEDED"

	// Server errors (5xx)
	ErrCodeDataServiceError   ErrorCode = "DATA_SERVICE_ERROR"
	ErrCodeCortexServiceError ErrorCode = "CORTEX_SERVICE_ERROR"
	ErrCodeInternalError      ErrorCode = "INTERNAL_ERROR"
)

// APIError represents a structured error response
type APIError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Status  int       `json:"-"`
}

// Error implements the error interface
func (apiError *APIError) Error() string {
	return apiError.Message
}

// ErrorResponse is the JSON structure returned to clients
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains the error information
type ErrorDetail struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

// NewAPIError creates a new APIError
func NewAPIError(code ErrorCode, message string, status int) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// Common error constructors for consistent error creation
func InvalidRequestBody(message string) *APIError {
	return NewAPIError(ErrCodeInvalidRequestBody, message, http.StatusBadRequest)
}

func MissingFields(message string) *APIError {
	return NewAPIError(ErrCodeMissingFields, message, http.StatusBadRequest)
}

func PlayerNotFound(gameName string, tagLine string) *APIError {
	return NewAPIError(ErrCodePlayerNotFound, "Player not found: "+gameName+"#"+tagLine, http.StatusNotFound)
}

func MatchesNotFound(message string) *APIError {
	return NewAPIError(ErrCodeMatchesNotFound, message, http.StatusNotFound)
}

func DataServiceError(message string) *APIError {
	return NewAPIError(ErrCodeDataServiceError, message, http.StatusBadGateway)
}

func CortexServiceError(message string) *APIError {
	return NewAPIError(ErrCodeCortexServiceError, message, http.StatusBadGateway)
}

func InternalError(message string) *APIError {
	return NewAPIError(ErrCodeInternalError, message, http.StatusInternalServerError)
}

func ValidationFailed(message string) *APIError {
	return NewAPIError(ErrCodeValidationFailed, message, http.StatusBadRequest)
}

// WriteError writes a JSON error response to the http.ResponseWriter
func WriteError(writer http.ResponseWriter, apiError *APIError) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(apiError.Status)

	errorResponse := ErrorResponse{
		Error: ErrorDetail{
			Code:    apiError.Code,
			Message: apiError.Message,
		},
	}

	json.NewEncoder(writer).Encode(errorResponse)
}
