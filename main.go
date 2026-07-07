package main

import (
	_ "embed"

	"timesheets/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

//go:embed Icon.png
var iconData []byte

func main() {
	a := app.New()
	a.SetIcon(fyne.NewStaticResource("Icon.png", iconData))
	w := a.NewWindow("Kindergarten Dienstplan-Generator")
	w.Resize(fyne.NewSize(700, 800))
	w.SetContent(ui.BuildForm(w))
	w.ShowAndRun()
}
