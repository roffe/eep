package cmd

import (
	"fmt"
	"time"

	"github.com/tarm/serial"
)

func openPort(port string) (*serial.Port, error) {
	config := &serial.Config{
		Name:        port,
		Baud:        57600,
		ReadTimeout: 1500 * time.Millisecond,
		Parity:      serial.ParityNone,
		Size:        8,
	}

	sr, err := serial.OpenPort(config)
	if err != nil {
		return nil, err
	}

	if err := waitAck(sr); err != nil {
		return nil, err
	}
	if err := sr.Flush(); err != nil {
		return nil, err
	}

	_, err = sr.Write([]byte{'\r', '\n'}) // empty buffer
	if err != nil {
		sr.Close()
		return nil, fmt.Errorf("failed to init adapter: %v", err)
	}

	return sr, nil
}
