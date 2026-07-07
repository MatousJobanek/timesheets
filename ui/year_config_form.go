package ui

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"timesheets/model"
	"timesheets/store"
)

var ycDateRegex = regexp.MustCompile(`^\d{2}\.\d{2}$`)

func NewYearConfigForm(state *AppState, yc *model.YearConfig) fyne.CanvasObject {
	dirty := false
	saveBtn := widget.NewButton("Speichern", nil)
	saveBtn.Importance = widget.HighImportance
	saveBtn.Disable()

	markDirty := func() {
		if !dirty {
			dirty = true
			saveBtn.Enable()
		}
	}

	holidaysInput := NewHolidaysInput(yc.Year)
	holidaysInput.SetDateRanges(yc.FreePeriods)
	holidaysInput.OnChanged = markDirty

	validate := func() []string {
		var errs []string
		for i, r := range holidaysInput.rows {
			if r.FromEntry.Text != "" && !ycDateRegex.MatchString(r.FromEntry.Text) {
				errs = append(errs, fmt.Sprintf("Freie Tage #%d: Von muss im Format TT.MM sein", i+1))
			}
			if r.ToEntry.Text != "" && !ycDateRegex.MatchString(r.ToEntry.Text) {
				errs = append(errs, fmt.Sprintf("Freie Tage #%d: Bis muss im Format TT.MM sein", i+1))
			}
			fromText := r.FromEntry.Text
			toText := r.ToEntry.Text
			if toText == "" {
				toText = fromText
			}
			if ycDateRegex.MatchString(fromText) && ycDateRegex.MatchString(toText) {
				yearStr := strconv.Itoa(yc.Year)
				toYearStr := yearStr
				if crossesYear(fromText, toText) {
					toYearStr = strconv.Itoa(yc.Year + 1)
				}
				from, e1 := time.Parse("02.01.2006", fromText+"."+yearStr)
				to, e2 := time.Parse("02.01.2006", toText+"."+toYearStr)
				if e1 == nil && e2 == nil && from.After(to) {
					errs = append(errs, fmt.Sprintf("Freie Tage #%d: Von muss vor Bis liegen", i+1))
				}
			}
		}
		return errs
	}

	saveBtn.OnTapped = func() {
		errs := validate()
		if len(errs) > 0 {
			dialog.ShowError(fmt.Errorf("%s", strings.Join(errs, "\n")), state.window)
			return
		}
		yc.FreePeriods = holidaysInput.GetDateRanges()
		if err := state.saveStore(); err != nil {
			dialog.ShowError(err, state.window)
			return
		}
		dirty = false
		saveBtn.Disable()
		state.refreshSidebar()
	}

	deleteBtn := widget.NewButton("Löschen", func() {
		dialog.ShowConfirm(
			"Löschen bestätigen",
			fmt.Sprintf("Freie Tage für %d wirklich löschen?", yc.Year),
			func(ok bool) {
				if !ok {
					return
				}
				deleteYearConfig(state.store, yc.Year)
				if err := state.saveStore(); err != nil {
					dialog.ShowError(err, state.window)
				}
				state.refreshSidebar()
				state.showPlaceholder()
			},
			state.window,
		)
	})

	section := widget.NewCard(
		fmt.Sprintf("Gemeinsame freie Tage — %d", yc.Year),
		"Schulferien, Fenstertage etc. die für alle Mitarbeiter gelten",
		holidaysInput.Container,
	)

	actions := widget.NewCard("Aktionen", "",
		container.NewGridWithColumns(2,
			saveBtn,
			deleteBtn,
		),
	)

	// Register validate+save callbacks for auto-save on switch/close
	state.currentValidate = validate
	state.currentSave = func() {
		if !dirty {
			return
		}
		yc.FreePeriods = holidaysInput.GetDateRanges()
	}

	return container.NewVBox(section, actions)
}

func deleteYearConfig(s *store.Store, year int) {
	for i, yc := range s.YearConfigs {
		if yc.Year == year {
			s.YearConfigs = append(s.YearConfigs[:i], s.YearConfigs[i+1:]...)
			return
		}
	}
}
