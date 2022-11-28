package gui

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"github.com/hirschmann-koxha-gbr/eep/update"
	"golang.org/x/mod/semver"
)

const VERSION = "v2.0.8"

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
	hwVersion       binding.String
	readDelayValue  binding.Float
	writeDelayValue binding.Float
	ignoreError     binding.Bool
}

func Run(a fyne.App) {
	state := &AppState{
		port:            a.Preferences().String("port"),
		hwVersion:       binding.NewString(),
		readDelayValue:  binding.NewFloat(),
		writeDelayValue: binding.NewFloat(),
		ignoreError:     binding.NewBool(),
	}

	prefs := a.Preferences()

	hw := prefs.StringWithFallback("hardware_version", "Uno")
	if err := state.hwVersion.Set(hw); err != nil {
		panic(err)
	}

	r := prefs.FloatWithFallback("read_pin_delay", 75)
	if err := state.readDelayValue.Set(r); err != nil {
		panic(err)
	}

	w := prefs.FloatWithFallback("write_pin_delay", 150)
	if err := state.writeDelayValue.Set(w); err != nil {
		panic(err)
	}

	ignoreError := prefs.BoolWithFallback("ignore_read_errors", false)
	if err := state.ignoreError.Set(ignoreError); err != nil {
		panic(err)
	}

	eep := &EEPGui{
		app:   a,
		state: state,
	}

	eep.mw = NewMainWindow(eep)

	go func() {
		latest, err := update.GetLatest()
		if err == nil {
			if semver.Compare(latest.TagName, VERSION) > 0 {
				dialog.ShowConfirm("Software update", "There is a new version available, would you like to visit the download page?", func(ok bool) {
					if ok {
						u, _ := url.Parse("https://github.com/Hirschmann-Koxha-GbR/eep/releases/latest")
						eep.app.OpenURL(u)
					}
				}, eep.mw.w)
			}
		}
	}()

	a.Run()
}
