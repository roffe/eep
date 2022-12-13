package gui

import (
	"encoding/hex"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func (vw *viewerWindow) addKey() {
	ideEntry := &widget.Entry{
		Wrapping:  fyne.TextWrapOff,
		Validator: hexValidator(8),
	}
	ideEntry.OnChanged = func(s string) {
		if len(s) > 8 {
			ideEntry.SetText(s[:8])
			return
		}
	}

	syncEntry := &widget.Entry{
		Wrapping:  fyne.TextWrapOff,
		Validator: hexValidator(8),
	}
	syncEntry.OnChanged = func(s string) {
		if len(s) > 8 {
			syncEntry.SetText(s[:8])
			return
		}
	}

	rec := canvas.NewRectangle(color.Transparent)
	rec.SetMinSize(fyne.NewSize(100, 10))

	dialog.ShowForm("Add key", "Add", "Cancel", []*widget.FormItem{
		{
			Text:     "IDE",
			HintText: "Enter IDE",
			Widget:   container.NewMax(rec, ideEntry),
		},
		{
			Text:     "Sync",
			HintText: "Enter Sync Data",
			Widget:   container.NewMax(rec, syncEntry),
		},
		{
			Widget: layout.NewSpacer(),
		},
	}, func(ok bool) {
		if ok {
			keyID, err := hex.DecodeString(ideEntry.Text)
			if err != nil {
				dialog.ShowError(err, vw)
				return
			}

			if err := vw.cimBin.SetKeyID(uint8(vw.cimBin.Keys.Count1), keyID); err != nil {
				dialog.ShowError(err, vw)
				return
			}

			syncData, err := hex.DecodeString(syncEntry.Text)
			if err != nil {
				dialog.ShowError(err, vw)
				return
			}

			if err := vw.cimBin.SetSyncData(uint8(vw.cimBin.Keys.Count1), syncData); err != nil {
				dialog.ShowError(err, vw)
				return
			}
			if err := vw.cimBin.SetKeyCount(vw.cimBin.Keys.Count1 + 1); err != nil {
				dialog.ShowError(err, vw)
				return
			}
			//vw.tabs.Refresh()
			vw.keyList.Refresh()
		}
	}, vw)
}
