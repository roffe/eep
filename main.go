package main

import (
	_ "embed"
	"log"
	"os"
	"os/signal"
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/hirschmann-koxha-gbr/eep/gui"
)

//go:embed Icon.png
var icon []byte
var appIcon = fyne.NewStaticResource("icon", icon)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main() {
	quitChan := make(chan os.Signal, 1)
	signal.Notify(quitChan, syscall.SIGTERM, syscall.SIGINT)
	a := app.NewWithID("com.cimtool")
	a.SetIcon(appIcon)
	a.Settings().SetTheme(&gui.Theme{})
	app := gui.New(a)
	a.Lifecycle().SetOnStarted(app.CheckUpdate)
	go quitHandler(quitChan, a)
	app.Run()
}

func quitHandler(quitChan chan os.Signal, app fyne.App) {
	<-quitChan
	app.Quit()
}
