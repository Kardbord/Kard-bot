package kardbot

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/TannerKvarfordt/Kard-bot/kardbot/config"
	"github.com/bwmarrin/discordgo"
	"github.com/gabriel-vasile/mimetype"

	log "github.com/sirupsen/logrus"
)

const (
	dalleFlowCmd        = "dalle-flow"
	dalleFlowPromptOpt  = "prompt"
	dalleFlowConfigFile = "./config/dalle-flow.json"
)

var (
	dalleFlowServer func() string
	dalleFlowOutput func() string
	dalleFlowScript func() string
)

func init() {
	cfg := struct {
		Server string `json:"dalle-flow-server"`
		Output string `json:"dalle-flow-output"`
		Script string `json:"dalle-flow-script"`
	}{}

	jsonCfg, err := config.NewJsonConfig(dalleFlowConfigFile)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(jsonCfg.Raw, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	dalleFlowServer = func() string { return cfg.Server }
	dalleFlowOutput = func() string { return cfg.Output }
	dalleFlowScript = func() string { return cfg.Script }
}

func dalleFlowOpts() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        dalleFlowPromptOpt,
			Description: "A prompt to generate an image from. This can be very specific.",
			Required:    true,
		},
	}
}

func handleDalleFlowCmd(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	}

	mdata, err := getInteractionMetaData(i)
	if err != nil {
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
		return
	}

	prompt := i.ApplicationCommandData().Options[0].StringValue()
	dalleFlowSubProc := exec.Command(dalleFlowScript(), "-p", prompt, "-s", dalleFlowServer(), "-o", dalleFlowOutput())
	combinedOutput, err := dalleFlowSubProc.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("%s: %s", err, combinedOutput)
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
		return
	}

	resultReader, err := os.Open(dalleFlowOutput())
	if err != nil {
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
		return
	}

	mimeType, err := mimetype.DetectFile(dalleFlowOutput())
	if err != nil {
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
		return
	}

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: fmt.Sprintf("%s\n-%s", prompt, mdata.AuthorMention),
		Files: []*discordgo.File{
			{
				Name:        fmt.Sprintf("Dalle-Flow-Output%s", mimeType.Extension()),
				ContentType: mimeType.String(),
				Reader:      resultReader,
			},
		},
		AllowedMentions: &discordgo.MessageAllowedMentions{
			Users: []string{mdata.AuthorID},
		},
	})
	if err != nil {
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
	}

	err = os.Remove(dalleFlowOutput())
	if err != nil {
		log.Error(err)
	}
}
