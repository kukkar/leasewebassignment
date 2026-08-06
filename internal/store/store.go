// Package store holds the server repository: the in-memory catalog and the
// filter-evaluation logic used to narrow it. Filtering is optimized around a
// simple invariant - the catalog only changes via ReplaceServers, so every
// value a filter needs (total storage, disk type family, RAM family) is
// parsed once there and cached, never re-derived on the read path.
package store

import (
	"context"
	"strings"

	"github.com/sahil/leasewebassignment/internal/model"
)

type Repository interface {
	ListServers(ctx context.Context, filter model.ServerFilter) ([]model.Server, error)
	ReplaceServers(ctx context.Context, servers []model.Server) error
	SaveUpload(ctx context.Context, filename string, content []byte) (string, error)
}

// indexedServer pairs a raw catalog record with the values filtering needs,
// computed once when the record is ingested rather than on every read. It's
// deliberately unexported - nothing outside this package ever sees it, so
// these derived fields can never leak into the JSON API contract.
type indexedServer struct {
	server         model.Server
	totalStorageGB int
	diskType       string
	ramFamily      string
}

// newIndexedServer parses the fields matchesFilter needs exactly once. This
// is the only place in a request's lifecycle that regex parsing happens -
// ListServers, called on every request, never parses anything.
func newIndexedServer(s model.Server) indexedServer {
	idx := indexedServer{server: s, ramFamily: model.RAMFamily(s.RAM)}
	if s.HDD != "" {
		if totalGB, diskType, err := model.ParseHDD(s.HDD); err == nil {
			idx.totalStorageGB = totalGB
			idx.diskType = diskType
		}
	}
	return idx
}

// preparedFilter is model.ServerFilter with its RAM values normalized once
// per request (via model.RAMFamily), rather than once per server compared -
// the filter's own values don't change while scanning the catalog.
type preparedFilter struct {
	model       string
	location    string
	diskType    string
	ramFamilies []string
	storageMin  *int
	storageMax  *int
}

func prepareFilter(f model.ServerFilter) preparedFilter {
	families := make([]string, len(f.RAM))
	for i, ram := range f.RAM {
		families[i] = model.RAMFamily(ram)
	}
	return preparedFilter{
		model:       f.Model,
		location:    f.Location,
		diskType:    f.DiskType,
		ramFamilies: families,
		storageMin:  f.StorageMin,
		storageMax:  f.StorageMax,
	}
}

// matchesFilter is the single filter-evaluation implementation shared by
// every Repository, so backends can never drift in what "matches the
// filter" means. It only ever compares precomputed fields - no parsing.
func matchesFilter(idx indexedServer, filter preparedFilter) bool {
	s := idx.server
	if filter.model != "" && !stringMatches(s.Model, filter.model) {
		return false
	}
	if len(filter.ramFamilies) > 0 && !containsString(filter.ramFamilies, idx.ramFamily) {
		return false
	}
	if filter.location != "" && !stringMatches(s.Location, filter.location) {
		return false
	}
	if filter.diskType != "" && !strings.EqualFold(filter.diskType, idx.diskType) {
		return false
	}
	if filter.storageMin != nil && idx.totalStorageGB < *filter.storageMin {
		return false
	}
	if filter.storageMax != nil && idx.totalStorageGB > *filter.storageMax {
		return false
	}
	return true
}

func containsString(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}

func stringMatches(value, filter string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	filter = strings.ToLower(strings.TrimSpace(filter))
	return strings.Contains(value, filter)
}
