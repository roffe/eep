package avr

import (
	"archive/zip"
	"bufio"
	"bytes"
	_ "embed"
	"io"
	"net/http"
	"os"
	"os/exec"
)

//go:embed firmware.hex
var firmwareHex []byte

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

	cb("%s", "Downloading avrdude ...")
	if err := getAvrdudeBytes(); err != nil {
		return nil, err
	}

	defer os.Remove("avrdude.exe")
	defer os.Remove("avrdude.conf")

	if err := os.WriteFile("firmware.hex", firmwareHex, 0644); err != nil {
		return nil, err
	}
	defer os.Remove("firmware.hex")

	opts := []string{
		"/C",
		"avrdude.exe",
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

	cmd := exec.Command("CMD.exe", opts...)
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

func getAvrdudeBytes() error {
	resp, err := http.Get("https://github.com/mariusgreuel/avrdude/releases/download/v7.0-windows/avrdude-v7.0-windows-windows-x64.zip")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}

	fname := []string{"avrdude.exe", "avrdude.conf"}

	for _, file := range fname {
		if err := unzip(zipReader, file); err != nil {
			return err
		}
	}

	return nil
}

func unzip(zipReader *zip.Reader, file string) error {
	zf, err := zipReader.Open(file)
	if err != nil {
		return err
	}
	defer zf.Close()

	b, err := io.ReadAll(zf)
	if err != nil {
		return err
	}
	if err := os.WriteFile(file, b, 0755); err != nil {
		return err
	}
	return nil
}
