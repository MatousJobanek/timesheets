package ui

import (
	"fmt"
	"regexp"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"timesheets/generator"
	"timesheets/model"
	"timesheets/store"
)

var empTimeRegex = regexp.MustCompile(`^\d{1,2}:\d{2}$`)
var empDateRegex = regexp.MustCompile(`^\d{1,2}\.\d{1,2}\.?$`)

type clockPair struct {
	begin int
	end   int
}

// validateTimePair checks one from–to pair. Both fields empty is valid and returns ok=false.
// A partial pair, a bad clock, or an end that is not after the start appends an error.
func validateTimePair(errs *[]string, prefix, day, beginLabel, endLabel, begin, end string) (clockPair, bool) {
	if begin == "" && end == "" {
		return clockPair{}, false
	}
	if begin == "" || end == "" {
		*errs = append(*errs, fmt.Sprintf("%s%s: %s und %s müssen beide ausgefüllt sein", prefix, day, beginLabel, endLabel))
		return clockPair{}, false
	}
	b, bok := parseClock(begin)
	e, eok := parseClock(end)
	if !bok {
		*errs = append(*errs, fmt.Sprintf("%s%s: %s muss im Format HH:MM sein", prefix, day, beginLabel))
	}
	if !eok {
		*errs = append(*errs, fmt.Sprintf("%s%s: %s muss im Format HH:MM sein", prefix, day, endLabel))
	}
	if !bok || !eok {
		return clockPair{}, false
	}
	if e <= b {
		*errs = append(*errs, fmt.Sprintf("%s%s: %s muss nach %s liegen", prefix, day, endLabel, beginLabel))
		return clockPair{}, false
	}
	return clockPair{begin: b, end: e}, true
}

func parseClock(s string) (int, bool) {
	if !empTimeRegex.MatchString(s) {
		return 0, false
	}
	var h, m int
	if _, err := fmt.Sscanf(s, "%d:%d", &h, &m); err != nil {
		return 0, false
	}
	if h > 23 || m > 59 {
		return 0, false
	}
	return h*60 + m, true
}

