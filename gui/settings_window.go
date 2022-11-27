package gui

import (
	"bufio"
	"bytes"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/hirschmann-koxha-gbr/eep/avr"
)

type SettingsWindow struct {
	e                *EEPGui
	w                fyne.Window
	ignoreError      *widget.Check
	readSliderLabel  *widget.Label
	readSlider       *widget.Slider
	writeSliderLabel *widget.Label
	writeSlider      *widget.Slider
	updateButton     *widget.Button
}

func NewSettingsWindow(e *EEPGui) *SettingsWindow {
	w := e.app.NewWindow("Settings")
	w.CenterOnScreen()
	w.SetOnClosed(func() {
		e.sw = nil
	})
	sw := &SettingsWindow{
		e:                e,
		w:                w,
		ignoreError:      widget.NewCheckWithData("Ignore read validation errors", e.state.ignoreError),
		readSliderLabel:  widget.NewLabel(""),
		readSlider:       widget.NewSliderWithData(0, 400, e.state.readDelayValue),
		writeSliderLabel: widget.NewLabel(""),
		writeSlider:      widget.NewSliderWithData(0, 400, e.state.writeDelayValue),
	}

	if f, err := sw.e.state.readDelayValue.Get(); err == nil {
		sw.readSliderLabel.SetText(delayLabel("Read", f))
	}

	if f, err := sw.e.state.writeDelayValue.Get(); err == nil {
		sw.writeSliderLabel.SetText(delayLabel("Write", f))
	}

	sw.ignoreError.OnChanged = func(b bool) {
		sw.e.app.Preferences().SetBool("ignore_read_errors", b)
	}

	sw.readSlider.OnChanged = func(f float64) {
		sw.readSliderLabel.SetText(delayLabel("Read", f))
		sw.e.app.Preferences().SetFloat("read_pin_delay", f)
		sw.e.state.readDelayValue.Set(f)
	}

	sw.writeSlider.OnChanged = func(f float64) {
		sw.writeSliderLabel.SetText(delayLabel("Write", f))
		sw.e.app.Preferences().SetFloat("write_pin_delay", f)
		sw.e.state.writeDelayValue.Set(f)
	}

	sw.updateButton = widget.NewButtonWithIcon("Update firmware", theme.WarningIcon(), func() {
		sw.updateButton.Disable()
		defer sw.updateButton.Enable()

		sw.e.mw.disableButtons()
		defer sw.e.mw.enableButtons()

		out, err := avr.Update(sw.e.state.port, sw.e.mw.output)
		if err != nil {
			sw.e.mw.output("Error updating: %v", err)
			return
		}

		r := bytes.NewReader(out)
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			sw.e.mw.output("%s", scanner.Text())
		}

	})

	sw.w.SetContent(sw.layout())
	w.Resize(fyne.NewSize(400, 185))
	w.Show()
	return sw
}

func (sw *SettingsWindow) layout() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabelWithStyle("CIM Tool Version: "+VERSION, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		sw.ignoreError,
		sw.readSliderLabel,
		sw.readSlider,
		sw.writeSliderLabel,
		sw.writeSlider,
		layout.NewSpacer(),
		sw.updateButton,
	)
}

func delayLabel(t string, f float64) string {
	return fmt.Sprintf("%s Pin Delay: %.0f", t, f)
}
