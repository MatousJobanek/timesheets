package model

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

func SaveConfig(path string, emp Employee) error {
	data, err := json.MarshalIndent(emp, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON-Serialisierung fehlgeschlagen: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func LoadConfig(path string) (Employee, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Employee{}, fmt.Errorf("Datei konnte nicht gelesen werden: %w", err)
	}
	var emp Employee
	if err := json.Unmarshal(data, &emp); err != nil {
		return Employee{}, fmt.Errorf("JSON-Deserialisierung fehlgeschlagen: %w", err)
	}
	return emp, nil
}

// DateToISO converts DD.MM.YYYY to YYYY-MM-DD.
func DateToISO(ddmmyyyy string) (string, error) {
	t, err := time.Parse("02.01.2006", strings.TrimSpace(ddmmyyyy))
	if err != nil {
		return "", fmt.Errorf("ungültiges Datum %q (erwartet DD.MM.YYYY): %w", ddmmyyyy, err)
	}
	return t.Format("2006-01-02"), nil
}

// DateFromISO converts YYYY-MM-DD to DD.MM.YYYY.
func DateFromISO(iso string) (string, error) {
	t, err := time.Parse("2006-01-02", strings.TrimSpace(iso))
	if err != nil {
		return "", fmt.Errorf("ungültiges ISO-Datum %q: %w", iso, err)
	}
	return t.Format("02.01.2006"), nil
}

// ParseISO parses YYYY-MM-DD to time.Time.
func ParseISO(iso string) (time.Time, error) {
	return time.Parse("2006-01-02", strings.TrimSpace(iso))
}
