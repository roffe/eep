package gui

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/hirschmann-koxha-gbr/eep/assets"
)

type aboutWindow struct {
	fyne.App
}

func (e *EEPGui) showAboutDialog() {
	aw := &aboutWindow{App: e.App}
	dialog.ShowCustom("About", "Close", aw.layout(), e.mw)
}

func (aw *aboutWindow) layout() fyne.CanvasObject {

	img := &canvas.Image{
		ScaleMode: canvas.ImageScaleFastest,
		FillMode:  canvas.ImageFillOriginal,
		Resource: &fyne.StaticResource{
			StaticName:    "hk.png",
			StaticContent: assets.HkBytes},
	}
	img.SetMinSize(fyne.NewSize(400, 400))

	return container.NewBorder(
		img,
		widget.NewButton("Visit homepage", func() {
			u, _ := url.Parse("https://hirschmann-koxha.de/en/")
			aw.OpenURL(u)
		}),
		nil,
		nil,
		&widget.Label{
			Text:      "Hirschmann & Koxha GbR",
			Alignment: fyne.TextAlignCenter,
		},
		&widget.Label{Text: ""},
	)
}
