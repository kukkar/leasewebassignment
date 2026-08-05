package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/sahil/leasewebassignment/internal/service"
)

const (
	CodeBadRequest       = "bad_request"
	CodeInvalidInput     = "invalid_input"
	CodeUnauthorized     = "unauthorized"
	CodeNotFound         = "not_found"
	CodeInternalError    = "internal_error"
	CodeCSVValidation    = "csv_validation_error"
	CodeMissingParameter = "missing_parameter"
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

func InternalError(message, details string) *APIError {
	return NewAPIError(CodeInternalError, message, http.StatusInternalServerError, details)
}

func CSVValidationError(message, details string) *APIError {
	return NewAPIError(CodeCSVValidation, message, http.StatusBadRequest, details)
}

type ErrorResponse struct {
	Error APIError `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, err error) {
	status, apiErr := MapError(err)
	WriteJSON(w, status, ErrorResponse{Error: apiErr})
}

func MapError(err error) (int, APIError) {
	if err == nil {
		return http.StatusOK, APIError{}
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode(), *apiErr
	}

	var csvErr *service.CSVParseError
	if errors.As(err, &csvErr) {
		return http.StatusBadRequest, *CSVValidationError("invalid csv payload", csvErr.Error())
	}

	if errors.Is(err, os.ErrNotExist) {
		return http.StatusNotFound, *NotFound("resource not found", err.Error())
	}

	var svcErr *service.ServiceError
	if errors.As(err, &svcErr) {
		return MapError(svcErr.Err)
	}

	return http.StatusInternalServerError, *InternalError("internal server error", err.Error())
}
