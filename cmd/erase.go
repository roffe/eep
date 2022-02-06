package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"
	"go.bug.st/serial"
)

func init() {
	rootCmd.AddCommand(eraseCmd)
}

// eraseCmd represents the erase command
var eraseCmd = &cobra.Command{
	Use:   "erase",
	Short: "erase eeprom",
	Args:  cobra.NoArgs,
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

		log.Printf("erasing eeprom type: %d, size: %d bytes, org: %d", chip, size, org)

		if err := erase(ctx, sr, chip, size, org); err != nil {
			log.Fatal(err)
		}
		if err := waitAck(sr, '\a'); err != nil {
			return err
		}

		log.Println("eeprom erased")

		return nil
	},
}

func erase(ctx context.Context, stream serial.Port, chip uint8, size uint16, org uint8) error {
	if err := sendCMD(stream, opErase, chip, 1, org); err != nil {
		return err
	}
	return nil
}
