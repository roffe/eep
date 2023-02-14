package gui

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/hirschmann-koxha-gbr/cim/pkg/cim"
)

type viewerWindow struct {
	e *EEPGui

	saved          bool
	askSaveOnClose bool

	data   []byte
	cimBin *cim.Bin

	keyList *widget.List

	toolbar    *widget.Toolbar
	infoTab    *container.TabItem
	versionTab *container.TabItem
	tabs       *container.AppTabs

	fyne.Window
}

func newViewerView(e *EEPGui, filename string, data []byte, askSaveOnClose bool) fyne.CanvasObject {
	vw := &viewerWindow{
		e:      e,
		data:   data,
		Window: e.mw,
	}

	if bin, err := cim.MustLoadBytes(filename, data); err == nil {
		vw.cimBin = bin
	} else {
		vw.toolbar = vw.newToolbar()
		return container.NewBorder(vw.toolbar, nil, nil, nil,
			newHexView(vw),
		)
	}

	return vw.layout()

}

func (vw *viewerWindow) save() {
	if vw.cimBin != nil {
		bin, err := vw.cimBin.XORBytes()
		if err != nil {
			dialog.ShowError(err, vw)
			return
		}
		vw.e.mw.saveFile("Save bin file",
			fmt.Sprintf("cim_%x_%s.bin", vw.cimBin.SnSticker, time.Now().Format("20060102-150405")),
			bin)
		return
	}
	vw.e.mw.saveFile("Sav raw bin file", fmt.Sprintf("cim_raw_%s.bin", time.Now().Format("20060102-150405")), vw.data)
}

func (vw *viewerWindow) closeIntercept() {
	if vw.askSaveOnClose && !vw.saved {
		dialog.ShowConfirm("Unsaved file", "Save file before closing?", func(ok bool) {
			if ok {
				vw.save()
			}
			vw.Close()
		}, vw)
		return
	}
	vw.Close()
}

func kv(w fyne.Window, k, valueFormat string, values ...interface{}) *fyne.Container {
	return container.NewHBox(
		//widget.NewLabelWithStyle(k+":", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Monospace: true}),
		newBoldEntry(k+":"),
		&widget.Label{
			Text:     fmt.Sprintf(valueFormat, values...),
			Wrapping: fyne.TextWrapOff,
		},
		layout.NewSpacer(),
		widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
			w.Clipboard().SetContent(fmt.Sprintf(valueFormat, values...))
		}),
	)
}

