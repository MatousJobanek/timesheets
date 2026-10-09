package generator

import (
	"fmt"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"timesheets/holidays"
	"timesheets/model"
)

const simpleHeaderGreen = "D1E6B9"
const simpleWeekendGrey = "EEEEEE"
const simpleFreeGrey = "F4F4F4"

const (
	simpleBannerRow   = 1
	simpleMonthRow    = 2
	simpleNameRow     = 3
	simpleHeaderRow   = 4
	simpleFirstDayRow = 5
)

// GenerateSimple writes one printable Stundenliste sheet per school-year month:
// planned duty as text, empty columns for handwritten changes, and no totals.
func GenerateSimple(emp model.Employee, freePeriods []model.DateRange) (*excelize.File, error) {
	f := excelize.NewFile()

	styles, err := createSimpleStyles(f)
	if err != nil {
		return nil, fmt.Errorf("Stile konnten nicht erstellt werden: %w", err)
	}

	publicHolidays := holidays.PublicHolidaysForSchoolYear(emp.Year)

	for i, month := range model.SchoolYearMonths() {
		calYear := model.CalendarYear(emp.Year, month)
		sheetName := fmt.Sprintf("%s %d", holidays.MonthName(month), calYear)

		if i == 0 {
			if err := f.SetSheetName("Sheet1", sheetName); err != nil {
				return nil, err
			}
		} else if _, err := f.NewSheet(sheetName); err != nil {
			return nil, err
		}

		if err := writeSimpleSheet(f, sheetName, emp, month, calYear, publicHolidays, freePeriods, styles); err != nil {
			return nil, fmt.Errorf("Fehler bei %s: %w", sheetName, err)
		}
	}

	return f, nil
}

type simpleStyles struct {
	banner   int
	month    int
	name     int
	header   int
	cell     int
	weekend  int
	free     int
	sigLine  int
	sigLabel int
}

func createSimpleStyles(f *excelize.File) (*simpleStyles, error) {
	s := &simpleStyles{}
	var err error

	green := excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{simpleHeaderGreen}}
	// Horizontal rules only: no vertical column separators and no frame around the table.
	rule := []excelize.Border{
		{Type: "bottom", Color: "000000", Style: 1},
	}

	s.banner, err = f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16, Color: "000000"},
		Fill:      green,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return nil, err
	}

	s.month, err = f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "000000"},
		Fill:      green,
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return nil, err
	}

	s.name, err = f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 12, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "top"},
	})
	if err != nil {
		return nil, err
	}

	s.header, err = f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Color: "000000"},
		Fill:      green,
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center", WrapText: true},
		Border:    rule,
	})
	if err != nil {
		return nil, err
	}

	s.cell, err = f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Color: "000000"},
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border:    rule,
	})
	if err != nil {
		return nil, err
	}

	s.weekend, err = f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Color: "000000"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{simpleWeekendGrey}},
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border:    rule,
	})
	if err != nil {
		return nil, err
	}

	s.free, err = f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Color: "000000"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{simpleFreeGrey}},
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border:    rule,
	})
	if err != nil {
		return nil, err
	}

	s.sigLine, err = f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	if err != nil {
		return nil, err
	}

	s.sigLabel, err = f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return nil, err
	}

	return s, nil
}

