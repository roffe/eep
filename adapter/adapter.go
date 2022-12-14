package adapter

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/avast/retry-go/v4"
	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
	"golang.org/x/mod/semver"
)

var speeds = []int{57600, 1000000, 57600}

type Client struct {
	port serial.Port

	rdelay uint8
	wdelay uint8

	onProgress func(progress float64)
	onMessage  func(msg string)
	onError    func(err error)
}

func New(rDelay, wDelay uint8) *Client {
	client := &Client{
		rdelay: rDelay,
		wdelay: wDelay,

		onProgress: func(float64) {},

		onMessage: func(msg string) {
			log.Println(msg)
		},
		onError: func(err error) {
			log.Println(err.Error())
		},
	}
	return client
}

func (c *Client) Port() serial.Port {
	return c.port
}

func (c *Client) Close() error {
	if c.port == nil {
		return nil
	}
	return c.port.Close()
}

func (c *Client) OnProgress(f func(progress float64)) *Client {
	c.onProgress = f
	return c
}

func (c *Client) OnMessage(f func(msg string)) *Client {
	c.onMessage = f
	return c
}

func (c *Client) OnError(f func(err error)) *Client {
	c.onError = f
	return c
}

func ListPorts() (string, []string, error) {
	var portsList []string
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return "", nil, err
	}
	if len(ports) == 0 {
		return "", nil, errors.New("no serial ports found")
	}
	var output strings.Builder

	output.WriteString("detected ports:\n")
	for i, port := range ports {
		pref := " "
		jun := "┗"
		if len(ports) > 1 && i+1 < len(ports) {
			pref = "┃"
			jun = "┣"
		}
		output.WriteString(fmt.Sprintf("  %s %s\n", jun, port.Name))
		if port.IsUSB {
			output.WriteString(fmt.Sprintf("  %s  ┣ USB ID: %s:%s\n", pref, port.VID, port.PID))
			output.WriteString(fmt.Sprintf("  %s  ┗ USB serial: %s\n", pref, port.SerialNumber))
			portsList = append(portsList, port.Name)
		}
	}
	return output.String(), portsList, nil
}

func (c *Client) Open(portName, clientVersion string) error {
	return c.openPort(portName, clientVersion)
}

func (c *Client) openPort(port, versionString string) error {
	mode := &serial.Mode{
		BaudRate: 1000000,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	c.onMessage(fmt.Sprintf("Open adapter on %q %dkbp/s", port, mode.BaudRate/1000))

	var sr serial.Port
	var adapterVersion string
	err := retry.Do(func() error {
		var err error
		sr, err = serial.Open(port, mode)
		if err != nil {
			return err
		}
		sr.ResetInputBuffer()
		sr.ResetOutputBuffer()
		if err := sr.SetReadTimeout(5 * time.Millisecond); err != nil {
			sr.Close()
			return err
		}
		if adapterVersion, err = getVersion(sr); err != nil {
			sr.Close()
			return err
		}

		if semver.Compare(versionString, adapterVersion) < 0 {
			c.onMessage(fmt.Sprintf("USB adapter is running newer wire version (%s). Please update CIM Tool", adapterVersion))
		}
		if semver.Compare(versionString, adapterVersion) > 0 {
			c.onMessage(fmt.Sprintf("USB adapter is running older wire version (%s). Please use settings to update your adapter firmware", adapterVersion))
		}
		c.port = sr
		return nil
	},
		retry.OnRetry(func(n uint, err error) {
			c.onMessage(fmt.Sprintf("trying %dkbp/s", speeds[n]/1000))
			mode.BaudRate = speeds[n]
		}),
		retry.Attempts(3),
		retry.Delay(200*time.Millisecond),
		retry.LastErrorOnly(true),
	)
	if err != nil {
		return err
	}
	return nil
}

func getVersion(stream serial.Port) (string, error) {
	start := time.Now()
	var version []byte
	for {
		readBuffer := make([]byte, 8)
		n, err := stream.Read(readBuffer)
		if err != nil {
			return "", err
		}
		if time.Since(start) > 3*time.Second {
			return "", errors.New("got no response from adapter")
		}
		if n == 0 {
			continue
		}
		for _, b := range readBuffer[:n] {
			if b == '\n' {
				return string(version[:]), nil
			}
			version = append(version, b)
		}
	}
}

func (c *Client) sendCMD(op string, chip uint8, size uint16, org uint8, delay uint8) error {
	cmd := fmt.Sprintf("%s,%d,%d,%d,%d\r", op, chip, size, org, delay)
	n, err := c.port.Write([]byte(cmd))
	if err != nil {
		return err
	}
	if n != len(cmd) {
		return errors.New("Failed to write all bytes to port") //lint:ignore ST1005 ignore this
	}
	return nil
}

func (c *Client) readBytes(size int) ([]byte, error) {
	out := make([]byte, size)
	readBuffer := make([]byte, 32)
	pos := 0
	lastRead := time.Now()

	c.onProgress(0)
	for pos < size {
		if time.Since(lastRead) > 2*time.Second {
			return nil, errors.New("Timeout reading eeprom") //lint:ignore ST1005 ignore this
		}
		n, err := c.port.Read(readBuffer)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			continue
		}
		lastRead = time.Now()
		c.onProgress(float64(pos))
	inner:
		for _, b := range readBuffer[:n] {
			out[pos] = b
			pos++
			if pos == size {
				break inner
			}
		}
	}

	c.onProgress(float64(size))
	return out, nil
}