func NewEmployeeForm(state *AppState, emp *model.Employee) fyne.CanvasObject {
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

	evenOddCheck := widget.NewCheck("Gerade/Ungerade Wochen verwenden", nil)
	evenOddCheck.SetChecked(emp.UseEvenOdd)

	normalGrid := NewWorkingHoursGrid()
	evenGrid := NewWorkingHoursGrid()
	oddGrid := NewWorkingHoursGrid()

	if emp.UseEvenOdd {
		evenGrid.SetSchedule(emp.EvenSchedule)
		oddGrid.SetSchedule(emp.OddSchedule)
	} else {
		normalGrid.SetSchedule(emp.WeekSchedule)
	}

	normalGrid.OnChanged = markDirty
	evenGrid.OnChanged = markDirty
	oddGrid.OnChanged = markDirty

	evenLabel := widget.NewLabel("Gerade Wochen")
	evenLabel.TextStyle = fyne.TextStyle{Bold: true}
	oddLabel := widget.NewLabel("Ungerade Wochen")
	oddLabel.TextStyle = fyne.TextStyle{Bold: true}

	evenOddContainer := container.NewVBox(
		evenLabel,
		evenGrid.Container,
		oddLabel,
		oddGrid.Container,
	)

	if !emp.UseEvenOdd {
		evenOddContainer.Hide()
	} else {
		normalGrid.Container.Hide()
	}

	evenOddCheck.OnChanged = func(checked bool) {
		markDirty()
		if checked {
			evenGrid.SetSchedule(normalGrid.GetSchedule())
			normalGrid.Container.Hide()
			evenOddContainer.Show()
		} else {
			normalGrid.SetSchedule(evenGrid.GetSchedule())
			normalGrid.Container.Show()
			evenOddContainer.Hide()
		}
	}

	// Shared free periods (read-only display)
	yc := store.FindYearConfig(state.store, emp.Year)
	var sharedPeriodsWidget fyne.CanvasObject
	if yc != nil && len(yc.FreePeriods) > 0 {
		rows := container.NewVBox()
		for _, fp := range yc.FreePeriods {
			from := dayMonthFromISO(fp.From)
			to := dayMonthFromISO(fp.To)
			text := fmt.Sprintf("%s: %s – %s", fp.Label, from, to)
			rows.Add(widget.NewLabel(text))
		}
		sharedPeriodsWidget = rows
	} else {
		sharedPeriodsWidget = widget.NewLabel("(keine gemeinsamen freien Tage definiert)")
	}

	// Employee-specific overrides
	overridesInput := NewHolidaysInput(emp.Year)
	overridesInput.SetDateRanges(emp.FreePeriodOverrides)
	overridesInput.OnChanged = markDirty

	// Build employee from form
	buildEmployee := func() model.Employee {
		updated := model.Employee{
			ID:         emp.ID,
			Name:       emp.Name,
			Year:       emp.Year,
			UseEvenOdd: evenOddCheck.Checked,
		}
		if updated.UseEvenOdd {
			updated.EvenSchedule = evenGrid.GetSchedule()
			updated.OddSchedule = oddGrid.GetSchedule()
		} else {
			updated.WeekSchedule = normalGrid.GetSchedule()
		}
		updated.FreePeriodOverrides = overridesInput.GetDateRanges()
		return updated
	}

	// Validation
	validate := func() []string {
		var errs []string

		validateGrid := func(g *WorkingHoursGrid, prefix string) {
			for i := 0; i < 5; i++ {
				day := dayLabels[i]
				begin1 := strings.TrimSpace(g.Begin1[i].Text)
				end1 := strings.TrimSpace(g.End1[i].Text)
				begin2 := strings.TrimSpace(g.Begin2[i].Text)
				end2 := strings.TrimSpace(g.End2[i].Text)

				if begin1 == "" && end1 == "" && begin2 == "" && end2 == "" {
					continue
				}

				b1, ok1 := validateTimePair(&errs, prefix, day, "Beginn 1", "Ende 1", begin1, end1)
				b2, ok2 := validateTimePair(&errs, prefix, day, "Beginn 2", "Ende 2", begin2, end2)

				if (begin2 != "" || end2 != "") && begin1 == "" && end1 == "" {
					errs = append(errs, fmt.Sprintf("%s%s: Beginn 2 erfordert Beginn 1 und Ende 1", prefix, day))
				}
				if ok1 && ok2 && b2.begin < b1.end {
					errs = append(errs, fmt.Sprintf("%s%s: Beginn 2 darf nicht vor Ende 1 liegen", prefix, day))
				}
			}
		}

		if evenOddCheck.Checked {
			validateGrid(evenGrid, "Gerade - ")
			validateGrid(oddGrid, "Ungerade - ")
		} else {
			validateGrid(normalGrid, "")
		}

		for i, r := range overridesInput.rows {
			if r.FromEntry.Text != "" && !empDateRegex.MatchString(r.FromEntry.Text) {
				errs = append(errs, fmt.Sprintf("Freie Tage #%d: Von muss im Format TT.MM sein", i+1))
			}
			if r.ToEntry.Text != "" && !empDateRegex.MatchString(r.ToEntry.Text) {
				errs = append(errs, fmt.Sprintf("Freie Tage #%d: Bis muss im Format TT.MM sein", i+1))
			}
			fromText := r.FromEntry.Text
			toText := r.ToEntry.Text
			if toText == "" {
				toText = fromText
			}
			if empDateRegex.MatchString(fromText) && empDateRegex.MatchString(toText) {
				from, e1 := model.ParseDDMMInSchoolYear(fromText, emp.Year)
				to, e2 := model.ParseDDMMInSchoolYear(toText, emp.Year)
				if e1 == nil && e2 == nil {
					if from.After(to) {
						to = to.AddDate(1, 0, 0)
					}
					if from.After(to) {
						errs = append(errs, fmt.Sprintf("Freie Tage #%d: Von muss vor Bis liegen", i+1))
					}
				}
			}
		}

		return errs
	}

	// Action buttons
	saveBtn.OnTapped = func() {
		errs := validate()
		if len(errs) > 0 {
			dialog.ShowError(fmt.Errorf("%s", strings.Join(errs, "\n")), state.window)
			return
		}
		updated := buildEmployee()
		store.UpdateEmployee(state.store, updated)
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
			fmt.Sprintf("Mitarbeiter %s (%s) wirklich löschen?", emp.Name, model.FormatSchoolYear(emp.Year)),
			func(ok bool) {
				if !ok {
					return
				}
				store.DeleteEmployee(state.store, emp.ID)
				if err := state.saveStore(); err != nil {
					dialog.ShowError(err, state.window)
				}
				state.refreshSidebar()
				state.showPlaceholder()
			},
			state.window,
		)
	})

	generateBtn := widget.NewButton("Stundenzettel erstellen", func() {
		errs := validate()
		if len(errs) > 0 {
			dialog.ShowError(fmt.Errorf("%s", strings.Join(errs, "\n")), state.window)
			return
		}

		updated := buildEmployee()
		var shared []model.DateRange
		if yc := store.FindYearConfig(state.store, updated.Year); yc != nil {
			shared = yc.FreePeriods
		}
		merged := store.MergeFreePeriods(shared, updated.FreePeriodOverrides)

		xlFile, err := generator.Generate(updated, merged)
		if err != nil {
			dialog.ShowError(err, state.window)
			return
		}

		saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, saveErr error) {
			if saveErr != nil || writer == nil {
				return
			}
			defer writer.Close()
			defer xlFile.Close()
			if _, writeErr := xlFile.WriteTo(writer); writeErr != nil {
				dialog.ShowError(writeErr, state.window)
			}
		}, state.window)
		saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".xlsx"}))
		saveDialog.SetFileName(fmt.Sprintf("Stundenzettel_%s_%s.xlsx", updated.Name, model.FormatSchoolYearFile(updated.Year)))
		saveDialog.Show()
	})

	// Layout
	section2 := widget.NewCard(fmt.Sprintf("%s — %s", emp.Name, model.FormatSchoolYear(emp.Year)), "",
		container.NewVBox(
			evenOddCheck,
			normalGrid.Container,
			evenOddContainer,
		),
	)

	section3 := widget.NewCard("Gemeinsame freie Tage (geerbt)", "", sharedPeriodsWidget)

	section4 := widget.NewCard("Zusätzliche freie Tage (nur dieser Mitarbeiter)", "",
		overridesInput.Container,
	)

	section5 := widget.NewCard("Aktionen", "",
		container.NewGridWithColumns(3,
			saveBtn,
			deleteBtn,
			generateBtn,
		),
	)

	// Register validate+save callbacks for auto-save on switch/close
	state.currentValidate = validate
	state.currentSave = func() {
		if !dirty {
			return
		}
		updated := buildEmployee()
		store.UpdateEmployee(state.store, updated)
	}

	return container.NewVBox(
		section2,
		section3,
		section4,
		section5,
	)
}
