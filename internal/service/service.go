package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/sahil/leasewebassignment/internal/model"
	"github.com/sahil/leasewebassignment/internal/store"
)

type Service interface {
	GetServers(ctx context.Context, filter model.ServerFilter) ([]model.Server, error)
	UploadServerData(ctx context.Context, reader io.Reader) error
}

type ServerService struct {
	repo store.Repository
}

func NewServerService(repo store.Repository) *ServerService {
	return &ServerService{repo: repo}
}

func (s *ServerService) GetServers(ctx context.Context, filter model.ServerFilter) ([]model.Server, error) {
	return s.repo.ListServers(ctx, filter)
}

func (s *ServerService) UploadServerData(ctx context.Context, reader io.Reader) error {
	servers, err := parseCSV(reader)
	if err != nil {
		return &ServiceError{Op: "upload", Err: err}
	}
	if err := s.repo.ReplaceServers(ctx, servers); err != nil {
		return &ServiceError{Op: "upload", Err: err}
	}
	return nil
}

func parseCSV(reader io.Reader) ([]model.Server, error) {
	r := csv.NewReader(reader)
	rows, err := r.ReadAll()
	if err != nil {
		return nil, &CSVParseError{Reason: "read csv", Err: err}
	}
	if len(rows) < 1 {
		return nil, &CSVParseError{Reason: "empty csv", Err: ErrInvalidUpload}
	}
	headers := rows[0]
	expected := []string{"Model", "RAM", "HDD", "Location", "Price"}
	for i, want := range expected {
		if i >= len(headers) || headers[i] != want {
			return nil, &CSVParseError{
				Column: "header",
				Reason: "unexpected header",
				Err:    fmt.Errorf("%w: got %q want %q", store.ErrInvalidCSVHeader, headers[i], want),
			}
		}
	}
	parsed := make([]model.Server, 0, len(rows)-1)
	for i, row := range rows[1:] {
		if len(row) < 5 {
			return nil, &CSVParseError{
				Row:    i + 2,
				Reason: "too few columns",
				Err:    ErrInvalidUpload,
			}
		}
		price, err := model.ParsePrice(row[4])
		if err != nil {
			return nil, &CSVParseError{
				Row:    i + 2,
				Column: "Price",
				Reason: "invalid price",
				Err:    err,
			}
		}
		parsed = append(parsed, model.Server{
			Model:    strings.TrimSpace(row[0]),
			RAM:      strings.TrimSpace(row[1]),
			HDD:      strings.TrimSpace(row[2]),
			Location: strings.TrimSpace(row[3]),
			Price:    price,
		})
	}
	return parsed, nil
}
