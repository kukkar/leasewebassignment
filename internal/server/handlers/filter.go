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
	QueryParamHDD        = "hdd"
	QueryParamLocation   = "location"
	QueryParamPriceMin   = "price_min"
	QueryParamPriceMax   = "price_max"
	QueryParamStorageMin = "storage_min"
	QueryParamStorageMax = "storage_max"
	QueryParamDiskType   = "disk_type"
)

type GetServersRequest struct {
	Model    string   `json:"model"`
	RAM      string   `json:"ram"`
	HDD      string   `json:"hdd"`
	Location string   `json:"location"`
	PriceMin *float64 `json:"price_min,omitempty"`
	PriceMax *float64 `json:"price_max,omitempty"`
	DiskType string   `json:"disk_type,omitempty"`
	// storage values are expressed as GB units
	StorageMin *int `json:"storage_min_gb,omitempty"`
	StorageMax *int `json:"storage_max_gb,omitempty"`
}

type GetServersRequestBuilder struct {
	values  url.Values
	request GetServersRequest
	err     error
}

var (
	allowedRAM = []string{"2GB", "4GB", "8GB", "12GB", "16GB", "24GB", "32GB", "48GB", "64GB", "96GB"}
	allowedHDD = []string{"SAS", "SATA", "SSD"}
)

func NewGetServersRequestBuilder(values url.Values) *GetServersRequestBuilder {
	return &GetServersRequestBuilder{values: values}
}

func (b *GetServersRequestBuilder) WithModel() *GetServersRequestBuilder {
	b.request.Model = b.values.Get(QueryParamModel)
	return b
}

func (b *GetServersRequestBuilder) WithRAM() *GetServersRequestBuilder {
	val := b.values.Get(QueryParamRAM)
	b.request.RAM = val
	if val != "" {
		ok := false
		for _, a := range allowedRAM {
			if strings.EqualFold(a, val) {
				ok = true
				break
			}
		}
		if !ok {
			b.err = api.InvalidInput("invalid ram value", fmt.Sprintf("ram must be one of: %s", strings.Join(allowedRAM, ", ")))
		}
	}
	return b
}

func (b *GetServersRequestBuilder) WithHDD() *GetServersRequestBuilder {
	val := b.values.Get(QueryParamHDD)
	b.request.HDD = val
	if val != "" {
		ok := false
		for _, a := range allowedHDD {
			if strings.EqualFold(a, val) {
				ok = true
				break
			}
		}
		if !ok {
			b.err = api.InvalidInput("invalid hdd type", fmt.Sprintf("hdd must be one of: %s", strings.Join(allowedHDD, ", ")))
		}
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
	if val != "" {
		ok := false
		for _, a := range allowedHDD {
			if strings.EqualFold(a, val) {
				ok = true
				break
			}
		}
		if !ok {
			b.err = api.InvalidInput("invalid disk type", fmt.Sprintf("disk_type must be one of: %s", strings.Join(allowedHDD, ", ")))
		}
	}
	return b
}

func (b *GetServersRequestBuilder) WithStorageMin() *GetServersRequestBuilder {
	if b.err != nil {
		return b
	}
	if value := b.values.Get(QueryParamStorageMin); value != "" {
		parsed, err := model.ParseStorageValue(value)
		if err != nil {
			b.err = api.InvalidInput("invalid storage_min", err.Error())
			return b
		}
		b.request.StorageMin = &parsed
	}
	return b
}

func (b *GetServersRequestBuilder) WithStorageMax() *GetServersRequestBuilder {
	if b.err != nil {
		return b
	}
	if value := b.values.Get(QueryParamStorageMax); value != "" {
		parsed, err := model.ParseStorageValue(value)
		if err != nil {
			b.err = api.InvalidInput("invalid storage_max", err.Error())
			return b
		}
		b.request.StorageMax = &parsed
	}
	return b
}

func (b *GetServersRequestBuilder) WithPriceMin() *GetServersRequestBuilder {
	if b.err != nil {
		return b
	}
	if value := b.values.Get(QueryParamPriceMin); value != "" {
		parsed, err := parseFloatQuery(value)
		if err != nil {
			b.err = err
			return b
		}
		b.request.PriceMin = parsed
	}
	return b
}

func (b *GetServersRequestBuilder) WithPriceMax() *GetServersRequestBuilder {
	if b.err != nil {
		return b
	}
	if value := b.values.Get(QueryParamPriceMax); value != "" {
		parsed, err := parseFloatQuery(value)
		if err != nil {
			b.err = err
			return b
		}
		b.request.PriceMax = parsed
	}
	return b
}

func (b *GetServersRequestBuilder) Build() (GetServersRequest, error) {
	return b.request, b.err
}

func (r GetServersRequest) ToFilter() model.ServerFilter {
	return model.ServerFilter{
		Model:      r.Model,
		RAM:        r.RAM,
		HDD:        r.HDD,
		Location:   r.Location,
		PriceMin:   r.PriceMin,
		PriceMax:   r.PriceMax,
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

func parseFloatQuery(value string) (*float64, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid numeric query parameter: %q", value)
	}
	return &parsed, nil
}
