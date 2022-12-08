package gui

import (
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

	toolbar *widget.Toolbar
	tabs    *container.AppTabs

	fyne.Window
}

var viewWindowSize = fyne.NewSize(550, 300)

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
	vw.Resize(viewWindowSize)
}

func (vw *viewerWindow) save() {
	if vw.cimBin != nil {
		bin, err := vw.cimBin.XORBytes()
		if err != nil {
			dialog.ShowError(err, vw)
			return
		}
		vw.e.mw.saveFile("Save bin file", bin)
		return
	}
	vw.e.mw.saveFile("Sav raw bin file", vw.data)
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
		widget.NewLabelWithStyle(k+":", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle(fmt.Sprintf(valueFormat, values...), fyne.TextAlignLeading, fyne.TextStyle{}),
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
			vw.e.mw.saveFile("Save bin file", bin)
			vw.saved = true
			return
		}
		vw.e.mw.saveFile("Save raw bin file", vw.data)
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
					if ok := vw.e.mw.writeCIM(vw.e.port, bin); !ok {
						return
					}
					dialog.ShowInformation("Write done", fmt.Sprintf("Write completed in %s", time.Since(start).Round(time.Millisecond).String()), vw)
				}()
			}
		}, vw)
	})

	resetAction := widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
		vw.cimBin.Unmarry()
		vw.SetContent(vw.layout())
		dialog.ShowInformation("Virginization complete", "The file has been virginized.\nNow flash the CIM and re-assemble the car and program new key(s) using Tech2", vw)
		vw.askSaveOnClose = true
		vw.saved = false
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

	home := func() {
		if vw.cimBin != nil {
			vw.SetContent(vw.layout())
			vw.Resize(viewWindowSize)
		}
	}

	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.HomeIcon(), home),
		saveAction,
	)

	if vw.cimBin != nil {
		toolbar.Append(writeAction)
		toolbar.Append(resetAction)
		toolbar.Append(hexAction)
	}

	return toolbar
}

func (vw *viewerWindow) layout() fyne.CanvasObject {
	vw.toolbar = vw.newToolbar()
	sasSelect := &widget.Select{
		Options: []string{"Yes", "No"},
		OnChanged: func(s string) {
			vw.cimBin.SetSasOpt(s=="Yes")
		},
		Selected: func()string{
			if vw.cimBin.GetSasOpt() {
				return "Yes"
			}
			return "No"
		}(),
	}
	

	iskEntry := &widget.Entry{
		Text:     fmt.Sprintf("%X%X", vw.cimBin.Keys.IskHI1, vw.cimBin.Keys.IskLO1),
		Wrapping: fyne.TextWrapOff,
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

	infoTab := container.NewTabItemWithIcon("Info", theme.SettingsIcon(), container.NewVBox(
		kv(vw, "MD5", "%s", vw.cimBin.MD5()),
		kv(vw, "CRC32", "%s", vw.cimBin.CRC32()),
		kv(vw, "Size", "%d", len(vw.data)),
		kv(vw, "VIN", "%s", vw.cimBin.Vin.Data),
		kv(vw, "MY", "%s", vw.cimBin.ModelYear()),
		container.NewHBox(widget.NewLabelWithStyle("SAS", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), sasSelect),
		container.NewHBox(
			newBoldEntry("ISK"),
			container.NewBorder(nil, nil, nil, nil, iskEntry),
			layout.NewSpacer(),
			widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
				vw.e.mw.Clipboard().SetContent(iskEntry.Text)
			}),
		),
		kv(vw, "PSK", "%X%X", vw.cimBin.PSK.Low, vw.cimBin.PSK.High),
	))

	versionTab := container.NewTabItemWithIcon("Versions", theme.QuestionIcon(), container.NewVBox(
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
			c,ok := obj.(*fyne.Container)
			if ok {
				c.Objects[0].(*widget.Label).SetText(fmt.Sprintf("Key #%d", item))
				c.Objects[2].(*widget.Label).SetText(vw.cimBin.Keys.Data1[item].Type())
				c.Objects[3].(*widget.Label).SetText(fmt.Sprintf("%02X", vw.cimBin.Keys.Data1[item].Value))
				c.Objects[5].(*widget.Button).OnTapped = func() {
					vw.SetContent(newKeyView(vw.e, vw, item, vw.cimBin))
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
			widget.NewButton("Add key", newAddKey(vw)),
			nil,
			nil,
			vw.keyList,
		),
	)

	vw.tabs = container.NewAppTabs(infoTab, versionTab, keysTab)

	return container.NewBorder(vw.toolbar, nil, nil, nil,
		vw.tabs,
	)
}
