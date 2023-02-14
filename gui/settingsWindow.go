package gui

import (
	"bufio"
	"bytes"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/hirschmann-koxha-gbr/eep/avr"
)

type settingsWindow struct {
	e                *EEPGui
	hwVerSelect      *widget.Select
	ignoreError      *widget.Check
	verifyWrite      *widget.Check
	readSliderLabel  *widget.Label
	readSlider       *widget.Slider
	writeSliderLabel *widget.Label
	writeSlider      *widget.Slider
	updateButton     *widget.Button

	fyne.Window
}

func newSettingsWindow(e *EEPGui) *settingsWindow {
	w := e.NewWindow("Settings")
	w.CenterOnScreen()
	w.SetOnClosed(func() {
		e.sw = nil
	})
	sw := &settingsWindow{
		e:      e,
		Window: w,
		hwVerSelect: widget.NewSelect([]string{"Uno", "Nano", "Nano (old bootloader)"}, func(s string) {
			e.hwVersion.Set(s)
			e.Preferences().SetString("hardware_version", s)
		}),
		ignoreError:      widget.NewCheckWithData("Ignore read validation errors", e.ignoreError),
		verifyWrite:      widget.NewCheckWithData("Verify written data", e.verifyWrite),
		readSliderLabel:  widget.NewLabel(""),
		readSlider:       widget.NewSliderWithData(0, 255, e.readDelayValue),
		writeSliderLabel: widget.NewLabel(""),
		writeSlider:      widget.NewSliderWithData(0, 255, e.writeDelayValue),
	}

	if f, err := sw.e.readDelayValue.Get(); err == nil {
		sw.readSliderLabel.SetText(delayLabel("Read", f))
	}

	if f, err := sw.e.writeDelayValue.Get(); err == nil {
		sw.writeSliderLabel.SetText(delayLabel("Write", f))
	}

	sw.hwVerSelect.Alignment = fyne.TextAlignCenter
	sw.hwVerSelect.PlaceHolder = "Select Arduino version"
	if hwVer, err := e.hwVersion.Get(); err == nil {
		sw.hwVerSelect.SetSelected(hwVer)
	}

	sw.ignoreError.OnChanged = func(b bool) {
		sw.e.Preferences().SetBool("ignore_read_errors", b)
		sw.e.ignoreError.Set(b)
	}

	sw.verifyWrite.OnChanged = func(b bool) {
		sw.e.Preferences().SetBool("verify_write", b)
		sw.e.verifyWrite.Set(b)
	}

	sw.readSlider.OnChanged = func(f float64) {
		sw.readSliderLabel.SetText(delayLabel("Read", f))
		sw.e.Preferences().SetFloat("read_pin_delay", f)
		sw.e.readDelayValue.Set(f)
	}

	sw.writeSlider.OnChanged = func(f float64) {
		sw.writeSliderLabel.SetText(delayLabel("Write", f))
		sw.e.Preferences().SetFloat("write_pin_delay", f)
		sw.e.writeDelayValue.Set(f)
	}

	sw.updateButton = widget.NewButtonWithIcon("Update firmware", theme.WarningIcon(), func() {
		sw.updateButton.Disable()
		defer sw.updateButton.Enable()

		sw.e.mw.disableButtons()
		defer sw.e.mw.enableButtons()

		hwVer, err := sw.e.hwVersion.Get()
		if err != nil {
			hwVer = "Uno"
		}

		dialog.ShowInformation("Update", fmt.Sprintf("Updating firmware for %s", hwVer), sw.e.mw)

		out, err := avr.Update(sw.e.port, hwVer, sw.e.mw.output)
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

	sw.SetContent(sw.layout())
	w.Resize(fyne.NewSize(400, 220))
	w.Show()
	return sw
}

func (sw *settingsWindow) layout() fyne.CanvasObject {
	return container.NewVBox(
		//&widget.Label{
		//	Text:      "CIM Tool Version: " + VERSION,
		//	TextStyle: fyne.TextStyle{Bold: true},
		//},
		container.NewHBox(widget.NewLabel("Arduino"), sw.hwVerSelect),
		sw.ignoreError,
		sw.verifyWrite,
		sw.readSliderLabel,
		sw.readSlider,
		sw.writeSliderLabel,
		sw.writeSlider,
		layout.NewSpacer(),
		sw.updateButton,
		//&widget.Button{
		//	Icon: theme.DocumentSaveIcon(),
		//	Text: "Save settings",
		//	OnTapped: func() {
		//		sw.Close()
		//	},
		//},
	)
}

func delayLabel(t string, f float64) string {
	return fmt.Sprintf("%s Pin Delay: %.0f", t, f)
}

func newSettingsView(e *EEPGui) fyne.CanvasObject {
	sw := &settingsWindow{
		e: e,
		hwVerSelect: widget.NewSelect([]string{"Uno", "Nano", "Nano (old bootloader)"}, func(s string) {
			e.hwVersion.Set(s)
			e.Preferences().SetString("hardware_version", s)
		}),
		ignoreError:      widget.NewCheckWithData("Ignore read validation errors", e.ignoreError),
		verifyWrite:      widget.NewCheckWithData("Verify written data", e.verifyWrite),
		readSliderLabel:  widget.NewLabel(""),
		readSlider:       widget.NewSliderWithData(0, 255, e.readDelayValue),
		writeSliderLabel: widget.NewLabel(""),
		writeSlider:      widget.NewSliderWithData(0, 255, e.writeDelayValue),
	}

	if f, err := sw.e.readDelayValue.Get(); err == nil {
		sw.readSliderLabel.SetText(delayLabel("Read", f))
	}

	if f, err := sw.e.writeDelayValue.Get(); err == nil {
		sw.writeSliderLabel.SetText(delayLabel("Write", f))
	}

	sw.hwVerSelect.Alignment = fyne.TextAlignCenter
	sw.hwVerSelect.PlaceHolder = "Select Arduino version"
	if hwVer, err := e.hwVersion.Get(); err == nil {
		sw.hwVerSelect.SetSelected(hwVer)
	}

	sw.ignoreError.OnChanged = func(b bool) {
		sw.e.Preferences().SetBool("ignore_read_errors", b)
		sw.e.ignoreError.Set(b)
	}

	sw.verifyWrite.OnChanged = func(b bool) {
		sw.e.Preferences().SetBool("verify_write", b)
		sw.e.verifyWrite.Set(b)
	}

	sw.readSlider.OnChanged = func(f float64) {
		sw.readSliderLabel.SetText(delayLabel("Read", f))
		sw.e.Preferences().SetFloat("read_pin_delay", f)
		sw.e.readDelayValue.Set(f)
	}

	sw.writeSlider.OnChanged = func(f float64) {
		sw.writeSliderLabel.SetText(delayLabel("Write", f))
		sw.e.Preferences().SetFloat("write_pin_delay", f)
		sw.e.writeDelayValue.Set(f)
	}

	sw.updateButton = widget.NewButtonWithIcon("Update firmware", theme.WarningIcon(), func() {
		sw.updateButton.Disable()
		sw.e.mw.disableButtons()
		go func() {
			sw.e.mw.appTabs.SelectIndex(1)
			defer sw.updateButton.Enable()
			defer sw.e.mw.enableButtons()

			hwVer, err := sw.e.hwVersion.Get()
			if err != nil {
				hwVer = "Uno"
			}

			out, err := avr.Update(sw.e.port, hwVer, sw.e.mw.output)
			if err != nil {
				sw.e.mw.output("Error updating: %v", err)
				return
			}

			r := bytes.NewReader(out)
			scanner := bufio.NewScanner(r)
			for scanner.Scan() {
				sw.e.mw.output("%s", scanner.Text())
			}
			sw.e.mw.appTabs.SelectIndex(len(sw.e.mw.appTabs.Items) - 1)
			dialog.ShowInformation("Update", "Firmware update complete", sw.e.mw)
		}()
	})

	return sw.layout()
}
