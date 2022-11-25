package gui

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/Hirschmann-Koxha-GbR/cim/pkg/cim"
)

type editWindow struct {
	e   *EEPGui
	w   fyne.Window
	bin *cim.Bin
}

func newEditWindow(e *EEPGui) {
	w := e.app.NewWindow("Editor")
	w.Resize(fyne.NewSize(450, 600))

	ew := &editWindow{e: e, w: w}

	w.SetContent(ew.layout())
	w.Show()
}

func (ew *editWindow) layout() fyne.CanvasObject {

	general := container.NewVBox()

	pins := []fyne.CanvasObject{
		widget.NewLabel("Keys"),
	}
	for i := 0; i < 4; i++ {
		e := widget.NewEntry()
		e.SetText("00000000")
		x := container.NewHSplit(
			widget.NewLabel("#"+strconv.Itoa(i)),
			e,
		)
		x.Offset = 0.1
		pins = append(pins, x)
	}

	sync := []fyne.CanvasObject{
		widget.NewLabel("Sync"),
	}
	for i := 0; i < 4; i++ {
		e := widget.NewEntry()
		e.SetText("00000000")
		x := container.NewHSplit(
			widget.NewLabel("#"+strconv.Itoa(i)),
			e,
		)
		x.Offset = 0.1
		sync = append(sync, x)
	}

	split := container.NewHSplit(
		container.NewVBox(
			pins...,
		),
		container.NewVBox(
			sync...,
		),
	)

	keys := container.NewMax(
		widget.NewLabel("Key errors"),
		split,
	)

	return container.NewAppTabs(
		container.NewTabItemWithIcon("General", theme.InfoIcon(), container.NewVScroll(
			general,
		)),
		container.NewTabItemWithIcon("Keys", theme.LoginIcon(), container.NewVScroll(
			keys,
		)),
	)
}
