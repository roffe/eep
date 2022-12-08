package gui

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type EditWindow struct {
	e *EEPGui
	w fyne.Window
	//bin *cim.Bin
}

func NewEditWindow(e *EEPGui) {
	w := e.NewWindow("Editor")
	w.Resize(fyne.NewSize(450, 600))

	ew := &EditWindow{e: e, w: w}

	w.SetContent(ew.layout())
	w.Show()
}

func (ew *EditWindow) layout() fyne.CanvasObject {

	general := container.NewVBox()

	pins := []fyne.CanvasObject{
		widget.NewLabel("Keys"),
	}
	for i := 0; i < 4; i++ {
		e := &widget.Entry{
			Text: "00000000",
		}
		pins = append(pins, e)
	}

	sync := []fyne.CanvasObject{
		widget.NewLabel("Sync"),
	}
	for i := 0; i < 4; i++ {
		e := &widget.Entry{
			Text: "00000000",
		}
		sync = append(sync, e)
	}

	split := container.NewHSplit(
		container.NewVBox(
			pins...,
		),
		container.NewVBox(
			sync...,
		),
	)

	keyErrors := widget.NewEntry()
	keyErrors.SetText(strconv.Itoa(13))
	keyErrors.SetMinRowsVisible(8)

	keys := container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Key errors:"),
			keyErrors,
		),
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
