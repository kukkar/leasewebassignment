// Package httperr is generic HTTP-API error-response scaffolding: a
// structured error type, a JSON envelope, constructors for the common HTTP
// status codes, and a way to write either as JSON. None of it references
// this application's own domain types (no model, service, or store
// package) - it would be identical dropped into any Go HTTP service. The
// one piece that *is* specific to this application - translating this
// app's internal error types (service.CSVParseError, service.ServiceError,
// store.StoreError) into this vocabulary - is deliberately kept out of
// here; see internal/api.MapError.
package httperr

import (
	"encoding/json"
	"net/http"
)

const (
	CodeBadRequest         = "bad_request"
	CodeInvalidInput       = "invalid_input"
	CodeUnauthorized       = "unauthorized"
	CodeNotFound           = "not_found"
	CodeMethodNotAllowed   = "method_not_allowed"
	CodeInternalError      = "internal_error"
	CodeCSVValidation      = "csv_validation_error"
	CodeMissingParameter   = "missing_parameter"
	CodeServiceUnavailable = "service_unavailable"
	CodeRequestTooLarge    = "request_too_large"
)

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Status  int    `json:"-"`
}

func (e *APIError) Error() string {
	return e.Message
}

func (e *APIError) StatusCode() int {
	if e == nil || e.Status == 0 {
		return http.StatusInternalServerError
	}
	return e.Status
}

func (e *APIError) ErrorCode() string {
	return e.Code
}

func NewAPIError(code, message string, status int, details string) *APIError {
	return &APIError{Code: code, Message: message, Details: details, Status: status}
}

func BadRequest(message, details string) *APIError {
	return NewAPIError(CodeBadRequest, message, http.StatusBadRequest, details)
}

func InvalidInput(message, details string) *APIError {
	return NewAPIError(CodeInvalidInput, message, http.StatusBadRequest, details)
}

func Unauthorized(message, details string) *APIError {
	return NewAPIError(CodeUnauthorized, message, http.StatusUnauthorized, details)
}

func NotFound(message, details string) *APIError {
	return NewAPIError(CodeNotFound, message, http.StatusNotFound, details)
}

func MethodNotAllowed(message, details string) *APIError {
	return NewAPIError(CodeMethodNotAllowed, message, http.StatusMethodNotAllowed, details)
}

func RequestTooLarge(message, details string) *APIError {
	return NewAPIError(CodeRequestTooLarge, message, http.StatusRequestEntityTooLarge, details)
}

func InternalError(message, details string) *APIError {
	return NewAPIError(CodeInternalError, message, http.StatusInternalServerError, details)
}

func CSVValidationError(message, details string) *APIError {
	return NewAPIError(CodeCSVValidation, message, http.StatusBadRequest, details)
}

func ServiceUnavailable(message, details string) *APIError {
	return NewAPIError(CodeServiceUnavailable, message, http.StatusServiceUnavailable, details)
}

type ErrorResponse struct {
	Error APIError `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
