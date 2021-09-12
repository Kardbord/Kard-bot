package kardbot

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/lus/dgc"
)

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

	if lvl, err := log.ParseLevel(levelStr); err == nil {
		info := fmt.Sprintf(`Set logging level to "%s"`, levelStr)
		log.Info(info)
		ctx.RespondText(info)
		log.SetLevel(lvl)
	} else {
		log.Error(err)
	}
}
