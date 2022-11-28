package avr

import (
	"bufio"
	_ "embed"
	"os"
	"os/exec"
)

//go:embed firmware.hex
var firmwareHex []byte

//go:embed avrdude.conf
var avrdudeConf []byte

//go:embed avrdude.exe
var avrdudeExe []byte

func Update(port, board string, cb func(format string, values ...interface{}) int) ([]byte, error) {
	var portSpeed string
	switch board {
	case "Nano":
		portSpeed = "115200"
	case "Nano (old bootloader)":
		portSpeed = "57600"
	case "Uno":
		fallthrough
	default:
		portSpeed = "115200"
	}

	if err := os.WriteFile("avrdude.conf", avrdudeConf, 0644); err != nil {
		return nil, err
	}
	defer os.Remove("avrdude.conf")

	if err := os.WriteFile("avrdude.exe", avrdudeExe, 0755); err != nil {
		return nil, err
	}
	defer os.Remove("avrdude.exe")

	if err := os.WriteFile("firmware.hex", firmwareHex, 0644); err != nil {
		return nil, err
	}
	defer os.Remove("firmware.hex")

	opts := []string{
		"-c",
		"arduino",
		"-P",
		port,
		"-b",
		portSpeed,
		"-p",
		"atmega328p",
		"-D",
		"-U",
		"flash:w:firmware.hex:i",
	}

	cmd := exec.Command("./avrdude.exe", opts...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	go func() {
		s := bufio.NewScanner(stdout)
		for s.Scan() {
			if s.Text() == "" {
				continue
			}
			cb("%s", s.Text())
		}
	}()

	go func() {
		s := bufio.NewScanner(stderr)
		for s.Scan() {
			if s.Text() == "" {
				continue
			}
			cb("%s", s.Text())
		}
	}()

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	return nil, nil
}
