package avr

import (
	_ "embed"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"go.bug.st/serial"
)

//go:embed firmware.hex
var firmwareHex []byte

// STK500v1 protocol constants (see optiboot / stk500.h)
const (
	stkOK       = 0x10
	stkInsync   = 0x14
	crcEOP      = 0x20
	cmdGetSync  = 0x30
	cmdEnterPM  = 0x50
	cmdLeavePM  = 0x51
	cmdLoadAddr = 0x55
	cmdProgPage = 0x64
	cmdReadSign = 0x75

	pageSize = 128 // ATmega328P flash page in bytes
)

func Update(port, board string, cb func(format string, values ...interface{})) ([]byte, error) {
	baud := 115200
	if board == "Nano (old bootloader)" {
		baud = 57600
	}

	firmware, err := parseIntelHex(firmwareHex)
	if err != nil {
		return nil, fmt.Errorf("parse firmware: %w", err)
	}

	cb("%s", "Opening "+port+" ...")
	p, err := serial.Open(port, &serial.Mode{BaudRate: baud})
	if err != nil {
		return nil, err
	}
	defer p.Close()
	// Short per-read timeout so we can hammer GET_SYNC inside the brief
	// (~1s) Optiboot window without overshooting it.
	p.SetReadTimeout(200 * time.Millisecond)

	// Arduino auto-reset: the reset cap triggers on the falling edge of DTR,
	// so assert high first to guarantee a clean high->low->high pulse
	// regardless of the line state the driver left on open.
	p.SetDTR(true)
	p.SetRTS(true)
	time.Sleep(50 * time.Millisecond)
	p.SetDTR(false)
	p.SetRTS(false)
	time.Sleep(250 * time.Millisecond)
	p.SetDTR(true)
	p.SetRTS(true)
	time.Sleep(50 * time.Millisecond)
	p.ResetInputBuffer()

	pr := &programmer{p: p}

	cb("%s", "Syncing with bootloader ...")
	if err := pr.sync(); err != nil {
		return nil, err
	}

	sig, err := pr.cmd([]byte{cmdReadSign}, 3)
	if err != nil {
		return nil, fmt.Errorf("read signature: %w", err)
	}
	cb("Device signature: %02X %02X %02X", sig[0], sig[1], sig[2])
	if sig[0] != 0x1E || sig[1] != 0x95 || sig[2] != 0x0F {
		return nil, fmt.Errorf("unexpected device signature %02X%02X%02X, expected 1E950F (ATmega328P)", sig[0], sig[1], sig[2])
	}

	if _, err := pr.cmd([]byte{cmdEnterPM}, 0); err != nil {
		return nil, fmt.Errorf("enter programming mode: %w", err)
	}

	cb("Writing %d bytes ...", len(firmware))
	for addr := 0; addr < len(firmware); addr += pageSize {
		end := addr + pageSize
		if end > len(firmware) {
			end = len(firmware)
		}
		if err := pr.writePage(addr, firmware[addr:end]); err != nil {
			return nil, fmt.Errorf("write page at 0x%X: %w", addr, err)
		}
		cb("Wrote 0x%04X", addr)
	}

	if _, err := pr.cmd([]byte{cmdLeavePM}, 0); err != nil {
		return nil, fmt.Errorf("leave programming mode: %w", err)
	}

	cb("%s", "Done")
	return nil, nil
}

type programmer struct {
	p serial.Port
}

// sync hammers GET_SYNC until the bootloader answers INSYNC/OK. Optiboot only
// listens for ~1s after reset, so we send fast with a short read timeout
// rather than waiting long on any single attempt.
func (pr *programmer) sync() error {
	deadline := time.Now().Add(5 * time.Second)
	resp := make([]byte, 2)
	for time.Now().Before(deadline) {
		pr.p.ResetInputBuffer()
		if _, err := pr.p.Write([]byte{cmdGetSync, crcEOP}); err != nil {
			return err
		}
		if err := pr.readFull(resp); err != nil {
			continue // timeout/no data, try again
		}
		if resp[0] == stkInsync && resp[1] == stkOK {
			return nil
		}
	}
	return fmt.Errorf("could not sync with bootloader (no response) - check the board/baud and that nothing else has the port open")
}

func (pr *programmer) writePage(addr int, data []byte) error {
	// STK500 addresses flash in words.
	word := addr / 2
	if _, err := pr.cmd([]byte{cmdLoadAddr, byte(word), byte(word >> 8)}, 0); err != nil {
		return err
	}
	payload := []byte{cmdProgPage, byte(len(data) >> 8), byte(len(data)), 'F'}
	payload = append(payload, data...)
	_, err := pr.cmd(payload, 0)
	return err
}

// cmd sends payload+CRC_EOP, then reads INSYNC, respLen data bytes, and OK.
func (pr *programmer) cmd(payload []byte, respLen int) ([]byte, error) {
	if _, err := pr.p.Write(append(payload, crcEOP)); err != nil {
		return nil, err
	}
	head := make([]byte, 1)
	if err := pr.readFull(head); err != nil {
		return nil, err
	}
	if head[0] != stkInsync {
		return nil, fmt.Errorf("expected INSYNC, got 0x%02X", head[0])
	}
	resp := make([]byte, respLen)
	if respLen > 0 {
		if err := pr.readFull(resp); err != nil {
			return nil, err
		}
	}
	tail := make([]byte, 1)
	if err := pr.readFull(tail); err != nil {
		return nil, err
	}
	if tail[0] != stkOK {
		return nil, fmt.Errorf("expected OK, got 0x%02X", tail[0])
	}
	return resp, nil
}

func (pr *programmer) readFull(b []byte) error {
	for got := 0; got < len(b); {
		n, err := pr.p.Read(b[got:])
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrNoProgress // read timeout with no data
		}
		got += n
	}
	return nil
}

// parseIntelHex decodes Intel HEX into a flat byte slice, padding gaps with 0xFF.
func parseIntelHex(raw []byte) ([]byte, error) {
	var out []byte
	for _, line := range splitLines(raw) {
		if len(line) == 0 || line[0] != ':' {
			continue
		}
		b, err := hex.DecodeString(string(line[1:]))
		if err != nil {
			return nil, err
		}
		if len(b) < 5 {
			return nil, fmt.Errorf("short record")
		}
		count := int(b[0])
		if len(b) != count+5 {
			return nil, fmt.Errorf("bad record length")
		}
		var sum byte
		for _, x := range b {
			sum += x
		}
		if sum != 0 {
			return nil, fmt.Errorf("checksum error")
		}
		addr := int(b[1])<<8 | int(b[2])
		switch b[3] {
		case 0x00: // data
			end := addr + count
			for len(out) < end {
				out = append(out, 0xFF)
			}
			copy(out[addr:end], b[4:4+count])
		case 0x01: // EOF
			return out, nil
		default:
			return nil, fmt.Errorf("unsupported record type 0x%02X", b[3])
		}
	}
	return out, nil
}

func splitLines(raw []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i := 0; i <= len(raw); i++ {
		if i == len(raw) || raw[i] == '\n' || raw[i] == '\r' {
			if i > start {
				lines = append(lines, raw[start:i])
			}
			start = i + 1
		}
	}
	return lines
}
