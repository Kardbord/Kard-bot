package kardbot

import (
	"encoding/json"
	"net/http"

	"github.com/TannerKvarfordt/Kard-bot/kardbot/config"
	"github.com/TannerKvarfordt/hfapigo"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

const (
	storyTimeCmd        = "story-time"
	StoryTimeConfigFile = "config/storytime.json"
)

var storyTimeTextGenModel = func() string {
	return hfapigo.RecommendedTextGenerationModel
}

func init() {
	cfg := struct {
		StoryTimeTextGenModel string `json:"story-time-text-generation-model,omitempty"`
	}{}

	jsonCfg, err := config.NewJsonConfig(StoryTimeConfigFile)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(jsonCfg.Raw, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	storyTimeTextGenModel = func() string { return cfg.StoryTimeTextGenModel }

	resp, err := http.Get(hfapigo.APIBaseURL + storyTimeTextGenModel())
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatal("invalid text generation model:", http.StatusText(resp.StatusCode))
	}
}

func storyTime(s *discordgo.Session, i *discordgo.InteractionCreate) {
	wg := bot().updateLastActive()
	defer wg.Wait()
	if isSelf, err := authorIsSelf(s, i); err != nil {
		log.Error(err)
		return
	} else if isSelf {
		log.Trace("Ignoring message from self")
		return
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Error(err)
		return
	}

	textResps, err := hfapigo.SendTextGenerationRequest(storyTimeTextGenModel(), &hfapigo.TextGenerationRequest{
		Inputs:     []string{i.ApplicationCommandData().Options[0].StringValue()},
		Parameters: *hfapigo.NewTextGenerationParameters().SetReturnFullText(true).SetMaxNewTokens(250),
		Options:    *hfapigo.NewOptions().SetWaitForModel(true).SetUseCache(false),
	})
	if err != nil {
		log.Error(err)
		return
	}
	if len(textResps) == 0 || len(textResps[0].GeneratedTexts) == 0 || textResps[0].GeneratedTexts[0] == "" {
		log.Error("Received no text generation responses")
	}

	_, err = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
		Content: firstN(textResps[0].GeneratedTexts[0], int(MaxDiscordMsgLen)),
		AllowedMentions: &discordgo.MessageAllowedMentions{
			Parse: []discordgo.AllowedMentionType{
				discordgo.AllowedMentionTypeUsers,
				discordgo.AllowedMentionTypeRoles,
				discordgo.AllowedMentionTypeEveryone,
			},
		},
	})
	if err != nil {
		log.Error(err)
	}
}
