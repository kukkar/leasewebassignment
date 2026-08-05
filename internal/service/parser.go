package service

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/sahil/leasewebassignment/internal/model"
	"github.com/sahil/leasewebassignment/internal/store"
)

// Parser parses an upload payload into domain servers.
type Parser interface {
	Parse(io.Reader) ([]model.Server, error)
}

// CSVParser is the default CSV parser implementation.
type CSVParser struct{}

func NewCSVParser() *CSVParser { return &CSVParser{} }

func (p *CSVParser) Parse(reader io.Reader) ([]model.Server, error) {
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
		got := ""
		if i < len(headers) {
			got = headers[i]
		}
		if got != want {
			return nil, &CSVParseError{
				Column: "header",
				Reason: "unexpected header",
				Err:    fmt.Errorf("%w: got %q want %q", store.ErrInvalidCSVHeader, got, want),
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
