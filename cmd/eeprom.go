package cmd

import (
	"errors"
	"fmt"
	"log"

	"go.bug.st/serial"
)

func sendCMD(stream serial.Port, op string, chip uint8, size uint16, org uint8) error {
	cmd := fmt.Sprintf("%s,%d,%d,%d\r", op, chip, size, org)
	log.Println(cmd)
	n, err := stream.Write([]byte(cmd))
	if err != nil {
		return err
	}
	if n != len(cmd) {
		return errors.New("failed to write all bytes to com")
	}
	return nil
}
