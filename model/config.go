package model

import (
	"fmt"
	"strings"
	"time"
)

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
