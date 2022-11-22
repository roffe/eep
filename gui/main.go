package gui

import (
	_ "embed"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/data/binding"
)

type EEPGui struct {
	app   fyne.App
	state *appState
	mw    *mainWindow
	hw    *helpWindow
	sw    *settingsWindow
}

type appState struct {
	port            string
	portList        []string
	readDelayValue  binding.Float
	writeDelayValue binding.Float
	ignoreError     binding.Bool
}

//go:embed Icon.png
var icon []byte
var appIcon = fyne.NewStaticResource("icon", icon)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func Run() {
	app := app.NewWithID("com.cimtool")
	app.SetIcon(appIcon)
	app.Settings().SetTheme(&gocanTheme{})

	eep := &EEPGui{
		app: app,
		state: &appState{
			port:            app.Preferences().String("port"),
			readDelayValue:  binding.NewFloat(),
			writeDelayValue: binding.NewFloat(),
			ignoreError:     binding.NewBool(),
		},
	}

	r := app.Preferences().Float("read_pin_delay")
	if r < 20 {
		r = 100
	}
	eep.state.readDelayValue.Set(r)

	w := app.Preferences().Float("write_pin_delay")
	if r < 100 {
		r = 200
	}
	eep.state.writeDelayValue.Set(w)
	eep.state.ignoreError.Set(app.Preferences().Bool("ignore_read_errors"))

	eep.mw = newMainWindow(eep)
	eep.hw = newHelpWindow(eep)
	eep.sw = newSettingsWindow(eep)

	eep.app.Run()
}
