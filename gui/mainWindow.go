package gui

import (
	"bufio"
	"fmt"
	"os"
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
	sdialog "github.com/sqweek/dialog"
)

type mainWindow struct {
	e *EEPGui

	hw fyne.Window
	aw fyne.Window

	logList binding.StringList
	log     *widget.List

	rescanButton *widget.Button
	portList     *widget.Select

	viewButton     *widget.Button
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

var mainSize = fyne.NewSize(1200, 600)

func newMainWindow(e *EEPGui) *mainWindow {
	m := &mainWindow{
		Window:      e.NewWindow("Saab CIM Tool " + VERSION),
		e:           e,
		logList:     binding.NewStringList(),
		progressBar: widget.NewProgressBar(),
	}

	m.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("File",
			&fyne.MenuItem{
				Icon:  theme.HelpIcon(),
				Label: "About",
				Action: func() {
					if m.aw != nil {
						m.aw.RequestFocus()
						return
					}
					m.aw = newAboutWindow(e)
					m.aw.Show()
				},
			},
			fyne.NewMenuItemSeparator(),
			&fyne.MenuItem{
				Icon:  theme.CancelIcon(),
				Label: "Quit",
				Action: func() {
					e.Quit()
				},
			},
		),
	))

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

	m.rescanButton = widget.NewButtonWithIcon("Rescan ports", theme.ViewRefreshIcon(), func() { m.portList.Options = m.listPorts() })

	m.portList = &widget.Select{
		PlaceHolder: m.e.port,
		Alignment:   fyne.TextAlignCenter,
		Options:     m.listPorts(),
		OnChanged: func(s string) {
			m.e.port = s
			m.e.Preferences().SetString("port", s)
		},
	}

	m.viewButton = widget.NewButtonWithIcon("View", theme.SearchIcon(), m.viewClickHandler)
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
	split := &container.Split{
		Offset:     0.85,
		Horizontal: true,
		Leading:    container.NewVScroll(m.log),
		Trailing: container.NewVBox(
			m.rescanButton,
			m.portList,
			m.viewButton,
			m.readButton,
			m.writeButton,
			m.eraseButton,
			layout.NewSpacer(),
			m.helpButton,
			m.copyButton,
			m.clearButton,
			m.settingsButton,
		),
	}
	return &container.Split{
		Offset:   1,
		Leading:  split,
		Trailing: m.progressBar,
	}
}

func (m *mainWindow) viewClickHandler() {
	m.viewButton.Disable()
	go func() {
		defer m.viewButton.Enable()
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
					newViewerWindow(m.e, filename, rawbin, false)
				}
			}, m)
			return
		}
		b, err := bin.XORBytes()
		if err != nil {
			dialog.ShowError(err, m)
			return
		}
		newViewerWindow(m.e, filename, b, false)
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
		rawBytes, bin, err := m.readCIM(m.e.port, 1)
		if err != nil {
			m.output(err.Error())
			if err.Error() == "Timeout reading eeprom" {
				return
			}
			if ignoreReadErrors {
				m.saveFile("Save raw bin file", rawBytes)
			} else {
				dialog.ShowConfirm("Error reading CIM", "There was errors reading, view anyway?", func(ok bool) {
					if ok {
						newViewerWindow(m.e, fmt.Sprintf("failed read from %s", time.Now().Format(time.RFC1123Z)), rawBytes, true)
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

		newViewerWindow(m.e, fmt.Sprintf("successful read from %s", time.Now().Format(time.RFC1123Z)), xorBytes, true)
	}()
}

func (m *mainWindow) writeClickHandler() {
	if m.e.port == "" {
		m.output("Please select a port first")
		return
	}

	filename, bin, err := loadFile()
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
				if ok := m.writeCIM(m.e.port, bin); !ok {
					return
				}
				//m.output("Flashed %s, took %s", filename, time.Since(start).String())
				dialog.ShowInformation("Write done", fmt.Sprintf("Flashed %s, took %s", filename, time.Since(start).Round(time.Millisecond).String()), m)
			}()

		}
	}, m)
}

func (m *mainWindow) eraseClickHandler() {
	if m.e.port == "" {
		m.output("Please select a port first")
		return
	}

	dialog.ShowConfirm("Erase CIM?", "Continue erasing CIM?", func(b bool) {
		if b {
			go func() {
				m.disableButtons()
				defer m.enableButtons()
				start := time.Now()
				sr, err := m.openPort(m.e.port)
				if sr != nil {
					defer sr.Close()
				}

				if err != nil {
					m.output("Failed to init adapter: %v", err)
					return
				}

				m.output("Erasing ... ")
				if err := m.erase(sr); err != nil {
					m.output(err.Error())
				}

				m.output("Erase took %s", time.Since(start).String())
			}()

		}
	}, m)
}

func (m *mainWindow) saveFile(title string, data []byte) bool {
	filename, err := sdialog.File().Filter("Bin file", "bin").Title(title).Save()
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
	//m.viewButton.Disable()
	m.readButton.Disable()
	m.writeButton.Disable()
	m.eraseButton.Disable()
}

func (m *mainWindow) enableButtons() {
	m.rescanButton.Enable()
	m.readButton.Enable()
	m.portList.Enable()
	//m.viewButton.Enable()
	m.readButton.Enable()
	m.writeButton.Enable()
	m.eraseButton.Enable()
}
