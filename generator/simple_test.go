package generator

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"

	"timesheets/holidays"
	"timesheets/model"
)

func TestGenerateSimpleSeptemberSheet(t *testing.T) {
	emp := model.Employee{
		Name: "Lorna Olsen",
		Year: 2026,
		WeekSchedule: model.WeekSchedule{
			Tuesday:   model.DaySchedule{Begin1: "8:00", End1: "12:00"},
			Thursday:  model.DaySchedule{Begin1: "10:30", End1: "13:00", Begin2: "13:30", End2: "15:00"},
			Wednesday: model.DaySchedule{Begin1: "09:00", End1: "13:00"},
		},
	}
	free := []model.DateRange{{
		Label: "Teamtage",
		From:  "2026-09-10",
		To:    "2026-09-11",
	}}

	f, err := GenerateSimple(emp, free)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	sheet := "September 2026"
	idx, err := f.GetSheetIndex(sheet)
	if err != nil {
		t.Fatal(err)
	}
	if idx < 0 {
		t.Fatalf("missing sheet %s", sheet)
	}

	assertCell(t, f, sheet, "A1", "")
	assertCell(t, f, sheet, "D1", "Stundenliste")
	assertCell(t, f, sheet, "D2", "September 2026")
	assertCell(t, f, sheet, "A3", "")
	assertCell(t, f, sheet, "D3", "Lorna Olsen")
	assertCell(t, f, sheet, "A4", "Tag")
	assertCell(t, f, sheet, "B4", "Wochentag")
	assertCell(t, f, sheet, "C4", "Dienst laut Dienstplan")
	assertCell(t, f, sheet, "D4", "Änderungen")
	assertCell(t, f, sheet, "E4", "Verhinderung/Mehr-\nstunden")

	if got := fillColor(t, f, sheet, "D1"); got != simpleHeaderGreen {
		t.Fatalf("banner fill = %s", got)
	}
	if colors := styleColors(t, f, sheet, "A1"); len(colors) != 0 {
		t.Fatalf("left of the title should have no fill, got %v", colors)
	}
	if got := fillColor(t, f, sheet, "A4"); got != simpleHeaderGreen {
		t.Fatalf("header fill = %s", got)
	}
	if got := fillColor(t, f, sheet, "E4"); got != simpleHeaderGreen {
		t.Fatalf("last header fill = %s", got)
	}
	if colors := styleColors(t, f, sheet, "D3"); len(colors) != 0 {
		t.Fatalf("name row should have no fill, got %v", colors)
	}
	if align := horizontalAlign(t, f, sheet, "D3"); align != "left" {
		t.Fatalf("name horizontal alignment = %q", align)
	}
	if align := verticalAlign(t, f, sheet, "D3"); align != "top" {
		t.Fatalf("name vertical alignment = %q", align)
	}
	nameHeight, err := f.GetRowHeight(sheet, simpleNameRow)
	if err != nil {
		t.Fatal(err)
	}
	if nameHeight != 30 {
		t.Fatalf("name row height = %v, want 30", nameHeight)
	}
	if align := horizontalAlign(t, f, sheet, "A4"); align != "left" {
		t.Fatalf("header alignment = %q", align)
	}
	assertHorizontalRule(t, f, sheet, "A4")
	assertHorizontalRule(t, f, sheet, cellRef(0, simpleFirstDayRow))
	assertHorizontalRule(t, f, sheet, "E4")

	tuesday := findSeptemberDay(t, time.Tuesday, nil)
	wednesday := findSeptemberDay(t, time.Wednesday, nil)
	thursday := findSeptemberDay(t, time.Thursday, func(d time.Time) bool {
		return d.Day() == 10 || d.Day() == 11
	})
	saturday := findSeptemberDay(t, time.Saturday, nil)
	teamDay := time.Date(2026, time.September, 10, 0, 0, 0, 0, time.Local)

	assertCell(t, f, sheet, cellRef(0, simpleRow(tuesday)), tuesday.Format("02.1.2006"))
	assertCell(t, f, sheet, cellRef(1, simpleRow(tuesday)), holidays.WeekdayName(time.Tuesday))
	assertCell(t, f, sheet, cellRef(2, simpleRow(tuesday)), "8:00 - 12:00")
	assertCell(t, f, sheet, cellRef(3, simpleRow(tuesday)), "")

	assertCell(t, f, sheet, cellRef(2, simpleRow(wednesday)), "9:00 - 13:00")
	assertCell(t, f, sheet, cellRef(2, simpleRow(thursday)), "10:30 - 13:00    13:30 - 15:00")
	assertCell(t, f, sheet, cellRef(2, simpleRow(teamDay)), "Teamtage")
	if got := fillColor(t, f, sheet, cellRef(0, simpleRow(teamDay))); got != simpleFreeGrey {
		t.Fatalf("free-day fill = %s", got)
	}
	assertCell(t, f, sheet, cellRef(2, simpleRow(saturday)), "")
	if got := fillColor(t, f, sheet, cellRef(0, simpleRow(saturday))); got != simpleWeekendGrey {
		t.Fatalf("weekend fill = %s", got)
	}
	if got := fillColor(t, f, sheet, cellRef(4, simpleRow(saturday))); got != simpleWeekendGrey {
		t.Fatalf("weekend last-column fill = %s", got)
	}
	if colors := styleColors(t, f, sheet, cellRef(0, simpleRow(tuesday))); len(colors) != 0 {
		t.Fatalf("weekday should have no fill, got %v", colors)
	}

	holiday, holidayName := weekdayHoliday(t)
	holidaySheet := fmt.Sprintf("%s %d", holidays.MonthName(holiday.Month()), holiday.Year())
	holidayCell := cellRef(0, simpleRow(holiday))
	assertCell(t, f, holidaySheet, cellRef(2, simpleRow(holiday)), holidayName)
	if got := fillColor(t, f, holidaySheet, holidayCell); got != simpleWeekendGrey {
		t.Fatalf("public holiday fill = %s", got)
	}

	labelRow := simpleFirstDayRow + 30 + 2
	assertCell(t, f, sheet, cellRef(3, labelRow), "Datum")
	assertCell(t, f, sheet, cellRef(4, labelRow), "Unterschrift")
	assertCell(t, f, sheet, cellRef(1, labelRow), "")
	assertCell(t, f, sheet, cellRef(2, labelRow), "")

	for col, want := range map[string]float64{"C": 34, "D": 16, "E": 30} {
		got, err := f.GetColWidth(sheet, col)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("column %s width = %v, want %v", col, got, want)
		}
	}

	rows, err := f.GetRows(sheet)
	if err != nil {
		t.Fatal(err)
	}
	var sawDatum, sawUnterschrift bool
	for _, row := range rows {
		for _, cell := range row {
			if strings.Contains(cell, "Gesamt") {
				t.Fatalf("simple sheet contains a summary: %q", cell)
			}
			if cell == "Datum" {
				sawDatum = true
			}
			if cell == "Unterschrift" {
				sawUnterschrift = true
			}
		}
	}
	if !sawDatum || !sawUnterschrift {
		t.Fatalf("signature labels missing (Datum=%v Unterschrift=%v)", sawDatum, sawUnterschrift)
	}

	layout, err := f.GetPageLayout(sheet)
	if err != nil {
		t.Fatal(err)
	}
	if layout.Orientation == nil || *layout.Orientation != "portrait" {
		t.Fatalf("orientation = %v", layout.Orientation)
	}
	if layout.Size == nil || *layout.Size != 9 {
		t.Fatalf("paper size = %v", layout.Size)
	}
	if layout.FitToWidth == nil || *layout.FitToWidth != 1 || layout.FitToHeight == nil || *layout.FitToHeight != 1 {
		t.Fatalf("fit = %v x %v", layout.FitToWidth, layout.FitToHeight)
	}
	props, err := f.GetSheetProps(sheet)
	if err != nil {
		t.Fatal(err)
	}
	if props.FitToPage == nil || !*props.FitToPage {
		t.Fatal("FitToPage is not enabled")
	}
}

