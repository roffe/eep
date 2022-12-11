package avr

import (
	"archive/zip"
	"bufio"
	"bytes"
	_ "embed"
	"io"
	"os"
	"os/exec"
)

//go:embed firmware.hex
var firmwareHex []byte

//go:embed avrdude.conf
var avrdudeConf []byte

//go:embed avrdude_.zip
var avrdudeZip []byte

func Update(port, board string, cb func(format string, values ...interface{})) ([]byte, error) {
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

	b, err := getAvrdudeBytes()
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile("avrdude.exe", b, 0755); err != nil {
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

func getAvrdudeBytes() ([]byte, error) {
	data := make([]byte, len(avrdudeZip))
	for i, bb := range avrdudeZip {
		data[i] = bb ^ 0x69
	}

	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(avrdudeZip)))
	if err != nil {
		return nil, err
	}
	zf, err := zipReader.Open("avrdude.exe")
	if err != nil {
		return nil, err
	}
	defer zf.Close()

	b, err := io.ReadAll(zf)
	if err != nil {
		return nil, err
	}
	return b, nil
}
