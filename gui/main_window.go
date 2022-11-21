package gui

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	sdialog "github.com/sqweek/dialog"
	"go.bug.st/serial/enumerator"
)

type mainWindow struct {
	w        fyne.Window
	app      fyne.App
	settings fyne.Window
	help     fyne.Window

	log *widget.List

	rescanButton *widget.Button
	portList     *widget.Select

	viewButton  *widget.Button
	readButton  *widget.Button
	writeButton *widget.Button
	eraseButton *widget.Button

	progressBar *widget.ProgressBar

	once sync.Once
}

func newMainWindow(app fyne.App) fyne.Window {
	window := app.NewWindow("Saab CIM Cloner by Hirschmann-Koxha GbR")
	//window.SetFixedSize(true)
	window.Resize(fyne.NewSize(900, 600))
	window.CenterOnScreen()
	mw := &mainWindow{
		w:           window,
		app:         app,
		log:         createLogList(),
		help:        newHelpWindow(app),
		settings:    newSettingsWindow(app),
		progressBar: widget.NewProgressBar(),
	}

	window.SetContent(mw.layout())
	window.SetMaster()
	return window
}

func (m *mainWindow) layout() *container.Split {
	m.once.Do(func() {
		m.init()
	})
	left := container.New(layout.NewMaxLayout(), m.log)

	right := container.NewVBox(
		m.rescanButton,
		m.portList,
		m.viewButton,
		m.readButton,
		m.writeButton,
		m.eraseButton,
		layout.NewSpacer(),
		widget.NewButtonWithIcon("Help", theme.HelpIcon(), func() {
			m.help.Show()
			m.help.RequestFocus()
		}),
		widget.NewButtonWithIcon("Copy log", theme.ContentCopyIcon(), func() {
			if content, err := listData.Get(); err == nil {
				m.w.Clipboard().SetContent(strings.Join(content, "\n"))
			}
		}),
		widget.NewButtonWithIcon("Clear log", theme.ContentClearIcon(), func() {
			listData.Set([]string{})
		}),
		widget.NewButtonWithIcon("Settings", theme.SettingsIcon(), func() {
			m.settings.Show()
			m.settings.RequestFocus()
		}),
	)

	split := container.NewHSplit(left, right)
	split.Offset = 0.99
	view := container.NewVSplit(split, m.progressBar)
	view.Offset = 1
	return view
}

func (m *mainWindow) init() {
	m.rescanButton = widget.NewButtonWithIcon("Rescan ports", theme.ViewRefreshIcon(), func() {
		m.portList.Options = m.listPorts()
	})

	m.portList = widget.NewSelect(m.listPorts(), func(s string) {
		state.port = s
		m.app.Preferences().SetString("port", s)
	})
	m.portList.Alignment = fyne.TextAlignCenter

	if state.port != "" {
		m.portList.PlaceHolder = state.port
	}
	m.viewButton = widget.NewButtonWithIcon("Load", theme.DocumentIcon(), func() {
		newViewerWindow(m.app, []byte{})
	})

	m.readButton = widget.NewButtonWithIcon("Read", theme.DownloadIcon(), m.readClickHandler)
	m.writeButton = widget.NewButtonWithIcon("Write", theme.UploadIcon(), m.writeClickHandler)
	m.eraseButton = widget.NewButtonWithIcon("Erase", theme.DeleteIcon(), m.eraseClickHandler)

}

func (m *mainWindow) eraseClickHandler() {
	go m.eraseClick()
}

func (m *mainWindow) eraseClick() {
	if state.port == "" {
		m.output("Please select a port first")
		return
	}
	m.disableButtons()
	defer m.enableButtons()

	ok := sdialog.Message("%s", "Do you want to erase CIM?").Title("Are you sure?").YesNo()
	if !ok {
		return
	}
	start := time.Now()
	sr, err := m.openPort(state.port)
	if sr != nil {
		defer sr.Close()
	}

	if err != nil {
		m.output(err.Error())
		return
	}

	m.output("Erasing ... ")
	if err := m.erase(sr); err != nil {
		m.output(err.Error())
	}

	m.output("Erase took %s", time.Since(start).String())
}

func (m *mainWindow) writeClickHandler() {
	go m.writeClick()
}

