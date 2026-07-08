package ui

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"timesheets/model"
)

type HolidayRow struct {
	LabelEntry *widget.Entry
	FromEntry  *widget.Entry
	ToEntry    *widget.Entry
}

type HolidaysInput struct {
	Container *fyne.Container
	rows      []*HolidayRow
	vbox      *fyne.Container
	year      int
	OnChanged func()
}

func NewHolidaysInput(year int) *HolidaysInput {
	hi := &HolidaysInput{year: year}
	hi.vbox = container.NewVBox()

	addBtn := widget.NewButton("Hinzufügen", func() {
		hi.AddRow("", "", "")
		if hi.OnChanged != nil {
			hi.OnChanged()
		}
	})

	hi.Container = container.NewVBox(
		hi.vbox,
		addBtn,
	)
	return hi
}

func (hi *HolidaysInput) AddRow(label, from, to string) {
	row := &HolidayRow{
		LabelEntry: widget.NewEntry(),
		FromEntry:  widget.NewEntry(),
		ToEntry:    widget.NewEntry(),
	}
	row.LabelEntry.SetPlaceHolder("Bezeichnung")
	row.FromEntry.SetPlaceHolder("TT.MM")
	row.ToEntry.SetPlaceHolder("= Von")

	notifyChanged := func(_ string) {
		if hi.OnChanged != nil {
			hi.OnChanged()
		}
	}
	row.LabelEntry.OnChanged = notifyChanged
	row.FromEntry.OnChanged = notifyChanged
	row.ToEntry.OnChanged = notifyChanged

	row.LabelEntry.SetText(label)
	row.FromEntry.SetText(from)
	row.ToEntry.SetText(to)

	removeBtn := widget.NewButton("Entfernen", nil)

	rowContainer := container.NewGridWithColumns(4,
		row.LabelEntry,
		row.FromEntry,
		row.ToEntry,
		removeBtn,
	)

	hi.rows = append(hi.rows, row)
	hi.vbox.Add(rowContainer)

	removeBtn.OnTapped = func() {
		hi.removeRow(row, rowContainer)
	}
}

func (hi *HolidaysInput) removeRow(row *HolidayRow, c fyne.CanvasObject) {
	for i, r := range hi.rows {
		if r == row {
			hi.rows = append(hi.rows[:i], hi.rows[i+1:]...)
			break
		}
	}
	hi.vbox.Remove(c)
	if hi.OnChanged != nil {
		hi.OnChanged()
	}
}

func (hi *HolidaysInput) GetDateRanges() []model.DateRange {
	var ranges []model.DateRange
	yearStr := strconv.Itoa(hi.year)
	nextYearStr := strconv.Itoa(hi.year + 1)
	for _, r := range hi.rows {
		toText := r.ToEntry.Text
		if toText == "" && r.FromEntry.Text != "" {
			toText = r.FromEntry.Text
		}
		fromFull := appendYear(r.FromEntry.Text, yearStr)
		toFull := appendYear(toText, yearStr)
		if crossesYear(r.FromEntry.Text, toText) {
			toFull = appendYear(toText, nextYearStr)
		}
		fromISO, _ := model.DateToISO(fromFull)
		toISO, _ := model.DateToISO(toFull)
		ranges = append(ranges, model.DateRange{
			Label: r.LabelEntry.Text,
			From:  fromISO,
			To:    toISO,
		})
	}
	return ranges
}

func (hi *HolidaysInput) SetDateRanges(ranges []model.DateRange) {
	hi.rows = nil
	hi.vbox.RemoveAll()
	for _, dr := range ranges {
		from := dayMonthFromISO(dr.From)
		to := dayMonthFromISO(dr.To)
		hi.AddRow(dr.Label, from, to)
	}
}

// normalizeDDMM takes flexible date input (e.g. "3.7.", "13.7", "3.07")
// and returns zero-padded "DD.MM" format.
func normalizeDDMM(input string) string {
	input = strings.TrimSpace(input)
	input = strings.TrimRight(input, ".")
	parts := strings.SplitN(input, ".", 2)
	if len(parts) != 2 {
		return input
	}
	day, _ := strconv.Atoi(parts[0])
	month, _ := strconv.Atoi(parts[1])
	return fmt.Sprintf("%02d.%02d", day, month)
}

// appendYear turns "DD.MM" (or flexible variant) into "DD.MM.YYYY"
func appendYear(ddmm, year string) string {
	if ddmm == "" {
		return ""
	}
	return fmt.Sprintf("%s.%s", normalizeDDMM(ddmm), year)
}

// dayMonthFromISO extracts DD.MM from an ISO date (YYYY-MM-DD)
func dayMonthFromISO(iso string) string {
	full, err := model.DateFromISO(iso)
	if err != nil || len(full) < 6 {
		return ""
	}
	// full is "DD.MM.YYYY", return "DD.MM"
	return full[:5]
}

// crossesYear returns true if a DD.MM range spans a year boundary
// (i.e. the To month is earlier than the From month).
func crossesYear(fromDDMM, toDDMM string) bool {
	from := normalizeDDMM(fromDDMM)
	to := normalizeDDMM(toDDMM)
	if len(from) < 5 || len(to) < 5 {
		return false
	}
	fromMonth, err1 := strconv.Atoi(from[3:5])
	toMonth, err2 := strconv.Atoi(to[3:5])
	if err1 != nil || err2 != nil {
		return false
	}
	return toMonth < fromMonth
}
