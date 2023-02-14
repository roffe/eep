package gui

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"github.com/hirschmann-koxha-gbr/eep/update"
	"golang.org/x/mod/semver"
)

const VERSION = "v2.0.16"

type EEPGui struct {
	port            string
	hwVersion       binding.String
	readDelayValue  binding.Float
	writeDelayValue binding.Float
	ignoreError     binding.Bool
	verifyWrite     binding.Bool

	mw *mainWindow
	sw *settingsWindow
	fyne.App
}

func New(app fyne.App) (*EEPGui, error) {
	eep := &EEPGui{
		App:             app,
		hwVersion:       binding.NewString(),
		readDelayValue:  binding.NewFloat(),
		writeDelayValue: binding.NewFloat(),
		ignoreError:     binding.NewBool(),
		verifyWrite:     binding.NewBool(),
	}

	if err := loadPrefs(eep); err != nil {
		return nil, err
	}
	eep.mw = newMainWindow(eep)

	return eep, nil
}

func loadPrefs(e *EEPGui) error {
	prefs := e.Preferences()

	if port := prefs.String("port"); port != "" {
		e.port = port
	}

	hw := prefs.StringWithFallback("hardware_version", "Uno")
	if err := e.hwVersion.Set(hw); err != nil {
		return err
	}

	readPinDelay := prefs.FloatWithFallback("read_pin_delay", 150)
	if err := e.readDelayValue.Set(readPinDelay); err != nil {
		return err
	}

	writePinDelay := prefs.FloatWithFallback("write_pin_delay", 150)
	if err := e.writeDelayValue.Set(writePinDelay); err != nil {
		return err
	}

	ignoreError := prefs.BoolWithFallback("ignore_read_errors", false)
	if err := e.ignoreError.Set(ignoreError); err != nil {
		return err
	}

	verifyWrite := prefs.BoolWithFallback("verify_write", true)
	if err := e.verifyWrite.Set(verifyWrite); err != nil {
		return err
	}
	return nil
}

func (e *EEPGui) CheckUpdate() {
	if latest, err := update.GetLatest(); err == nil {
		if semver.Compare(latest.TagName, VERSION) > 0 {
			dialog.ShowConfirm("Software update", "There is a new version available, would you like to visit the download page?", e.openWebpage, e.mw)
		}
	}
}

var releasepageURL = &url.URL{Scheme: "https", Host: "github.com", Path: "/Hirschmann-Koxha-GbR/eep/releases/latest"}

func (e *EEPGui) openWebpage(ok bool) {
	if ok {
		e.OpenURL(releasepageURL)
	}
}
