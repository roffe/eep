package gui

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2/widget"
	"github.com/roffe/cim/pkg/cim"
	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

const (
	opWrite = "w"
	opRead  = "r"
	opErase = "e"
)

func (m *mainWindow) listPorts() []string {
	var portsList []string
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		m.output(err.Error())
		return []string{}
	}
	if len(ports) == 0 {
		m.output("No serial ports found!")
		return []string{}
	}

	/*
		for i := 0; i < 6; i++ {
			ports = append(ports, &enumerator.PortDetails{
				Name:         fmt.Sprintf("Dummy%d", i),
				VID:          strconv.Itoa(i),
				PID:          strconv.Itoa(i),
				SerialNumber: "foo",
				IsUSB:        true,
			})
		}
	*/

	m.output("Detected ports")
	for i, port := range ports {
		pref := " "
		jun := "┗"
		if len(ports) > 1 && i+1 < len(ports) {
			pref = "┃"
			jun = "┣"
		}

		m.output("  %s %s", jun, port.Name)
		if port.IsUSB {
			m.output("  %s  ┣ USB ID: %s:%s", pref, port.VID, port.PID)
			m.output("  %s  ┗ USB serial: %s", pref, port.SerialNumber)
			portsList = append(portsList, port.Name)
		}
	}
	m.e.state.portList = portsList
	return portsList
}

func (m *mainWindow) openPort(port string) (serial.Port, error) {
	mode := &serial.Mode{
		BaudRate: 57600,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}
	m.output("Init adapter on %q... ", port)
	sr, err := serial.Open(port, mode)
	if err != nil {
		return nil, err
	}

	if err := sr.SetReadTimeout(5 * time.Millisecond); err != nil {
		return nil, err
	}

	if err := waitAck(sr, '\n'); err != nil {
		return sr, err
	}
	m.append("Done")
	return sr, nil
}

func (m *mainWindow) writeCIM(port string, data []byte) bool {
	sr, err := m.openPort(m.e.state.port)
	if sr != nil {
		defer sr.Close()
	}
	if err != nil {
		m.output("Failed to init adapter: %v", err)
		return false
	}
	if err := m.write(context.TODO(), sr, data); err != nil {
		m.output("Failed to write: %v", err)
		return false
	}
	return true
}

func (m *mainWindow) readCIM(port string, count int) ([]byte, *cim.Bin, error) {
	sr, err := m.openPort(m.e.state.port)
	if sr != nil {
		defer sr.Close()
	}
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to init adapter: %v", err) //lint:ignore ST1005 ignore
	}

	sr.ResetInputBuffer()

	m.progressBar.Max = 512 * float64(count)
	m.progressBar.SetValue(0)

	start := time.Now()
	m.output("Reading CIM ...")
	rawBytes, bin, err := m.readN(sr)
	if err != nil {
		m.output("Read took %s", time.Since(start).String())
		return rawBytes, nil, err
	}
	if err := bin.Validate(); err != nil {
		m.output("Read took %s", time.Since(start).String())
		return rawBytes, nil, err
	}

	m.output("Read took %s", time.Since(start).String())
	return rawBytes, bin, nil
}

func (m *mainWindow) readN(sr serial.Port) ([]byte, *cim.Bin, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()
	bin, err := m.read(ctx, sr, 66, 512, 8, m.progressBar)
	if err != nil {
		return bin, nil, err
	}
	cb, err := cim.LoadBytes("read.bin", bin)
	return bin, cb, err
}

func waitAck(stream serial.Port, char byte) error {
	start := time.Now()
	readBuffer := make([]byte, 1)
	for {
		n, err := stream.Read(readBuffer)
		if err != nil {
			return err
		}
		if time.Since(start) > 3*time.Second {
			return errors.New("got no ack")
		}
		if n == 0 {
			continue
		}
		if readBuffer[0] == char {
			return nil
		}
	}
}

