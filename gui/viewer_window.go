package gui

import (
	"bytes"
	"fmt"
	"io"
	"log"
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
	e       *EEPGui
	w       fyne.Window
	toolbar *widget.Toolbar
	tabs    *container.AppTabs

	data []byte

	cimBin *cim.Bin
	saved  bool
}

var viewWindowSize = fyne.NewSize(550, 300)

func (vw *viewerWindow) Save() {
	if vw.cimBin != nil {
		b, err := vw.cimBin.XORBytes()
		if err != nil {
			dialog.ShowError(err, vw.w)
			return
		}
		vw.e.mw.saveFile("Save bin file", b)
	} else {
		vw.e.mw.saveFile("Save bin file", vw.data)
	}

}

func newViewerWindow(e *EEPGui, filename string, data []byte, askSaveOnClose bool) *viewerWindow {
	w := e.app.NewWindow("Viewing " + filename)
	vw := &viewerWindow{
		e:    e,
		w:    w,
		data: data,
	}

	vw.toolbar = vw.newToolbar()

	bin, err := cim.MustLoadBytes(filename, data)
	if err == nil {
		vw.cimBin = bin
	} else {
		w.SetContent(newHexView(vw))
		w.Show()
		return vw
	}

	w.SetCloseIntercept(func() {
		if askSaveOnClose && !vw.saved {
			dialog.ShowConfirm("Unsaved file", "Save file before closing?", func(b bool) {
				if b {
					vw.Save()
				}
				vw.w.Close()
			}, vw.w)
		} else {
			w.Close()

		}
	})

	var containers []*container.TabItem

	if vw.cimBin != nil {
		sasSelect := widget.NewSelect([]string{"Yes", "No"}, func(s string) {
			switch s {
			case "Yes":
				vw.cimBin.SetSasOpt(true)
			case "No":
				fallthrough
			default:
				vw.cimBin.SetSasOpt(false)
			}
		})

		sasSelect.SetSelected(func() string {
			if vw.cimBin.GetSasOpt() {
				return "Yes"
			}
			return "No"
		}())

		infoTab := container.NewTabItemWithIcon("Info", theme.SettingsIcon(), container.NewVBox(
			kv(vw.w, "MD5", "%s", vw.cimBin.MD5()),
			kv(vw.w, "CRC32", "%s", vw.cimBin.CRC32()),
			kv(vw.w, "Size", "%d", len(data)),
			kv(vw.w, "VIN", "%s", vw.cimBin.Vin.Data),
			kv(vw.w, "MY", "%s", vw.cimBin.ModelYear()),
			container.NewHBox(widget.NewLabelWithStyle("SAS", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), sasSelect),
			newEditEntry(vw.w, "ISK", fmt.Sprintf("%X%X", vw.cimBin.Keys.IskHI1, vw.cimBin.Keys.IskLO1),
				func(s string) {
					if len(s) == 12 {
						if err == nil {
							if err := vw.cimBin.Keys.SetISKHigh(s[:8]); err != nil {
								dialog.ShowError(err, vw.w)
								return
							}
							if err := vw.cimBin.Keys.SetISKLow(s[8:12]); err != nil {
								dialog.ShowError(err, vw.w)
								return
							}
						}
					}
				}),
			kv(vw.w, "PSK", "%X%X", vw.cimBin.PSK.Low, vw.cimBin.PSK.High),
		))

		versionTab := container.NewTabItemWithIcon("Versions", theme.QuestionIcon(), container.NewVBox(
			kv(vw.w, "End model (HW+SW)", "%d%s", vw.cimBin.PartNo1, vw.cimBin.PartNo1Rev),
			kv(vw.w, "Base model (HW+boot)", "%d%s", vw.cimBin.PnBase1, vw.cimBin.PnBase1Rev),
			kv(vw.w, "Delphi part number", "%d", vw.cimBin.DelphiPN),
			kv(vw.w, "SAAB part number", "%d", vw.cimBin.PartNo),
			kv(vw.w, "Configuration Version", "%d", vw.cimBin.ConfigurationVersion),
		))

		keyList := new(widget.List)
		keyList.Length = func() int {
			return int(vw.cimBin.Keys.Count1)
		}
		keyList.UpdateItem = func(id widget.ListItemID, item fyne.CanvasObject) {
			log.Println(id)
		}
		keyList.CreateItem = func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				layout.NewSpacer(),
				widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Italic: true}),
				widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{}),
				layout.NewSpacer(),
				widget.NewButtonWithIcon("Show", theme.HelpIcon(), func() {}),
				widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {}),
			)
		}

		keyList.UpdateItem = func(item widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*fyne.Container).Objects[0].(*widget.Label).SetText(fmt.Sprintf("Key #%d", item))
			obj.(*fyne.Container).Objects[2].(*widget.Label).SetText(vw.cimBin.Keys.Data1[item].Type())
			obj.(*fyne.Container).Objects[3].(*widget.Label).SetText(fmt.Sprintf("%02X", vw.cimBin.Keys.Data1[item].Value))
			obj.(*fyne.Container).Objects[5].(*widget.Button).OnTapped = func() {
				vw.w.SetContent(newKeyView(vw.e, vw, item, vw.cimBin))
			}
			obj.(*fyne.Container).Objects[6].(*widget.Button).OnTapped = func() {
				vw.cimBin.Keys.Count1--
				vw.cimBin.Keys.Count2--
				vw.cimBin.DeleteKey(item)
				data, err = vw.cimBin.XORBytes()
				if err != nil {
					dialog.ShowError(err, vw.w)
				}
				keyList.Refresh()
			}
		}

		keysTab := container.NewTabItemWithIcon("Keys", theme.LoginIcon(), container.NewBorder(
			nil,
			widget.NewButton("Add key", newAddKey(vw)),
			nil,
			nil,
			keyList,
		))

		vinSetting := widget.NewEntry()
		vinSetting.SetText(vw.cimBin.Vin.Data)
		vinSetting.Refresh()
		vinSetting.OnSubmitted = func(s string) {
			if err := vw.cimBin.Vin.Set(s); err != nil {
				dialog.ShowError(err, vw.w)
			}
		}
		vinSetting.Wrapping = fyne.TextWrapOff

		containers = append(containers, infoTab, versionTab, keysTab)
	}

	vw.tabs = container.NewAppTabs(containers...)

	w.SetContent(vw.layout())
	w.Resize(viewWindowSize)
	w.Show()
	return vw
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
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {
			vw.e.mw.saveFile("Save bin file", vw.data)
			vw.saved = true
		}),
		widget.NewToolbarAction(theme.UploadIcon(), func() {
			bin, err := vw.cimBin.XORBytes()
			if err != nil {
				dialog.ShowError(err, vw.w)
				return
			}
			dialog.ShowConfirm("Write to cim?", "Continue writing to CIM?", func(ok bool) {
				if ok {
					vw.e.mw.output("Flashing CIM ... ")
					start := time.Now()
					go func() {
						vw.e.mw.disableButtons()
						defer vw.e.mw.enableButtons()
						if ok := vw.e.mw.writeCIM(vw.e.state.port, bin); !ok {
							return
						}
						vw.e.mw.output("Flashed %s, took %s", "CIM", time.Since(start).String())
					}()

				}
			}, vw.w)

		}),
		widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
			fw, err := cim.MustLoadBytes("input file", vw.data)
			if err != nil {
				dialog.ShowError(err, vw.w)
				return
			}

			fw.Unmarry()
			b, err := fw.XORBytes()
			if err != nil {
				dialog.ShowError(err, vw.w)
				return
			}

			if vw.e.mw.saveFile("Save virginized bin", b) {
				dialog.ShowInformation("File file saved", "The virginized bin file has been saved.", vw.w)
			}
		}),
		widget.NewToolbarAction(theme.SearchIcon(), func() {
			b, err := vw.cimBin.Bytes()
			if err != nil {
				dialog.ShowError(err, vw.w)
				return
			}
			vw.data = b
			vw.w.SetContent(newHexView(vw))
		}),
	)
}

