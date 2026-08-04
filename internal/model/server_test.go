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
