package kardbot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"

	"github.com/Kardbord/Kard-bot/kardbot/config"
	"github.com/Kardbord/ubiquity/stringutils"
	"github.com/bwmarrin/discordgo"
	owoify_go "github.com/deadshot465/owoify-go"
	log "github.com/sirupsen/logrus"
)

const (
	pastaCmd       = "pasta"
	pastaOptionUwu = "uwu"
	pastaOptionTTS = "tts"
)

// Returns a map of pasta names to pasta objects.
// Includes the "random" option.
var pastaMenu func() map[string]pasta

// Returns a list of paths to pasta files.
// Does not include the "random" option
var pastaList func() []string

const PastaConfigFile = "config/pasta.json"

func init() {
	cfg := struct {
		Pastas []pasta `json:"pastas"`
	}{}

	jsonCfg, err := config.NewJsonConfig(PastaConfigFile)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(jsonCfg.Raw, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	var pastas = make(map[string]pasta)
	var pastaFiles = make([]string, len(cfg.Pastas))
	for i, p := range cfg.Pastas {
		pastas[p.Name] = p
		pastaFiles[i] = p.File
	}
	if len(pastas) == 0 {
		log.Warn("No pastas found in config :(")
	}
	// Random is a special case
	pastas["random"] = pasta{Name: "random", File: pastaFiles[rand.Intn(len(pastaFiles))]}
	pastaMenu = func() map[string]pasta { return pastas }
	pastaList = func() []string { return pastaFiles }
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
	if p.Name == "random" {
		p.File = pastaList()[rand.Intn(len(pastaList()))]
	}
	fd, err := os.Open(p.File)
	if err != nil {
		return "", err
	}
	if fd == nil {
		return "", errors.New("fd is nil")
	}
	defer func() { _ = fd.Close() }()

	const bufSize = MaxDiscordMsgLen - 3
	buf := make([]byte, bufSize)

	n, err := fd.Read(buf)
	if err == io.EOF || err == nil {
		if n == int(bufSize) {
			return string(buf[:n]) + "...", nil
		}
		return string(buf[:n]), nil
	} else {
		// Something went wrong while reading the file
		return "", err
	}
}

func servePasta(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if isSelf, err := authorIsSelf(s, i); err != nil {
		log.Error(err)
		return
	} else if isSelf {
		log.Trace("Ignoring message from self")
		return
	}

	selection := i.ApplicationCommandData().Options[0].StringValue()
	content := ""
	if p, ok := pastaMenu()[selection]; ok {
		var err error
		content, err = p.makePasta()
		if err != nil {
			log.Error(err)
			interactionRespondEphemeralError(s, i, true, err)
			return
		}
	} else {
		err := fmt.Errorf("invalid selection: %s", selection)
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	}

	tts := false
	uwulvl := ""
	for j := 2; j > 0; j-- {
		if len(i.ApplicationCommandData().Options) > j {
			for k := 1; k <= j; k++ {
				switch i.ApplicationCommandData().Options[k].Name {
				case pastaOptionTTS:
					tts = i.ApplicationCommandData().Options[k].BoolValue()
				case pastaOptionUwu:
					uwulvl = i.ApplicationCommandData().Options[k].StringValue()
				}
			}
			break
		}
	}

	if uwulvl != "" {
		content = owoify_go.Owoify(content, uwulvl)
		if len(content) > int(MaxDiscordMsgLen) {
			content = stringutils.FirstN(content, MaxDiscordMsgLen-3) + "..."
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
		interactionRespondEphemeralError(s, i, true, err)
	}
}
