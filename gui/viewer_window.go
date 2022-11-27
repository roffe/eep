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
			fmt.Sprintf(" MD5: %s", vw.cimBin.MD5()),
			fmt.Sprintf(" CRC32: %s", vw.cimBin.CRC32()),
			fmt.Sprintf(" VIN: %s", vw.cimBin.Vin.Data),
			fmt.Sprintf(" MY: %s", vw.cimBin.ModelYear()),
			fmt.Sprintf(" SAS: %t", vw.cimBin.SasOpt()),
			fmt.Sprintf(" End model (HW+SW): %d%s", vw.cimBin.PartNo1, vw.cimBin.PartNo1Rev),
			fmt.Sprintf(" Base model (HW+boot): %d%s", vw.cimBin.PnBase1, vw.cimBin.PnBase1Rev),
			fmt.Sprintf(" Delphi part number: %d", vw.cimBin.DelphiPN),
			fmt.Sprintf(" SAAB part number: %d", vw.cimBin.PartNo),
			fmt.Sprintf(" Configuration Version: %d", vw.cimBin.ConfigurationVersion),
		}
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
			name:  "Unknown Data 3 #1",
			start: 0x57,
			end:   0x80,
			color: colorUnknown,
		},
		{
			name:  "Unknown Data 3 #1 CRC",
			start: 81,
			end:   0x82,
			color: colorChecksum,
		},
		{
			name:  "Unknown Data 3 #2",
			start: 0x83,
			end:   0xac,
			color: colorUnknown,
		},
		{
			name:  "Unknown Data 3 #2 CRC",
			start: 0xad,
			end:   0xae,
			color: colorChecksum,
		},
		{
			name:  "PIN Data #1",
			start: 0xaf,
			end:   0xb2,
			color: rgb(56, 89, 217),
		},
		{
			name:  "PIN Unknown #1",
			start: 0xb3,
			end:   0xb6,
			color: colorUnknown,
		},
		{
			name:  "PIN CRC #1",
			start: 0xb7,
			end:   0xb8,
			color: colorChecksum,
		},

		{
			name:  "PIN Data #2",
			start: 0xb9,
			end:   0xbc,
			color: rgb(56, 89, 217),
		},
		{
			name:  "PIN Unknown #2",
			start: 0xbd,
			end:   0xc0,
			color: colorUnknown,
		},
		{
			name:  "PIN CRC #2",
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
		{
			name:  "KEYS ISK High #1",
			start: 0xfd,
			end:   0x100,
			color: rgb(37, 132, 20),
		},
		{
			name:  "KEYS ISK Low #1",
			start: 0x101,
			end:   0x102,
			color: rgb(137, 132, 120),
		},
		{
			name:  "KEYS Data #1",
			start: 0x103,
			end:   0x116,
			color: rgb(170, 120, 100),
		},
		{
			name:  "KEYS Count #1",
			start: 0x117,
			end:   0x117,
			color: rgb(170, 120, 100),
		},
		{
			name:  "KEYS Constant #1",
			start: 0x118,
			end:   0x11e,
			color: rgb(60, 40, 90),
		},
		{
			name:  "KEYS Errors #1",
			start: 0x11f,
			end:   0x11f,
			color: rgb(60, 40, 90),
		},
		{
			name:  "KEYS #1 CRC",
			start: 0x120,
			end:   0x121,
			color: colorChecksum,
		},

		{
			name:  "KEYS ISK High #2",
			start: 0x122,
			end:   0x124,
			color: rgb(37, 132, 20),
		},
		{
			name:  "KEYS ISK Low #2",
			start: 0x126,
			end:   0x127,
			color: rgb(137, 132, 120),
		},
		{
			name:  "KEYS Data #2",
			start: 0x128,
			end:   0x13b,
			color: rgb(170, 120, 100),
		},
		{
			name:  "KEYS Count #2",
			start: 0x13c,
			end:   0x13c,
			color: rgb(170, 120, 100),
		},
		{
			name:  "KEYS Constant #2",
			start: 0x13d,
			end:   0x143,
			color: rgb(60, 40, 90),
		},
		{
			name:  "KEYS Errors #2",
			start: 0x144,
			end:   0x144,
			color: rgb(60, 40, 90),
		},
		{
			name:  "KEYS #2 CRC",
			start: 0x145,
			end:   0x146,
			color: colorChecksum,
		},
		{
			name:  "Unknown Data 5",
			start: 0x147,
			end:   0x15d,
			color: colorUnknown,
		},
		{
			name:  "Unknown Data 5 CRC",
			start: 0x15e,
			end:   0x15f,
			color: colorChecksum,
		},
		{
			name:  "Sync Data",
			start: 0x160,
			end:   0x173,
			color: rgb(200, 220, 130),
		},
		{
			name:  "Sync Data CRC",
			start: 0x174,
			end:   0x175,
			color: colorChecksum,
		},
		{
			name:  "Unknown Data 6 #1",
			start: 0x176,
			end:   0x189,
			color: colorUnknown,
		},
		{
			name:  "Unknown Data 6 #1 CRC",
			start: 0x18a,
			end:   0x18b,
			color: colorChecksum,
		},
		{
			name:  "Unknown Data 6 #2",
			start: 0x18c,
			end:   0x19f,
			color: colorUnknown,
		},
		{
			name:  "Unknown Data 6 #2 CRC",
			start: 0x1a0,
			end:   0x1a1,
			color: colorChecksum,
		},
		{
			name:  "Unknown Data 7 #1",
			start: 0x1a2,
			end:   0x1a6,
			color: colorUnknown,
		},
		{
			name:  "Unknown Data 7 #1 CRC",
			start: 0x1a7,
			end:   0x1a8,
			color: colorChecksum,
		},
		{
			name:  "Unknown Data 7 #2",
			start: 0x1a9,
			end:   0x1ad,
			color: colorUnknown,
		},
		{
			name:  "Unknown Data 7 #2 CRC",
			start: 0x1ae,
			end:   0x1af,
			color: colorChecksum,
		},
		{
			name:  "Unknown Data 8",
			start: 0x1b0,
			end:   0x1b5,
			color: colorUnknown,
		},
		{
			name:  "Unknown Data 8 CRC",
			start: 0x1b6,
			end:   0x1b7,
			color: colorChecksum,
		},
		{
			name:  "Unknown Data 9",
			start: 0x1b8,
			end:   0x1bc,
			color: colorUnknown,
		},
		{
			name:  "Unknown Data 9 CRC",
			start: 0x1bd,
			end:   0x1be,
			color: colorChecksum,
		},
		{
			name:  "Unnown Data 2 #1",
			start: 0x1bf,
			end:   0x1c3,
			color: colorUnknown,
		},
		{
			name:  "Unnown Data 2 #1 CRC",
			start: 0x1c4,
			end:   0x1c5,
			color: colorChecksum,
		}, {
			name:  "Unnown Data 2 #1",
			start: 0x1c6,
			end:   0x1ca,
			color: colorUnknown,
		},
		{
			name:  "Unnown Data 2 #1 CRC",
			start: 0x1cb,
			end:   0x1cc,
			color: colorChecksum,
		},
		{
			name:  "SN Sticker",
			start: 0x1cd,
			end:   0x1d1,
			color: rgb(66, 166, 66),
		},
		{
			name:  "Factory Programming Date",
			start: 0x1d2,
			end:   0x1d4,
			color: rgb(166, 166, 166),
		},
		{
			name:  "Unknown Bytes 2",
			start: 0x1d5,
			end:   0x1d7,
			color: colorUnknown,
		},
		{
			name:  "Delphi PN",
			start: 0x1d8,
			end:   0x1db,
			color: rgb(200, 10, 14),
		},
		{
			name:  "Unknown Bytes 3",
			start: 0x1dc,
			end:   0x1dd,
			color: colorUnknown,
		},
		{
			name:  "Part No",
			start: 0x1de,
			end:   0x1e1,
			color: rgb(123, 31, 220),
		},
		{
			name:  "Unknown Data 14",
			start: 0x1e2,
			end:   0x1e4,
			color: colorUnknown,
		},
		{
			name:  "PSK Low",
			start: 0x1e5,
			end:   0x1e8,
			color: rgb(98, 42, 138),
		},
		{
			name:  "PSK Low",
			start: 0x1e9,
			end:   0x1ea,
			color: rgb(198, 72, 138),
		},
		{
			name:  "PSK Constant",
			start: 0x1eb,
			end:   0x1ee,
			color: rgb(33, 66, 77),
		},
		{
			name:  "PSK Unknown",
			start: 0x1ef,
			end:   0x1f0,
			color: colorUnknown,
		},
		{
			name:  "PSK Checksum",
			start: 0x1f1,
			end:   0x1f2,
			color: colorChecksum,
		},

		{
			name:  "SAS Calibration #1",
			start: 0x1f3,
			end:   0x1f6,
			color: rgb(65, 187, 35),
		},
		{
			name:  "SAS Calibration #1 CRC",
			start: 0x1f7,
			end:   0x1f8,
			color: colorChecksum,
		},

		{
			name:  "SAS Calibration #2",
			start: 0x1f3,
			end:   0x1f6,
			color: rgb(65, 187, 35),
		},
		{
			name:  "SAS Calibration #2 CRC",
			start: 0x1f9,
			end:   0x1fc,
			color: colorChecksum,
		},
		{
			name:  "EOF",
			start: 0x1ff,
			end:   0x1ff,
			color: rgb(255, 0, 0),
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
		row.Cells = append(row.Cells, widget.TextGridCell{
			Style: widget.TextGridStyleWhitespace,
		})
		rows = append(rows, row)
	}
	return rows
}
