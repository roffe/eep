package gui

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/hirschmann-koxha-gbr/eep/assets"
)

type aboutWindow struct {
	e *EEPGui
	fyne.Window
}

func newAboutWindow(e *EEPGui) fyne.Window {
	w := &aboutWindow{
		e:      e,
		Window: e.NewWindow("About"),
	}
	w.SetOnClosed(func() {
		e.mw.aw = nil
	})
	w.SetContent(w.layout())
	w.Show()
	return w
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
		nil,
		&widget.Label{
			Text:      "Hirschmann & Koxha GbR",
			Alignment: fyne.TextAlignCenter,
		},
		nil,
		nil,
		container.NewPadded(img, widget.NewButton("", func() {
			u, _ := url.Parse("https://hirschmann-koxha.de/en/")
			aw.e.OpenURL(u)
		})),
		//widget.NewButtonWithIcon("", fyne.NewStaticResource("hk.png", assets.HkBytes), func() {
		//	log.Println("fisring")
		//}),
		//container.NewCenter(container.NewMax(&canvas.Image{
		//	ScaleMode: canvas.ImageScaleFastest,
		//	FillMode:  canvas.ImageFillOriginal,
		//	Resource: &fyne.StaticResource{
		//		StaticName:    "hk.png",
		//		StaticContent: assets.HkBytes},
		//})),
	)
}
