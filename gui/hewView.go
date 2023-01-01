package gui

import (
	"bytes"
	"fmt"
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func newHexView(vw *viewerWindow) fyne.CanvasObject {
	grid := &widget.TextGrid{
		Rows: generateGrid(vw.data),
	}
	/*
			home := func() {
				vw.w.SetContent(vw.layout())
				vw.w.Resize(viewWindowSize)
			}
			xorHex := func() {
				for i := range vw.data {
				vw.data[i] ^= 0xff
			}
			grid.Rows = generateGrid(vw.data)
			grid.Refresh()
		}
	*/

	//vw.toolbar.Append(widget.NewToolbarAction(theme.ContentClearIcon(), xorHex))
	/*
		return container.NewBorder(
			//widget.NewToolbar(
			//	widget.NewToolbarAction(theme.HomeIcon(), home),
			//	widget.NewToolbarAction(theme.DocumentSaveIcon(), vw.save),
			//	widget.NewToolbarAction(theme.ContentClearIcon(), xorHex),
			//),
			vw.toolbar,
			nil,
			nil,
			nil,
			grid,
		)
	*/
	return grid
}

func generateGrid(data []byte) []widget.TextGridRow {
	rowWidth := 32
	var rows []widget.TextGridRow
	r := bytes.NewReader(data)
	buff := make([]byte, rowWidth)
	pos := 0
	rPos := 0
	rowNo := 0

	headerRow := widget.TextGridRow{}

	for i := 0; i < 4; i++ {
		headerRow.Cells = append(headerRow.Cells, widget.TextGridCell{
			Style: widget.TextGridStyleWhitespace,
		})
	}

	for i := 0; i < 32; i++ {
		dd := fmt.Sprintf("%02X", i)
		headerRow.Cells = append(headerRow.Cells,
			widget.TextGridCell{
				Rune:  rune(dd[0]),
				Style: &widget.CustomTextGridStyle{},
			},
			widget.TextGridCell{
				Rune:  rune(dd[1]),
				Style: &widget.CustomTextGridStyle{},
			},
			widget.TextGridCell{
				Style: widget.TextGridStyleWhitespace,
			},
		)

	}

	headerRow.Cells = append(headerRow.Cells, widget.TextGridCell{
		Rune:  rune('║'),
		Style: widget.TextGridStyleWhitespace,
	})

	rows = append(rows, headerRow)

	for {
		n, err := r.Read(buff)
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		var row widget.TextGridRow

		dd := fmt.Sprintf("%03X", rowNo)

		row.Cells = append(row.Cells,
			widget.TextGridCell{
				Rune:  rune(dd[0]),
				Style: &widget.CustomTextGridStyle{},
			}, widget.TextGridCell{
				Rune:  rune(dd[1]),
				Style: &widget.CustomTextGridStyle{},
			},
		)

		if len(dd) == 3 {
			row.Cells = append(row.Cells, widget.TextGridCell{
				Rune:  rune(dd[2]),
				Style: &widget.CustomTextGridStyle{},
			})
		} else {
			row.Cells = append(row.Cells, widget.TextGridCell{
				Style: widget.TextGridStyleWhitespace,
			})
		}

		row.Cells = append(row.Cells, widget.TextGridCell{
			Style: widget.TextGridStyleWhitespace,
		})

		for x, bb := range buff[:n] {
			hexChar := fmt.Sprintf("%02X", bb)

			row.Cells = append(row.Cells,
				widget.TextGridCell{
					Rune: rune(hexChar[0]),
					Style: &widget.CustomTextGridStyle{
						FGColor: viewColor(pos),
					},
				},
				widget.TextGridCell{
					Rune: rune(hexChar[1]),
					Style: &widget.CustomTextGridStyle{
						FGColor: viewColor(pos),
					},
				},
			)

			if x < rowWidth-1 {
				row.Cells = append(row.Cells, widget.TextGridCell{
					Style: widget.TextGridStyleWhitespace,
				})
			}
			pos++
		}

		if n < rowWidth {
			for ex := n; ex < rowWidth; ex++ {
				row.Cells = append(row.Cells,
					widget.TextGridCell{
						Style: widget.TextGridStyleWhitespace,
					},
					widget.TextGridCell{
						Style: widget.TextGridStyleWhitespace,
					},
				)
				if ex < 31 {
					row.Cells = append(row.Cells, widget.TextGridCell{
						Style: widget.TextGridStyleWhitespace,
					})
				}
			}
		}

		row.Cells = append(row.Cells,
			widget.TextGridCell{
				Style: widget.TextGridStyleWhitespace,
			},
			widget.TextGridCell{
				Rune:  rune('║'),
				Style: widget.TextGridStyleWhitespace,
			},
			widget.TextGridCell{
				Style: widget.TextGridStyleWhitespace,
			},
		)

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
		rowNo += 0x20
	}
	return rows
}
