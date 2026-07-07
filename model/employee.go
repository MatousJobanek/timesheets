package model

import "time"

type Employee struct {
	Name         string       `json:"name"`
	Year         int          `json:"year"`
	UseEvenOdd   bool         `json:"use_even_odd"`
	WeekSchedule WeekSchedule `json:"week_schedule"`
	EvenSchedule WeekSchedule `json:"even_schedule"`
	OddSchedule  WeekSchedule `json:"odd_schedule"`
	FreePeriods  []DateRange  `json:"free_periods"`
}

type WeekSchedule struct {
	Monday    DaySchedule `json:"monday"`
	Tuesday   DaySchedule `json:"tuesday"`
	Wednesday DaySchedule `json:"wednesday"`
	Thursday  DaySchedule `json:"thursday"`
	Friday    DaySchedule `json:"friday"`
}

// ForDay returns the DaySchedule for the given weekday.
// Returns a zero-value DaySchedule (TotalMinutes=0) for Saturday/Sunday.
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
	StartTime    string `json:"start_time"`    // "07:00" (HH:MM format)
	TotalMinutes int    `json:"total_minutes"` // 480 = 8h, 260 = 4h20m, 0 = not working
}

type DateRange struct {
	Label string `json:"label"` // "Weihnachtsferien"
	From  string `json:"from"`  // "2026-12-24" (ISO 8601)
	To    string `json:"to"`    // "2027-01-06" (ISO 8601)
}
