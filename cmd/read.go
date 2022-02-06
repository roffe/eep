package cmd

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"
	"unicode/utf16"

	"github.com/spf13/cobra"
	"go.bug.st/serial"
)

func init() {
	rootCmd.AddCommand(readCmd)
}

// readCmd represents the read command
var readCmd = &cobra.Command{
	Use:   "read <filename>",
	Short: "read eeprom",
	Long:  "specify filname \"-\" to output to stdout",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		chip, size, org, port, err := getFlags()
		if err != nil {
			return err
		}

		sr, err := openPort(port)
		if err != nil {
			return err
		}
		defer sr.Close()

		log.Printf("reading %d bytes from type: %d, org: %d", size, chip, org)

		bin, err := read(ctx, sr, chip, size, org)
		if err != nil {
			log.Fatal(err)
		}

		xor, err := rootCmd.PersistentFlags().GetBytesHex(optXor)
		if err != nil {
			return err
		}
		switch args[0] {
		case "-":
			prettyPrintBin(bin, org, xor[0])
		default:
			for i := 0; i < len(bin); i++ {
				bin[i] = bin[i] ^ xor[0]
			}

			if err := writeFile(args[0], size, bin); err != nil {
				return err
			}
		}

		return nil
	},
}

func writeFile(fname string, size uint16, bin []byte) error {
	f, err := os.Create(fname)
	if err != nil {
		return err
	}

	n, err := f.Write(bin)
	if err != nil {
		return err
	}

	if n != int(size) {
		log.Printf("/!\\ only wrote %d out of %d requested bytes", n, size)
	}

	log.Printf("wrote %d bytes to %s", n, fname)
	return nil
}

func prettyPrintBin(bin []byte, org uint8, xor byte) {
	pp := 0
	switch org {
	case 8:
		for _, b := range bin {
			fmt.Printf("%02X ", b^xor)
			pp++
			if pp == 25 {
				fmt.Println()
				pp = 0
			}
		}
	case 16:
		r := bytes.NewReader(bin)
		var chars []uint16
	outer:
		for {
			var c uint16
			if err := binary.Read(r, binary.BigEndian, &c); err != nil {
				if err == io.EOF {
					break outer
				}
				panic(err)
			}
			chars = append(chars, c)
		}
		runes := utf16.Decode(chars)
		for _, rr := range runes {
			fmt.Printf("%s ", string(rr))
			pp++
			if pp == 25 {
				fmt.Println()
				pp = 0
			}
		}

	default:
		panic("unknown org")
	}

	fmt.Println()
}

func read(ctx context.Context, stream serial.Port, chip uint8, size uint16, org uint8) ([]byte, error) {
	if err := sendCMD(stream, opRead, chip, size, org); err != nil {
		return nil, err
	}
	out, err := readBytes(ctx, stream, size)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func readBytes(ctx context.Context, stream serial.Port, size uint16) ([]byte, error) {
	out := make([]byte, size)
	//buff := bytes.NewBuffer(nil)
	readBuffer := make([]byte, 32)
	pos := 0
	lastRead := time.Now()

	for pos < int(size) {
		select {
		case <-ctx.Done():
			return out, ctx.Err()
		default:
		}
		if time.Since(lastRead) > 5*time.Second {
			return nil, errors.New("timeout reading eeprom")
		}
		n, err := stream.Read(readBuffer)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			log.Println("0 read")
			continue
		}
		lastRead = time.Now()
	inner:
		for _, b := range readBuffer[:n] {
			out[pos] = b
			pos++
			if uint16(pos) == size {
				break inner
			}
		}
	}

	return out, nil
}
