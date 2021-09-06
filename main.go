package main

import (
	"fmt"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/TannerKvarfordt/Kard-bot/kardbot"
)

func init() {
	log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			split := strings.Split(f.File, "Kard-bot/")
			filename := "Kard-bot/" + split[len(split)-1]
			return "", fmt.Sprintf("%s:%d", filename, f.Line)
		},
	})
	// TODO: make this configurable
	log.SetLevel(log.DebugLevel)
}

func main() {
	kbot := kardbot.NewKardbot()
	kbot.Run(true)
}
