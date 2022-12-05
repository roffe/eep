package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func newHexView(vw *viewerWindow) fyne.CanvasObject {
	grid := widget.NewTextGrid()
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.HomeIcon(), func() {
			vw.w.SetContent(vw.layout())
			vw.w.Resize(viewWindowSize)
		}),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), vw.Save),
		widget.NewToolbarAction(theme.ContentClearIcon(), func() {
			for i := range vw.data {
				vw.data[i] ^= 0xff
			}

			for i, r := range generateGrid(vw.data) {
				for j, c := range r.Cells {
					grid.SetCell(i, j, c)
				}
			}
		}),

		widget.NewToolbarSpacer(),
	)
	for i, row := range generateGrid(vw.data) {
		grid.SetRow(i, row)
	}
	cc := container.NewBorder(toolbar, nil, nil, nil, grid)
	return cc
}
