package handlers

import (
	"net/http"
)

func (h *Handler) GetServers(r *http.Request) (*HandlerResult, error) {
	req, err := NewGetServersRequestBuilder(r.URL.Query(), h.AllowedRAM, h.AllowedDiskTypes).
		WithModel().
		WithRAM().
		WithLocation().
		WithDiskType().
		WithStorageMin().
		WithStorageMax().
		WithPagination().
		Build()
	if err != nil {
		return nil, err
	}

	servers, err := h.Service.GetServers(r.Context(), req.ToFilter())
	if err != nil {
		return nil, err
	}
	return &HandlerResult{
		Status:    http.StatusOK,
		Body:      mapServersResponse(servers, req.Limit, req.Offset),
		Cacheable: true,
	}, nil
}

func (h *Handler) Upload(r *http.Request) (*HandlerResult, error) {
	uploadReq, apiErr := NewUploadServerRequest(r)
	if apiErr != nil {
		return nil, apiErr
	}
	defer func() { _ = uploadReq.File.Close() }()

	if err := h.Service.UploadServerData(r.Context(), uploadReq.FileName, uploadReq.File); err != nil {
		return nil, err
	}
	// A successful upload is a valid recovery path if the server booted
	// without startup data - see GET /readyz.
	h.Ready.Store(true)
	return &HandlerResult{Status: http.StatusNoContent, Body: nil}, nil
}