func (m *mainWindow) writeClick() {
	if state.port == "" {
		m.output("Please select a port first")
		return
	}
	m.disableButtons()
	defer m.enableButtons()
	filename, err := sdialog.File().Filter("Bin file", "bin").Title("Load bin file").Load()
	if err != nil {
		if err.Error() == "Cancelled" {
			return
		}
		m.output(err.Error())
		return
	}

	bin, err := os.ReadFile(filename)
	if err != nil {
		m.output(err.Error())
		return
	}
	ok := sdialog.Message("%s", "Do you want to continue?").Title("Are you sure?").YesNo()
	if !ok {
		return
	}

	if err := m.writeCIM(state.port, bin); err != nil {
		m.output(err.Error())
		return
	}

}

func addSuffix(s, suffix string) string {
	if !strings.HasSuffix(s, suffix) {
		return s + suffix
	}
	return s
}

func (m *mainWindow) readClickHandler() {
	filename, err := sdialog.File().Filter("Bin file", "bin").Title("Save bin file").Save()
	if err != nil {
		if err.Error() == "Cancelled" {
			return
		}
		m.output(err.Error())
		return
	}
	filename = addSuffix(filename, ".bin")
	go m.readClick(filename)
}
func (m *mainWindow) readClick(filename string) {
	m.disableButtons()
	defer m.enableButtons()
	if state.port == "" {
		m.output("Please select a port first")
		return
	}
	//m.output("Reading CIM ...")
	bin, err := m.readCIM(state.port, 5)
	if err != nil {
		m.output(err.Error())
		return
	}

	m.printKV("MD5", bin.MD5())
	m.printKV("CRC32", bin.CRC32())
	m.printKV("VIN", bin.Vin.Data)
	m.printKV("MY", bin.Vin.Data[9:10])
	m.printKV("End model (HW+SW)", fmt.Sprintf("%d%s", bin.PartNo1, bin.PartNo1Rev))
	m.printKV("Base model (HW+boot)", fmt.Sprintf("%d%s", bin.PnBase1, bin.PnBase1Rev))
	m.printKV("Delphi part number", fmt.Sprintf("%d", bin.DelphiPN))
	m.printKV("SAAB part number", fmt.Sprintf("%d", bin.PartNo))
	m.printKV("Configuration Version:", fmt.Sprintf("%d", bin.ConfigurationVersion))

	b, err := bin.XORBytes()
	if err != nil {
		m.output(err.Error())
		return
	}

	if err := os.WriteFile(filename, b, 0644); err == nil {
		m.output("Saved as " + filename)
	} else {
		m.output(err.Error())
	}

}

func (m *mainWindow) printKV(k, v string) {
	m.output(k + ": " + v)
}

func (m *mainWindow) listPorts() []string {
	var portsList []string
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		m.output(err.Error())
		return []string{}
	}
	if len(ports) == 0 {
		m.output("No serial ports found!")
		return []string{}
	}
	for _, port := range ports {
		m.output(fmt.Sprintf("Found port: %s\n", port.Name))
		if port.IsUSB {
			m.append("\t\t\t\t├ USB ID:     %s:%s\n", port.VID, port.PID)
			m.append("\t\t\t\t└ USB serial: %s\n", port.SerialNumber)
			m.output("")
			portsList = append(portsList, port.Name)
		}
	}
	state.portList = portsList
	return portsList
}

func (m *mainWindow) output(format string, values ...interface{}) int {
	var text string
	if format != "" {
		text = fmt.Sprintf("%s - %s", time.Now().Format("15:04:05.000"), fmt.Sprintf(format, values...))
	}
	listData.Append(text)
	m.log.Refresh()
	m.log.ScrollToBottom()
	return listData.Length()
}

func (m *mainWindow) append(format string, values ...interface{}) {
	di, err := listData.GetValue(listData.Length() - 1)
	if err != nil {
		panic(err)
	}
	listData.SetValue(listData.Length()-1, di+fmt.Sprintf(format, values...))
	m.log.Refresh()
}

func (m *mainWindow) appendPos(index int, format string, values ...interface{}) {
	di, err := listData.GetValue(listData.Length() - 1)
	if err != nil {
		panic(err)
	}
	listData.SetValue(index, di+fmt.Sprintf(format, values...))
	m.log.Refresh()
}

func (m *mainWindow) disableButtons() {
	m.rescanButton.Disable()
	m.portList.Disable()
	m.viewButton.Disable()
	m.readButton.Disable()
	m.writeButton.Disable()
	m.eraseButton.Disable()
}

func (m *mainWindow) enableButtons() {
	m.rescanButton.Enable()
	m.readButton.Enable()
	m.portList.Enable()
	m.viewButton.Enable()
	m.readButton.Enable()
	m.writeButton.Enable()
	m.eraseButton.Enable()
}
