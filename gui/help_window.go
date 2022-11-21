package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func newHelpWindow(app fyne.App) fyne.Window {
	w := app.NewWindow("Help")
	w.Resize(fyne.NewSize(600, 800))
	w.SetCloseIntercept(func() {
		w.Hide()
	})

	generalTab := container.NewTabItemWithIcon("General", theme.QuestionIcon(),
		container.NewHBox(
			widget.NewLabel("General Help"),
			layout.NewSpacer(),
		),
	)

	pinDelayTab := container.NewTabItemWithIcon("Pin Delay", theme.ComputerIcon(),
		container.NewHBox(
			widget.NewLabel("Pin Delay"),
			layout.NewSpacer(),
		),
	)

	tabs := container.NewDocTabs(
		generalTab,
		pinDelayTab,
	)

	tabs.CloseIntercept = func(ti *container.TabItem) {}

	w.SetContent(
		tabs,
	)

	return w
}
