package cmd

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/cobra"
	"go.bug.st/serial"
)

// writeCmd represents the write command
var writeCmd = &cobra.Command{
	Use:   "write <filename>",
	Short: "write eeprom content",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		bin, err := os.ReadFile(args[0])
		if err != nil {
			return err
		}

		chip, size, org, port, err := getFlags()
		if err != nil {
			return err
		}

		sr, err := openPort(port)
		if err != nil {
			return err
		}
		defer sr.Close()
		log.Printf("erasing eeprom type: %d, size: %d bytes, org: %d", chip, size, org)
		if err := erase(ctx, sr, chip, size, org); err != nil {
			log.Fatal(err)
		}

		if err := write(ctx, sr, chip, size, org, bin); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(writeCmd)
}

func write(ctx context.Context, stream serial.Port, chip uint8, size uint16, org uint8, data []byte) error {
	if err := sendCMD(stream, opWrite, chip, size, org); err != nil {
		return err
	}
	if err := waitAck(stream, '\f'); err != nil {
		return err
	}

	bar := pb.StartNew(int(size))

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

	for _, b := range data {
		sendLock <- struct{}{}
		if _, err := stream.Write([]byte{b}); err != nil {
			return err
		}
		bar.Increment()
	}
	done = true
	bar.Finish()
	time.Sleep(100 * time.Millisecond)
	return nil
}
