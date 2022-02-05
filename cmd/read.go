package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

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

		switch args[0] {
		case "-":
			prettyPrintBin(bin)
		default:
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

	log.Printf("read %d bytes to %s", n, fname)
	return nil
}

func prettyPrintBin(bin []byte) {
	pp := 0
	for _, b := range bin {
		fmt.Printf("%02X ", b)
		pp++
		if pp == 25 {
			fmt.Println()
			pp = 0
		}
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
	readBuffer := make([]byte, 16)
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

		/*
			outer:
				for _, b := range readBuffer[:n] {
					buff.WriteByte(b)
					if b == 0xA || b == 0x0D {
						if buff.Len() == 1 {
							buff.Reset()
							continue
						}

							str := strings.ReplaceAll(buff.String(), " ", "")
							str = strings.ReplaceAll(str, "\r", "")
							bb, err := hex.DecodeString(str)
							if err != nil {
								return nil, err
							}

								for _, db := range buff.Bytes() {
							out[pos] = db
							pos++
							if uint16(pos) == size {
								break outer
							}
						}
						buff.Reset()
					}
				}
		*/
	}

	return out, nil
}
