package kardbot

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/TannerKvarfordt/gopenai/images"
	"github.com/TannerKvarfordt/gopenai/moderations"
	"github.com/bwmarrin/discordgo"

	log "github.com/sirupsen/logrus"
)

const (
	dalle2Cmd       = "dalle2"
	dalle2PromptOpt = "prompt"
	dalle2SizeOpt   = "size"
)

func dalle2Opts() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        dalle2PromptOpt,
			Description: "A prompt to generate an image from. This can be very specific.",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        dalle2SizeOpt,
			Description: "The size of the image to be generated",
			Required:    true,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{
					Name:  images.SmallImage,
					Value: images.SmallImage,
				},
				{
					Name:  images.MediumImage,
					Value: images.MediumImage,
				},
				{
					Name:  images.LargeImage,
					Value: images.LargeImage,
				},
			},
		},
	}
}

func handleDalle2Cmd(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
	size := i.ApplicationCommandData().Options[1].StringValue()
	imageCount := uint64(1)
	resp, modr, err := images.MakeModeratedCreationRequest(&images.CreationRequest{
		Prompt:         prompt,
		N:              &imageCount,
		Size:           size,
		ResponseFormat: images.ResponseFormatB64JSON,
		User:           mdata.AuthorID,
	}, nil)
	if err != nil {
		targetErr := moderations.NewModerationFlagError()
		if errors.As(err, &targetErr) {
			contentFlags, err := json.MarshalIndent(modr.Results[0].Categories, "", "  ")
			if err != nil {
				log.Error(err)
				contentFlags = []byte("Whoops, couldn't retrieve the details of your violation.")
			}
			interactionFollowUpEphemeralError(s, i, false, fmt.Errorf("sorry! Your prompt does not appear to conform to [Open AI's Usage Policies](<https://beta.openai.com/docs/usage-policies>)\n```JSON\n%s\n```", contentFlags))
		} else {
			log.Error(err)
			interactionFollowUpEphemeralError(s, i, true, err)
		}
		return
	}

	unbased, err := base64.StdEncoding.DecodeString(resp.Data[0].B64JSON)
	if err != nil {
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
		return
	}

	errMsg := fmt.Sprintf("> %s\n\nImage generated using [DALL·E 2](<https://openai.com/dall-e-2/>).", prompt)
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &errMsg,
		Files: []*discordgo.File{
			{
				Name:        "Dalle-2-Output.png",
				ContentType: "image/png",
				Reader:      bytes.NewReader(unbased),
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
}