func writeSimpleSheet(f *excelize.File, sheet string, emp model.Employee, month time.Month, calYear int, publicHolidays map[time.Time]string, freePeriods []model.DateRange, styles *simpleStyles) error {
	// Compact title block on the right, flush with the table, as on the paper form.
	if err := f.MergeCell(sheet, "D1", "E1"); err != nil {
		return err
	}
	if err := f.MergeCell(sheet, "D2", "E2"); err != nil {
		return err
	}
	if err := f.MergeCell(sheet, "D3", "E3"); err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, "D1", "Stundenliste"); err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, "D2", fmt.Sprintf("%s %d", holidays.MonthName(month), calYear)); err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, "D3", emp.Name); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "D1", "E1", styles.banner); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "D2", "E2", styles.month); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "D3", "E3", styles.name); err != nil {
		return err
	}
	if err := f.SetRowHeight(sheet, simpleBannerRow, 24); err != nil {
		return err
	}
	if err := f.SetRowHeight(sheet, simpleMonthRow, 20); err != nil {
		return err
	}
	// Twice a normal row, so the name sits in the top-left with space beneath it.
	if err := f.SetRowHeight(sheet, simpleNameRow, 30); err != nil {
		return err
	}

	headers := []string{
		"Tag",
		"Wochentag",
		"Dienst laut Dienstplan",
		"Änderungen",
		"Verhinderung/Mehr-\nstunden",
	}
	for i, h := range headers {
		cell := cellRef(i, simpleHeaderRow)
		if err := f.SetCellValue(sheet, cell, h); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, cell, cell, styles.header); err != nil {
			return err
		}
	}
	if err := f.SetRowHeight(sheet, simpleHeaderRow, 32); err != nil {
		return err
	}

	firstDay := time.Date(calYear, month, 1, 0, 0, 0, 0, time.Local)
	daysInMonth := daysIn(month, calYear)
	for day := 0; day < daysInMonth; day++ {
		date := firstDay.AddDate(0, 0, day)
		row := simpleFirstDayRow + day

		if err := f.SetCellValue(sheet, cellRef(0, row), date.Format("02.1.2006")); err != nil {
			return err
		}
		if err := f.SetCellValue(sheet, cellRef(1, row), holidays.WeekdayName(date.Weekday())); err != nil {
			return err
		}
		if duty := simpleDuty(emp, date, publicHolidays, freePeriods); duty != "" {
			if err := f.SetCellValue(sheet, cellRef(2, row), duty); err != nil {
				return err
			}
		}
		rowStyle := simpleRowStyle(styles, date, publicHolidays, freePeriods)
		for col := 0; col <= 4; col++ {
			cell := cellRef(col, row)
			if err := f.SetCellStyle(sheet, cell, cell, rowStyle); err != nil {
				return err
			}
		}
	}

	lastDataRow := simpleFirstDayRow + daysInMonth - 1
	if err := writeSignature(f, sheet, lastDataRow+2, styles); err != nil {
		return err
	}

	widths := map[string]float64{
		"A": 14, "B": 16, "C": 34, "D": 16, "E": 30,
	}
	for col, w := range widths {
		if err := f.SetColWidth(sheet, col, col, w); err != nil {
			return err
		}
	}

	return setSimplePage(f, sheet)
}

func writeSignature(f *excelize.File, sheet string, lineRow int, styles *simpleStyles) error {
	labelRow := lineRow + 1

	// Datum stays in the fourth column. Unterschrift stays in the fifth.
	if err := f.SetCellStyle(sheet, cellRef(3, lineRow), cellRef(3, lineRow), styles.sigLine); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, cellRef(4, lineRow), cellRef(4, lineRow), styles.sigLine); err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, cellRef(3, labelRow), "Datum"); err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, cellRef(4, labelRow), "Unterschrift"); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, cellRef(3, labelRow), cellRef(3, labelRow), styles.sigLabel); err != nil {
		return err
	}
	return f.SetCellStyle(sheet, cellRef(4, labelRow), cellRef(4, labelRow), styles.sigLabel)
}

func simpleRowStyle(styles *simpleStyles, date time.Time, publicHolidays map[time.Time]string, freePeriods []model.DateRange) int {
	if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
		return styles.weekend
	}
	if _, ok := publicHolidays[date]; ok {
		return styles.weekend
	}
	if freePeriodLabel(date, freePeriods) != "" {
		return styles.free
	}
	return styles.cell
}

func setSimplePage(f *excelize.File, sheet string) error {
	size := 9 // A4
	orientation := "portrait"
	fit := 1
	if err := f.SetPageLayout(sheet, &excelize.PageLayoutOptions{
		Size:        &size,
		Orientation: &orientation,
		FitToWidth:  &fit,
		FitToHeight: &fit,
	}); err != nil {
		return err
	}
	fitToPage := true
	if err := f.SetSheetProps(sheet, &excelize.SheetPropsOptions{FitToPage: &fitToPage}); err != nil {
		return err
	}
	left, right := 0.25, 0.25
	top, bottom := 0.75, 0.75
	header, footer := 0.3, 0.3
	center := true
	return f.SetPageMargins(sheet, &excelize.PageLayoutMarginsOptions{
		Left:         &left,
		Right:        &right,
		Top:          &top,
		Bottom:       &bottom,
		Header:       &header,
		Footer:       &footer,
		Horizontally: &center,
	})
}

func simpleDuty(emp model.Employee, date time.Time, publicHolidays map[time.Time]string, freePeriods []model.DateRange) string {
	if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
		return ""
	}
	if name, ok := publicHolidays[date]; ok {
		return name
	}
	if label := freePeriodLabel(date, freePeriods); label != "" {
		return label
	}
	return formatDuty(resolveSchedule(emp, date))
}

func formatDuty(sched model.DaySchedule) string {
	var parts []string
	if pair := formatPair(sched.Begin1, sched.End1); pair != "" {
		parts = append(parts, pair)
	}
	if pair := formatPair(sched.Begin2, sched.End2); pair != "" {
		parts = append(parts, pair)
	}
	return strings.Join(parts, "    ")
}

func formatPair(begin, end string) string {
	begin = strings.TrimSpace(begin)
	end = strings.TrimSpace(end)
	if begin == "" || end == "" {
		return ""
	}
	return formatClock(begin) + " - " + formatClock(end)
}

func formatClock(hhmm string) string {
	h, m := parseTime(hhmm)
	return fmt.Sprintf("%d:%02d", h, m)
}
