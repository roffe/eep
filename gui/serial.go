package gui

import (
	"fmt"
	"time"

	"github.com/hirschmann-koxha-gbr/cim/pkg/cim"
	"github.com/hirschmann-koxha-gbr/eep/adapter"
)

func (m *mainWindow) newAdapter() *adapter.Client {
	onMessage := func(msg string) {
		m.output(msg)
	}
	onProgress := func(progress float64) {
		m.progressBar.SetValue(progress)
	}
	onError := func(err error) {
		m.output(err.Error())
	}
	rd, err := m.e.readDelayValue.Get()
	if err != nil {
		panic(err)
	}
	wd, err := m.e.writeDelayValue.Get()
	if err != nil {
		panic(err)
	}
	return adapter.New(uint8(rd), uint8(wd)).OnMessage(onMessage).OnProgress(onProgress).OnError(onError)

}

func (m *mainWindow) writeCIM(port string, data []byte) error {
	input, err := cim.MustLoadBytes("read.bin", data)
	if err != nil {
		return fmt.Errorf("Failed to load CIM: %w", err) //lint:ignore ST1005 ignore
	}

	xorBytes, err := input.XORBytes()
	if err != nil {
		return fmt.Errorf("Failed to XOR CIM: %w", err) //lint:ignore ST1005 ignore
	}

	client := m.newAdapter()
	if err := client.Open(m.e.port, VERSION); err != nil {
		return fmt.Errorf("Failed to init adapter: %w", err) //lint:ignore ST1005 ignore
	}
	defer client.Close()

	m.progressBar.Max = float64(len(xorBytes))

	if err := client.WriteCIM(xorBytes); err != nil {
		return fmt.Errorf("Failed to write CIM: %w", err) //lint:ignore ST1005 ignore
	}

	if verify, err := m.e.verifyWrite.Get(); verify && err == nil {
		time.Sleep(200 * time.Millisecond)
		rawBytes, err := client.ReadCIM()
		if err != nil {
			return fmt.Errorf("Failed to read CIM: %w", err) //lint:ignore ST1005 ignore
		}
		readback, err := cim.MustLoadBytes("read.bin", rawBytes)
		if err != nil {
			return fmt.Errorf("Failed to load CIM: %w", err) //lint:ignore ST1005 ignore
		}

		if input.MD5() == readback.MD5() && input.CRC32() == readback.CRC32() {
			m.output("Write verified OK")
		} else {
			m.output("Write verify failed")
		}
	}
	return nil
}

func (m *mainWindow) readCIM() ([]byte, *cim.Bin, error) {
	client := m.newAdapter()
	if err := client.Open(m.e.port, VERSION); err != nil {
		return nil, nil, fmt.Errorf("Failed to init adapter: %v", err) //lint:ignore ST1005 ignore
	}
	defer client.Close()

	m.progressBar.Max = 512

	start := time.Now()
	m.output("Reading CIM ...")

	rawBytes, err := client.ReadCIM()
	if err != nil {
		return rawBytes, nil, fmt.Errorf("Failed to read CIM: %w", err) //lint:ignore ST1005 ignore
	}
	defer m.output("Read took %s", time.Since(start).String())
	bin, err := cim.LoadBytes("read.bin", rawBytes)
	if err != nil {
		return rawBytes, nil, fmt.Errorf("Failed to load CIM: %w", err) //lint:ignore ST1005 ignore
	}
	if err := bin.Validate(); err != nil {
		return rawBytes, nil, fmt.Errorf("Failed to validate CIM: %w", err) //lint:ignore ST1005 ignore
	}
	return rawBytes, bin, nil
}

func (m mainWindow) readMIU() ([]byte, error) {
	client := m.newAdapter()
	if err := client.Open(m.e.port, VERSION); err != nil {
		return nil, fmt.Errorf("Failed to init adapter: %v", err) //lint:ignore ST1005 ignore
	}
	defer client.Close()

	m.progressBar.Max = 256

	start := time.Now()
	m.output("Reading MIU ...")

	rawBytes, err := client.ReadMIU()
	if err != nil {
		return rawBytes, fmt.Errorf("Failed to read MIU: %w", err) //lint:ignore ST1005 ignore
	}
	defer m.output("Read took %s", time.Since(start).String())

	return rawBytes, nil
}
