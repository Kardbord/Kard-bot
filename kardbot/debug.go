package kardbot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
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

func updateLogLevel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	wg := bot().updateLastActive()
	defer wg.Wait()

	if isSelf, err := authorIsSelf(s, i); err != nil {
		log.Error(err)
		return
	} else if isSelf {
		log.Trace("Ignoring message from self")
		return
	}

	author, authorID, err := getInteractionCreateAuthorNameAndID(i)
	if err != nil {
		log.Error(err)
		return
	}

	if isOwner, err := authorIsOwner(i); err != nil {
		log.Error(err)
		return
	} else if !isOwner {
		log.Warnf("User %s (%s) does not have privilege to update log level", author, authorID)
		return
	}

	levelStr := strings.ToLower(i.ApplicationCommandData().Options[0].StringValue())

	if lvl, err := log.ParseLevel(levelStr); err == nil {
		info := fmt.Sprintf(`Set logging level to "%s"`, levelStr)
		log.Info(info)
		log.SetLevel(lvl)
		if bot().EnableDGLogging {
			s.LogLevel = logrusToDiscordGo()[lvl]
		}
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: info,
			},
		})
		if err != nil {
			log.Error(err)
		}
	} else {
		log.Error(err)
	}
}

func logLevelChoices() []*discordgo.ApplicationCommandOptionChoice {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, len(log.AllLevels))

	i := 0
	for _, lvl := range log.AllLevels {
		choices[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:  lvl.String(),
			Value: lvl.String(),
		}
		i++
	}
	return choices
}
