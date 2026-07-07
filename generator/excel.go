package generator

import (
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
	"timesheets/holidays"
	"timesheets/model"
)

func Generate(emp model.Employee) (*excelize.File, error) {
	f := excelize.NewFile()

	styles, err := createStyles(f)
	if err != nil {
		return nil, fmt.Errorf("Stile konnten nicht erstellt werden: %w", err)
	}

	publicHolidays := holidays.PublicHolidays(emp.Year)

	for month := time.January; month <= time.December; month++ {
		sheetName := fmt.Sprintf("%s %d", holidays.MonthName(month), emp.Year)

		if month == time.January {
			f.SetSheetName("Sheet1", sheetName)
		} else {
			if _, err := f.NewSheet(sheetName); err != nil {
				return nil, err
			}
		}

		if err := writeSheet(f, sheetName, emp, month, publicHolidays, styles); err != nil {
			return nil, fmt.Errorf("Fehler bei %s: %w", sheetName, err)
		}
	}

	return f, nil
}

func writeSheet(f *excelize.File, sheet string, emp model.Employee, month time.Month, publicHolidays map[time.Time]string, styles *Styles) error {
	// Header row 1: employee name and month/year
	f.SetCellValue(sheet, "A1", fmt.Sprintf("%s — %s %d", emp.Name, holidays.MonthName(month), emp.Year))
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14},
	})
	f.SetCellStyle(sheet, "A1", "H1", titleStyle)
	f.MergeCell(sheet, "A1", "H1")

	// Column headers in row 3
	headers := []string{"Datum", "Wochentag", "Beginn 1", "Ende 1", "Beginn 2", "Ende 2", "Stunden", "Anmerkung"}
	for i, h := range headers {
		cell := cellRef(i, 3)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, styles.Header)
	}

	// Data rows starting at row 4
	firstDay := time.Date(emp.Year, month, 1, 0, 0, 0, 0, time.Local)
	daysInMonth := daysIn(month, emp.Year)
	dataStartRow := 4

	for day := 0; day < daysInMonth; day++ {
		date := firstDay.AddDate(0, 0, day)
		row := dataStartRow + day

		f.SetCellValue(sheet, cellRef(0, row), date.Format("02.01.2006"))
		f.SetCellValue(sheet, cellRef(1, row), holidays.WeekdayName(date.Weekday()))

		rowStyle := styles.Normal
		timeStyle := styles.TimeFormat
		anmerkung := ""

		if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
			rowStyle = styles.Weekend
			timeStyle = styles.TimeWeekend
			anmerkung = "Wochenende"
		} else if name, ok := publicHolidays[date]; ok {
			rowStyle = styles.PublicHoliday
			timeStyle = styles.TimeHoliday
			anmerkung = name
		} else if label := freePeriodLabel(date, emp.FreePeriods); label != "" {
			rowStyle = styles.SchoolHoliday
			timeStyle = styles.TimeSchool
			anmerkung = label
		} else {
			schedule := resolveSchedule(emp, date)
			if schedule.TotalMinutes > 0 {
				writeTimeBlocks(f, sheet, row, schedule, timeStyle)
			}
		}

		if anmerkung != "" {
			f.SetCellValue(sheet, cellRef(7, row), anmerkung)
		}

		// Apply row style to non-time columns
		for col := 0; col <= 1; col++ {
			f.SetCellStyle(sheet, cellRef(col, row), cellRef(col, row), rowStyle)
		}
		for col := 7; col <= 7; col++ {
			f.SetCellStyle(sheet, cellRef(col, row), cellRef(col, row), rowStyle)
		}
		// Time columns get time-aware style
		for col := 2; col <= 6; col++ {
			f.SetCellStyle(sheet, cellRef(col, row), cellRef(col, row), timeStyle)
		}
	}

	// SUM formula for Stunden column
	lastDataRow := dataStartRow + daysInMonth - 1
	sumRow := lastDataRow + 2
	sumCell := cellRef(6, sumRow)
	f.SetCellFormula(sheet, sumCell, fmt.Sprintf("SUM(G%d:G%d)", dataStartRow, lastDataRow))
	f.SetCellStyle(sheet, sumCell, sumCell, styles.TimeFormat)

	labelCell := cellRef(5, sumRow)
	f.SetCellValue(sheet, labelCell, "Gesamt:")
	sumLabelStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	f.SetCellStyle(sheet, labelCell, labelCell, sumLabelStyle)

	// Column widths
	colWidths := map[string]float64{
		"A": 12, "B": 13, "C": 9, "D": 9, "E": 9, "F": 9, "G": 9, "H": 22,
	}
	for col, w := range colWidths {
		f.SetColWidth(sheet, col, col, w)
	}

	return nil
}

func writeTimeBlocks(f *excelize.File, sheet string, row int, sched model.DaySchedule, timeStyle int) {
	startH, startM := parseTime(sched.StartTime)
	startMinutes := startH*60 + startM

	if sched.TotalMinutes <= 360 {
		// No break needed
		endMinutes := startMinutes + sched.TotalMinutes
		f.SetCellValue(sheet, cellRef(2, row), timeToExcel(startMinutes))
		f.SetCellValue(sheet, cellRef(3, row), timeToExcel(endMinutes))
		f.SetCellValue(sheet, cellRef(6, row), minutesToExcel(sched.TotalMinutes))
	} else {
		// Break after 4 hours (240 min)
		end1Minutes := startMinutes + 240
		begin2Minutes := end1Minutes + 30
		end2Minutes := begin2Minutes + (sched.TotalMinutes - 240)

		f.SetCellValue(sheet, cellRef(2, row), timeToExcel(startMinutes))
		f.SetCellValue(sheet, cellRef(3, row), timeToExcel(end1Minutes))
		f.SetCellValue(sheet, cellRef(4, row), timeToExcel(begin2Minutes))
		f.SetCellValue(sheet, cellRef(5, row), timeToExcel(end2Minutes))
		f.SetCellValue(sheet, cellRef(6, row), minutesToExcel(sched.TotalMinutes))
	}
}

func resolveSchedule(emp model.Employee, date time.Time) model.DaySchedule {
	if !emp.UseEvenOdd {
		return emp.WeekSchedule.ForDay(date.Weekday())
	}
	_, week := date.ISOWeek()
	if week%2 == 0 {
		return emp.EvenSchedule.ForDay(date.Weekday())
	}
	return emp.OddSchedule.ForDay(date.Weekday())
}

func freePeriodLabel(date time.Time, periods []model.DateRange) string {
	for _, p := range periods {
		from, err1 := model.ParseISO(p.From)
		to, err2 := model.ParseISO(p.To)
		if err1 != nil || err2 != nil {
			continue
		}
		if !date.Before(from) && !date.After(to) {
			return p.Label
		}
	}
	return ""
}

// timeToExcel converts minutes-since-midnight to an Excel time fraction (0.0–1.0).
func timeToExcel(totalMinutes int) float64 {
	return float64(totalMinutes) / 1440.0
}

// minutesToExcel converts a duration in minutes to an Excel time fraction.
func minutesToExcel(minutes int) float64 {
	return float64(minutes) / 1440.0
}

func parseTime(hhmm string) (int, int) {
	var h, m int
	fmt.Sscanf(hhmm, "%d:%d", &h, &m)
	return h, m
}

func cellRef(col, row int) string {
	return fmt.Sprintf("%c%d", 'A'+col, row)
}

func daysIn(m time.Month, year int) int {
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.Local).Day()
}
