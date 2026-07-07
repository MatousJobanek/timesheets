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
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"timesheets/generator"
	"timesheets/model"
	"timesheets/store"
)

var empTimeRegex = regexp.MustCompile(`^\d{1,2}:\d{2}$`)
var empDateRegex = regexp.MustCompile(`^\d{2}\.\d{2}$`)

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

	nameEntry := widget.NewEntry()
	nameEntry.SetText(emp.Name)
	nameEntry.OnChanged = func(_ string) { markDirty() }

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
			normalGrid.Container.Hide()
			evenOddContainer.Show()
		} else {
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
			Name:       nameEntry.Text,
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
		if strings.TrimSpace(nameEntry.Text) == "" {
			errs = append(errs, "Name darf nicht leer sein")
		}

		validateGrid := func(g *WorkingHoursGrid, prefix string) {
			for i := 0; i < 5; i++ {
				start := g.StartEntries[i].Text
				hours := g.HoursEntries[i].Text
				mins := g.MinEntries[i].Text

				h := strToInt(hours)
				m := strToInt(mins)
				total := h*60 + m

				if total == 0 && start == "" {
					continue
				}
				if total > 0 && !empTimeRegex.MatchString(start) {
					errs = append(errs, fmt.Sprintf("%s%s: Beginn muss im Format HH:MM sein", prefix, dayLabels[i]))
				}
				if strToInt(hours) < 0 || strToInt(mins) < 0 {
					errs = append(errs, fmt.Sprintf("%s%s: Stunden/Minuten dürfen nicht negativ sein", prefix, dayLabels[i]))
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
				yearStr := strconv.Itoa(emp.Year)
				toYearStr := yearStr
				if crossesYear(fromText, toText) {
					toYearStr = strconv.Itoa(emp.Year + 1)
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
			fmt.Sprintf("Mitarbeiter %s (%d) wirklich löschen?", emp.Name, emp.Year),
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

	generateBtn := widget.NewButton("Dienstplan erstellen", func() {
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
			path := writer.URI().Path()
			if writeErr := xlFile.SaveAs(path); writeErr != nil {
				dialog.ShowError(writeErr, state.window)
			}
		}, state.window)
		saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".xlsx"}))
		saveDialog.SetFileName(fmt.Sprintf("Dienstplan_%s_%d.xlsx", updated.Name, updated.Year))
		saveDialog.Show()
	})

	// Layout
	section1 := widget.NewCard(fmt.Sprintf("Mitarbeiterdaten — %d", emp.Year), "",
		container.NewGridWithColumns(2,
			widget.NewLabel("Name:"), nameEntry,
		),
	)

	section2 := widget.NewCard("Arbeitszeiten", "",
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
		section1,
		section2,
		section3,
		section4,
		section5,
	)
}
