package kardbot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"

	"github.com/lus/dgc"
)

// Map logrus log levels to discordgo log levels
func logrusToDiscordGo() map[log.Level]int {
	return map[log.Level]int{
		log.PanicLevel: discordgo.LogError,
		log.FatalLevel: discordgo.LogError,
		log.ErrorLevel: discordgo.LogError,
		log.WarnLevel:  discordgo.LogWarning,
		log.InfoLevel:  discordgo.LogInformational,
		log.DebugLevel: discordgo.LogInformational,
		log.TraceLevel: discordgo.LogDebug,
	}
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

	if lvl, err := log.ParseLevel(levelStr); err == nil {
		info := fmt.Sprintf(`Set logging level to "%s"`, levelStr)
		log.Info(info)
		ctx.RespondText(info)
		log.SetLevel(lvl)
		if bot().EnableDGLogging {
			// TODO: make this thread safe somehow (logrus is already thread safe)
			ctx.Session.LogLevel = logrusToDiscordGo()[lvl]
		}
	} else {
		log.Error(err)
	}
}
