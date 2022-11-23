package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/Hirschmann-Koxha-GbR/eep/assets"
)

type helpWindow struct {
	e *EEPGui
	w fyne.Window

	tabs *container.DocTabs
}

func newHelpWindow(e *EEPGui) *helpWindow {
	w := e.app.NewWindow("Help")
	//w.Resize(fyne.NewSize(1000, 850))
	w.SetCloseIntercept(func() {
		w.Hide()
	})
	hw := &helpWindow{e: e, w: w}
	hw.w.Resize(fyne.NewSize(1000, 850))

	introTab := container.NewTabItemWithIcon("Intro", theme.QuestionIcon(),
		container.NewVScroll(
			container.NewVBox(
				widget.NewLabelWithStyle("To access the storage of your CIM, you need to connect your SOP8 clip to the EEPROM", fyne.TextAlignCenter, fyne.TextStyle{}),
				container.NewCenter(assets.PCB),
				widget.NewSeparator(),
				widget.NewLabelWithStyle("It is necessary to remove any conformal coating from the legs of the EEPROM. This can be achieved by the use of a sharp knife or razor blade", fyne.TextAlignCenter, fyne.TextStyle{}),
				widget.NewLabelWithStyle("then clean with IPA / acetone using cotton swabs", fyne.TextAlignCenter, fyne.TextStyle{}),
				widget.NewLabel(""),
				widget.NewLabelWithStyle("Be careful, while attempting to do this. Excessive force could break the legs and therefore, would require a new EEPROM to be soldered in", fyne.TextAlignCenter, fyne.TextStyle{}),
				widget.NewLabelWithStyle("Before you put the SOP8 clip on you need to make sure that the wire is connected in the right orientation.\nMake sure that the red wire is in the corner of the EEPROM with the indentation shown below.", fyne.TextAlignCenter, fyne.TextStyle{}),
				widget.NewLabel(""),
				container.NewCenter(assets.EEPROM),
				widget.NewLabel(""),
				widget.NewLabelWithStyle("The result should look like this", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
				container.NewCenter(assets.OVERVIEW),
			),
		),
	)

	failedTab := container.NewTabItemWithIcon("Failed reads", theme.ErrorIcon(),
		container.NewVBox(
			widget.NewLabelWithStyle("Failed reads could be the result of:", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabel(""),
			widget.NewLabel("* Conformal coating still on the EEPROM"),
			widget.NewLabel("* SOP8 clip not seated properly"),
			widget.NewLabel("* Too low pin delays"),
			widget.NewLabel("* Defective/corrupted EEPROM"),
		))

	settingsTab := container.NewTabItemWithIcon("Settings", theme.ComputerIcon(),
		container.NewVBox(
			widget.NewLabelWithStyle("Ignore read validation errors", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel("Enable to allow saving corupt eeproms that fails CRC checks on read"),
			widget.NewLabel(""),
			widget.NewLabelWithStyle("Read/Write pin delay", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewRichTextWithText("In case you encounter failed reads or flashes, you might want to try adjusting the delays.\nThe higher the delay, the slower, yet more stable the read out and flash will be.\nDepending on the age and wear of the EEPROM you might have to adjust these values"),
		),
	)

	hw.tabs = container.NewDocTabs(
		introTab,
		failedTab,
		settingsTab,
	)
	hw.tabs.CloseIntercept = func(ti *container.TabItem) {}

	hw.w.SetContent(hw.layout())

	return hw
}

func (hw *helpWindow) layout() fyne.CanvasObject {
	header := widget.NewLabelWithStyle("Welcome to the CIM Tool!", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	footer := container.NewVBox()
	return container.New(
		layout.NewBorderLayout(header, footer, nil, nil),
		header,
		footer,
		hw.tabs,
	)
}
