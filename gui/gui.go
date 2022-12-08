package gui

import (
	"log"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"github.com/hirschmann-koxha-gbr/eep/update"
	"golang.org/x/mod/semver"
)

const VERSION = "v2.0.10"

type EEPGui struct {
	port            string
	portList        []string
	hwVersion       binding.String
	readDelayValue  binding.Float
	writeDelayValue binding.Float
	ignoreError     binding.Bool

	mw *mainWindow
	sw *settingsWindow
	fyne.App
}

func New(a fyne.App) *EEPGui {
	eep := &EEPGui{
		App:             a,
		port:            a.Preferences().String("port"),
		hwVersion:       binding.NewString(),
		readDelayValue:  binding.NewFloat(),
		writeDelayValue: binding.NewFloat(),
		ignoreError:     binding.NewBool(),
	}

	eep.loadPrefs()
	eep.mw = newMainWindow(eep)

	return eep
}

func (e *EEPGui) loadPrefs() {
	prefs := e.Preferences()

	hw := prefs.StringWithFallback("hardware_version", "Uno")
	if err := e.hwVersion.Set(hw); err != nil {
		log.Fatal(err)
	}

	readPinDelay := prefs.FloatWithFallback("read_pin_delay", 150)
	if err := e.readDelayValue.Set(readPinDelay); err != nil {
		log.Fatal(err)
	}

	writePinDelay := prefs.FloatWithFallback("write_pin_delay", 150)
	if err := e.writeDelayValue.Set(writePinDelay); err != nil {
		log.Fatal(err)
	}

	ignoreError := prefs.BoolWithFallback("ignore_read_errors", false)
	if err := e.ignoreError.Set(ignoreError); err != nil {
		log.Fatal(err)
	}
}

func (e *EEPGui) CheckUpdate() {
	latest, err := update.GetLatest()
	if err == nil {
		if semver.Compare(latest.TagName, VERSION) > 0 {
			dialog.ShowConfirm("Software update", "There is a new version available, would you like to visit the download page?", func(ok bool) {
				if ok {
					u, _ := url.Parse("https://github.com/Hirschmann-Koxha-GbR/eep/releases/latest")
					e.OpenURL(u)
				}
			}, e.mw)
		}
	}
}

func (e *EEPGui) Start() {
	e.mw.Show()
	e.Run()
}
