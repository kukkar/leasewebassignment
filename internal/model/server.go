package model

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	ramPattern = regexp.MustCompile(`^(\d+)(GB)$`)
	hddPattern = regexp.MustCompile(`^(\d+)x(\d+)(GB|TB)([A-Za-z0-9]+)$`)
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
type ServerFilter struct {
	Model      string
	RAM        string
	HDD        string
	Location   string
	PriceMin   *float64
	PriceMax   *float64
	DiskType   string
	StorageMin *int
	StorageMax *int
	Search     string
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
	cleaned := strings.TrimSpace(strings.TrimPrefix(value, "€"))
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
	return count * size, unit, nil
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
