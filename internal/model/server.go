// Package model holds the domain types (Server, ServerFilter) and the
// parsing/normalization functions for the source CSV's field formats -
// price ("€49.99"), RAM ("16GBDDR3"), and HDD ("2x2TBSATA2"). Values here
// carry a technology/generation suffix the filter API doesn't use directly
// (RAMFamily, NormalizeDiskType), so parsing and filter-family extraction
// are kept as distinct steps rather than collapsed into one.
package model

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	ramPattern       = regexp.MustCompile(`^(\d+)(GB)$`)
	hddPattern       = regexp.MustCompile(`^(\d+)x(\d+)(GB|TB)([A-Za-z0-9]+)$`)
	ramFamilyPattern = regexp.MustCompile(`(?i)^\s*(\d+GB)`)
)

// Server represents a server catalog entry.
type Server struct {
	Model    string  `json:"model"`
	RAM      string  `json:"ram"`
	HDD      string  `json:"hdd"`
	Location string  `json:"location"`
	Price    float64 `json:"price"`
}

// ServerFilter defines allowed filter criteria.
// RAM is a slice because the assignment's filter spec models RAM as a
// checkbox group (multi-select): a server matches if its RAM equals ANY of
// the selected values. All other fields are single-value AND filters.
type ServerFilter struct {
	Model      string
	RAM        []string
	Location   string
	DiskType   string
	StorageMin *int
	StorageMax *int
}

func ValidateRAM(value string) error {
	if value == "" {
		return nil
	}
	if !ramPattern.MatchString(value) {
		return fmt.Errorf("invalid ram value: %q", value)
	}
	return nil
}

func ValidateHDD(value string) error {
	if value == "" {
		return nil
	}
	if !hddPattern.MatchString(value) {
		return fmt.Errorf("invalid hdd value: %q", value)
	}
	return nil
}

func ParsePrice(value string) (float64, error) {
	if value == "" {
		return 0, nil
	}
	cleaned := strings.TrimSpace(value)
	if strings.HasPrefix(cleaned, "€") {
		cleaned = strings.TrimPrefix(cleaned, "€")
	} else if strings.HasPrefix(cleaned, "S$") {
		cleaned = strings.TrimPrefix(cleaned, "S$")
	} else if strings.HasPrefix(cleaned, "$") {
		cleaned = strings.TrimPrefix(cleaned, "$")
	}
	cleaned = strings.ReplaceAll(cleaned, ",", "")
	return strconv.ParseFloat(cleaned, 64)
}

func ParseRAM(value string) (int, error) {
	matches := ramPattern.FindStringSubmatch(strings.TrimSpace(value))
	if len(matches) != 3 {
		return 0, fmt.Errorf("invalid ram format: %q", value)
	}
	return strconv.Atoi(matches[1])
}

func ParseHDD(value string) (int, string, error) {
	matches := hddPattern.FindStringSubmatch(strings.TrimSpace(value))
	if len(matches) != 5 {
		return 0, "", fmt.Errorf("invalid hdd format: %q", value)
	}
	count, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, "", err
	}
	size, err := strconv.Atoi(matches[2])
	if err != nil {
		return 0, "", err
	}
	unit := matches[3]
	total := count * size
	if strings.EqualFold(unit, "TB") {
		total = total * 1024
	}
	return total, NormalizeDiskType(matches[4]), nil
}

// diskTypeGenerationSuffix matches a trailing generation/revision number on a
// disk type label, e.g. the "2" in "SATA2".
var diskTypeGenerationSuffix = regexp.MustCompile(`\d+$`)

// NormalizeDiskType maps a raw disk type label parsed from the HDD column
// (e.g. "SATA2", "SATA3") to the disk type family used by the filter API
// (e.g. "SATA"). The source dataset labels drives with a generation suffix,
// but the assignment's own filter spec only exposes SAS/SATA/SSD as filter
// values, so labels must collapse to that family for `disk_type` to ever match.
func NormalizeDiskType(raw string) string {
	return diskTypeGenerationSuffix.ReplaceAllString(strings.ToUpper(strings.TrimSpace(raw)), "")
}

// RAMFamily extracts the "<digits>GB" family from a raw RAM label for
// filter comparison purposes. Like disk type, the source data appends a
// memory technology suffix directly onto the value with no separator (e.g.
// "16GBDDR3", "128GBDDR4"), which the assignment's RAM checkbox values
// (plain "16GB") would never equal under exact comparison. Falls back to the
// upper-cased, trimmed input if no leading "<digits>GB" is found.
func RAMFamily(raw string) string {
	if m := ramFamilyPattern.FindStringSubmatch(raw); m != nil {
		return strings.ToUpper(m[1])
	}
	return strings.ToUpper(strings.TrimSpace(raw))
}

func ParseStorageValue(value string) (int, error) {
	if value == "" || strings.EqualFold(value, "0") {
		return 0, nil
	}
	trimmed := strings.TrimSpace(value)
	if strings.HasSuffix(trimmed, "TB") {
		parts := strings.TrimSuffix(trimmed, "TB")
		amount, err := strconv.Atoi(parts)
		if err != nil {
			return 0, err
		}
		return amount * 1024, nil
	}
	if strings.HasSuffix(trimmed, "GB") {
		parts := strings.TrimSuffix(trimmed, "GB")
		return strconv.Atoi(parts)
	}
	return 0, fmt.Errorf("invalid storage unit: %q", value)
}
