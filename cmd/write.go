/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"os"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/cobra"
	"github.com/tarm/serial"
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

		if err := write(ctx, sr, chip, size, org, bin); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(writeCmd)
}

func write(ctx context.Context, stream *serial.Port, chip uint8, size uint16, org uint8, data []byte) error {
	if err := sendCMD(stream, opWrite, chip, size, org); err != nil {
		return err
	}

	if err := waitAck(stream); err != nil {
		return err
	}

	bar := pb.StartNew(int(size))

	for _, b := range data {
		if _, err := stream.Write([]byte{b}); err != nil {
			return err
		}
		time.Sleep(20 * time.Microsecond)
		bar.Increment()
	}

	bar.Finish()

	return nil
}
