package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/roffe/eep/cmd"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	ctx, cancel := context.WithCancel(context.TODO())

	go func() {
		<-sig
		cancel()
		time.Sleep(10 * time.Second)
		os.Exit(1)
	}()

	cmd.Execute(ctx)
}
