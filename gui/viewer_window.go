package gui

import "fyne.io/fyne/v2"

func newViewerWindow(app fyne.App, data []byte) fyne.Window {
	w := app.NewWindow("Viewer")
	w.Resize(fyne.NewSize(512, 512))
	w.Show()
	return w
}
