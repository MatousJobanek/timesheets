package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"timesheets/model"
	"timesheets/store"
)

type AppState struct {
	store    *store.Store
	storeDir string
	window   fyne.Window

	sidebar     *Sidebar
	detailPanel *fyne.Container

	currentEmployeeID string
	currentYearConfig int // 0 = none selected

	// Set by the active form to allow validate+save on switch/close
	currentValidate func() []string
	currentSave     func()
}

func BuildApp(a fyne.App, w fyne.Window) fyne.CanvasObject {
	storeDir := a.Storage().RootURI().Path()

	s, err := store.Load(storeDir)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Store laden fehlgeschlagen: %w", err), w)
		s = &store.Store{}
	}

	state := &AppState{
		store:    s,
		storeDir: storeDir,
		window:   w,
	}

	state.detailPanel = container.NewStack(newPlaceholder())

	state.sidebar = NewSidebar(state)

	split := container.NewHSplit(state.sidebar.Container, container.NewVScroll(state.detailPanel))
	split.SetOffset(0.25)

	w.SetCloseIntercept(func() {
		if state.currentValidate != nil && state.currentSave != nil {
			errs := state.currentValidate()
			if len(errs) > 0 {
				dialog.ShowCustomConfirm(
					"Ungültige Daten",
					"Schließen",
					"Abbrechen",
					widget.NewLabel("Aktuelle Eingaben sind ungültig:\n"+strings.Join(errs, "\n")),
					func(close bool) {
						if close {
							w.Close()
						}
					},
					w,
				)
				return
			}
			state.currentSave()
		}
		if err := store.Save(state.storeDir, state.store); err != nil {
			dialog.ShowCustomConfirm(
				"Speichern fehlgeschlagen",
				"Schließen",
				"Abbrechen",
				widget.NewLabel(err.Error()),
				func(close bool) {
					if close {
						w.Close()
					}
				},
				w,
			)
			return
		}
		w.Close()
	})

	return split
}

func (state *AppState) saveCurrentIfNeeded() {
	if state.currentValidate == nil || state.currentSave == nil {
		return
	}
	errs := state.currentValidate()
	if len(errs) == 0 {
		state.currentSave()
	}
}

// trySaveAndSwitch validates the current form. If valid, saves and calls switchFn.
// If invalid, shows a warning dialog asking the user whether to discard changes.
func (state *AppState) trySaveAndSwitch(switchFn func()) {
	if state.currentValidate == nil || state.currentSave == nil {
		switchFn()
		return
	}
	errs := state.currentValidate()
	if len(errs) == 0 {
		state.currentSave()
		if err := state.saveStore(); err != nil {
			dialog.ShowError(err, state.window)
		}
		switchFn()
		return
	}
	dialog.ShowCustomConfirm(
		"Ungültige Daten",
		"Verwerfen",
		"Zurück",
		widget.NewLabel("Aktuelle Eingaben sind ungültig:\n"+strings.Join(errs, "\n")),
		func(discard bool) {
			if discard {
				switchFn()
			}
		},
		state.window,
	)
}

func (state *AppState) saveStore() error {
	return store.Save(state.storeDir, state.store)
}

func (state *AppState) showEmployee(id string) {
	doSwitch := func() {
		state.currentEmployeeID = id
		state.currentYearConfig = 0
		state.currentValidate = nil
		state.currentSave = nil

		var emp *model.Employee
		for i := range state.store.Employees {
			if state.store.Employees[i].ID == id {
				emp = &state.store.Employees[i]
				break
			}
		}
		if emp == nil {
			state.showPlaceholder()
			return
		}

		form := NewEmployeeForm(state, emp)
		state.detailPanel.RemoveAll()
		state.detailPanel.Add(form)
		state.detailPanel.Refresh()
		state.refreshSidebar()
	}

	state.trySaveAndSwitch(doSwitch)
}

func (state *AppState) showYearConfig(year int) {
	doSwitch := func() {
		state.currentEmployeeID = ""
		state.currentYearConfig = year
		state.currentValidate = nil
		state.currentSave = nil

		yc := store.FindYearConfig(state.store, year)
		if yc == nil {
			state.showPlaceholder()
			return
		}

		form := NewYearConfigForm(state, yc)
		state.detailPanel.RemoveAll()
		state.detailPanel.Add(form)
		state.detailPanel.Refresh()
		state.refreshSidebar()
	}

	state.trySaveAndSwitch(doSwitch)
}

func (state *AppState) showPlaceholder() {
	state.currentEmployeeID = ""
	state.currentYearConfig = 0
	state.detailPanel.RemoveAll()
	state.detailPanel.Add(newPlaceholder())
	state.detailPanel.Refresh()
}

func (state *AppState) refreshSidebar() {
	state.sidebar.Refresh(state.store)
}

func newPlaceholder() fyne.CanvasObject {
	label := widget.NewLabel("Mitarbeiter auswählen oder erstellen.")
	label.Alignment = fyne.TextAlignCenter
	return container.New(layout.NewCenterLayout(), label)
}