func (vw *viewerWindow) newToolbar() *widget.Toolbar {
	saveAction := widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {
		if vw.cimBin != nil {
			bin, err := vw.cimBin.XORBytes()
			if err != nil {
				dialog.ShowError(err, vw)
				return
			}
			vw.e.mw.saveFile("Save bin file", fmt.Sprintf("cim_%x_%s.bin", vw.cimBin.SnSticker, time.Now().Format("20060102-15_04_05")), bin)
			vw.saved = true
			return
		}
		vw.e.mw.saveFile("Save raw bin file", fmt.Sprintf("cim_raw_%s.bin", time.Now().Format("20060102-15_04_05")), vw.data)
		vw.saved = true
	})
	writeAction := widget.NewToolbarAction(theme.UploadIcon(), func() {
		if vw.cimBin == nil {
			dialog.ShowError(errors.New("Not valid cim eeprom"), vw) //lint:ignore ST1005 ignore this error
			return
		}
		bin, err := vw.cimBin.XORBytes()
		if err != nil {
			dialog.ShowError(err, vw)
			return
		}
		dialog.ShowConfirm("Write to CIM?", "Continue writing to CIM?", func(ok bool) {
			if ok {
				vw.e.mw.output("Flashing CIM ... ")
				start := time.Now()
				go func() {
					vw.e.mw.disableButtons()
					defer vw.e.mw.enableButtons()
					if err := vw.e.mw.writeCIM(vw.e.port, bin); err != nil {
						dialog.ShowError(err, vw)
						return
					}
					dialog.ShowInformation("Write done", fmt.Sprintf("Write successfull, took %s", time.Since(start).Round(time.Millisecond).String()), vw)
				}()
			}
		}, vw)
	})

	/*
		hexAction := widget.NewToolbarAction(theme.SearchIcon(), func() {
			var err error
			vw.data, err = vw.cimBin.Bytes()
			if err != nil {
				dialog.ShowError(err, vw)
				return
			}
			vw.e.mw.docTab.Items[vw.e.mw.docTab.SelectedIndex()].Content = newHexView(vw)
			//vw.SetContent(newHexView(vw))
			//vw.Resize(fyne.NewSize(920, 240))
		})
	*/

	//homeAction := widget.NewToolbarAction(theme.HomeIcon(), func() {
	//	if vw.cimBin != nil {
	//		vw.SetContent(vw.layout())
	//		vw.Resize(viewWindowSize)
	//	}
	//})

	//editAction := widget.NewToolbarAction(theme.WarningIcon(), func() {
	//	vw.SetContent(newEditView(vw))
	//})

	toolbar := widget.NewToolbar(
		//homeAction,
		saveAction,
		writeAction,
		//widget.NewToolbarSeparator(),
	)

	//toolbar.Append(widget.NewToolbarSpacer())
	//toolbar.Append(editAction)
	return toolbar
}

func (vw *viewerWindow) layout() fyne.CanvasObject {
	vw.toolbar = vw.newToolbar()
	vw.infoTab = container.NewTabItemWithIcon("Info", theme.InfoIcon(), vw.renderInfoTab())
	vw.versionTab = container.NewTabItemWithIcon("Versions", theme.QuestionIcon(), widget.NewForm(
		widget.NewFormItem("End model (HW+SW)", widget.NewLabel(fmt.Sprintf("%d%s", vw.cimBin.PartNo1, vw.cimBin.PartNo1Rev))),
		widget.NewFormItem("Base model (HW+boot)", widget.NewLabel(fmt.Sprintf("%d%s", vw.cimBin.PnBase1, vw.cimBin.PnBase1Rev))),
		widget.NewFormItem("Delphi part number", widget.NewLabel(fmt.Sprintf("%d", vw.cimBin.DelphiPN))),
		widget.NewFormItem("SAAB part number", widget.NewLabel(fmt.Sprintf("%d", vw.cimBin.PartNo))),
		widget.NewFormItem("Configuration Version", widget.NewLabel(fmt.Sprintf("%d", vw.cimBin.ConfigurationVersion))),
	))
	vw.keyList = &widget.List{
		Length: func() int {
			return int(vw.cimBin.Keys.Count1)
		},
		CreateItem: func() fyne.CanvasObject {
			return container.NewHBox(
				&widget.Label{TextStyle: fyne.TextStyle{Bold: true}},
				layout.NewSpacer(),
				&widget.Label{TextStyle: fyne.TextStyle{Italic: true}},
				&widget.Label{},
				layout.NewSpacer(),
				widget.NewButtonWithIcon("Show", theme.HelpIcon(), func() {}),
				widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {}),
			)
		},
		UpdateItem: func(item widget.ListItemID, obj fyne.CanvasObject) {
			c, ok := obj.(*fyne.Container)
			if ok {
				c.Objects[0].(*widget.Label).SetText(fmt.Sprintf("Key #%d", item))
				c.Objects[2].(*widget.Label).SetText(vw.cimBin.Keys.Data1[item].Type())
				c.Objects[3].(*widget.Label).SetText(fmt.Sprintf("%02X", vw.cimBin.Keys.Data1[item].Value))
				c.Objects[5].(*widget.Button).OnTapped = func() {
					dialog.ShowCustom(fmt.Sprintf("Key #%d", item), "OK", newKeyView(vw.e, vw, item, vw.cimBin), vw)
				}
				obj.(*fyne.Container).Objects[6].(*widget.Button).OnTapped = func() {
					vw.cimBin.Keys.Count1--
					vw.cimBin.Keys.Count2--
					vw.cimBin.DeleteKey(item)
					var err error
					vw.data, err = vw.cimBin.XORBytes()
					if err != nil {
						dialog.ShowError(err, vw)
					}
					vw.keyList.Refresh()
				}
			}
		},
	}

	keysTab := container.NewTabItemWithIcon("Keys", theme.LoginIcon(),
		container.NewBorder(
			nil,
			widget.NewButtonWithIcon("Add key", theme.ContentAddIcon(), func() {
				vw.addKey()
			}),
			nil,
			nil,
			vw.keyList,
		),
	)

	hexTab := container.NewTabItemWithIcon("Hex", theme.SearchIcon(), newHexView(vw))
	vw.tabs = container.NewAppTabs(vw.infoTab, vw.versionTab, keysTab, hexTab)

	return container.NewBorder(vw.toolbar, nil, nil, nil,
		vw.tabs,
	)
}

