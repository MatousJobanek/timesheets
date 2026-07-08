package ui

import (
	"strconv"
	"strings"

	"timesheets/model"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func strToInt(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, _ := strconv.Atoi(s)
	return v
}

func intToStr(v int) string {
	return strconv.Itoa(v)
}

var dayLabels = []string{"Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag"}

type WorkingHoursGrid struct {
	Container    fyne.CanvasObject
	StartEntries [5]*widget.Entry
	HoursEntries [5]*widget.Entry
	MinEntries   [5]*widget.Entry
	OnChanged    func()
}

func NewWorkingHoursGrid() *WorkingHoursGrid {
	g := &WorkingHoursGrid{}

	header := container.NewGridWithColumns(4,
		widget.NewLabel("Tag"),
		widget.NewLabel("Beginn (HH:MM)"),
		widget.NewLabel("Stunden"),
		widget.NewLabel("Minuten"),
	)

	rows := []fyne.CanvasObject{header}

	for i := 0; i < 5; i++ {
		g.StartEntries[i] = widget.NewEntry()
		g.StartEntries[i].SetPlaceHolder("HH:MM")
		g.StartEntries[i].OnChanged = func(_ string) {
			if g.OnChanged != nil {
				g.OnChanged()
			}
		}
		g.HoursEntries[i] = widget.NewEntry()
		g.HoursEntries[i].OnChanged = func(_ string) {
			if g.OnChanged != nil {
				g.OnChanged()
			}
		}
		g.MinEntries[i] = widget.NewEntry()
		g.MinEntries[i].OnChanged = func(_ string) {
			if g.OnChanged != nil {
				g.OnChanged()
			}
		}

		row := container.NewGridWithColumns(4,
			widget.NewLabel(dayLabels[i]),
			g.StartEntries[i],
			g.HoursEntries[i],
			g.MinEntries[i],
		)
		rows = append(rows, row)
	}

	g.Container = container.NewVBox(rows...)
	return g
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
		g.StartEntries[i].SetText(ds.StartTime)
		if ds.TotalMinutes == 0 {
			g.HoursEntries[i].SetText("")
			g.MinEntries[i].SetText("")
		} else {
			g.HoursEntries[i].SetText(intToStr(ds.TotalMinutes / 60))
			g.MinEntries[i].SetText(intToStr(ds.TotalMinutes % 60))
		}
	}
}

func (g *WorkingHoursGrid) getDaySchedule(idx int) model.DaySchedule {
	start := g.StartEntries[idx].Text
	hours := strToInt(g.HoursEntries[idx].Text)
	mins := strToInt(g.MinEntries[idx].Text)
	total := hours*60 + mins
	return model.DaySchedule{
		StartTime:    start,
		TotalMinutes: total,
	}
}
