package main

import (
	"log"
	"os"

	"github.com/Hirschmann-Koxha-GbR/eep/cmd"
	"github.com/Hirschmann-Koxha-GbR/eep/gui"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main() {
	if len(os.Args) > 1 {
		cmd.Execute()
	} else {
		gui.Run()
	}
}
