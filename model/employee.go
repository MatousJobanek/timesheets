package model

import (
	"strings"
	"time"
)

// ExportFormatSimple is the Stundenliste layout. An empty ExportFormat means
// the detailed Stundenzettel, including employees saved before the choice existed.
const ExportFormatSimple = "simple"

type Employee struct {
	ID                  string       `json:"id"`
	Name                string       `json:"name"`
	Year                int          `json:"year"`
	UseEvenOdd          bool         `json:"use_even_odd"`
	WeekSchedule        WeekSchedule `json:"week_schedule"`
	EvenSchedule        WeekSchedule `json:"even_schedule"`
	OddSchedule         WeekSchedule `json:"odd_schedule"`
	FreePeriodOverrides []DateRange  `json:"free_period_overrides"`
	ExportFormat        string       `json:"export_format,omitempty"`
}

type YearConfig struct {
	Year        int         `json:"year"`
	FreePeriods []DateRange `json:"free_periods"`
}

type WeekSchedule struct {
	Monday    DaySchedule `json:"monday"`
	Tuesday   DaySchedule `json:"tuesday"`
	Wednesday DaySchedule `json:"wednesday"`
	Thursday  DaySchedule `json:"thursday"`
	Friday    DaySchedule `json:"friday"`
}

// ForDay returns the DaySchedule for the given weekday.
// Returns a zero-value DaySchedule for Saturday and Sunday.
func (ws WeekSchedule) ForDay(d time.Weekday) DaySchedule {
	switch d {
	case time.Monday:
		return ws.Monday
	case time.Tuesday:
		return ws.Tuesday
	case time.Wednesday:
		return ws.Wednesday
	case time.Thursday:
		return ws.Thursday
	case time.Friday:
		return ws.Friday
	default:
		return DaySchedule{}
	}
}

type DaySchedule struct {
	Begin1 string `json:"begin_1,omitempty"` // "07:00"
	End1   string `json:"end_1,omitempty"`
	Begin2 string `json:"begin_2,omitempty"`
	End2   string `json:"end_2,omitempty"`
}

// Works reports whether the first from–to pair is filled.
// The second pair is an optional later block; the gap between them is the break.
func (d DaySchedule) Works() bool {
	return strings.TrimSpace(d.Begin1) != "" && strings.TrimSpace(d.End1) != ""
}

type DateRange struct {
	Label string `json:"label"` // "Weihnachtsferien"
	From  string `json:"from"`  // "2026-12-24" (ISO 8601)
	To    string `json:"to"`    // "2027-01-06" (ISO 8601)
}
