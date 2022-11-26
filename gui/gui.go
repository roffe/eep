package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
)

type EEPGui struct {
	app   fyne.App
	state *AppState
	mw    *MainWindow
	hw    *HelpWindow
	sw    *SettingsWindow
}

type AppState struct {
	port            string
	portList        []string
	readDelayValue  binding.Float
	writeDelayValue binding.Float
	ignoreError     binding.Bool
}

func Run(a fyne.App) {
	state := &AppState{
		port:            a.Preferences().String("port"),
		readDelayValue:  binding.NewFloat(),
		writeDelayValue: binding.NewFloat(),
		ignoreError:     binding.NewBool(),
	}

	r := a.Preferences().FloatWithFallback("read_pin_delay", 75)
	if err := state.readDelayValue.Set(r); err != nil {
		panic(err)
	}

	w := a.Preferences().FloatWithFallback("write_pin_delay", 150)
	if err := state.writeDelayValue.Set(w); err != nil {
		panic(err)
	}

	ignoreError := a.Preferences().BoolWithFallback("ignore_read_errors", false)
	if err := state.ignoreError.Set(ignoreError); err != nil {
		panic(err)
	}

	eep := &EEPGui{
		app:   a,
		state: state,
	}

	eep.mw = NewMainWindow(eep)
	a.Run()
}
