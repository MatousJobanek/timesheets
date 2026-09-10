package ui

import (
	"fmt"
	"sort"
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

type tappableLabel struct {
	widget.BaseWidget
	label        *widget.RichText
	onRightClick func()
}

func newTappableLabel(text string, onRightClick func()) *tappableLabel {
	t := &tappableLabel{
		label:        widget.NewRichTextFromMarkdown(fmt.Sprintf("**%s**", text)),
		onRightClick: onRightClick,
	}
	t.ExtendBaseWidget(t)
	return t
}

func (t *tappableLabel) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.label)
}

func (t *tappableLabel) TappedSecondary(_ *fyne.PointEvent) {
	if t.onRightClick != nil {
		t.onRightClick()
	}
}

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
			btn := widget.NewButton(fmt.Sprintf("    %s", model.FormatSchoolYear(y)), func() {
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

		n := name
		empHeader := newTappableLabel(name, func() {
			sb.showRenameDialog(n)
		})
		sb.list.Add(empHeader)

		for _, entry := range entries {
			e := entry
			btn := widget.NewButton(fmt.Sprintf("    %s", model.FormatSchoolYear(e.year)), func() {
				sb.state.showEmployee(e.id)
			})
			btn.Alignment = widget.ButtonAlignLeading
			btn.Importance = widget.LowImportance
			if e.id == sb.state.currentEmployeeID {
				btn.Importance = widget.HighImportance
			}
			sb.list.Add(btn)
		}

		addYearBtn := widget.NewButton("    + Jahr", func() {
			sb.showAddYearForEmployeeDialog(n)
		})
		addYearBtn.Alignment = widget.ButtonAlignLeading
		addYearBtn.Importance = widget.LowImportance
		sb.list.Add(addYearBtn)
	}
}

func (sb *Sidebar) showRenameDialog(oldName string) {
	nameEntry := widget.NewEntry()
	nameEntry.SetText(oldName)

	form := dialog.NewForm(
		"Umbenennen",
		"Umbenennen",
		"Abbrechen",
		[]*widget.FormItem{
			widget.NewFormItem("Name", nameEntry),
		},
		func(ok bool) {
			if !ok {
				return
			}
			newName := strings.TrimSpace(nameEntry.Text)
			if newName == "" || newName == oldName {
				return
			}
			store.RenameEmployees(sb.state.store, oldName, newName)
			if err := sb.state.saveStore(); err != nil {
				dialog.ShowError(err, sb.state.window)
			}
			sb.state.refreshSidebar()

			if sb.state.currentEmployeeID != "" {
				for _, emp := range sb.state.store.Employees {
					if emp.ID == sb.state.currentEmployeeID && emp.Name == newName {
						sb.state.showEmployee(emp.ID)
						break
					}
				}
			}
		},
		sb.state.window,
	)
	form.Resize(fyne.NewSize(400, 150))
	form.Show()
}

func (sb *Sidebar) showNewEmployeeDialog() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Vor- und Nachname")

	yearEntry := newStartYearEntry()

	form := dialog.NewForm(
		"Neuer Mitarbeiter",
		"Erstellen",
		"Abbrechen",
		[]*widget.FormItem{
			widget.NewFormItem("Name", nameEntry),
			widget.NewFormItem("Startjahr", yearEntry),
		},
		func(ok bool) {
			if !ok {
				return
			}
			name := nameEntry.Text
			year, err := parseStartYear(yearEntry.Text)
			if err != nil || name == "" {
				if err != nil {
					dialog.ShowError(err, sb.state.window)
				}
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
	yearEntry := newStartYearEntry()

	form := dialog.NewForm(
		fmt.Sprintf("Neues Schuljahr für %s", name),
		"Erstellen",
		"Abbrechen",
		[]*widget.FormItem{
			widget.NewFormItem("Startjahr", yearEntry),
		},
		func(ok bool) {
			if !ok {
				return
			}
			year, err := parseStartYear(yearEntry.Text)
			if err != nil {
				dialog.ShowError(err, sb.state.window)
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
	yearEntry := newStartYearEntry()

	form := dialog.NewForm(
		"Freie Tage für neues Schuljahr",
		"Erstellen",
		"Abbrechen",
		[]*widget.FormItem{
			widget.NewFormItem("Startjahr", yearEntry),
		},
		func(ok bool) {
			if !ok {
				return
			}
			year, err := parseStartYear(yearEntry.Text)
			if err != nil {
				dialog.ShowError(err, sb.state.window)
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

func newStartYearEntry() *widget.Entry {
	yearEntry := widget.NewEntry()
	yearEntry.SetPlaceHolder("z.B. 2026 (August–Juli)")
	yearEntry.SetText(strconv.Itoa(model.CurrentSchoolYearStart(time.Now())))
	return yearEntry
}

func parseStartYear(text string) (int, error) {
	year, err := strconv.Atoi(strings.TrimSpace(text))
	if err != nil || year < 2000 || year > 2100 {
		return 0, fmt.Errorf("ungültiges Startjahr %q (erwartet z.B. 2026)", text)
	}
	return year, nil
}
