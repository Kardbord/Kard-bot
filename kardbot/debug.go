package kardbot

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/lus/dgc"
)

var logLevelMap = map[string]log.Level{
	"panic":   log.PanicLevel,
	"fatal":   log.FatalLevel,
	"error":   log.ErrorLevel,
	"err":     log.ErrorLevel,
	"warning": log.WarnLevel,
	"warn":    log.WarnLevel,
	"info":    log.InfoLevel,
	"debug":   log.DebugLevel,
	"trace":   log.TraceLevel,
}

func getLogLevelKeys() []string {
	keys := make([]string, 0, len(logLevelMap))
	for k := range logLevelMap {
		keys = append(keys, k)
	}
	return keys
}

func updateLogLevel(ctx *dgc.Ctx) {
	if isOwner, err := authorIsOwner(ctx); err != nil {
		log.Error(err)
		return
	} else if !isOwner {
		log.Warnf("User %s (%s) does not have privilege to update log level", ctx.Event.Author.Username, ctx.Event.Author.ID)
		return
	}

	args, err := getArgsExpectCount(ctx, 1, true)
	if err != nil {
		log.Error(err)
		return
	}
	levelStr := strings.ToLower(args.Get(0).Raw())

	if l, ok := logLevelMap[levelStr]; ok {
		info := fmt.Sprintf(`Set logging level to "%s"`, levelStr)
		log.Info(info)
		ctx.RespondText(info)
		log.SetLevel(l)
	} else {
		log.Errorf("Invalid error level provided: %s", levelStr)
	}
}
