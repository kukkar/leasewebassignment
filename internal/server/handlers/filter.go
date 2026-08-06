package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/sahil/leasewebassignment/internal/api"
	"github.com/sahil/leasewebassignment/internal/model"
)

const (
	QueryParamModel      = "model"
	QueryParamRAM        = "ram"
	QueryParamLocation   = "location"
	QueryParamStorageMin = "storage_min"
	QueryParamStorageMax = "storage_max"
	QueryParamDiskType   = "disk_type"
	QueryParamLimit      = "limit"
	QueryParamOffset     = "offset"

	DefaultLimit = 50
	MaxLimit     = 200
)

type GetServersRequest struct {
	Model    string   `json:"model"`
	RAM      []string `json:"ram,omitempty"`
	Location string   `json:"location"`
	DiskType string   `json:"disk_type,omitempty"`
	// storage values are expressed as GB units
	StorageMin *int `json:"storage_min_gb,omitempty"`
	StorageMax *int `json:"storage_max_gb,omitempty"`
	Limit      int  `json:"limit"`
	Offset     int  `json:"offset"`
}

// GetServersRequestBuilder validates query parameters into a
// GetServersRequest. Every With* method runs regardless of whether an
// earlier one failed, and Build() reports every problem found in a single
// response - a caller with two invalid params fixes both on the next
// request instead of rediscovering them one at a time.
type GetServersRequestBuilder struct {
	values           url.Values
	request          GetServersRequest
	errs             []string
	allowedRAM       []string
	allowedDiskTypes []string
}

func NewGetServersRequestBuilder(values url.Values, allowedRAM, allowedDiskTypes []string) *GetServersRequestBuilder {
	allowedRAM, allowedDiskTypes = withDefaults(allowedRAM, allowedDiskTypes)
	return &GetServersRequestBuilder{values: values, allowedRAM: allowedRAM, allowedDiskTypes: allowedDiskTypes}
}

func (b *GetServersRequestBuilder) WithModel() *GetServersRequestBuilder {
	b.request.Model = b.values.Get(QueryParamModel)
	return b
}

// WithRAM reads all repeated `ram` query params (the RAM filter is a
// checkbox group in the assignment's spec: a server matches if its RAM is
// ANY of the selected values), validating each against the allowed list.
// An invalid value is recorded and skipped; the rest are still processed.
func (b *GetServersRequestBuilder) WithRAM() *GetServersRequestBuilder {
	for _, val := range b.values[QueryParamRAM] {
		if val == "" {
			continue
		}
		ok := false
		for _, a := range b.allowedRAM {
			if strings.EqualFold(a, val) {
				ok = true
				break
			}
		}
		if !ok {
			b.errs = append(b.errs, fmt.Sprintf("ram: %q is invalid, must be one of: %s", val, strings.Join(b.allowedRAM, ", ")))
			continue
		}
		b.request.RAM = append(b.request.RAM, val)
	}
	return b
}

func (b *GetServersRequestBuilder) WithLocation() *GetServersRequestBuilder {
	b.request.Location = b.values.Get(QueryParamLocation)
	return b
}

func (b *GetServersRequestBuilder) WithDiskType() *GetServersRequestBuilder {
	val := b.values.Get(QueryParamDiskType)
	b.request.DiskType = val
	if val == "" {
		return b
	}
	ok := false
	for _, a := range b.allowedDiskTypes {
		if strings.EqualFold(a, val) {
			ok = true
			break
		}
	}
	if !ok {
		b.errs = append(b.errs, fmt.Sprintf("disk_type: %q is invalid, must be one of: %s", val, strings.Join(b.allowedDiskTypes, ", ")))
	}
	return b
}

func (b *GetServersRequestBuilder) WithStorageMin() *GetServersRequestBuilder {
	value := b.values.Get(QueryParamStorageMin)
	if value == "" {
		return b
	}
	parsed, err := model.ParseStorageValue(value)
	if err != nil {
		b.errs = append(b.errs, fmt.Sprintf("storage_min: %v", err))
		return b
	}
	b.request.StorageMin = &parsed
	return b
}

func (b *GetServersRequestBuilder) WithStorageMax() *GetServersRequestBuilder {
	value := b.values.Get(QueryParamStorageMax)
	if value == "" {
		return b
	}
	parsed, err := model.ParseStorageValue(value)
	if err != nil {
		b.errs = append(b.errs, fmt.Sprintf("storage_max: %v", err))
		return b
	}
	b.request.StorageMax = &parsed
	return b
}

// WithPagination parses limit/offset, defaulting to DefaultLimit/0 and
// capping limit at MaxLimit so a single request can't force the server to
// marshal and transmit the entire catalog in one response.
func (b *GetServersRequestBuilder) WithPagination() *GetServersRequestBuilder {
	limit := DefaultLimit
	if raw := b.values.Get(QueryParamLimit); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			b.errs = append(b.errs, fmt.Sprintf("limit: %q must be a positive integer", raw))
		} else if parsed > MaxLimit {
			limit = MaxLimit
		} else {
			limit = parsed
		}
	}

	offset := 0
	if raw := b.values.Get(QueryParamOffset); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			b.errs = append(b.errs, fmt.Sprintf("offset: %q must be a non-negative integer", raw))
		} else {
			offset = parsed
		}
	}

	b.request.Limit = limit
	b.request.Offset = offset
	return b
}

func (b *GetServersRequestBuilder) Build() (GetServersRequest, error) {
	if len(b.errs) > 0 {
		return GetServersRequest{}, api.InvalidInput("invalid query parameters", strings.Join(b.errs, "; "))
	}
	return b.request, nil
}

func (r GetServersRequest) ToFilter() model.ServerFilter {
	return model.ServerFilter{
		Model:      r.Model,
		RAM:        r.RAM,
		Location:   r.Location,
		DiskType:   r.DiskType,
		StorageMin: r.StorageMin,
		StorageMax: r.StorageMax,
	}
}

type UploadServerRequest struct {
	FileName string
	File     io.ReadCloser
}

func NewUploadServerRequest(r *http.Request) (*UploadServerRequest, *api.APIError) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return nil, api.InvalidInput("invalid multipart payload", err.Error())
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, api.InvalidInput("missing file field", err.Error())
	}
	if header == nil || header.Filename == "" {
		_ = file.Close()
		return nil, api.InvalidInput("upload filename required", "form field 'file' must contain a filename")
	}

	return &UploadServerRequest{FileName: header.Filename, File: file}, nil
}
