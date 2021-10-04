package kardbot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/TannerKvarfordt/Kard-bot/kardbot/dg_helpers"
	"github.com/bwmarrin/discordgo"
	"github.com/lucasb-eyer/go-colorful"

	log "github.com/sirupsen/logrus"
)

func botInfo(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if isSelf, err := authorIsSelf(s, i); err != nil {
		log.Error(err)
		return
	} else if isSelf {
		log.Trace("Ignoring message from self")
		return
	}

	hexColor, _ := strconv.ParseInt(strings.Replace(colorful.FastHappyColor().Hex(), "#", "", -1), 16, 32)
	embed := dg_helpers.NewEmbed().
		SetColor(int(hexColor)).
		SetTitle(s.State.User.Username).
		SetURL("https://github.com/TannerKvarfordt/Kard-bot").
		SetDescription(fmt.Sprintf("Hello! I'm %s! You can find my code or submit an issue about my behavior on [GitHub](https://github.com/TannerKvarfordt/Kard-bot). Below is some information about the commands I offer.", s.State.User.Username)).
		SetThumbnail("https://raw.githubusercontent.com/TannerKvarfordt/Kard-bot/main/Robo_cat.png")

	for _, cmd := range getCommands() {
		embed.AddField("/"+cmd.Name, cmd.Description)
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed.Truncate().MessageEmbed},
		},
	})
	if err != nil {
		log.Error(err)
	}
}
