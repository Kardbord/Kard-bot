package main

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/TannerKvarfordt/Kard-bot/kardbot"
	"github.com/TannerKvarfordt/Kard-bot/kardbot/config"
)

const MainConfigFile = "config/setup.json"

func init() {
	log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.UnixDate,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			split := strings.Split(f.File, "Kard-bot/")
			filename := "Kard-bot/" + split[len(split)-1]
			return "", fmt.Sprintf("%s:%d", filename, f.Line)
		},
	})

	cfg := struct {
		DefaultLogLvl string `json:"default-log-level"`
	}{"info"}

	jsonCfg, err := config.NewJsonConfig(MainConfigFile)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(jsonCfg.Raw, &cfg)
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
	kardbot.RunAndBlock()
}
