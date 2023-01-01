package gui

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/hirschmann-koxha-gbr/eep/assets"
)

func aboutView(aw fyne.App) fyne.CanvasObject {
	img := &canvas.Image{
		ScaleMode: canvas.ImageScaleFastest,
		FillMode:  canvas.ImageFillOriginal,
		Resource: &fyne.StaticResource{
			StaticName:    "hk.png",
			StaticContent: assets.HkBytes},
	}
	img.SetMinSize(fyne.NewSize(400, 400))

	return container.NewBorder(
		nil,
		widget.NewButton("Visit homepage", func() {
			u, _ := url.Parse("https://hirschmann-koxha.de/en/")
			aw.OpenURL(u)
		}),
		nil,
		nil,
		container.NewCenter(
			img,
			&widget.Label{
				Text:      "Hirschmann & Koxha GbR",
				Alignment: fyne.TextAlignCenter,
			},
		),
	)
}
