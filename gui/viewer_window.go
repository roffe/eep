package gui

import (
	"bytes"
	"fmt"
	"image/color"
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type viewerWindow struct {
	e       *EEPGui
	w       fyne.Window
	toolbar *widget.Toolbar
	grid    *widget.TextGrid
	data    []byte
	saved   bool
}

func newViewerWindow(e *EEPGui, filename string, data []byte, askSaveOnClose bool) *viewerWindow {
	w := e.app.NewWindow("Viewing " + filename)
	vw := &viewerWindow{
		e:    e,
		w:    w,
		data: data,
		grid: widget.NewTextGrid(),
	}
	vw.toolbar = vw.newToolbar()

	w.SetCloseIntercept(func() {
		if askSaveOnClose && !vw.saved {
			dialog.ShowConfirm("Unsaved file", "Save file before closing?", func(b bool) {
				if b {
					e.mw.saveFile(vw.data)
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

	w.SetContent(vw.layout())
	w.Show()
	return vw
}

func (vw *viewerWindow) newToolbar() *widget.Toolbar {
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {
			vw.e.mw.saveFile(vw.data)
			vw.saved = true
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
			for i := range vw.data {
				vw.data[i] ^= 0xff
			}
			for i, r := range generateGrid(vw.data) {
				for j, c := range r.Cells {
					vw.grid.SetCell(i, j, c)
				}
			}
		}),
	)
}

func (vw *viewerWindow) layout() fyne.CanvasObject {
	return container.NewBorder(vw.toolbar, nil, nil, nil,
		container.New(layout.NewMaxLayout(),
			vw.grid,
		),
	)
}

func generateGrid(data []byte) []widget.TextGridRow {
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
