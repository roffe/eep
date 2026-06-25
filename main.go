package main

import (
	_ "embed"
	"log"

	"fyne.io/fyne/v2/app"
	"github.com/roffe/eep/gui"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main() {
	application := app.NewWithID("com.roffe.cimtool")

	//application.Settings().SetTheme(&gui.Theme{})
	application.Settings().SetTheme(&gui.MyTheme{})
	ui, err := gui.New(application)
	if err != nil {
		log.Fatal(err)
	}
	//application.Lifecycle().SetOnStarted(ui.CheckUpdate)
	//application.Run()
	ui.Run()
}