func sendCMD(stream serial.Port, op string, chip uint8, size uint16, org uint8, delay uint8) error {
	cmd := fmt.Sprintf("%s,%d,%d,%d,%d\r", op, chip, size, org, delay)
	//log.Println(cmd)
	n, err := stream.Write([]byte(cmd))
	if err != nil {
		return err
	}
	if n != len(cmd) {
		return errors.New("Failed to write all bytes to port") //lint:ignore ST1005 ignore this
	}
	return nil
}

func (m *mainWindow) read(ctx context.Context, stream serial.Port, chip uint8, size uint16, org uint8, p *widget.ProgressBar) ([]byte, error) {
	f, err := m.e.state.readDelayValue.Get()
	if err != nil {
		return nil, err
	}
	if err := sendCMD(stream, opRead, chip, size, org, uint8(f)); err != nil {
		return nil, err
	}
	out, err := readBytes(ctx, stream, p)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func readBytes(ctx context.Context, stream serial.Port, p *widget.ProgressBar) ([]byte, error) {
	out := make([]byte, 512)
	//buff := bytes.NewBuffer(nil)
	readBuffer := make([]byte, 32)
	pos := 0
	lastRead := time.Now()
	for pos < 512 {
		select {
		case <-ctx.Done():
			return out, ctx.Err()
		default:
		}
		if time.Since(lastRead) > 2*time.Second {
			log.Printf("%X", out)
			return nil, errors.New("Timeout reading eeprom") //lint:ignore ST1005 ignore this
		}
		n, err := stream.Read(readBuffer)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			continue
		}
		lastRead = time.Now()
		p.SetValue(float64(pos))
	inner:
		for _, b := range readBuffer[:n] {
			out[pos] = b
			pos++
			if pos == 512 {
				break inner
			}
		}
	}
	p.SetValue(512)
	return out, nil
}

func (m *mainWindow) erase(stream serial.Port) error {
	f, err := m.e.state.writeDelayValue.Get()
	if err != nil {
		return err
	}
	if err := stream.ResetInputBuffer(); err != nil {
		return err
	}
	if err := sendCMD(stream, opErase, 66, 1, 8, uint8(f)); err != nil {
		return err
	}
	if err := waitAck(stream, '\a'); err != nil {
		return err
	}
	time.Sleep(20 * time.Millisecond)
	stream.ResetInputBuffer()
	return nil
}

func (m *mainWindow) write(ctx context.Context, stream serial.Port, data []byte) error {
	f, err := m.e.state.writeDelayValue.Get()
	if err != nil {
		return err
	}
	if err := sendCMD(stream, opWrite, 66, 512, 8, uint8(f)); err != nil {
		return err
	}
	if err := waitAck(stream, '\f'); err != nil {
		return err
	}

	m.progressBar.Max = 512
	m.progressBar.SetValue(0)

	sendLock := make(chan struct{}, 1)
	var done bool

	go func() {
		buff := make([]byte, 1)
		for !done {
			n, err := stream.Read(buff)
			if err != nil {
				log.Println(err)
				return
			}
			if done {
				return
			}
			if n == 0 {
				continue
			}
			if buff[0] == '\f' {
				select {
				case <-sendLock:
				default:
					panic("korv")
				}
			}

		}
	}()
	/*
		r := bytes.NewReader(data)
		bs := 0
		chunkSize := 16
		payload := make([]byte, chunkSize)
		for {
			bs += chunkSize
			n, err := r.Read(payload)
			if err != nil {
				if err == io.EOF {
					break
				}
				return err

			}
			if n != chunkSize {
				return errors.New("invalid chunk size")
			}
			sendLock <- struct{}{}
			nw, err := stream.Write(payload[:n])
			if err != nil {
				return err
			}
			m.progressBar.SetValue(float64(bs))
			if nw != chunkSize {
				log.Println(err)
			}

		}
	*/
	for i, b := range data {
		m.progressBar.SetValue(float64(i))
		sendLock <- struct{}{}
		if _, err := stream.Write([]byte{b}); err != nil {
			return err
		}
	}

	done = true
	time.Sleep(100 * time.Millisecond)
	m.progressBar.SetValue(512)
	return nil
}
