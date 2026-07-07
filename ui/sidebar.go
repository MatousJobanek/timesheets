package ui

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"timesheets/model"
	"timesheets/store"
)

type Sidebar struct {
	Container *fyne.Container
	list      *fyne.Container
	state     *AppState
}

func NewSidebar(state *AppState) *Sidebar {
	sb := &Sidebar{state: state}
	sb.list = container.NewVBox()
	sb.buildList(state.store)

	newBtn := widget.NewButton("+ Neu", func() {
		sb.showNewEmployeeDialog()
	})

	sb.Container = container.NewBorder(nil, newBtn, nil, nil, container.NewVScroll(sb.list))
	return sb
}

func (sb *Sidebar) Refresh(s *store.Store) {
	sb.list.RemoveAll()
	sb.buildList(s)
	sb.list.Refresh()
}

func (sb *Sidebar) buildList(s *store.Store) {
	// "Freie Tage" section
	header := widget.NewRichTextFromMarkdown("**Freie Tage**")
	sb.list.Add(header)

	if len(s.YearConfigs) > 0 {
		years := make([]int, len(s.YearConfigs))
		for i, yc := range s.YearConfigs {
			years[i] = yc.Year
		}
		sort.Ints(years)

		for _, year := range years {
			y := year
			btn := widget.NewButton(fmt.Sprintf("    %d", y), func() {
				sb.state.showYearConfig(y)
			})
			btn.Alignment = widget.ButtonAlignLeading
			btn.Importance = widget.LowImportance
			if y == sb.state.currentYearConfig {
				btn.Importance = widget.HighImportance
			}
			sb.list.Add(btn)
		}
	}

	addYearConfigBtn := widget.NewButton("    + Jahr", func() {
		sb.showNewYearConfigDialog()
	})
	addYearConfigBtn.Alignment = widget.ButtonAlignLeading
	addYearConfigBtn.Importance = widget.LowImportance
	sb.list.Add(addYearConfigBtn)
	sb.list.Add(widget.NewSeparator())

	// Employee sections grouped by name
	type empEntry struct {
		id   string
		year int
	}
	grouped := make(map[string][]empEntry)
	for _, emp := range s.Employees {
		grouped[emp.Name] = append(grouped[emp.Name], empEntry{id: emp.ID, year: emp.Year})
	}

	names := make([]string, 0, len(grouped))
	for name := range grouped {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		entries := grouped[name]
		sort.Slice(entries, func(i, j int) bool { return entries[i].year < entries[j].year })

		empHeader := widget.NewRichTextFromMarkdown(fmt.Sprintf("**%s**", name))
		sb.list.Add(empHeader)

		for _, entry := range entries {
			e := entry
			btn := widget.NewButton(fmt.Sprintf("    %d", e.year), func() {
				sb.state.showEmployee(e.id)
			})
			btn.Alignment = widget.ButtonAlignLeading
			btn.Importance = widget.LowImportance
			if e.id == sb.state.currentEmployeeID {
				btn.Importance = widget.HighImportance
			}
			sb.list.Add(btn)
		}

		n := name
		addYearBtn := widget.NewButton("    + Jahr", func() {
			sb.showAddYearForEmployeeDialog(n)
		})
		addYearBtn.Alignment = widget.ButtonAlignLeading
		addYearBtn.Importance = widget.LowImportance
		sb.list.Add(addYearBtn)
	}
}

func (sb *Sidebar) showNewEmployeeDialog() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Vor- und Nachname")

	yearEntry := widget.NewEntry()
	yearEntry.SetText(strconv.Itoa(time.Now().Year()))

	form := dialog.NewForm(
		"Neuer Mitarbeiter",
		"Erstellen",
		"Abbrechen",
		[]*widget.FormItem{
			widget.NewFormItem("Name", nameEntry),
			widget.NewFormItem("Jahr", yearEntry),
		},
		func(ok bool) {
			if !ok {
				return
			}
			name := nameEntry.Text
			year, err := strconv.Atoi(yearEntry.Text)
			if err != nil || name == "" {
				return
			}
			emp := model.Employee{
				Name: name,
				Year: year,
			}
			id := store.AddEmployee(sb.state.store, emp)
			if err := sb.state.saveStore(); err != nil {
				dialog.ShowError(err, sb.state.window)
			}
			sb.state.refreshSidebar()
			sb.state.showEmployee(id)
		},
		sb.state.window,
	)
	form.Resize(fyne.NewSize(400, 200))
	form.Show()
}

func (sb *Sidebar) showAddYearForEmployeeDialog(name string) {
	yearEntry := widget.NewEntry()
	yearEntry.SetText(strconv.Itoa(time.Now().Year()))

	form := dialog.NewForm(
		fmt.Sprintf("Neues Jahr für %s", name),
		"Erstellen",
		"Abbrechen",
		[]*widget.FormItem{
			widget.NewFormItem("Jahr", yearEntry),
		},
		func(ok bool) {
			if !ok {
				return
			}
			year, err := strconv.Atoi(yearEntry.Text)
			if err != nil {
				return
			}
			emp := model.Employee{
				Name: name,
				Year: year,
			}
			id := store.AddEmployee(sb.state.store, emp)
			if err := sb.state.saveStore(); err != nil {
				dialog.ShowError(err, sb.state.window)
			}
			sb.state.refreshSidebar()
			sb.state.showEmployee(id)
		},
		sb.state.window,
	)
	form.Resize(fyne.NewSize(300, 150))
	form.Show()
}

func (sb *Sidebar) showNewYearConfigDialog() {
	yearEntry := widget.NewEntry()
	yearEntry.SetText(strconv.Itoa(time.Now().Year()))

	form := dialog.NewForm(
		"Freie Tage für neues Jahr",
		"Erstellen",
		"Abbrechen",
		[]*widget.FormItem{
			widget.NewFormItem("Jahr", yearEntry),
		},
		func(ok bool) {
			if !ok {
				return
			}
			year, err := strconv.Atoi(yearEntry.Text)
			if err != nil {
				return
			}
			store.EnsureYearConfig(sb.state.store, year)
			if err := sb.state.saveStore(); err != nil {
				dialog.ShowError(err, sb.state.window)
			}
			sb.state.refreshSidebar()
			sb.state.showYearConfig(year)
		},
		sb.state.window,
	)
	form.Resize(fyne.NewSize(300, 150))
	form.Show()
}