func simpleRow(date time.Time) int {
	return simpleFirstDayRow + date.Day() - 1
}

func findSeptemberDay(t *testing.T, weekday time.Weekday, skip func(time.Time) bool) time.Time {
	t.Helper()
	d := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.Local)
	for d.Month() == time.September {
		if d.Weekday() == weekday && (skip == nil || !skip(d)) {
			return d
		}
		d = d.AddDate(0, 0, 1)
	}
	t.Fatalf("no %s in September 2026", weekday)
	return time.Time{}
}

func assertCell(t *testing.T, f *excelize.File, sheet, cell, want string) {
	t.Helper()
	got, err := f.GetCellValue(sheet, cell)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("%s = %q, want %q", cell, got, want)
	}
}

func fillColor(t *testing.T, f *excelize.File, sheet, cell string) string {
	t.Helper()
	colors := styleColors(t, f, sheet, cell)
	if len(colors) == 0 {
		t.Fatalf("cell %s has no fill", cell)
	}
	return strings.TrimPrefix(strings.ToUpper(colors[0]), "FF")
}

func styleColors(t *testing.T, f *excelize.File, sheet, cell string) []string {
	t.Helper()
	style := cellStyle(t, f, sheet, cell)
	if style.Fill.Color == nil {
		return nil
	}
	return style.Fill.Color
}

func horizontalAlign(t *testing.T, f *excelize.File, sheet, cell string) string {
	t.Helper()
	style := cellStyle(t, f, sheet, cell)
	if style.Alignment == nil {
		return ""
	}
	return style.Alignment.Horizontal
}

func verticalAlign(t *testing.T, f *excelize.File, sheet, cell string) string {
	t.Helper()
	style := cellStyle(t, f, sheet, cell)
	if style.Alignment == nil {
		return ""
	}
	return style.Alignment.Vertical
}

func weekdayHoliday(t *testing.T) (time.Time, string) {
	t.Helper()
	start := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(2027, time.August, 1, 0, 0, 0, 0, time.Local)
	for date, name := range holidays.PublicHolidaysForSchoolYear(2026) {
		if date.Before(start) || !date.Before(end) {
			continue
		}
		if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
			continue
		}
		return date, name
	}
	t.Fatal("no weekday public holiday in the school year")
	return time.Time{}, ""
}

func assertHorizontalRule(t *testing.T, f *excelize.File, sheet, cell string) {
	t.Helper()
	style := cellStyle(t, f, sheet, cell)
	var sawBottom bool
	for _, border := range style.Border {
		if border.Style == 0 {
			continue
		}
		switch border.Type {
		case "bottom":
			sawBottom = true
		case "left", "right", "top":
			t.Fatalf("%s has a %s border", cell, border.Type)
		}
	}
	if !sawBottom {
		t.Fatalf("%s has no bottom rule", cell)
	}
}

func cellStyle(t *testing.T, f *excelize.File, sheet, cell string) *excelize.Style {
	t.Helper()
	id, err := f.GetCellStyle(sheet, cell)
	if err != nil {
		t.Fatal(err)
	}
	style, err := f.GetStyle(id)
	if err != nil {
		t.Fatal(err)
	}
	return style
}
