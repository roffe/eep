package gui

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"strings"

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
	grid    *widget.TextGrid
	data    []byte
	cimBin  *cim.Bin
	saved   bool
	xor     bool
}

func newViewerWindow(e *EEPGui, filename string, data []byte, askSaveOnClose bool) *viewerWindow {
	w := e.app.NewWindow("Viewing " + filename)
	vw := &viewerWindow{
		e:    e,
		w:    w,
		data: data,
		grid: widget.NewTextGrid(),
		xor:  true,
	}
	vw.toolbar = vw.newToolbar()

	w.SetCloseIntercept(func() {
		if askSaveOnClose && !vw.saved {
			dialog.ShowConfirm("Unsaved file", "Save file before closing?", func(b bool) {
				if b {
					e.mw.saveFile("Save bin file", vw.data)
				}
				w.Close()
			}, vw.w)
		} else {
			w.Close()

		}
	})

	for i, row := range generateGrid(vw.data) {
		vw.grid.SetRow(i, row)
	}
	bin, err := cim.MustLoadBytes(filename, data)
	if err == nil {
		vw.cimBin = bin
	}

	w.SetContent(vw.layout())
	w.Show()
	return vw
}

func (vw *viewerWindow) newToolbar() *widget.Toolbar {
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {
			vw.e.mw.saveFile("Save bin file", vw.data)
			vw.saved = true
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.ContentClearIcon(), func() {
			tmpData := make([]byte, len(vw.data))
			if vw.xor {
				copy(tmpData, vw.data)
				vw.xor = false
			} else {
				copy(tmpData, vw.data)
				for i := range tmpData {
					tmpData[i] ^= 0xff
				}
				vw.xor = true
			}
			for i, r := range generateGrid(tmpData) {
				for j, c := range r.Cells {
					vw.grid.SetCell(i, j, c)
				}
			}
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
	)
}

func (vw *viewerWindow) layout() fyne.CanvasObject {
	var containers []fyne.CanvasObject
	if vw.cimBin != nil {
		infoItems := []string{
			fmt.Sprintf("MD5: %s", vw.cimBin.MD5()),
			fmt.Sprintf("CRC32: %s", vw.cimBin.CRC32()),
			fmt.Sprintf("VIN: %s", vw.cimBin.Vin.Data),
			fmt.Sprintf("MY: %s", vw.cimBin.ModelYear()),
			fmt.Sprintf("SAS: %t", vw.cimBin.SasOpt()),
			fmt.Sprintf("End model (HW+SW): %d%s", vw.cimBin.PartNo1, vw.cimBin.PartNo1Rev),
			fmt.Sprintf("Base model (HW+boot): %d%s", vw.cimBin.PnBase1, vw.cimBin.PnBase1Rev),
			fmt.Sprintf("Delphi part number: %d", vw.cimBin.DelphiPN),
			fmt.Sprintf("SAAB part number: %d", vw.cimBin.PartNo),
			fmt.Sprintf("Configuration Version: %d", vw.cimBin.ConfigurationVersion),
		}
		/*
			list := widget.NewList(func() int {
				return len(infoItems)
			}, func() fyne.CanvasObject {
				w := widget.NewLabel("")
				w.Wrapping = fyne.TextWrapOff
				w.TextStyle.Monospace = true
				return w
			}, func(item widget.ListItemID, obj fyne.CanvasObject) {
				obj.(*widget.Label).SetText(infoItems[item])
			})*/

		tg := widget.NewTextGridFromString(strings.Join(infoItems, "\n"))
		containers = append(containers, tg)
	}

	containers = append(containers, vw.grid)

	return container.NewBorder(vw.toolbar, nil, nil, nil,
		container.New(layout.NewHBoxLayout(),
			containers...,
		),
	)
}

type colorDesc struct {
	name  string
	start int
	end   int
	color color.RGBA
}

var (
	colorChecksum = rgb(0, 255, 0)
	colorUnknown  = rgb(33, 33, 33)

	colorList = []colorDesc{
		{
			name:  "Programming date",
			start: 0x0,
			end:   0x3,
			color: rgb(128, 128, 0),
		},

		{
			name:  "Sas Option",
			start: 0x4,
			end:   0x4,
			color: rgb(50, 200, 0),
		},
		{
			name:  "Unknown Bytes 1",
			start: 0x5,
			end:   0xa,
			color: colorUnknown,
		},
		{
			name:  "PartNo 1",
			start: 0xb,
			end:   0xe,
			color: rgb(160, 18, 34),
		},
		{
			name:  "PartNo 1 Revision",
			start: 0xf,
			end:   0x10,
			color: rgb(60, 60, 10),
		},
		{
			name:  "Configuration Version",
			start: 0x11,
			end:   0x14,
			color: rgb(255, 0, 255),
		},
		{
			name:  "PNBase",
			start: 0x15,
			end:   0x18,
			color: rgb(45, 72, 200),
		},
		{
			name:  "PNBase Revision",
			start: 0x19,
			end:   0x1a,
			color: rgb(100, 100, 43),
		},
		{
			name:  "VIN Data",
			start: 0x1b,
			end:   0x2b,
			color: rgb(200, 30, 76),
		},
		{
			name:  "VIN Value",
			start: 0x2c,
			end:   0x2c,
			color: rgb(240, 240, 10),
		},
		{
			name:  "VIN Unknown",
			start: 0x2d,
			end:   0x35,
			color: rgb(35, 156, 63),
		},
		{
			name:  "VIN SPS Count",
			start: 0x36,
			end:   0x36,
			color: rgb(66, 22, 88),
		},
		{
			name:  "VIN Checksum",
			start: 0x37,
			end:   0x38,
			color: colorChecksum,
		},
		{
			name:  "Programming ID",
			start: 0x39,
			end:   0x56,
			color: rgb(72, 140, 38),
		},
		{
			name:  "Unknown Data 3 Bank #1",
			start: 0x57,
			end:   0x80,
			color: colorUnknown,
		},
		{
			name:  "Unknown Data 3 Bank #1 CRC",
			start: 81,
			end:   0x82,
			color: colorChecksum,
		},
		{
			name:  "Unknown Data 3 Bank #2",
			start: 0x83,
			end:   0xac,
			color: colorUnknown,
		},
		{
			name:  "Unknown Data 3 Bank #2 CRC",
			start: 0xad,
			end:   0xae,
			color: colorChecksum,
		},
		{
			name:  "PIN Data Bank #1",
			start: 0xaf,
			end:   0xb2,
			color: rgb(56, 89, 217),
		},
		{
			name:  "PIN Unknown Bank #1",
			start: 0xb3,
			end:   0xb6,
			color: colorUnknown,
		},
		{
			name:  "PIN CRC Bank #1",
			start: 0xb7,
			end:   0xb8,
			color: colorChecksum,
		},

		{
			name:  "PIN Data Bank #2",
			start: 0xb9,
			end:   0xbc,
			color: rgb(56, 89, 217),
		},
		{
			name:  "PIN Unknown Bank #2",
			start: 0xbd,
			end:   0xc0,
			color: colorUnknown,
		},
		{
			name:  "PIN CRC Bank #2",
			start: 0xc1,
			end:   0xc2,
			color: colorChecksum,
		},
		{
			name:  "Unknown Data 3",
			start: 0xc3,
			end:   0xc4,
			color: colorUnknown,
		},
		{
			name:  "Unknown Data 1",
			start: 0xc7,
			end:   0xf0,
			color: colorUnknown,
		},
		{
			name:  "Unknwon Data 2 CRC",
			start: 0xf1,
			end:   0xf2,
			color: colorChecksum,
		},
		{
			name:  "Const 1 Data",
			start: 0xf3,
			end:   0xfa,
			color: rgb(40, 5, 113),
		},
		{
			name:  "Const 1 CRC",
			start: 0xfb,
			end:   0xfc,
			color: colorChecksum,
		},
		/*
			{
				name:  "",
				start: 0x,
				end:   0x,
				color: rgb(),
			},
		*/
	}
)

func rgb(r, g, b uint8) color.RGBA {
	return color.RGBA{R: r, G: g, B: b, A: 1}
}

func viewColor(pos int) color.RGBA {
	for _, c := range colorList {
		if (c.start == c.end) && c.start == pos {
			return c.color
		}
		if pos >= c.start && pos <= c.end {
			return c.color
		}
	}
	return rgb(255, 255, 255)
}

func generateGrid(data []byte) []widget.TextGridRow {
	var rows []widget.TextGridRow
	r := bytes.NewReader(data)
	buff := make([]byte, 32)
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

		for x, bb := range buff[:n] {
			asd := fmt.Sprintf("%02X", bb)
			row.Cells = append(row.Cells, widget.TextGridCell{
				Rune: rune(asd[0]),
				Style: &widget.CustomTextGridStyle{
					FGColor: viewColor(pos),
				},
			})

			row.Cells = append(row.Cells, widget.TextGridCell{
				Rune: rune(asd[1]),
				Style: &widget.CustomTextGridStyle{
					FGColor: viewColor(pos),
				},
			})
			if x < 31 {
				row.Cells = append(row.Cells, widget.TextGridCell{
					Style: widget.TextGridStyleWhitespace,
				})
			}
			pos++
		}

		if n < 32 {
			for ex := n; ex < 32; ex++ {
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
