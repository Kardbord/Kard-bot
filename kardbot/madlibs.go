package kardbot

import (
	"fmt"
	"strings"

	"github.com/TannerKvarfordt/hfapigo"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

const (
	madlibCmd            = "madlib"
	madlibBlank          = "<>"
	madlibModel          = "roberta-base" // https://huggingface.co/bert-base-multilingual-cased
	madlibModelMaskToken = "<mask>"
)

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

	input := strings.ReplaceAll(i.ApplicationCommandData().Options[0].StringValue(), madlibBlank, madlibModelMaskToken)
	resp, err := hfapigo.SendFillMaskRequest(madlibModel, &hfapigo.FillMaskRequest{
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
		output = strings.Replace(output, madlibModelMaskToken, strings.TrimSpace(mask.Masks[0].TokenStr), 1)
	}

	_, err = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
		Content: output,
	})
	if err != nil {
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
	}
}
