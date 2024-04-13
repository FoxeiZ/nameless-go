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

//	func main() {
//		extractors.Extract(
//			"https://www.youtube.com/watch?v=B0dNXxLbh8s",
//			extractors.Options{},
//		)
//	}
// func main() {
// 	u, err := url.Parse("https://youtu.be/bGzGvY85kBU?si=yJV9iITO6iIPToob")
// 	if err != nil {
// 		panic(err)
// 	}
// 	q := u.Query()
// 	ep := u.EscapedPath()
// 	splitPath := strings.Split(u.Path, "/")

// 	fmt.Println(u.Host)

// 	fmt.Println(q)
// 	fmt.Println(ep)

// 	fmt.Println(splitPath)
// }
