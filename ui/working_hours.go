package ui

import (
	"strings"

	"timesheets/model"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var dayLabels = []string{"Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag"}

type WorkingHoursGrid struct {
	Container fyne.CanvasObject
	Begin1    [5]*widget.Entry
	End1      [5]*widget.Entry
	Begin2    [5]*widget.Entry
	End2      [5]*widget.Entry
	OnChanged func()
}

func NewWorkingHoursGrid() *WorkingHoursGrid {
	g := &WorkingHoursGrid{}

	header := container.NewGridWithColumns(5,
		widget.NewLabel("Tag"),
		widget.NewLabel("Beginn 1"),
		widget.NewLabel("Ende 1"),
		widget.NewLabel("Beginn 2"),
		widget.NewLabel("Ende 2"),
	)

	rows := []fyne.CanvasObject{header}

	for i := 0; i < 5; i++ {
		g.Begin1[i] = g.newTimeEntry()
		g.End1[i] = g.newTimeEntry()
		g.Begin2[i] = g.newTimeEntry()
		g.End2[i] = g.newTimeEntry()

		row := container.NewGridWithColumns(5,
			widget.NewLabel(dayLabels[i]),
			g.Begin1[i],
			g.End1[i],
			g.Begin2[i],
			g.End2[i],
		)
		rows = append(rows, row)
	}

	g.Container = container.NewVBox(rows...)
	return g
}

func (g *WorkingHoursGrid) newTimeEntry() *widget.Entry {
	e := widget.NewEntry()
	e.SetPlaceHolder("HH:MM")
	e.OnChanged = func(_ string) {
		if g.OnChanged != nil {
			g.OnChanged()
		}
	}
	return e
}

func (g *WorkingHoursGrid) GetSchedule() model.WeekSchedule {
	return model.WeekSchedule{
		Monday:    g.getDaySchedule(0),
		Tuesday:   g.getDaySchedule(1),
		Wednesday: g.getDaySchedule(2),
		Thursday:  g.getDaySchedule(3),
		Friday:    g.getDaySchedule(4),
	}
}

func (g *WorkingHoursGrid) SetSchedule(ws model.WeekSchedule) {
	schedules := []model.DaySchedule{ws.Monday, ws.Tuesday, ws.Wednesday, ws.Thursday, ws.Friday}
	for i, ds := range schedules {
		g.Begin1[i].SetText(ds.Begin1)
		g.End1[i].SetText(ds.End1)
		g.Begin2[i].SetText(ds.Begin2)
		g.End2[i].SetText(ds.End2)
	}
}

func (g *WorkingHoursGrid) getDaySchedule(idx int) model.DaySchedule {
	return model.DaySchedule{
		Begin1: strings.TrimSpace(g.Begin1[idx].Text),
		End1:   strings.TrimSpace(g.End1[idx].Text),
		Begin2: strings.TrimSpace(g.Begin2[idx].Text),
		End2:   strings.TrimSpace(g.End2[idx].Text),
	}
}