const (
	opWrite = "w"
	opRead  = "r"
	opErase = "e"
)

func (c *Client) ReadCIM() ([]byte, error) {
	c.port.ResetInputBuffer()
	c.port.ResetOutputBuffer()
	if err := c.sendCMD(opRead, 66, 512, 8, c.rdelay); err != nil {
		return nil, err
	}
	return c.readBytes(512)
}

func (c *Client) EraseCIM() error {
	if err := c.port.ResetInputBuffer(); err != nil {
		return err
	}
	if err := c.sendCMD(opErase, 66, 1, 8, c.wdelay); err != nil {
		return err
	}
	if err := c.waitAck('\a', 2*time.Second); err != nil {
		return err
	}
	time.Sleep(20 * time.Millisecond)
	c.port.ResetInputBuffer()
	return nil
}

func (c *Client) WriteCIM(data []byte) error {
	if err := c.sendCMD(opWrite, 66, 512, 8, c.wdelay); err != nil {
		return err
	}
	if err := c.waitAck('\f', 3*time.Second); err != nil {
		return err
	}

	c.onProgress(0)

	sendLock := make(chan struct{}, 1)
	var done bool

	go func() {
		buff := make([]byte, 1)
		for !done {
			n, err := c.port.Read(buff)
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
					c.onError(errors.New("unexpected ack"))
				}
			}

		}
	}()

	r := bytes.NewReader(data)
	buffSize := 16
	buff := make([]byte, buffSize)
	rr := 0
	for {
		n, err := r.Read(buff)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if n != buffSize {
			return errors.New("invalid size on read") // this should never happen
		}
		rr += n
		c.onProgress(float64(rr))
		sendLock <- struct{}{}
		if _, err := c.port.Write(buff[:n]); err != nil {
			return err
		}
	}

	done = true
	time.Sleep(75 * time.Millisecond)
	c.onProgress(float64(len(data)))
	return nil
}

func (c *Client) waitAck(char byte, timeout time.Duration) error {
	start := time.Now()
	readBuffer := make([]byte, 1)
	for {
		n, err := c.port.Read(readBuffer)
		if err != nil {
			return err
		}
		if time.Since(start) > timeout {
			return errors.New("got no response from adapter")
		}
		if n == 0 {
			continue
		}
		if readBuffer[0] == char {
			return nil
		}
	}
}
