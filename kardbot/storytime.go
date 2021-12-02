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

// Configuration parameters that will be passed to the HF API
// See https://api-inference.huggingface.co/docs/python/html/detailed_parameters.html#text-generation-task
type storyTimeConfig struct {
	TextGenModel      string   `json:"story-time-text-generation-model,omitempty"`
	TopK              *int     `json:"story-time-topK,omitempty"`
	TopP              *float64 `json:"story-time-topP,omitempty"`
	Temperature       *float64 `json:"story-time-temperature,omitempty"`
	RepetitionPenalty *float64 `json:"story-time-repetition-penalty,omitempty"`
	MaxNewTokens      *int     `json:"story-time-max-new-tokens,omitempty"`
	MaxTime           *float64 `json:"story-time-max-time-s,omitempty"`
}

var storyTimeCfg func() storyTimeConfig

func init() {
	cfg := storyTimeConfig{}
	jsonCfg, err := config.NewJsonConfig(StoryTimeConfigFile)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(jsonCfg.Raw, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.TextGenModel == "" {
		log.Fatal("No story time text generation model specified")
	}

	resp, err := http.Get(hfapigo.APIBaseURL + cfg.TextGenModel)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatal("invalid text generation model:", http.StatusText(resp.StatusCode))
	}

	storyTimeCfg = func() storyTimeConfig {
		var (
			topK       *int     = nil
			topP       *float64 = nil
			temp       *float64 = nil
			repPen     *float64 = nil
			maxNewToks *int     = nil
			maxTime    *float64 = nil
		)

		// Make copies of pointer values so that they can't be accidentally modified
		if cfg.TopK != nil {
			tmpTopK := *cfg.TopK
			topK = &tmpTopK
		}
		if cfg.TopP != nil {
			tmpTopP := *cfg.TopP
			topP = &tmpTopP
		}
		if cfg.Temperature != nil {
			tmpTemp := *cfg.Temperature
			temp = &tmpTemp
		}
		if cfg.RepetitionPenalty != nil {
			tmpRepPen := *cfg.RepetitionPenalty
			repPen = &tmpRepPen
		}
		if cfg.MaxNewTokens != nil {
			tmpMaxNewToks := *cfg.MaxNewTokens
			maxNewToks = &tmpMaxNewToks
		}
		if cfg.MaxTime != nil {
			tmpMaxTime := *cfg.MaxTime
			maxTime = &tmpMaxTime
		}

		return storyTimeConfig{
			TextGenModel:      cfg.TextGenModel,
			TopK:              topK,
			TopP:              topP,
			Temperature:       temp,
			RepetitionPenalty: repPen,
			MaxNewTokens:      maxNewToks,
			MaxTime:           maxTime,
		}
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

	cfg := storyTimeCfg()
	textResps, err := hfapigo.SendTextGenerationRequest(cfg.TextGenModel, &hfapigo.TextGenerationRequest{
		Inputs:  []string{i.ApplicationCommandData().Options[0].StringValue()},
		Options: *hfapigo.NewOptions().SetWaitForModel(true).SetUseCache(false),
		Parameters: *(&hfapigo.TextGenerationParameters{
			TopK:              cfg.TopK,
			TopP:              cfg.TopP,
			Temperature:       cfg.Temperature,
			RepetitionPenalty: cfg.RepetitionPenalty,
			MaxNewTokens:      cfg.MaxNewTokens,
			MaxTime:           cfg.MaxTime,
		}).SetReturnFullText(true),
	})
	if err != nil {
		log.Error(err)
		return
	}
	if len(textResps) == 0 || len(textResps[0].GeneratedTexts) == 0 || textResps[0].GeneratedTexts[0] == "" {
		log.Error("Received no text generation responses")
		return
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
