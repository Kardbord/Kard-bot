package kardbot

import (
	"github.com/bwmarrin/discordgo"

	owoify_go "github.com/deadshot465/owoify-go"
	log "github.com/sirupsen/logrus"
)

const (
	uwuSubCmdCustom = "custom"
	uwuSubCmdPasta  = "pasta"

	owoLevel = "owo"
	uwuLevel = "uwu"
	uvuLevel = "uvu"
)

var uwuChoices func() []*discordgo.ApplicationCommandOptionChoice

func init() {
	uwulevels := []string{
		owoLevel,
		uwuLevel,
		uvuLevel,
	}

	uwuchoices := make([]*discordgo.ApplicationCommandOptionChoice, len(uwulevels))
	uwuchoices[0] = &discordgo.ApplicationCommandOptionChoice{
		Name:  "Vanilla UwU",
		Value: owoLevel,
	}
	uwuchoices[1] = &discordgo.ApplicationCommandOptionChoice{
		Name:  "Moderate UwU",
		Value: uwuLevel,
	}
	uwuchoices[2] = &discordgo.ApplicationCommandOptionChoice{
		Name:  "Litewawwy unweadabwal",
		Value: uvuLevel,
	}

	uwuChoices = func() []*discordgo.ApplicationCommandOptionChoice {
		return uwuchoices
	}
}

func uwuify(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if isSelf, err := authorIsSelf(s, i); err != nil {
		log.Error(err)
		return
	} else if isSelf {
		log.Trace("Ignoring message from self")
		return
	}

	content := i.ApplicationCommandData().Options[0].Options[1].StringValue()
	if i.ApplicationCommandData().Options[0].Name == uwuSubCmdPasta {
		if p, ok := pastaMenu()[content]; ok {
			var err error
			content, err = p.makePasta()
			if err != nil {
				log.Error(err)
				return
			}
		}
	}

	uwulvl := i.ApplicationCommandData().Options[0].Options[0].StringValue()
	content = owoify_go.Owoify(content, uwulvl)

	if len(content) > int(MaxDiscordMsgLen) {
		content = firstN(content, int(MaxDiscordMsgLen)-3) + "..."
	}

	tts := false
	if len(i.ApplicationCommandData().Options[0].Options) > 2 {
		if i.ApplicationCommandData().Options[0].Options[2].BoolValue() {
			tts = true
		}
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			TTS:     tts,
		},
	})
	if err != nil {
		log.Error(err)
	}
}
