package kardbot

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/TannerKvarfordt/Kard-bot/kardbot/dg_helpers"
	"github.com/bwmarrin/discordgo"
	"github.com/gabriel-vasile/mimetype"
	"github.com/lucasb-eyer/go-colorful"

	log "github.com/sirupsen/logrus"
)

const roboCatPng string = "Robo_cat.png"

func botInfo(s *discordgo.Session, i *discordgo.InteractionCreate) {
	wg := bot().updateLastActive()
	defer wg.Wait()

	if isSelf, err := authorIsSelf(s, i); err != nil {
		log.Error(err)
		return
	} else if isSelf {
		log.Trace("Ignoring message from self")
		return
	}

	log.Debugf("Opening %s", roboCatPng)
	fd, err := os.Open(roboCatPng)
	if err != nil {
		log.Error(err)
		return
	}
	defer fd.Close()
	mimeType, err := mimetype.DetectReader(fd)
	if err != nil {
		log.Error(err)
		return
	}
	_, err = fd.Seek(0, 0)
	if err != nil {
		log.Error(err)
		return
	}

	hexColor, _ := strconv.ParseInt(strings.Replace(colorful.FastHappyColor().Hex(), "#", "", -1), 16, 32)
	embed := dg_helpers.NewEmbed().
		SetColor(int(hexColor)).
		SetTitle(s.State.User.Username).
		SetURL("https://github.com/TannerKvarfordt/Kard-bot").
		SetDescription(fmt.Sprintf("Hello! I'm %s! You can find my code or submit an issue about my behavior on [GitHub](https://github.com/TannerKvarfordt/Kard-bot). Below is some information about the commands I offer.", s.State.User.Username)).
		SetThumbnail("attachment://" + roboCatPng)

	for _, cmd := range getCommands() {
		embed.AddField("/"+cmd.Name, cmd.Description)
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed.Truncate().MessageEmbed},
			Files: []*discordgo.File{
				{
					Name:        roboCatPng,
					ContentType: mimeType.String(),
					Reader:      fd,
				},
			},
		},
	})
	if err != nil {
		log.Error(err)
	}
}
