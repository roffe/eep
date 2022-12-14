package gui

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
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

var viewWindowSize = fyne.NewSize(550, 320)

func newViewerWindow(e *EEPGui, filename string, data []byte, askSaveOnClose bool) {
	vw := &viewerWindow{
		e:      e,
		Window: e.NewWindow("Viewing " + filename),
		data:   data,
	}
	defer vw.Show()

	if bin, err := cim.MustLoadBytes(filename, data); err == nil {
		vw.cimBin = bin
	} else {
		vw.toolbar = vw.newToolbar()
		vw.SetContent(newHexView(vw))
		return
	}

	vw.SetCloseIntercept(vw.closeIntercept)
	vw.SetContent(vw.layout())
	vw.CenterOnScreen()
	vw.Resize(viewWindowSize)
}

func (vw *viewerWindow) save() {
	if vw.cimBin != nil {
		bin, err := vw.cimBin.XORBytes()
		if err != nil {
			dialog.ShowError(err, vw)
			return
		}
		vw.e.mw.saveFile("Save bin file",
			fmt.Sprintf("cim_%d_%s.bin", vw.cimBin.SnSticker, time.Now().Format("20060102-150405")),
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
			vw.e.mw.saveFile("Save bin file", fmt.Sprintf("cim_%d_%s.bin", vw.cimBin.SnSticker, time.Now().Format("20060102-15_04_05")), bin)
			vw.saved = true
			return
		}
		vw.e.mw.saveFile("Save raw bin file", fmt.Sprintf("cim_raw_%s.bin", time.Now().Format("20060102-15_04_05")), vw.data)
		vw.saved = true
	})
	writeAction := widget.NewToolbarAction(theme.UploadIcon(), func() {
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
					dialog.ShowInformation("Write done", fmt.Sprintf("Write completed in %s", time.Since(start).Round(time.Millisecond).String()), vw)
				}()
			}
		}, vw)
	})

	hexAction := widget.NewToolbarAction(theme.SearchIcon(), func() {
		var err error
		vw.data, err = vw.cimBin.Bytes()
		if err != nil {
			dialog.ShowError(err, vw)
			return
		}
		vw.SetContent(newHexView(vw))
		vw.Resize(fyne.NewSize(920, 240))
	})

	homeAction := widget.NewToolbarAction(theme.HomeIcon(), func() {
		if vw.cimBin != nil {
			vw.SetContent(vw.layout())
			vw.Resize(viewWindowSize)
		}
	})

	editAction := widget.NewToolbarAction(theme.WarningIcon(), func() {
		vw.SetContent(newEditView(vw))
	})

	toolbar := widget.NewToolbar(
		homeAction,
		saveAction,
		widget.NewToolbarSeparator(),
	)

	if vw.cimBin != nil {
		toolbar.Append(writeAction)
		//toolbar.Append(resetAction)
		toolbar.Append(hexAction)
	}
	toolbar.Append(widget.NewToolbarSpacer())
	toolbar.Append(editAction)
	return toolbar
}

func (vw *viewerWindow) layout() fyne.CanvasObject {
	vw.toolbar = vw.newToolbar()

	vw.infoTab = container.NewTabItemWithIcon("Info", theme.InfoIcon(), vw.renderInfoTab())

	vw.versionTab = container.NewTabItemWithIcon("Versions", theme.QuestionIcon(), container.NewVBox(
		kv(vw, "End model (HW+SW)", "%d%s", vw.cimBin.PartNo1, vw.cimBin.PartNo1Rev),
		kv(vw, "Base model (HW+boot)", "%d%s", vw.cimBin.PnBase1, vw.cimBin.PnBase1Rev),
		kv(vw, "Delphi part number", "%d", vw.cimBin.DelphiPN),
		kv(vw, "SAAB part number", "%d", vw.cimBin.PartNo),
		kv(vw, "Configuration Version", "%d", vw.cimBin.ConfigurationVersion),
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
					//vw.SetContent(newKeyView(vw.e, vw, item, vw.cimBin))
					dialog.ShowCustom("Edit key", "OK", newKeyView(vw.e, vw, item, vw.cimBin), vw)
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

	vw.tabs = container.NewAppTabs(vw.infoTab, vw.versionTab, keysTab)

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
		vw.SetContent(vw.layout())
		dialog.ShowInformation("Virginization complete", "The file has been virginized.\nNow flash the eeprom, re-assemble the car and add the CIM with Tech2", vw)
		vw.askSaveOnClose = true
		vw.saved = false
	})

	iskEntry := &widget.Entry{
		Text:      fmt.Sprintf("%X%X", vw.cimBin.Keys.IskHI1, vw.cimBin.Keys.IskLO1),
		Wrapping:  fyne.TextWrapOff,
		TextStyle: fyne.TextStyle{Monospace: true},
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
		Text:      fmt.Sprintf("%X%X", vw.cimBin.PSK.Low, vw.cimBin.PSK.High),
		Wrapping:  fyne.TextWrapOff,
		TextStyle: fyne.TextStyle{Monospace: true},
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

			if err := vw.cimBin.PSK.SetLow(decoded[:4]); err != nil {
				dialog.ShowError(err, vw)
				return
			}
			if err := vw.cimBin.PSK.SetHigh(decoded[4:6]); err != nil {
				dialog.ShowError(err, vw)
				return
			}
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

	pin := string(vw.cimBin.Pin.Data1[:])
	if bytes.Equal(vw.cimBin.Pin.Data1, []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		pin = "not set"
	}

	vin := func() string {
		if vw.cimBin.Vin.Data == strings.Repeat(" ", 17) {
			return "not set"
		}
		return vw.cimBin.Vin.Data
	}()

	left := container.NewVBox(
		kv(vw, "MD5  ", "%s", vw.cimBin.MD5()),
		kv(vw, "CRC32", "%s", vw.cimBin.CRC32()),
		kv(vw, "Size ", "%d", len(vw.data)),

		layout.NewSpacer(),

		container.NewHBox(
			container.NewHBox(
				newBoldEntry("VIN  :"),
				widget.NewLabelWithStyle(vin, fyne.TextAlignLeading, fyne.TextStyle{}),
			),
			container.NewHBox(
				widget.NewLabelWithStyle("MY:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabelWithStyle(vw.cimBin.ModelYear(), fyne.TextAlignLeading, fyne.TextStyle{}),
			),
			layout.NewSpacer(),
			widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
				vw.e.mw.Clipboard().SetContent(vw.cimBin.Vin.Data)
			}),
		),

		kv(vw, "PIN  ", "%s", pin),
		kv(vw, "PIN (hex)", "%X", vw.cimBin.Pin.Data1),

		layout.NewSpacer(),

		container.NewHBox(
			newBoldEntry("SAS  :"),
			sasSelect,
		),
		container.NewHBox(
			newBoldEntry("ISK  :"),
			container.NewBorder(nil, nil, nil, nil, iskEntry),
			layout.NewSpacer(),
			widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
				vw.e.mw.Clipboard().SetContent(iskEntry.Text)
			}),
		),
		container.NewHBox(
			newBoldEntry("PSK  :"),
			container.NewBorder(nil, nil, nil, nil, pskEntry),
			layout.NewSpacer(),
			widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
				vw.Clipboard().SetContent(iskEntry.Text)
			}),
		),
		layout.NewSpacer(),
		vButton,
	)

	//right := container.NewVBox(
	//
	//	layout.NewSpacer(),
	//)

	// return container.NewGridWithColumns(2, left, right)
	return left

}
