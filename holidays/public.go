package holidays

import (
	"time"

	"github.com/wlbr/feiertage"
)

// Austrian German month names
var MonthNames = [12]string{
	"Jänner", "Februar", "März", "April", "Mai", "Juni",
	"Juli", "August", "September", "Oktober", "November", "Dezember",
}

// Austrian German weekday names (Sunday=0 in time.Weekday)
var WeekdayNames = [7]string{
	"Sonntag", "Montag", "Dienstag", "Mittwoch",
	"Donnerstag", "Freitag", "Samstag",
}

func MonthName(m time.Month) string {
	return MonthNames[m-1]
}

func WeekdayName(d time.Weekday) string {
	return WeekdayNames[d]
}

// PublicHolidays returns all Austrian public holidays for a given year.
func PublicHolidays(year int) map[time.Time]string {
	region := feiertage.Tirol(year)
	holidays := make(map[time.Time]string, len(region.Feiertage))
	for _, f := range region.Feiertage {
		date := time.Date(f.Year(), f.Month(), f.Day(), 0, 0, 0, 0, time.Local)
		holidays[date] = f.Text
	}
	return holidays
}

// IsPublicHoliday checks if the given date is an Austrian public holiday.
// Returns the holiday name if it is, empty string otherwise.
func IsPublicHoliday(date time.Time) (bool, string) {
	normalized := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
	holidays := PublicHolidays(date.Year())
	if name, ok := holidays[normalized]; ok {
		return true, name
	}
	return false, ""
}
