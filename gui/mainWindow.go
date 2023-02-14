package gui

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/hirschmann-koxha-gbr/cim/pkg/cim"
	"github.com/hirschmann-koxha-gbr/eep/adapter"
	sdialog "github.com/sqweek/dialog"
)

type mainWindow struct {
	e *EEPGui

	hw      fyne.Window
	appTabs *container.AppTabs
	docTab  *container.DocTabs

	logList binding.StringList
	log     *widget.List

	rescanButton *widget.Button
	portList     *widget.Select

	openButton     *widget.Button
	readButton     *widget.Button
	writeButton    *widget.Button
	eraseButton    *widget.Button
	helpButton     *widget.Button
	copyButton     *widget.Button
	clearButton    *widget.Button
	settingsButton *widget.Button

	progressBar *widget.ProgressBar

	fyne.Window
}

var mainSize = fyne.NewSize(1280, 700)

func newMainWindow(e *EEPGui) *mainWindow {
	m := &mainWindow{
		Window:      e.NewWindow("Saab CIM Tool " + VERSION),
		e:           e,
		logList:     binding.NewStringList(),
		progressBar: widget.NewProgressBar(),
	}

	m.docTab = container.NewDocTabs()
	m.docTab.CloseIntercept = func(i *container.TabItem) {
		if i.Text == "Start" {
			return
		}
		dialog.ShowConfirm("Close", "Are you sure you want to close this tab?", func(b bool) {
			if b {
				m.docTab.Remove(i)
			}
		}, m.Window)
	}

	m.docTab.Append(
		container.NewTabItemWithIcon("Start", theme.DocumentIcon(), container.NewCenter(
			widget.NewLabelWithStyle("Welcome to the Saab CIM Tool", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		)),
	)
	m.log = widget.NewListWithData(
		m.logList,
		func() fyne.CanvasObject {
			return &widget.Label{
				TextStyle: fyne.TextStyle{Monospace: true},
			}
		},
		func(item binding.DataItem, obj fyne.CanvasObject) {
			i := item.(binding.String)
			txt, err := i.Get()
			if err != nil {
				panic(err)
			}
			if v, ok := obj.(*widget.Label); ok {
				v.SetText(txt)
			}
		},
	)

	message, ports, err := adapter.ListPorts()
	if err != nil {
		m.output(err.Error())
	}
	if message != "" {
		m.output(message)
	}

	m.rescanButton = widget.NewButtonWithIcon("Refresh ports", theme.ViewRefreshIcon(), func() {
		message, ports, err := adapter.ListPorts()
		if err != nil {
			m.output(err.Error())
			return
		}
		m.portList.Options = ports
		m.output(message)
	})

	m.portList = &widget.Select{
		PlaceHolder: m.e.port,
		Alignment:   fyne.TextAlignCenter,
		Options:     ports,
		OnChanged: func(s string) {
			m.e.port = s
			m.e.Preferences().SetString("port", s)
		},
	}

	m.openButton = widget.NewButtonWithIcon("Open", theme.FolderOpenIcon(), m.viewClickHandler)
	m.readButton = widget.NewButtonWithIcon("Read", theme.DownloadIcon(), m.readClickHandler)
	m.writeButton = widget.NewButtonWithIcon("Write", theme.UploadIcon(), m.writeClickHandler)
	m.eraseButton = widget.NewButtonWithIcon("Erase", theme.DeleteIcon(), m.eraseClickHandler)
	m.helpButton = widget.NewButtonWithIcon("Help", theme.HelpIcon(), func() {
		if m.hw == nil {
			m.hw = newHelpWindow(e)
		} else {
			m.hw.RequestFocus()
		}
	})
	m.copyButton = widget.NewButtonWithIcon("Copy log", theme.ContentCopyIcon(), func() {
		if content, err := m.logList.Get(); err == nil {
			m.Clipboard().SetContent(strings.Join(content, "\n"))
		}
	})
	m.clearButton = widget.NewButtonWithIcon("Clear log", theme.ContentClearIcon(), func() {
		m.logList.Set([]string{})
	})
	m.settingsButton = widget.NewButtonWithIcon("Settings", theme.SettingsIcon(), func() {
		if m.e.sw == nil {
			m.e.sw = newSettingsWindow(m.e)
		} else {
			m.e.sw.RequestFocus()
		}
	})

	m.SetContent(m.layout())
	m.Resize(mainSize)
	m.SetMaster()
	m.Show()
	return m
}

func (m *mainWindow) layout() fyne.CanvasObject {

	m.appTabs = container.NewAppTabs(
		container.NewTabItemWithIcon("Home", theme.HomeIcon(),
			m.docTab,
		),
		container.NewTabItemWithIcon("Log", theme.DocumentIcon(), m.log),
		//container.NewTabItemWithIcon("Help", theme.HelpIcon(), newHelpView(m.e)),
		container.NewTabItemWithIcon("About", theme.InfoIcon(), aboutView(m.e)),
		container.NewTabItemWithIcon("Settings", theme.SettingsIcon(), newSettingsView(m.e)),
	)

	split := &container.Split{
		Offset:     0.90,
		Horizontal: true,
		Leading:    m.appTabs,
		Trailing: container.NewVBox(
			m.rescanButton,
			m.portList,
			m.openButton,
			m.readButton,
			m.writeButton,
			m.eraseButton,
			/*
				widget.NewButtonWithIcon("Read MIU", theme.DownloadIcon(), func() {
					b, err := m.readMIU()
					if err != nil {
						dialog.ShowError(err, m.Window)
						return
					}

					log.Printf("%X", md5.Sum(b))
					if err := os.WriteFile(time.Now().Format("15_04_05")+".bin", b, 0644); err != nil {
						dialog.ShowError(err, m.Window)
					}
				}),
			*/
			layout.NewSpacer(),
			m.helpButton,
			m.copyButton,
			m.clearButton,
		),
	}

	return container.NewBorder(
		nil,
		m.progressBar,
		nil,
		nil,
		split,
	)
}

func (m *mainWindow) viewClickHandler() {
	m.openButton.Disable()
	go func() {
		defer m.openButton.Enable()
		filename, err := sdialog.File().Filter("Bin file", "bin").Title("Select file to view").Load()
		if err != nil {
			if err.Error() == "Cancelled" {
				return
			}
			m.output(err.Error())
			return
		}

		bin, err := cim.MustLoad(filename)
		if err != nil {
			dialog.ShowConfirm("File verification failed", fmt.Sprintf("File verification failed: %v. View anyway?", err), func(ok bool) {
				if ok {
					rawbin, err := os.ReadFile(filename)
					if err != nil {
						dialog.ShowError(err, m)
						return
					}
					m.docTab.Append(container.NewTabItemWithIcon(filepath.Base(filename), theme.FileIcon(), newViewerView(m.e, filename, rawbin, false)))
					m.appTabs.SelectIndex(0)
					m.docTab.SelectIndex(len(m.docTab.Items) - 1)
				}
			}, m)
			return
		}
		b, err := bin.XORBytes()
		if err != nil {
			dialog.ShowError(err, m)
			return
		}
		d := container.NewTabItemWithIcon(filepath.Base(filename), theme.FileIcon(), newViewerView(m.e, filename, b, false))
		m.docTab.Append(d)
		m.appTabs.SelectIndex(0)
		m.docTab.SelectIndex(len(m.docTab.Items) - 1)
	}()
}

func (m *mainWindow) readClickHandler() {
	m.disableButtons()
	go func() {
		defer m.enableButtons()
		if m.e.port == "" {
			m.output("Please select a port first")
			return
		}
		ignoreReadErrors, _ := m.e.ignoreError.Get()
		rawBytes, bin, err := m.readCIM()
		if err != nil {
			m.output(err.Error())
			if err.Error() == "Timeout reading eeprom" {
				return
			}
			if ignoreReadErrors {
				m.saveFile("Save raw bin file", fmt.Sprintf("cim_raw_%s.bin", time.Now().Format("20060102-150405")), rawBytes)
			} else {
				m.appTabs.SelectIndex(1)
				dialog.ShowConfirm("Error reading CIM", "There was errors reading, view anyway?", func(ok bool) {
					if ok {
						m.docTab.Append(container.NewTabItemWithIcon(fmt.Sprintf("Raw read at %s", time.Now().Format("15:04:05")), theme.FileIcon(), newViewerView(m.e, fmt.Sprintf("failed read from %s", time.Now().Format(time.RFC1123Z)), rawBytes, true)))
						m.appTabs.SelectIndex(0)
						m.docTab.SelectIndex(len(m.docTab.Items) - 1)
					}
				}, m)
			}
			return
		}
		xorBytes, err := bin.XORBytes()
		if err != nil {
			dialog.ShowError(err, m)
			return
		}
		m.docTab.Append(container.NewTabItemWithIcon(fmt.Sprintf("Read at %s", time.Now().Format("15:04:05")), theme.FileIcon(), newViewerView(m.e, fmt.Sprintf("successful read from %s", time.Now().Format(time.RFC1123Z)), xorBytes, true)))
		m.appTabs.SelectIndex(0)
		m.docTab.SelectIndex(len(m.docTab.Items) - 1)
	}()
}

func (m *mainWindow) writeClickHandler() {
	if m.e.port == "" {
		m.output("Please select a port first")
		return
	}

	_, bin, err := loadFile()
	if err != nil {
		if err.Error() == "Cancelled" {
			return
		}
		m.output(err.Error())
		return
	}

	dialog.ShowConfirm("Write to CIM?", "Continue writing to CIM?", func(ok bool) {
		if ok {
			start := time.Now()
			go func() {
				m.disableButtons()
				defer m.enableButtons()
				if err := m.writeCIM(m.e.port, bin); err != nil {
					dialog.ShowError(err, m)
					m.appTabs.SelectIndex(1)
					return
				}
				dialog.ShowInformation("Write done", fmt.Sprintf("Write successfull, took %s", time.Since(start).Round(time.Millisecond).String()), m)
			}()

		}
	}, m)
}

func (m *mainWindow) eraseClickHandler() {
	if m.e.port == "" {
		m.output("Please select a port first")
		return
	}
	dialog.ShowConfirm("Erase CIM?", "Continue erasing CIM?", func(ok bool) {
		if ok {
			go func() {
				m.disableButtons()
				defer m.enableButtons()

				start := time.Now()

				client := m.newAdapter()
				if err := client.Open(m.e.port, VERSION); err != nil {
					m.output("Failed to init adapter: %v", err)
					return
				}
				defer client.Close()

				m.output("Erasing ... ")
				if err := client.EraseCIM(); err != nil {
					m.output(err.Error())
					return
				}
				m.output("Erase took %s", time.Since(start).String())
				dialog.ShowInformation("Erase done", fmt.Sprintf("Erase successfull, took %s", time.Since(start).Round(time.Millisecond).String()), m)
			}()

		}
	}, m)
}

func (m *mainWindow) saveFile(title, suggestedFilename string, data []byte) bool {
	filename, err := sdialog.File().Filter("Bin file", "bin").SetStartFile(suggestedFilename).Title(title).Save()
	if err != nil {
		if err.Error() == "Cancelled" {
			return false
		}
		m.output(err.Error())
		return false
	}
	filename = addSuffix(filename, ".bin")

	if err := os.WriteFile(filename, data, 0644); err == nil {
		m.output("Saved to %s", filename)
	} else {
		m.output(err.Error())
		return false
	}
	return true
}

func loadFile() (string, []byte, error) {
	filename, err := sdialog.File().Filter("Bin file", "bin").Title("Load bin file").Load()
	if err != nil {
		return "", nil, err
	}

	bin, err := os.ReadFile(filename)
	if err != nil {
		return "", nil, err
	}
	return filename, bin, nil
}

func addSuffix(s, suffix string) string {
	if !strings.HasSuffix(s, suffix) {
		return s + suffix
	}
	return s
}

func (m *mainWindow) output(format string, values ...interface{}) {
	scanner := bufio.NewScanner(strings.NewReader(fmt.Sprintf(format, values...)))
	for scanner.Scan() {
		var text string
		if format != "" {
			text = fmt.Sprintf("%s - %s", time.Now().Format("15:04:05.000"), scanner.Text())
		}
		m.logList.Append(text)
		m.log.ScrollToBottom()
	}
}

func (m *mainWindow) disableButtons() {
	m.rescanButton.Disable()
	m.portList.Disable()
	//m.openButton.Disable()
	m.readButton.Disable()
	m.writeButton.Disable()
	m.eraseButton.Disable()
}

func (m *mainWindow) enableButtons() {
	m.rescanButton.Enable()
	m.readButton.Enable()
	m.portList.Enable()
	//m.openButton.Enable()
	m.readButton.Enable()
	m.writeButton.Enable()
	m.eraseButton.Enable()
}
