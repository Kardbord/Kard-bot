package kardbot

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/TannerKvarfordt/Kard-bot/kardbot/config"
	"github.com/TannerKvarfordt/hfapigo"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

const (
	madlibCmd        = "madlib"
	madlibBlank      = "<>"
	madlibConfigFile = "config/madlib.json"
)

type madlibConfig struct {
	Model     string `json:"model,omitempty"`
	ModelMask string `json:"model-mask,omitempty"`
}

var madlibCfg = madlibConfig{
	Model:     "roberta-base", // https://huggingface.co/roberta-base
	ModelMask: "<mask>",
}

func init() {
	jsonCfg, err := config.NewJsonConfig(madlibConfigFile)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(jsonCfg.Raw, &madlibCfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Madlib using %s with mask=%s", madlibCfg.Model, madlibCfg.ModelMask)
}

func handleMadLibCmd(s *discordgo.Session, i *discordgo.InteractionCreate) {
	wg := bot().updateLastActive()
	defer wg.Wait()

	if isSelf, err := authorIsSelf(s, i); err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	} else if isSelf {
		log.Trace("Ignoring message from self")
		return
	}

	if !strings.Contains(i.ApplicationCommandData().Options[0].StringValue(), madlibBlank) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("No blanks provided. Provide a prompt string containing at least one of the following mask: `%s`.\nFor example: `The quick brown %s jumps over the lazy %s.`", madlibBlank, madlibBlank, madlibBlank),
				Flags:   InteractionResponseFlagEphemeral,
			},
		})
		return
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	}

	input := strings.ReplaceAll(i.ApplicationCommandData().Options[0].StringValue(), madlibBlank, madlibCfg.ModelMask)
	resp, err := hfapigo.SendFillMaskRequest(madlibCfg.Model, &hfapigo.FillMaskRequest{
		Inputs:  []string{input},
		Options: *hfapigo.NewOptions().SetWaitForModel(true),
	})
	if err != nil {
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
		return
	}

	if len(resp) < strings.Count(i.ApplicationCommandData().Options[0].StringValue(), madlibBlank) {
		err := fmt.Errorf("too few responses received")
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
		return
	}

	output := input
	for _, mask := range resp {
		if len(mask.Masks) == 0 {
			err := fmt.Errorf("received empty response")
			log.Error(err)
			interactionFollowUpEphemeralError(s, i, true, err)
			return
		}
		output = strings.Replace(output, madlibCfg.ModelMask, strings.TrimSpace(mask.Masks[0].TokenStr), 1)
	}

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: output,
	})
	if err != nil {
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
	}
}
