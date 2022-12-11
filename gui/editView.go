package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type editView struct {
	vw *viewerWindow
}

func newEditView(vw *viewerWindow) fyne.CanvasObject {
	ev := &editView{
		vw: vw,
	}
	return ev.layout()
}

func (ev *editView) layout() fyne.CanvasObject {
	return container.NewBorder(ev.vw.toolbar, nil, nil, nil,
		container.NewVBox(
			widget.NewLabel("Here the full eeprom editor will soon come"),
			layout.NewSpacer(),
			widget.NewButtonWithIcon("Close", theme.ContentClearIcon(), func() {
				ev.vw.SetContent(ev.vw.layout())
			}),
		),
	)
}
