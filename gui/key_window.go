package gui

import (
	"encoding/hex"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/hirschmann-koxha-gbr/cim/pkg/cim"
)

func newKeyView(e *EEPGui, vw *viewerWindow, index int, fw *cim.Bin) fyne.CanvasObject {
	return container.NewVBox(
		kv(vw.w, "Type", "%s", fw.Keys.Data1[index].Type()),
		newEditEntry(vw.e.mw.w, "P0 IDE",
			fmt.Sprintf("%X", fw.Keys.Data1[index].Value),
			func(s string) {
				if len(s) == 8 {
					b, err := hex.DecodeString(s)
					if err != nil {
						dialog.ShowError(err, vw.w)
						return
					}
					if err := fw.SetKeyID(uint8(index), b); err != nil {
						dialog.ShowError(err, vw.w)
					}
				}
			}),
		kv(vw.e.mw.w, "P1 ISK Hi", "%X", fw.Keys.IskHI1), // P1 ISK Hi, first 4 bytes
		/*
			fmt.Sprintf("%X", fw.Keys.IskHI1),
				func(s string) {
					if len(s) == 8 {
						if err := fw.Keys.SetISKHigh(s); err != nil {
							dialog.ShowError(err, vw.w)
						}
					}
				}),
		*/
		kv(vw.e.mw.w, "P2 ISK Lo", "%X", fw.Keys.IskLO1), // P2 ISK Lo, 2 bytes reserved and two are remaining ISK bytes
		/*
			fmt.Sprintf("%X", fw.Keys.IskLO1),
				func(s string) {
					if len(s) == 8 {
						if err := fw.Keys.SetISKLow(s); err != nil {
							dialog.ShowError(err, vw.w)
						}
					}
				}),
		*/
		// widget.NewLabel("P3 is like control bytes (keyfob settings)"),
		// widget.NewLabel("P4 is unknown yet"),
		// widget.NewLabel("P5 is 00WWYYYY - probably production date"),
		// widget.NewLabel("P6 is revision - i have always seen only 00004141 (AA)"),
		// widget.NewLabel("P7 is part number, you can google it"),

		//widget.NewLabel("P4 PSK: "fw.PSK.Low)),

		kv(vw.w, "P4 PSK Hi", "%X%X", fw.PSK.High, fw.PSK.Constant[:2]), //PSK first 4 bytes (like on cim dump analyzer)
		kv(vw.w, "P5 PSK Lo", "%X%X", fw.PSK.High, fw.PSK.Low[:2]),      // P5 is next two bytes of PSK but prefixed with its first two bytes

		kv(vw.w, "P6 PCF", "%s", "6732F2C5"),

		newEditEntry(vw.w, "P7 Sync", fmt.Sprintf("%X", fw.Sync.Data[index]), // sync from eeprom
			func(s string) {
				if len(s) == 8 {
					b, err := hex.DecodeString(s)
					if err == nil {
						if len(b) == 16 {
							if err := fw.SetSyncData(uint8(index), b); err != nil {
								dialog.ShowError(err, vw.w)
							}
						}
					}
				}
			}),

		layout.NewSpacer(),
		widget.NewButton("Close", func() {
			vw.w.SetContent(vw.layout())
		}),
	)
}

func newEditEntry(w fyne.Window, key, value string, onChange func(s string)) fyne.CanvasObject {
	entry := &widget.Entry{
		Text:      value,
		Wrapping:  fyne.TextWrapOff,
		OnChanged: onChange,
	}
	return container.NewHBox(
		newBoldEntry(key),
		container.NewBorder(nil, nil, nil, nil, entry),
		layout.NewSpacer(),
		widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
			w.Clipboard().SetContent(entry.Text)
		}),
	)
}

func newBoldEntry(str string) *widget.Label {
	return widget.NewLabelWithStyle(str, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
}

func kve(w fyne.Window, k, valueFormat string, values ...interface{}) *fyne.Container {
	text := fmt.Sprintf(valueFormat, values...)
	ew := &widget.Entry{
		Text:     text,
		Wrapping: fyne.TextWrapOff,
	}

	return container.NewHBox(
		widget.NewLabelWithStyle(k+":", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		ew,
		layout.NewSpacer(),
		widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
			w.Clipboard().SetContent(fmt.Sprintf(valueFormat, values...))
		}),
	)
}
