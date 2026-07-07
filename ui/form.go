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
)

var timeRegex = regexp.MustCompile(`^\d{1,2}:\d{2}$`)
var dateRegex = regexp.MustCompile(`^\d{2}\.\d{2}\.\d{4}$`)

func BuildForm(w fyne.Window) fyne.CanvasObject {
	// Section 1: Mitarbeiterdaten
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Vor- und Nachname")
	yearEntry := widget.NewEntry()
	yearEntry.SetPlaceHolder(strconv.Itoa(time.Now().Year()))

	// Section 2: Arbeitszeiten
	evenOddCheck := widget.NewCheck("Gerade/Ungerade Wochen verwenden", nil)

	normalGrid := NewWorkingHoursGrid()
	evenGrid := NewWorkingHoursGrid()
	oddGrid := NewWorkingHoursGrid()

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
	evenOddContainer.Hide()

	evenOddCheck.OnChanged = func(checked bool) {
		if checked {
			normalGrid.Container.Hide()
			evenOddContainer.Show()
		} else {
			normalGrid.Container.Show()
			evenOddContainer.Hide()
		}
	}

	// Section 3: Freie Tage
	holidaysInput := NewHolidaysInput()

	// Build Employee from form
	buildEmployee := func() model.Employee {
		year, _ := strconv.Atoi(yearEntry.Text)
		emp := model.Employee{
			Name:       nameEntry.Text,
			Year:       year,
			UseEvenOdd: evenOddCheck.Checked,
		}
		if emp.UseEvenOdd {
			emp.EvenSchedule = evenGrid.GetSchedule()
			emp.OddSchedule = oddGrid.GetSchedule()
		} else {
			emp.WeekSchedule = normalGrid.GetSchedule()
		}
		emp.FreePeriods = holidaysInput.GetDateRanges()
		return emp
	}

	// Populate form from Employee
	populateForm := func(emp model.Employee) {
		nameEntry.SetText(emp.Name)
		yearEntry.SetText(strconv.Itoa(emp.Year))
		evenOddCheck.SetChecked(emp.UseEvenOdd)
		if emp.UseEvenOdd {
			evenGrid.SetSchedule(emp.EvenSchedule)
			oddGrid.SetSchedule(emp.OddSchedule)
		} else {
			normalGrid.SetSchedule(emp.WeekSchedule)
		}
		holidaysInput.SetDateRanges(emp.FreePeriods)
	}

	// Validation
	validate := func() []string {
		var errs []string
		if strings.TrimSpace(nameEntry.Text) == "" {
			errs = append(errs, "Name darf nicht leer sein")
		}
		year, err := strconv.Atoi(yearEntry.Text)
		if err != nil || year < 1900 || year > 2100 {
			errs = append(errs, "Jahr muss eine gültige Jahreszahl sein")
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
				if total > 0 && !timeRegex.MatchString(start) {
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

		for i, r := range holidaysInput.rows {
			if r.FromEntry.Text != "" && !dateRegex.MatchString(r.FromEntry.Text) {
				errs = append(errs, fmt.Sprintf("Freie Tage #%d: Von muss im Format TT.MM.JJJJ sein", i+1))
			}
			if r.ToEntry.Text != "" && !dateRegex.MatchString(r.ToEntry.Text) {
				errs = append(errs, fmt.Sprintf("Freie Tage #%d: Bis muss im Format TT.MM.JJJJ sein", i+1))
			}
			if dateRegex.MatchString(r.FromEntry.Text) && dateRegex.MatchString(r.ToEntry.Text) {
				from, e1 := time.Parse("02.01.2006", r.FromEntry.Text)
				to, e2 := time.Parse("02.01.2006", r.ToEntry.Text)
				if e1 == nil && e2 == nil && from.After(to) {
					errs = append(errs, fmt.Sprintf("Freie Tage #%d: Von muss vor Bis liegen", i+1))
				}
			}
		}

		return errs
	}

	// Action buttons
	saveBtn := widget.NewButton("Konfiguration speichern", func() {
		emp := buildEmployee()
		saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil || writer == nil {
				return
			}
			defer writer.Close()
			path := writer.URI().Path()
			if saveErr := model.SaveConfig(path, emp); saveErr != nil {
				dialog.ShowError(saveErr, w)
			}
		}, w)
		saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
		saveDialog.SetFileName(fmt.Sprintf("%s_%d.json", emp.Name, emp.Year))
		saveDialog.Show()
	})

	loadBtn := widget.NewButton("Konfiguration laden", func() {
		openDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()
			path := reader.URI().Path()
			emp, loadErr := model.LoadConfig(path)
			if loadErr != nil {
				dialog.ShowError(loadErr, w)
				return
			}
			populateForm(emp)
		}, w)
		openDialog.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
		openDialog.Show()
	})

	generateBtn := widget.NewButton("Dienstplan erstellen", func() {
		errs := validate()
		if len(errs) > 0 {
			dialog.ShowError(fmt.Errorf("%s", strings.Join(errs, "\n")), w)
			return
		}

		emp := buildEmployee()
		xlFile, err := generator.Generate(emp)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, saveErr error) {
			if saveErr != nil || writer == nil {
				return
			}
			defer writer.Close()
			path := writer.URI().Path()
			if writeErr := xlFile.SaveAs(path); writeErr != nil {
				dialog.ShowError(writeErr, w)
			}
		}, w)
		saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".xlsx"}))
		saveDialog.SetFileName(fmt.Sprintf("Dienstplan_%s_%d.xlsx", emp.Name, emp.Year))
		saveDialog.Show()
	})
	generateBtn.Importance = widget.HighImportance

	// Layout sections
	section1 := widget.NewCard("Mitarbeiterdaten", "",
		container.NewGridWithColumns(2,
			widget.NewLabel("Name:"), nameEntry,
			widget.NewLabel("Jahr:"), yearEntry,
		),
	)

	section2 := widget.NewCard("Arbeitszeiten", "",
		container.NewVBox(
			evenOddCheck,
			normalGrid.Container,
			evenOddContainer,
		),
	)

	section3 := widget.NewCard("Freie Tage (Schulferien, Fenstertage)", "",
		holidaysInput.Container,
	)

	section4 := widget.NewCard("Aktionen", "",
		container.NewGridWithColumns(3,
			saveBtn,
			loadBtn,
			generateBtn,
		),
	)

	content := container.NewVBox(
		section1,
		section2,
		section3,
		section4,
	)

	return container.NewVScroll(content)
}

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
