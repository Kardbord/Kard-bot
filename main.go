package main

import (
	"log"

	"github.com/TannerKvarfordt/Kard-bot/kardbot"
)

func init() {
	// TODO: replace log with logrus
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	kbot := kardbot.NewKardbot()
	kbot.Run(true)
}
