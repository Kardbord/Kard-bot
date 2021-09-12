package main

import (
	"encoding/json"
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

	cfg := struct {
		DefaultLogLvl string `json:"default-log-level"`
	}{"info"}

	err := json.Unmarshal(kardbot.RawJSONConfig(), &cfg)
	if err != nil {
		log.Fatal(err)
	}

	if lvl, err := log.ParseLevel(cfg.DefaultLogLvl); err == nil {
		log.SetLevel(lvl)
	} else {
		log.SetLevel(log.InfoLevel)
		log.Warnf(`Could not read default log level from config (%s). Defaulting to "%s".`, cfg.DefaultLogLvl, log.InfoLevel)
	}
}

func main() {
	log.RegisterExitHandler(kardbot.Stop)
	kardbot.Run()
	kardbot.BlockThenStop()
}
