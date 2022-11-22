package gui

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func newViewerWindow(app fyne.App, filename string, data []byte) fyne.Window {
	w := app.NewWindow("Viewing " + filename)
	w.Resize(fyne.NewSize(256, 256))
	w.SetFixedSize(true)
	w.CenterOnScreen()
	grid := widget.NewTextGrid()

	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentCreateIcon(), func() {}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.ContentCutIcon(), func() {}),
		widget.NewToolbarAction(theme.ContentCopyIcon(), func() {}),
		widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
			for i := range data {
				data[i] ^= 0xff
			}
			for i, r := range foo(data) {
				for j, c := range r.Cells {
					grid.SetCell(i, j, c)
				}
			}
		}),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.HelpIcon(), func() {}),
	)

	b, err := os.ReadFile("test.bin")
	if err != nil {
		panic(err)
	}

	rows := foo(b)
	for i, row := range rows {
		grid.SetRow(i, row)
	}

	content := container.NewBorder(toolbar, nil, nil, nil, container.New(layout.NewMaxLayout(),
		grid,
	))

	w.SetContent(
		content,
	)
	w.Show()
	return w
}

func foo(data []byte) []widget.TextGridRow {
	var rows []widget.TextGridRow
	r := bytes.NewReader(data)
	buff := make([]byte, 32)
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
			row.Cells = append(row.Cells, widget.TextGridCell{
				Rune: rune(asd[0]),
				Style: &widget.CustomTextGridStyle{
					FGColor: color.RGBA{
						R: 60, G: 128, B: 0, A: 1,
					},
				},
			})
			row.Cells = append(row.Cells, widget.TextGridCell{
				Rune: rune(asd[1]),
				Style: &widget.CustomTextGridStyle{
					FGColor: color.RGBA{
						R: 60, G: 128, B: 0, A: 1,
					},
				},
			})
			if x < 31 {
				row.Cells = append(row.Cells, widget.TextGridCell{
					Style: widget.TextGridStyleWhitespace,
				})
			}
		}
		row.Cells = append(row.Cells, widget.TextGridCell{
			Rune:  rune('║'),
			Style: widget.TextGridStyleWhitespace,
		})

		for _, bb := range buff[:n] {
			row.Cells = append(row.Cells, widget.TextGridCell{
				Rune: rune(bb),
				Style: &widget.CustomTextGridStyle{
					FGColor: color.RGBA{
						R: 60, G: 128, B: 128, A: 1,
					},
				},
			})
		}

		rows = append(rows, row)
	}
	return rows
}
