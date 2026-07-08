package ui

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"timesheets/model"
	"timesheets/store"
)

// scrollInterceptor is a transparent overlay that captures scroll events
// and forwards them to the parent scroll container. This prevents Entry
// widgets from consuming vertical scroll events.
type scrollInterceptor struct {
	widget.BaseWidget
	scroll *container.Scroll
}

func newScrollInterceptor(scroll *container.Scroll) *scrollInterceptor {
	si := &scrollInterceptor{scroll: scroll}
	si.ExtendBaseWidget(si)
	return si
}

func (si *scrollInterceptor) CreateRenderer() fyne.WidgetRenderer {
	r := canvas.NewRectangle(color.Transparent)
	return widget.NewSimpleRenderer(r)
}

func (si *scrollInterceptor) Scrolled(ev *fyne.ScrollEvent) {
	if si.scroll != nil {
		si.scroll.Scrolled(ev)
	}
}

type AppState struct {
	store    *store.Store
	storeDir string
	window   fyne.Window

	sidebar *Sidebar
	scroll  *container.Scroll

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

	state.scroll = container.NewVScroll(newPlaceholder())

	state.sidebar = NewSidebar(state)

	split := container.NewHSplit(state.sidebar.Container, state.scroll)
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

func (state *AppState) setScrollContent(content fyne.CanvasObject) {
	interceptor := newScrollInterceptor(state.scroll)
	state.scroll.Content = container.NewStack(content, interceptor)
	state.scroll.ScrollToTop()
	state.scroll.Refresh()
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
		state.setScrollContent(form)
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
		state.setScrollContent(form)
		state.refreshSidebar()
	}

	state.trySaveAndSwitch(doSwitch)
}

func (state *AppState) showPlaceholder() {
	state.currentEmployeeID = ""
	state.currentYearConfig = 0
	state.setScrollContent(newPlaceholder())
}

func (state *AppState) refreshSidebar() {
	state.sidebar.Refresh(state.store)
}

func newPlaceholder() fyne.CanvasObject {
	label := widget.NewLabel("Mitarbeiter auswählen oder erstellen.")
	label.Alignment = fyne.TextAlignCenter
	return container.New(layout.NewCenterLayout(), label)
}
