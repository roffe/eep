/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"errors"
	"os"

	"github.com/spf13/cobra"
	"go.bug.st/serial"
)

const (
	optChip  = "chip"
	optSize  = "size"
	optOrg   = "org"
	optPort  = "port"
	optXor   = "xor"
	optErase = "erase"

	defaultChip = 66
	defaultSize = 512
	defaultOrg  = 8
	defaultPort = "COM8"

	opWrite = "w"
	opRead  = "r"
	opErase = "e"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "eep",
	Short:        "eeprom tool",
	Long:         `a CLI to interface with the arduino eeprom programmer`,
	SilenceUsage: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ctx context.Context) {
	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	pf := rootCmd.PersistentFlags()

	pf.Uint8P(optChip, "c", defaultChip, "chip type")
	pf.Uint16P(optSize, "s", defaultSize, "chip size")
	pf.Uint8P(optOrg, "o", defaultOrg, "chip org")
	pf.StringP(optPort, "p", defaultPort, "com port")
	pf.BytesHexP(optXor, "x", []byte{0x00}, "xor output")
	pf.BoolP(optErase, "e", false, "erase before write (default false")

	//cobra.MarkFlagRequired(pf, optChip)
	//cobra.MarkFlagRequired(pf, optSize)
	//cobra.MarkFlagRequired(pf, optOrg)

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func getFlags() (uint8, uint16, uint8, string, error) {
	chip, err := rootCmd.PersistentFlags().GetUint8(optChip)
	if err != nil {
		return 0, 0, 0, "", err
	}
	size, err := rootCmd.PersistentFlags().GetUint16(optSize)
	if err != nil {
		return 0, 0, 0, "", err
	}
	org, err := rootCmd.PersistentFlags().GetUint8(optOrg)
	if err != nil {
		return 0, 0, 0, "", err
	}
	port, err := rootCmd.PersistentFlags().GetString(optPort)
	if err != nil {
		return 0, 0, 0, "", err
	}
	return chip, size, org, port, nil
}

func waitAck(stream serial.Port) error {
	readBuffer := make([]byte, 1)
	x := 0
	for x < 5 {

		n, err := stream.Read(readBuffer)
		if err != nil {
			return err
		}
		if n == 0 {
			continue
			//return errors.New("got no ack")
		}
		if readBuffer[0] == '\f' {
			return nil
		} else {
			//log.Printf("%q", readBuffer[0])
			//return errors.New("invalid ack byte")
			x++
		}
	}
	return errors.New("got no ack")
}
