package kardbot

import (
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/TannerKvarfordt/Kard-bot/kardbot/config"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

// returns a map of pasta names to pasta objects
var pastaMenu func() map[string]pasta

func init() {
	cfg := struct {
		Pastas []pasta `json:"pastas"`
	}{}
	err := json.Unmarshal(config.RawJSONConfig(), &cfg)
	if err != nil {
		log.Fatal(err)
	}

	var pastas = make(map[string]pasta)
	for _, p := range cfg.Pastas {
		pastas[p.Name] = p
	}
	if len(pastas) == 0 {
		log.Warn("No pastas found in config :(")
	}
	pastaMenu = func() map[string]pasta { return pastas }
}

type pasta struct {
	Name string `json:"name"`
	File string `json:"file"`
}

func pastaChoices() []*discordgo.ApplicationCommandOptionChoice {
	options := make([]*discordgo.ApplicationCommandOptionChoice, len(pastaMenu()))

	i := 0
	for k := range pastaMenu() {
		options[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:  k,
			Value: k,
		}
		i++
	}

	return options
}

// Load and return the copy pasta based on the pasta's member fields
func (p *pasta) makePasta() (string, error) {
	fd, err := os.Open(p.File)
	if err != nil {
		return "", err
	}
	if fd == nil {
		return "", errors.New("fd is nil")
	}
	defer func() { _ = fd.Close() }()

	buf := make([]byte, MaxDiscordMsgLen-3)

	n, err := fd.Read(buf)
	if err == io.EOF {
		// We read the entire copy-pasta into the buffer
		return string(buf[:n]), nil
	} else if err != nil {
		// Something went wrong while reading the file
		return "", err
	} else {
		// We did not read the entire copy-pasta into the buffer, add ellipsis
		// TODO: Allow multi-page pastas for pastas that are over the discord message size limit; toggle pages via a button
		return string(buf[:n]) + "...", nil
	}
}

func servePasta(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// TODO: move this check into a helper function
	_, authorID, err := getInteractionCreateAuthorNameAndID(i)
	if err != nil {
		log.Error(err)
		return
	}
	if authorID == s.State.User.ID {
		log.Trace("Ignoring message from self")
		return
	}

	selection := i.ApplicationCommandData().Options[0].StringValue()
	if p, ok := pastaMenu()[selection]; ok {
		content, err := p.makePasta()
		if err != nil {
			log.Error(err)
			return
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
			},
		})
	} else {
		log.Error("invalid selection: ", selection)
	}
}
