package generator

import (
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
	"timesheets/holidays"
	"timesheets/model"
)

func Generate(emp model.Employee, freePeriods []model.DateRange) (*excelize.File, error) {
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

		if err := writeSheet(f, sheetName, emp, month, publicHolidays, freePeriods, styles); err != nil {
			return nil, fmt.Errorf("Fehler bei %s: %w", sheetName, err)
		}
	}

	return f, nil
}

func writeSheet(f *excelize.File, sheet string, emp model.Employee, month time.Month, publicHolidays map[time.Time]string, freePeriods []model.DateRange, styles *Styles) error {
	// Header row 1: employee name and month/year
	f.SetCellValue(sheet, "A1", fmt.Sprintf("%s — %s %d", emp.Name, holidays.MonthName(month), emp.Year))
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14},
	})
	f.SetCellStyle(sheet, "A1", "H1", titleStyle)
	f.MergeCell(sheet, "A1", "H1")

	// Column headers in row 3 (with top border for the table frame)
	headers := []string{"Datum", "Wochentag", "Beginn 1", "Ende 1", "Beginn 2", "Ende 2", "Stunden", "Anmerkung"}
	for i, h := range headers {
		cell := cellRef(i, 3)
		f.SetCellValue(sheet, cell, h)
		style := styles.withBorder(f, styles.Header, true, false, i == 0, i == 7)
		f.SetCellStyle(sheet, cell, cell, style)
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
		} else if label := freePeriodLabel(date, freePeriods); label != "" {
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

		// Apply row style with outer border on edge cells
		isLastRow := day == daysInMonth-1
		for col := 0; col <= 7; col++ {
			var base int
			if col <= 1 || col == 7 {
				base = rowStyle
			} else {
				base = timeStyle
			}
			needLeft := col == 0
			needRight := col == 7
			if needLeft || needRight || isLastRow {
				style := styles.withBorder(f, base, false, isLastRow, needLeft, needRight)
				f.SetCellStyle(sheet, cellRef(col, row), cellRef(col, row), style)
			} else {
				f.SetCellStyle(sheet, cellRef(col, row), cellRef(col, row), base)
			}
		}
	}

	// Summary section
	lastDataRow := dataStartRow + daysInMonth - 1
	sumLabelStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})

	// Total hours (Gesamt)
	sumRow := lastDataRow + 2
	f.SetCellValue(sheet, cellRef(5, sumRow), "Gesamt:")
	f.SetCellStyle(sheet, cellRef(5, sumRow), cellRef(5, sumRow), sumLabelStyle)
	f.SetCellFormula(sheet, cellRef(6, sumRow), fmt.Sprintf("SUM(G%d:G%d)", dataStartRow, lastDataRow))
	f.SetCellStyle(sheet, cellRef(6, sumRow), cellRef(6, sumRow), styles.TimeFormat)

	// Per-weekday hours breakdown
	weekdaySums := [5]int{}
	for day := 0; day < daysInMonth; day++ {
		date := firstDay.AddDate(0, 0, day)
		wd := date.Weekday()
		if wd == time.Saturday || wd == time.Sunday {
			continue
		}
		if _, ok := publicHolidays[date]; ok {
			continue
		}
		if freePeriodLabel(date, freePeriods) != "" {
			continue
		}
		schedule := resolveSchedule(emp, date)
		if schedule.TotalMinutes > 0 {
			weekdaySums[wd-1] += schedule.TotalMinutes
		}
	}

	weekdayNames := []string{"Montag:", "Dienstag:", "Mittwoch:", "Donnerstag:", "Freitag:"}
	for i, name := range weekdayNames {
		row := sumRow + 2 + i
		f.SetCellValue(sheet, cellRef(5, row), name)
		f.SetCellStyle(sheet, cellRef(5, row), cellRef(5, row), sumLabelStyle)
		f.SetCellValue(sheet, cellRef(6, row), minutesToExcel(weekdaySums[i]))
		f.SetCellStyle(sheet, cellRef(6, row), cellRef(6, row), styles.TimeFormat)
	}

	// Column widths
	colWidths := map[string]float64{
		"A": 12, "B": 13, "C": 9, "D": 9, "E": 9, "F": 13, "G": 9, "H": 22,
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
