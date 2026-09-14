package main

import (
	_ "embed"

	"timesheets/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

//go:embed Icon.png
var iconData []byte

const appID = "io.github.matousjobanek.stundenzettel"

func main() {
	a := app.NewWithID(appID)
	a.SetIcon(fyne.NewStaticResource("Icon.png", iconData))
	w := a.NewWindow("Stundenzettel-Generator")
	w.Resize(fyne.NewSize(900, 800))
	w.SetContent(ui.BuildApp(a, w))
	w.ShowAndRun()
}
