package main

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/Kardbord/Kard-bot/kardbot"
	"github.com/Kardbord/Kard-bot/kardbot/config"

	// For runtime profiling if enabled in config
	"net/http"
	_ "net/http/pprof"
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

// Start an http server for pprof profiling data
// if configured. No-op if not.
// See https://pkg.go.dev/net/http/pprof
var pprofServe = func() {
	log.Info("pprof not enabled (this is normal)")
}

var PprofCfg = pprofConfig{}

type pprofConfig struct {
	Enabled              bool   `json:"enabled"`
	Address              string `json:"address"`
	BlockProfileRate     int    `json:"block-profile-rate"`
	MutexProfileFraction int    `json:"mutex-profile-fraction"`
}

func init() {
	jsonCfg, err := config.NewJsonConfig(MainConfigFile)
	if err != nil {
		log.Fatal(err)
	}

	{
		cfg := struct {
			pprofConfig `json:"pprof"`
		}{}
		err = json.Unmarshal(jsonCfg.Raw, &cfg)
		if err != nil {
			log.Fatal(err)
		}
		PprofCfg = cfg.pprofConfig
	}

	if !PprofCfg.Enabled {
		return
	}

	if PprofCfg.Address == "" {
		log.Warn("pprof is enabled but address is not set, not starting pprof server")
		return
	}

	log.Infof("Setting block profile rate to %d", PprofCfg.BlockProfileRate)
	runtime.SetBlockProfileRate(PprofCfg.BlockProfileRate)

	log.Infof("Setting mutex profile fraction to %d", PprofCfg.MutexProfileFraction)
	runtime.SetMutexProfileFraction(PprofCfg.MutexProfileFraction)

	pprofServe = func() {
		log.Infof("Starting pprof server at %s/debug/pprof/", PprofCfg.Address)
		log.Info(http.ListenAndServe(PprofCfg.Address, nil))
	}
}

func main() {
	go pprofServe()
	kardbot.RunAndBlock(PprofCfg.Enabled)
}
