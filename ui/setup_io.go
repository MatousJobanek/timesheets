package ui

import (
	"fmt"
	"io"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"timesheets/store"
)

func (state *AppState) installFileMenu() {
	exportItem := fyne.NewMenuItem("Einrichtung exportieren…", state.exportSetup)
	importItem := fyne.NewMenuItem("Einrichtung importieren…", state.importSetup)
	state.window.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("Datei", exportItem, importItem),
	))
}

func (state *AppState) prepareStoreForIO() bool {
	if state.currentValidate == nil || state.currentSave == nil {
		return true
	}
	errs := state.currentValidate()
	if len(errs) > 0 {
		dialog.ShowError(fmt.Errorf("%s", strings.Join(errs, "\n")), state.window)
		return false
	}
	state.currentSave()
	return true
}

func (state *AppState) exportSetup() {
	if !state.prepareStoreForIO() {
		return
	}
	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, saveErr error) {
		if saveErr != nil {
			dialog.ShowError(saveErr, state.window)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()
		data, err := store.ExportJSON(state.store)
		if err != nil {
			dialog.ShowError(err, state.window)
			return
		}
		if _, err := writer.Write(data); err != nil {
			dialog.ShowError(err, state.window)
		}
	}, state.window)
	saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
	saveDialog.SetFileName("einrichtung.json")
	saveDialog.Show()
}

func (state *AppState) importSetup() {
	if !state.prepareStoreForIO() {
		return
	}
	openDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, openErr error) {
		if openErr != nil {
			dialog.ShowError(openErr, state.window)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()
		data, err := io.ReadAll(reader)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Datei lesen fehlgeschlagen: %w", err), state.window)
			return
		}
		imported, err := store.ImportJSON(data)
		if err != nil {
			dialog.ShowError(err, state.window)
			return
		}
		apply := func() {
			state.replaceStore(imported)
		}
		if state.store.IsEmpty() {
			apply()
			return
		}
		dialog.ShowCustomConfirm(
			"Einrichtung ersetzen?",
			"Ersetzen",
			"Abbrechen",
			widget.NewLabel("Vorhandene Mitarbeiter und Jahre werden durch die importierte Einrichtung ersetzt."),
			func(ok bool) {
				if ok {
					apply()
				}
			},
			state.window,
		)
	}, state.window)
	openDialog.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
	openDialog.Show()
}

func (state *AppState) replaceStore(imported *store.Store) {
	state.store = imported
	state.currentValidate = nil
	state.currentSave = nil
	if err := state.saveStore(); err != nil {
		dialog.ShowError(err, state.window)
		return
	}
	state.showPlaceholder()
	state.refreshSidebar()
}
