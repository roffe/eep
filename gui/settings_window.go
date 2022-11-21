package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func newSettingsWindow(app fyne.App) fyne.Window {
	w := app.NewWindow("Settings")
	w.SetCloseIntercept(func() {
		w.Hide()
	})
	w.Resize(fyne.NewSize(300, 110))
	w.SetFixedSize(true)
	w.CenterOnScreen()

	readSliderLabel := widget.NewLabel("")
	if f, err := state.readDelayValue.Get(); err == nil {
		readSliderLabel.SetText(delayLabel("Read", f))
	}
	readSlider := widget.NewSliderWithData(20, 300, state.readDelayValue)
	readSlider.OnChanged = func(f float64) {
		readSliderLabel.SetText(delayLabel("Read", f))
		app.Preferences().SetFloat("read_pin_delay", f)
		state.readDelayValue.Set(f)
	}

	writeSliderLabel := widget.NewLabel("")
	if f, err := state.writeDelayValue.Get(); err == nil {
		writeSliderLabel.SetText(delayLabel("Write", f))
	}
	writeSlider := widget.NewSliderWithData(100, 300, state.writeDelayValue)
	writeSlider.OnChanged = func(f float64) {
		writeSliderLabel.SetText(delayLabel("Write", f))
		app.Preferences().SetFloat("write_pin_delay", f)
		state.writeDelayValue.Set(f)
	}

	ignoreError := widget.NewCheckWithData("Ignore read validation errors", state.ignoreError)
	ignoreError.OnChanged = func(b bool) {
		app.Preferences().SetBool("ignore_read_errors", b)
	}
	w.SetContent(container.NewVBox(
		ignoreError,
		readSliderLabel,
		readSlider,
		writeSliderLabel,
		writeSlider,
		widget.NewSeparator(),
	))
	return w
}

func delayLabel(t string, f float64) string {
	return fmt.Sprintf("%s Pin Delay: %.0f", t, f)
}
