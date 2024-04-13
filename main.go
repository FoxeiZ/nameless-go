package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	bot "github.com/foxeiz/namelessgo/src"
	_ "github.com/foxeiz/namelessgo/src/extractors/youtube"
)

func main() {
	bot.Start()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Println("Shutting down...")
	bot.Close()
}

// func main() {
// 	extractors.Extract(
// 		"https://www.youtube.com/watch?v=B0dNXxLbh8s",
// 		extractors.Options{},
// 	)
// }
