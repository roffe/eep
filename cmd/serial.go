package cmd

import (
	"fmt"
	"log"
	"time"

	"go.bug.st/serial"
)

func openPort(port string) (serial.Port, error) {

	mode := &serial.Mode{
		BaudRate: 57600,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	sr, err := serial.Open(port, mode)
	if err != nil {
		log.Fatal(err)
	}

	if err := sr.SetReadTimeout(1500 * time.Millisecond); err != nil {
		return nil, err
	}

	if err := waitAck(sr); err != nil {
		return nil, err
	}

	if err := sr.ResetInputBuffer(); err != nil {
		return nil, err
	}

	_, err = sr.Write([]byte{'\r', '\n'}) // empty buffer
	if err != nil {
		sr.Close()
		return nil, fmt.Errorf("failed to init adapter: %v", err)
	}

	return sr, nil
}
