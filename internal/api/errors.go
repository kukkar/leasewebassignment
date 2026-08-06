// Package api translates this application's own error types - across
// every layer (store, service) - into the generic httperr vocabulary
// (internal/platform/httperr) and writes the result. This is deliberately
// the only piece of HTTP error handling that isn't generic: MapError knows
// the concrete types service.CSVParseError, service.ServiceError, and
// store.StoreError, which is exactly what makes it specific to this
// application rather than reusable scaffolding - see httperr's package
// comment for the reasoning behind that split.
package api

import (
	"errors"
	"net/http"
	"os"

	"github.com/sahil/leasewebassignment/internal/platform/httperr"
	"github.com/sahil/leasewebassignment/internal/service"
	"github.com/sahil/leasewebassignment/internal/store"
)

func WriteError(w http.ResponseWriter, err error) {
	status, apiErr := MapError(err)
	httperr.WriteJSON(w, status, httperr.ErrorResponse{Error: apiErr})
}

func MapError(err error) (int, httperr.APIError) {
	if err == nil {
		return http.StatusOK, httperr.APIError{}
	}

	var apiErr *httperr.APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode(), *apiErr
	}

	var csvErr *service.CSVParseError
	if errors.As(err, &csvErr) {
		return http.StatusBadRequest, *httperr.CSVValidationError("invalid csv payload", csvErr.Error())
	}

	if errors.Is(err, os.ErrNotExist) {
		return http.StatusNotFound, *httperr.NotFound("resource not found", err.Error())
	}

	var svcErr *service.ServiceError
	if errors.As(err, &svcErr) {
		return MapError(svcErr.Err)
	}

	// Symmetric with the ServiceError branch above: peels a StoreError's own
	// "store <op>:" prefix so a fallback 500's details show the innermost
	// cause, not a wrapper describing which internal layer noticed it.
	var storeErr *store.StoreError
	if errors.As(err, &storeErr) {
		return MapError(storeErr.Err)
	}

	return http.StatusInternalServerError, *httperr.InternalError("internal server error", err.Error())
}
