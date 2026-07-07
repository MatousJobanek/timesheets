package ui

import (
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
}

func NewHolidaysInput() *HolidaysInput {
	hi := &HolidaysInput{}
	hi.vbox = container.NewVBox()

	addBtn := widget.NewButton("Hinzufügen", func() {
		hi.AddRow("", "", "")
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
	row.FromEntry.SetPlaceHolder("TT.MM.JJJJ")
	row.ToEntry.SetPlaceHolder("TT.MM.JJJJ")

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
}

func (hi *HolidaysInput) GetDateRanges() []model.DateRange {
	var ranges []model.DateRange
	for _, r := range hi.rows {
		fromISO, _ := model.DateToISO(r.FromEntry.Text)
		toISO, _ := model.DateToISO(r.ToEntry.Text)
		ranges = append(ranges, model.DateRange{
			Label: r.LabelEntry.Text,
			From:  fromISO,
			To:    toISO,
		})
	}
	return ranges
}

func (hi *HolidaysInput) SetDateRanges(ranges []model.DateRange) {
	// Clear existing
	hi.rows = nil
	hi.vbox.RemoveAll()
	// Add new rows
	for _, dr := range ranges {
		from, _ := model.DateFromISO(dr.From)
		to, _ := model.DateFromISO(dr.To)
		hi.AddRow(dr.Label, from, to)
	}
}
