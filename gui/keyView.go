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
	ideEntry := &widget.Entry{
		Text:     fmt.Sprintf("%X", fw.Keys.Data1[index].Value),
		Wrapping: fyne.TextWrapOff,
	}
	ideEntry.OnChanged = func(s string) {
		if len(s) > 8 {
			ideEntry.SetText(s[:8])
		}
		if len(s) == 8 {
			b, err := hex.DecodeString(s)
			if err != nil {
				dialog.ShowError(err, vw)
				return
			}
			if err := fw.SetKeyID(uint8(index), b); err != nil {
				dialog.ShowError(err, vw)
				return
			}
		}
	}

	syncEntry := &widget.Entry{
		Text:     fmt.Sprintf("%X", fw.Sync.Data[index]),
		Wrapping: fyne.TextWrapOff,
	}
	syncEntry.OnChanged = func(s string) {
		if len(s) > 8 {
			syncEntry.SetText(s[:8])
		}
		if len(s) == 8 {
			b, err := hex.DecodeString(s)
			if err == nil {
				if err := fw.SetSyncData(uint8(index), b); err != nil {
					dialog.ShowError(err, vw)
					return
				}
			}
		}
	}

	return container.NewVBox(
		kv(vw, "Type", "%s", fw.Keys.Data1[index].Type()),
		container.NewHBox(
			newBoldEntry("P0 IDE"),
			container.NewBorder(nil, nil, nil, nil, ideEntry),
			layout.NewSpacer(),
			widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
				e.mw.Clipboard().SetContent(ideEntry.Text)
			}),
		),
		kv(vw.e.mw, "P1 ISK Hi", "%X", fw.Keys.IskHI1),           // P1 ISK Hi, first 4 bytes
		kv(vw.e.mw, "P2 ISK Lo", "%X", fw.Keys.IskLO1),           // P2 ISK Lo, 2 bytes reserved and two are remaining ISK bytes
		kv(vw.e.mw, "P3", "%s", "96AA4854"),                      // P3
		kv(vw, "P4 PSK Hi", "%X", fw.PSK.High),                   //PSK first 4 bytes (like on cim dump analyzer)
		kv(vw, "P5 PSK Lo", "%X%X", fw.PSK.High[:2], fw.PSK.Low), // P5 is next two bytes of PSK but prefixed with its first two bytes
		kv(vw, "P6 PCF", "%s", "6732F2C5"),
		container.NewHBox(
			newBoldEntry("P7 Sync"),
			container.NewBorder(nil, nil, nil, nil, syncEntry),
			layout.NewSpacer(),
			widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
				e.mw.Clipboard().SetContent(syncEntry.Text)
			}),
		),
	)

}

func newBoldEntry(str string) *widget.Label {
	return widget.NewLabelWithStyle(str, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
}

/*
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
*/
