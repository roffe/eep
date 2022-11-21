package gui

import (
	_ "embed"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

type appState struct {
	port     string
	portList []string
	delay    uint8
}

var state = &appState{
	delay: 0,
}

//go:embed Icon.png
var icon []byte
var appIcon = fyne.NewStaticResource("icon", icon)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func Run() {
	app := newApp()
	newMainWindow(app).Show()
	app.Run()
}

func newApp() fyne.App {
	app := app.NewWithID("com.cimtool")
	app.SetIcon(appIcon)
	app.Settings().SetTheme(&gocanTheme{})

	state.port = app.Preferences().String("port")
	return app
}

var listData = binding.NewStringList()

func createLogList() *widget.List {
	return widget.NewListWithData(
		listData,
		func() fyne.CanvasObject {
			w := widget.NewLabel("")
			w.TextStyle.Monospace = true
			return w
		},
		func(item binding.DataItem, obj fyne.CanvasObject) {
			i := item.(binding.String)
			txt, err := i.Get()
			if err != nil {
				panic(err)
			}
			obj.(*widget.Label).SetText(txt)
		},
	)
}
