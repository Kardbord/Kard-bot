package kardbot

import (
	"fmt"
	"math/rand"

	"github.com/bwmarrin/discordgo"

	log "github.com/sirupsen/logrus"
)

const oddsCmd = "what-are-the-odds"

func whatAreTheOdds(s *discordgo.Session, i *discordgo.InteractionCreate) {
	event := i.ApplicationCommandData().Options[0].StringValue()
	event = sentenceEndPunctRegex().ReplaceAllString(event, "")

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeUsers, discordgo.AllowedMentionTypeRoles, discordgo.AllowedMentionTypeEveryone},
			},
			Content: fmt.Sprintf("There is a %d%% chance %s.", rand.Intn(101), event),
		},
	})

	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
	}
}