func (vw *viewerWindow) layout() fyne.CanvasObject {
	return container.NewBorder(vw.toolbar, nil, nil, nil,
		vw.tabs,
	)
}

func generateGrid(data []byte) []widget.TextGridRow {
	rowWidth := 32
	var rows []widget.TextGridRow
	r := bytes.NewReader(data)
	buff := make([]byte, rowWidth)
	pos := 0
	rPos := 0
	for {
		n, err := r.Read(buff)
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		var row widget.TextGridRow
		for x, bb := range buff[:n] {
			asd := fmt.Sprintf("%02X", bb)
			rw1 := widget.TextGridCell{
				Rune: rune(asd[0]),
				Style: &widget.CustomTextGridStyle{
					FGColor: viewColor(pos),
				},
			}
			row.Cells = append(row.Cells, rw1)

			rw2 := widget.TextGridCell{
				Rune: rune(asd[1]),
				Style: &widget.CustomTextGridStyle{
					FGColor: viewColor(pos),
				},
			}

			row.Cells = append(row.Cells, rw2)
			if x < rowWidth-1 {
				row.Cells = append(row.Cells, widget.TextGridCell{
					Style: widget.TextGridStyleWhitespace,
				})
			}
			pos++
		}

		if n < rowWidth {
			for ex := n; ex < rowWidth; ex++ {
				row.Cells = append(row.Cells, widget.TextGridCell{
					Style: widget.TextGridStyleWhitespace,
				})
				row.Cells = append(row.Cells, widget.TextGridCell{
					Style: widget.TextGridStyleWhitespace,
				})
				if ex < 31 {
					row.Cells = append(row.Cells, widget.TextGridCell{
						Style: widget.TextGridStyleWhitespace,
					})
				}
			}
		}

		row.Cells = append(row.Cells, widget.TextGridCell{
			Style: widget.TextGridStyleWhitespace,
		})
		row.Cells = append(row.Cells, widget.TextGridCell{
			Rune:  rune('║'),
			Style: widget.TextGridStyleWhitespace,
		})
		row.Cells = append(row.Cells, widget.TextGridCell{
			Style: widget.TextGridStyleWhitespace,
		})

		for _, bb := range buff[:n] {
			row.Cells = append(row.Cells, widget.TextGridCell{
				Rune: rune(bb),
				Style: &widget.CustomTextGridStyle{
					FGColor: viewColor(rPos),
				},
			})
			rPos++
		}
		rows = append(rows, row)
	}
	return rows
}
