package gui

import (
	"encoding/hex"
	"errors"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func newAddKey(vw *viewerWindow) func() {
	return func() {
		ideEntry := widget.NewEntry()
		ideEntry.OnChanged = func(s string) {
			if len(s) > 8 {
				ideEntry.SetText(s[:8])
				return
			}
		}
		ideEntry.Validator = func(s string) error {
			if len(s) != 8 {
				return errors.New("invalid length")
			}
			if _, err := hex.DecodeString(s); err != nil {
				return errors.New("invalid hex value")
			}
			return nil
		}

		syncEntry := widget.NewEntry()
		syncEntry.OnChanged = func(s string) {
			if len(s) > 8 {
				syncEntry.SetText(s[:8])
				return
			}
		}
		syncEntry.Validator = func(s string) error {
			if len(s) != 8 {
				return errors.New("invalid length")
			}
			if _, err := hex.DecodeString(s); err != nil {
				return errors.New("invalid hex value")
			}
			return nil
		}

		form := widget.NewForm(
			&widget.FormItem{
				Text:     "IDE",
				HintText: "Enter Key ID",
				Widget:   ideEntry,
			},
			&widget.FormItem{
				Text:     "Sync",
				HintText: "Enter Key Sync Data",
				Widget:   syncEntry,
			},
		)

		form.SubmitText = "Add key"

		form.OnSubmit = func() {
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
			vw.SetContent(vw.layout())
		}

		vw.SetContent(container.NewVBox(
			form,
			layout.NewSpacer(),
			widget.NewButtonWithIcon("Close", theme.DeleteIcon(), func() {
				vw.SetContent(vw.layout())
			}),
		))
	}
}
