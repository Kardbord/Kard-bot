package kardbot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Kardbord/Kard-bot/kardbot/config"
	"github.com/Kardbord/Kard-bot/kardbot/dg_helpers"
	"github.com/Kardbord/hfapigo/v3"
	"github.com/Kardbord/ubiquity/stringutils"
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

	err = validateStoryTimeModel(cfg.TextGenModel)
	if err != nil {
		log.Fatal(err)
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

const (
	storyTimePromptOpt = "prompt"
	storyTimeModelOpt  = "model"
	storyTimeHelpOpt   = "help"
)

func storyTime(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if isSelf, err := authorIsSelf(s, i); err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	} else if isSelf {
		log.Trace("Ignoring message from self")
		return
	}

	cfg := storyTimeCfg()
	model := cfg.TextGenModel
	input := ""
	for _, opt := range i.ApplicationCommandData().Options {
		switch opt.Name {
		case storyTimePromptOpt:
			input = opt.StringValue()
		case storyTimeModelOpt:
			model = opt.StringValue()
		case storyTimeHelpOpt:
			if !opt.BoolValue() {
				continue
			}
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{buildStoryTimeHelpEmbed()},
					Flags:  discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				log.Error(err)
				interactionRespondEphemeralError(s, i, true, err)
			}
			return
		default:
			log.Warn("Unknown option: ", opt.Name)
		}
	}

	err := validateStoryTimeModel(model)
	if err != nil {
		err2 := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("%s is not a valid model", model),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err2 != nil {
			log.Error(err)
			interactionRespondEphemeralError(s, i, true, err2)
		}
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	}

	textResps, err := hfapigo.SendTextGenerationRequest(model, &hfapigo.TextGenerationRequest{
		Input:   input,
		Options: *hfapigo.NewOptions().SetWaitForModel(true).SetUseCache(false),
		Parameters: *(&hfapigo.TextGenerationParameters{
			TopK:              cfg.TopK,
			TopP:              cfg.TopP,
			Temperature:       cfg.Temperature,
			RepetitionPenalty: cfg.RepetitionPenalty,
			MaxNewTokens:      cfg.MaxNewTokens,
		}).SetReturnFullText(true),
	})
	if err != nil {
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
		return
	}
	if len(textResps) == 0 || len(textResps[0].GeneratedText) == 0 || textResps[0].GeneratedText == "" {
		err = fmt.Errorf("received no text generation responses")
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
		return
	}

	content := stringutils.FirstN(textResps[0].GeneratedText, MaxDiscordMsgLen)
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
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
		interactionFollowUpEphemeralError(s, i, true, err)
	}
}

func validateStoryTimeModel(model string) error {
	resp, err := http.Get(hfapigo.APIBaseURL + model)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid text generation model: %d - %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	type modelData struct {
		PipelineTag string `json:"pipeline_tag,omitempty"`
	}

	md := modelData{}
	err = json.Unmarshal(respBody, &md)
	if err != nil {
		return err
	}

	if md.PipelineTag != "text-generation" {
		return fmt.Errorf("%s is not a valid text generation model", model)
	}

	return nil
}

func buildStoryTimeHelpEmbed() *discordgo.MessageEmbed {
	optsHelp := fmt.Sprintf(
		"• **%s (Optional)**: A prompt to generate a story from.\n"+
			"• **%s (Optional)**: The text generation AI model to use. For example, _gpt2_ or _EleutherAI/gpt-neo-125M_. "+
			"For other options, see the [Hugging Face Model Repository](https://huggingface.co/models?pipeline_tag=text-generation).",
		storyTimePromptOpt, storyTimeModelOpt)

	about := fmt.Sprintf("Most available text-generation AI models are next-word prediction models. Think of your phone trying to predict your next word as you type, "+
		"except it isn't familiar with your personal typing patterns. The stories generated here are not intended (or likely) to be sensical or good, but it's fun to see what the bot comes up with. "+
		"The bot is probably going to stray from your prompt to a large degree. Some models will probably be better than others at staying on topic. The default model ([%s](https://huggingface.co/%s)) "+
		"was selected with this in mind, but there may be a better one out there. Feel free to experiment!", storyTimeCfg().TextGenModel, storyTimeCfg().TextGenModel)

	tips := "For best results, you'll want to provide a sensical prompt at least a few words long. The longer the prompt, the more likely the bot will stay at least somewhat on topic. " +
		"Usually about a single sentence-length prompt will do fairly well. Another trick is to leave the last sentence of your prompt incomplete, and let the bot finish it for you. " +
		"Below are a few examples of the kind of prompts that will do well.\n" +
		"\n• Ogres are like\n" +
		"• What makes you think she is a witch? Well, she turned me into a\n" +
		"• I'll build my own lunar lander! With blackjack, and\n" +
		"\nOf course, your mileage may vary."

	return dg_helpers.NewEmbed().
		SetTitle(fmt.Sprintf("`/%s` Help", storyTimeCmd)).
		AddField("Options", optsHelp).
		AddField("About", about).
		AddField("Usage Tips", tips).
		Truncate().
		MessageEmbed
}
