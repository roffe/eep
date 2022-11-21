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
	port            string
	portList        []string
	readDelayValue  binding.Float
	writeDelayValue binding.Float
	ignoreError     binding.Bool
}

var state = &appState{
	readDelayValue:  binding.NewFloat(),
	writeDelayValue: binding.NewFloat(),
	ignoreError:     binding.NewBool(),
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

	r := app.Preferences().Float("read_pin_delay")
	if r < 20 {
		r = 100
	}
	state.readDelayValue.Set(r)

	w := app.Preferences().Float("write_pin_delay")
	if r < 100 {
		r = 200
	}
	state.writeDelayValue.Set(w)
	state.ignoreError.Set(app.Preferences().Bool("ignore_read_errors"))
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
