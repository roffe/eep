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
)

const (
	opWrite = "w"
	opRead  = "r"
	opErase = "e"
)

func (m *mainWindow) writeCIM(port string, data []byte) error {
	sr, err := m.openPort(state.port)
	if sr != nil {
		defer sr.Close()
	}
	if err != nil {
		return err
	}

	m.output("Flashing CIM ... ")
	start := time.Now()
	if err := m.write(context.TODO(), sr, data); err != nil {
		m.append("ERROR")
		return err
	}

	m.append("took %s", time.Since(start).String())
	return nil
}

func (m *mainWindow) readCIM(port string, count int) (*cim.Bin, error) {
	var out *cim.Bin
	sr, err := m.openPort(state.port)
	if sr != nil {
		defer sr.Close()
	}
	if err != nil {
		return nil, err
	}
	failCount := 0
	m.progressBar.Max = 512 * float64(count)
	m.progressBar.SetValue(0)
	m.output("Reading CIM")
	start := time.Now()
	for i := 0; i < count; i++ {
		sr.ResetInputBuffer()
		bin, err := m.readN(sr)
		if err != nil {
			m.output(err.Error())
			time.Sleep(100 * time.Millisecond)
			failCount++
			continue
		}
		if err := bin.Validate(); err != nil {
			m.output(err.Error())
			time.Sleep(100 * time.Millisecond)
			failCount++
			continue
		}
		m.append(".")
		out = bin
	}

	if failCount > 0 {
		return nil, errors.New("error reading CIM")
	}

	m.output("Read took %s", time.Since(start).String())
	return out, nil
}

func (m *mainWindow) readN(sr serial.Port) (*cim.Bin, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()
	bin, err := read(ctx, sr, 66, 512, 8, m.progressBar)
	if err != nil {
		return nil, err
	}
	return cim.LoadBytes("read.bin", bin)
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

	if err := m.waitAck(sr, '\n'); err != nil {
		return nil, err
	}
	m.append("Done")
	return sr, nil
}

func (m *mainWindow) waitAck(stream serial.Port, char byte) error {
	start := time.Now()
	readBuffer := make([]byte, 1)
	for {
		n, err := stream.Read(readBuffer)
		if err != nil {
			return err
		}
		if time.Since(start) > 2*time.Second {
			return errors.New("got no ack")
		}
		if n == 0 {
			continue
			//return errors.New("got no ack")
		}
		if readBuffer[0] == char {
			//log.Println("got ack")
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
		return errors.New("failed to write all bytes to com")
	}
	return nil
}

func read(ctx context.Context, stream serial.Port, chip uint8, size uint16, org uint8, p *widget.ProgressBar) ([]byte, error) {
	f, err := state.delayValue.Get()
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
	readBuffer := make([]byte, 64)
	pos := 0
	lastRead := time.Now()
	for pos < int(512) {
		select {
		case <-ctx.Done():
			return out, ctx.Err()
		default:
		}
		if time.Since(lastRead) > 5*time.Second {
			log.Println("read timeout")
			return nil, errors.New("timeout reading eeprom")
		}
		n, err := stream.Read(readBuffer)
		if err != nil {
			log.Println("error reading")
			return nil, err
		}
		if n == 0 {
			continue
		}
		lastRead = time.Now()
	inner:
		for _, b := range readBuffer[:n] {
			out[pos] = b
			pos++
			p.Value++
			p.Refresh()
			if uint16(pos) == 512 {
				break inner
			}
		}
	}
	return out, nil
}

func (m *mainWindow) erase(stream serial.Port) error {
	f, err := state.delayValue.Get()
	if err != nil {
		return err
	}

	if err := sendCMD(stream, opErase, 66, 1, 8, uint8(f)); err != nil {
		return err
	}
	if err := m.waitAck(stream, '\a'); err != nil {
		return err
	}
	return nil
}

func (m *mainWindow) write(ctx context.Context, stream serial.Port, data []byte) error {
	f, err := state.delayValue.Get()
	if err != nil {
		return err
	}
	if err := sendCMD(stream, opWrite, 66, 512, 8, uint8(f)); err != nil {
		return err
	}
	if err := m.waitAck(stream, '\f'); err != nil {
		return err
	}

	m.progressBar.Max = 512
	m.progressBar.SetValue(0)

	sendLock := make(chan struct{}, 1)
	var done bool

	go func() {
		buff := make([]byte, 1)
		for {
			n, err := stream.Read(buff)
			if err != nil {
				log.Println(err)
				return
			}
			if done {
				log.Println("Done")
				return
			}
			if n == 0 {
				continue
			}
			if buff[0] == '\f' {
				select {
				case <-sendLock:
				default:
				}
			}

		}
	}()

	for i, b := range data {
		sendLock <- struct{}{}
		if _, err := stream.Write([]byte{b}); err != nil {
			return err
		}
		m.progressBar.SetValue(float64(i + 1))
	}
	done = true
	time.Sleep(100 * time.Millisecond)
	return nil
}
