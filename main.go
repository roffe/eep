package main

import (
	_ "embed"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"github.com/hirschmann-koxha-gbr/eep/gui"
)

//go:embed Icon.png
var icon []byte
var appIcon = fyne.NewStaticResource("icon", icon)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main() {
	application := app.NewWithID("com.cimtool")
	application.SetIcon(appIcon)
	//application.Settings().SetTheme(&gui.Theme{})
	application.Settings().SetTheme(theme.DarkTheme())
	ui, err := gui.New(application)
	if err != nil {
		log.Fatal(err)
	}
	application.Lifecycle().SetOnStarted(ui.CheckUpdate)
	application.Run()
}
