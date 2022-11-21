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
	w.Resize(fyne.NewSize(300, 100))
	w.SetFixedSize(true)
	w.CenterOnScreen()
	sliderLabel := widget.NewLabel("")

	if f, err := state.delayValue.Get(); err == nil {
		sliderLabel.SetText(delayLabel(f))
	}

	slider := widget.NewSliderWithData(100, 300, state.delayValue)

	slider.OnChanged = func(f float64) {
		sliderLabel.SetText(delayLabel(f))
		app.Preferences().SetFloat("pin_delay", f)
		state.delayValue.Set(f)
	}

	w.SetContent(container.NewVBox(
		sliderLabel,
		slider,
		widget.NewSeparator(),
	))
	return w
}

func delayLabel(f float64) string {
	return fmt.Sprintf("Pin Delay: %.0f", f)
}
