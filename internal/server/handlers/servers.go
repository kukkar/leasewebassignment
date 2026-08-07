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
	// ParseMultipartForm (inside NewUploadServerRequest) may have spilled
	// part of the upload to an OS temp file once it exceeded the in-memory
	// threshold (MaxUploadMemory). Without this, that temp file is never
	// deleted - a slow disk leak, one file per large-enough upload.
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()

	if err := h.Service.UploadServerData(r.Context(), uploadReq.FileName, uploadReq.File); err != nil {
		return nil, err
	}
	// A successful upload is a valid recovery path if the server booted
	// without startup data - see GET /readyz.
	h.Ready.Store(true)
	return &HandlerResult{Status: http.StatusNoContent, Body: nil}, nil
}
