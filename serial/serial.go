package serial

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/avast/retry-go/v4"
	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
	"golang.org/x/mod/semver"
)

type Client struct {
}

func ListPorts() (string, []string, error) {
	var portsList []string
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return "", nil, err
	}
	if len(ports) == 0 {
		return "", nil, errors.New("no serial ports found")
	}
	var output strings.Builder

	output.WriteString("detected ports:\n")
	for i, port := range ports {
		pref := " "
		jun := "┗"
		if len(ports) > 1 && i+1 < len(ports) {
			pref = "┃"
			jun = "┣"
		}
		output.WriteString(fmt.Sprintf("  %s %s\n", jun, port.Name))
		if port.IsUSB {
			output.WriteString(fmt.Sprintf("  %s  ┣ USB ID: %s:%s\n", pref, port.VID, port.PID))
			output.WriteString(fmt.Sprintf("  %s  ┗ USB serial: %s\n", pref, port.SerialNumber))
			portsList = append(portsList, port.Name)
		}
	}
	return output.String(), portsList, nil
}

var speeds = []int{57600, 1000000, 57600}

func OpenPort(port, versionString string) (serial.Port, string, error) {
	mode := &serial.Mode{
		BaudRate: 1000000,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}
	//m.output("Open adapter on %q %dkbp/s", port, mode.BaudRate/1000)

	var sr serial.Port
	var adapterVersion string

	err := retry.Do(func() error {
		var err error
		sr, err = serial.Open(port, mode)
		if err != nil {
			return err
		}
		sr.ResetInputBuffer()
		sr.ResetOutputBuffer()
		if err := sr.SetReadTimeout(5 * time.Millisecond); err != nil {
			sr.Close()
			return err
		}
		if versionString, err = getVersion(sr); err != nil {
			sr.Close()
			return err
		}

		if semver.Compare(versionString, adapterVersion) < 0 {
			//m.output("USB adapter is running newer wire version (%s). Please update CIM Tool", ver)
		}
		if semver.Compare(versionString, adapterVersion) > 0 {
			//m.output("USB adapter is running old wire version (%s). Please use settings to update your adapter firmware", ver)
		}

		return nil
	},
		retry.OnRetry(func(n uint, err error) {
			//m.output("trying %dkbp/s", speeds[n]/1000)
			mode.BaudRate = speeds[n]
		}),
		retry.Attempts(3),
		retry.Delay(200*time.Millisecond),
		retry.LastErrorOnly(true),
	)
	if err != nil {
		return nil, "", err
	}
	return sr, adapterVersion, nil
}

func getVersion(stream serial.Port) (string, error) {
	start := time.Now()
	var version []byte
	for {
		readBuffer := make([]byte, 8)
		n, err := stream.Read(readBuffer)
		if err != nil {
			return "", err
		}
		if time.Since(start) > 3*time.Second {
			return "", errors.New("got no response from adapter")
		}
		if n == 0 {
			continue
		}
		for _, b := range readBuffer[:n] {
			if b == '\n' {
				return string(version[:]), nil
			}
			version = append(version, b)
		}
	}
}
