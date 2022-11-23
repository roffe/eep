package main

import (
	"log"

	"github.com/Hirschmann-Koxha-GbR/eep/gui"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main() {
	gui.Run()
}
