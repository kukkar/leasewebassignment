package model

import "testing"

func TestParsePrice(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected float64
		wantErr  bool
	}{
		{"valid euros", "€49.99", 49.99, false},
		{"valid dollars", "$105.99", 105.99, false},
		{"valid singapore dollars", "S$565.99", 565.99, false},
		{"valid plain", "119.00", 119.00, false},
		{"invalid string", "abc", 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParsePrice(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ParsePrice(%q) error = %v", tc.input, err)
			}
			if err == nil && got != tc.expected {
				t.Fatalf("ParsePrice(%q) = %v, want %v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestParseRAM(t *testing.T) {
	cases := []struct {
		input    string
		expected int
		wantErr  bool
	}{
		{"16GB", 16, false},
		{"64GB", 64, false},
		{"invalid", 0, true},
	}
	for _, tc := range cases {
		got, err := ParseRAM(tc.input)
		if (err != nil) != tc.wantErr {
			t.Fatalf("ParseRAM(%q) error = %v", tc.input, err)
		}
		if err == nil && got != tc.expected {
			t.Fatalf("ParseRAM(%q) = %d, want %d", tc.input, got, tc.expected)
		}
	}
}

func TestValidateRAM(t *testing.T) {
	if err := ValidateRAM("16GB"); err != nil {
		t.Fatalf("expected valid ram, got %v", err)
	}
	if err := ValidateRAM("bad"); err == nil {
		t.Fatal("expected invalid ram error")
	}
}

func TestValidateHDD(t *testing.T) {
	if err := ValidateHDD("2x2TBSATA2"); err != nil {
		t.Fatalf("expected valid hdd, got %v", err)
	}
	if err := ValidateHDD("bad"); err == nil {
		t.Fatal("expected invalid hdd error")
	}
}

// TestParseHDD_NormalizesDiskTypeFamily locks in the acceptance criterion
// that disk_type filtering works against the assignment's SAS/SATA/SSD
// dropdown values even though the source data labels drives with a
// generation suffix (e.g. "SATA2").
func TestParseHDD_NormalizesDiskTypeFamily(t *testing.T) {
	cases := []struct {
		hdd          string
		wantTotalGB  int
		wantDiskType string
	}{
		{"2x2TBSATA2", 4096, "SATA"},
		{"8x300GBSAS", 2400, "SAS"},
		{"4x480GBSSD", 1920, "SSD"},
		{"24x1TBSATA2", 24576, "SATA"},
	}
	for _, tc := range cases {
		t.Run(tc.hdd, func(t *testing.T) {
			gotTotal, gotType, err := ParseHDD(tc.hdd)
			if err != nil {
				t.Fatalf("ParseHDD(%q) error = %v", tc.hdd, err)
			}
			if gotTotal != tc.wantTotalGB {
				t.Fatalf("ParseHDD(%q) total = %d, want %d", tc.hdd, gotTotal, tc.wantTotalGB)
			}
			if gotType != tc.wantDiskType {
				t.Fatalf("ParseHDD(%q) diskType = %q, want %q", tc.hdd, gotType, tc.wantDiskType)
			}
		})
	}
}

func TestNormalizeDiskType(t *testing.T) {
	cases := map[string]string{
		"SATA2": "SATA",
		"SATA3": "SATA",
		"SAS":   "SAS",
		"ssd":   "SSD",
		"  sas": "SAS",
	}
	for input, want := range cases {
		if got := NormalizeDiskType(input); got != want {
			t.Fatalf("NormalizeDiskType(%q) = %q, want %q", input, got, want)
		}
	}
}

// TestRAMFamily locks in the acceptance criterion that ram filtering works
// against the real dataset, where every RAM value has a memory technology
// suffix glued directly onto it (e.g. "16GBDDR3"), not the clean "16GB"
// the assignment's checkbox filter values use.
func TestRAMFamily(t *testing.T) {
	cases := map[string]string{
		"16GBDDR3":  "16GB",
		"128GBDDR4": "128GB",
		"4GBDDR3":   "4GB",
		"16GB":      "16GB",
		"  32gb  ":  "32GB",
	}
	for input, want := range cases {
		if got := RAMFamily(input); got != want {
			t.Fatalf("RAMFamily(%q) = %q, want %q", input, got, want)
		}
	}

	// A "4GB" filter must not spuriously match "48GBDDR3" - this is the
	// substring-matching bug the family extraction is specifically meant to avoid.
	if RAMFamily("48GBDDR3") == RAMFamily("4GB") {
		t.Fatalf("RAMFamily(%q) must not equal RAMFamily(%q)", "48GBDDR3", "4GB")
	}
}

func TestParseStorageValue(t *testing.T) {
	cases := []struct {
		input    string
		expected int
		wantErr  bool
	}{
		{"0", 0, false},
		{"250GB", 250, false},
		{"1TB", 1024, false},
		{"bad", 0, true},
	}
	for _, tc := range cases {
		got, err := ParseStorageValue(tc.input)
		if (err != nil) != tc.wantErr {
			t.Fatalf("ParseStorageValue(%q) error = %v", tc.input, err)
		}
		if err == nil && got != tc.expected {
			t.Fatalf("ParseStorageValue(%q) = %d, want %d", tc.input, got, tc.expected)
		}
	}
}
