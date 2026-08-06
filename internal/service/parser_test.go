package service

import (
	"strings"
	"testing"
)

// TestCSVParser_IgnoresTrailingSpreadsheetColumns locks in a real production
// bug found in data/servers.csv: the source spreadsheet export carries a
// "Filters" reference table in columns 6-9 on every single row (mostly
// empty, but always present due to a fixed column count). The parser must
// not treat those as a continuation of Location.
func TestCSVParser_IgnoresTrailingSpreadsheetColumns(t *testing.T) {
	csvData := "Model,RAM,HDD,Location,Price,,Filters,,\n" +
		"Dell R210Intel Xeon X3440,16GBDDR3,2x2TBSATA2,AmsterdamAMS-01,€49.99,,Name,Type,Values\n" +
		"HP DL180G62x Intel Xeon E5620,32GBDDR3,8x2TBSATA2,AmsterdamAMS-01,€119.00,,Storage,Range slider,\"0, 250GB, 500GB\"\n" +
		"Dell R210-IIIntel Xeon E3-1230v2,16GBDDR3,2x2TBSATA2,AmsterdamAMS-01,€72.99,,,,\n"

	servers, err := NewCSVParser().Parse(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(servers) != 3 {
		t.Fatalf("expected 3 servers, got %d", len(servers))
	}
	for _, s := range servers {
		if s.Location != "AmsterdamAMS-01" {
			t.Fatalf("expected clean location %q, got %q (model %q)", "AmsterdamAMS-01", s.Location, s.Model)
		}
	}
}
