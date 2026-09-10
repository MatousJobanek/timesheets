package model

import (
	"fmt"
	"strings"
	"time"
)

// FormatSchoolYear returns a display label like "2026/2027".
func FormatSchoolYear(start int) string {
	return fmt.Sprintf("%d/%d", start, start+1)
}

// FormatSchoolYearFile returns a filename-safe label like "2026-2027".
func FormatSchoolYearFile(start int) string {
	return fmt.Sprintf("%d-%d", start, start+1)
}

// CurrentSchoolYearStart returns the school-year start year for now.
// August–December belong to now.Year(); January–July belong to now.Year()-1.
func CurrentSchoolYearStart(now time.Time) int {
	if now.Month() >= time.August {
		return now.Year()
	}
	return now.Year() - 1
}

// CalendarYear returns the calendar year of month within school year start.
func CalendarYear(start int, month time.Month) int {
	if month >= time.August {
		return start
	}
	return start + 1
}

// SchoolYearMonths is August through July.
func SchoolYearMonths() []time.Month {
	return []time.Month{
		time.August, time.September, time.October, time.November, time.December,
		time.January, time.February, time.March, time.April, time.May, time.June, time.July,
	}
}

// ParseDDMMInSchoolYear parses a DD.MM date into the matching calendar year
// of the given school-year start (Aug–Dec → start, Jan–Jul → start+1).
func ParseDDMMInSchoolYear(ddmm string, start int) (time.Time, error) {
	ddmm = strings.TrimSpace(ddmm)
	ddmm = strings.TrimRight(ddmm, ".")
	t, err := time.Parse("2.1", ddmm)
	if err != nil {
		return time.Time{}, fmt.Errorf("ungültiges Datum %q (erwartet DD.MM): %w", ddmm, err)
	}
	year := CalendarYear(start, t.Month())
	return time.Date(year, t.Month(), t.Day(), 0, 0, 0, 0, time.Local), nil
}
