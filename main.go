package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"timesheets/ui"
)

func main() {
	a := app.New()
	w := a.NewWindow("Kindergarten Dienstplan-Generator")
	w.Resize(fyne.NewSize(700, 800))
	w.SetContent(ui.BuildForm(w))
	w.ShowAndRun()
}