func hexValidator(length int) func(s string) error {
	return func(s string) error {
		if len(s) != length {
			return errors.New("invalid length")
		}
		if _, err := hex.DecodeString(s); err != nil {
			return errors.New("invalid hex")
		}
		return nil
	}

}

func (vw *viewerWindow) renderInfoTab() fyne.CanvasObject {
	vButton := widget.NewButtonWithIcon("Virginize", theme.SearchReplaceIcon(), func() {
		vw.cimBin.Unmarry()
		vw.infoTab.Content = vw.renderInfoTab()
		dialog.ShowInformation("Virginization complete", "The file has been virginized.\nNow flash the eeprom, re-assemble the car and add the CIM with Tech2", vw)
		vw.askSaveOnClose = true
		vw.saved = false
	})

	iskEntry := &widget.Entry{
		Text:      fmt.Sprintf("%X%X", vw.cimBin.Keys.IskHI1, vw.cimBin.Keys.IskLO1),
		Wrapping:  fyne.TextWrapOff,
		Validator: hexValidator(12),
	}
	iskEntry.OnChanged = func(s string) {
		if len(s) > 12 {
			iskEntry.SetText(s[:12])
		}
		if len(s) == 12 {
			if err := vw.cimBin.Keys.SetISKHigh(s[:8]); err != nil {
				dialog.ShowError(err, vw)
				return
			}
			if err := vw.cimBin.Keys.SetISKLow(s[8:12]); err != nil {
				dialog.ShowError(err, vw)
				return
			}
		}
	}

	pskEntry := &widget.Entry{
		Text:      fmt.Sprintf("%X%X", vw.cimBin.PSK.High, vw.cimBin.PSK.Low),
		Wrapping:  fyne.TextWrapOff,
		Validator: hexValidator(12),
	}
	pskEntry.OnChanged = func(s string) {
		if len(s) > 12 {
			pskEntry.SetText(s[:12])
		}
		if len(s) == 12 {
			decoded, err := hex.DecodeString(s)
			if err != nil {
				dialog.ShowError(err, vw)
				return
			}
			if err := vw.cimBin.PSK.SetHigh(decoded[:4]); err != nil {
				dialog.ShowError(err, vw)
				return
			}
			if err := vw.cimBin.PSK.SetLow(decoded[4:6]); err != nil {
				dialog.ShowError(err, vw)
				return
			}
		}
	}

	vinEntry := &widget.Entry{Text: vw.cimBin.Vin.Data}
	vinEntry.OnChanged = func(s string) {
		if len(s) > 17 {
			vinEntry.SetText(s[:17])
		}
		if len(s) == 17 {
			vw.cimBin.Vin.Data = s
			if err := vw.cimBin.Vin.Set(s); err != nil {
				dialog.ShowError(err, vw)
			}
		}
	}

	pin := string(vw.cimBin.Pin.Data1[:])
	if bytes.Equal(vw.cimBin.Pin.Data1, []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		pin = "not set"
	}

	pinHex := &widget.Label{Text: fmt.Sprintf("%X", vw.cimBin.Pin.Data1)}

	pinEntry := &widget.Entry{Text: pin}
	pinEntry.OnChanged = func(s string) {
		if len(s) > 4 {
			pinEntry.SetText(s[:4])
		}
		if len(s) == 4 {
			if err := vw.cimBin.SetPin(fmt.Sprintf("%X", s)); err != nil {
				dialog.ShowError(err, vw)
			}
			pinHex.SetText(fmt.Sprintf("%X", vw.cimBin.Pin.Data1))
		}
	}

	sasSelect := &widget.Select{
		Options: []string{"Yes", "No"},
		OnChanged: func(s string) {
			vw.cimBin.SetSasOpt(s == "Yes")
		},
		Selected: func() string {
			if vw.cimBin.GetSasOpt() {
				return "Yes"
			}
			return "No"
		}(),
	}

	snstickerEntry := &widget.Entry{Text: fmt.Sprintf("%X", vw.cimBin.SnSticker)}
	snstickerEntry.OnChanged = func(s string) {
		if len(s) > 10 {
			snstickerEntry.SetText(s[:10])
			return
		}
		if len(s) == 10 {
			b, err := hex.DecodeString(s)
			if err != nil {
				dialog.ShowError(err, vw)
				return
			}
			vw.cimBin.SnSticker = b
		}
	}

	form := widget.NewForm(
		widget.NewFormItem("MD5", container.NewHBox(
			widget.NewLabel(vw.cimBin.MD5()),
			layout.NewSpacer(),
			widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
				vw.Clipboard().SetContent(vw.cimBin.MD5())
			}),
		)),
		widget.NewFormItem("CRC32", container.NewHBox(
			widget.NewLabel(vw.cimBin.CRC32()),
			layout.NewSpacer(),
			widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
				vw.Clipboard().SetContent(vw.cimBin.CRC32())
			}),
		)),
		widget.NewFormItem("S/N Sticker", container.NewHBox(
			container.NewHBox(
				snstickerEntry,
			),
			layout.NewSpacer(),
			widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
				vw.Clipboard().SetContent(vinEntry.Text)
			}),
		)),
		widget.NewFormItem("VIN", container.NewHBox(
			container.NewHBox(
				vinEntry,
			),
			container.NewHBox(
				widget.NewLabelWithStyle("MY:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabelWithStyle(vw.cimBin.ModelYear(), fyne.TextAlignLeading, fyne.TextStyle{}),
			),
			layout.NewSpacer(),
			widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
				vw.Clipboard().SetContent(vinEntry.Text)
			}),
		)),
		widget.NewFormItem("PIN", container.NewHBox(
			pinEntry,
			layout.NewSpacer(),
			widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
				vw.Clipboard().SetContent(pinEntry.Text)
			}),
		)),
		widget.NewFormItem("PIN (hex)", container.NewHBox(
			pinHex,
			layout.NewSpacer(),
			widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
				vw.Clipboard().SetContent(fmt.Sprintf("%X", pinHex.Text))
			}),
		)),
		widget.NewFormItem("SAS", sasSelect),
		widget.NewFormItem("ISK", container.NewHBox(
			iskEntry,
			layout.NewSpacer(),
			widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
				vw.Clipboard().SetContent(iskEntry.Text)
			}),
		)),
		widget.NewFormItem("PSK", container.NewHBox(
			pskEntry,
			layout.NewSpacer(),
			widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
				vw.Clipboard().SetContent(iskEntry.Text)
			}),
		)),
	)
	return container.NewBorder(nil, vButton, nil, nil, form)
}
