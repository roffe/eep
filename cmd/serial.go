package cmd

import (
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

	if err := sr.SetReadTimeout(5 * time.Millisecond); err != nil {
		return nil, err
	}

	if err := waitAck(sr, '\n'); err != nil {
		return nil, err
	}

	return sr, nil
}
