package handlers

import (
	"net/http"

	"github.com/sahil/leasewebassignment/internal/api"
)

func (h *Handler) GetServers(r *http.Request) (*HandlerResult, error) {
	req, err := NewGetServersRequestBuilder(r.URL.Query()).
		WithModel().
		WithRAM().
		WithHDD().
		WithLocation().
		WithDiskType().
		WithStorageMin().
		WithStorageMax().
		WithPriceMin().
		WithPriceMax().
		Build()
	if err != nil {
		return nil, api.InvalidInput("invalid query parameters", err.Error())
	}

	servers, err := h.Service.GetServers(r.Context(), req.ToFilter())
	if err != nil {
		return nil, err
	}
	return &HandlerResult{Status: http.StatusOK, Body: mapServersResponse(servers)}, nil
}

func (h *Handler) Upload(r *http.Request) (*HandlerResult, error) {
	uploadReq, apiErr := NewUploadServerRequest(r)
	if apiErr != nil {
		return nil, apiErr
	}
	defer uploadReq.File.Close()

	if err := h.Service.UploadServerData(r.Context(), uploadReq.FileName, uploadReq.File); err != nil {
		return nil, err
	}
	return &HandlerResult{Status: http.StatusNoContent, Body: nil}, nil
}
