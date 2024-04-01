package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	bot "github.com/foxeiz/namelessgo/src"
)

func main() {
	bot.Start()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Println("Shutting down...")
	bot.Close()
}
