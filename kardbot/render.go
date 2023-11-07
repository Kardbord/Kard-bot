package kardbot

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image/jpeg"
	"image/png"
	"math/rand"
	"mime"
	"strings"

	"github.com/TannerKvarfordt/Kard-bot/kardbot/config"
	"github.com/TannerKvarfordt/gopenai/images"
	"github.com/TannerKvarfordt/gopenai/moderations"
	"github.com/TannerKvarfordt/hfapigo"
	"github.com/bwmarrin/discordgo"

	log "github.com/sirupsen/logrus"
)

const (
	renderCmd = "render"

	hfSubCmd                    = "hugging-face"
	hfPromptOpt                 = "prompt"
	hfModelOpt                  = "model"
	hfModelOptCustom            = "custom-model"
	hfModelOptNegativePrompt    = "negative-prompt"
	hfModelOptHeight            = "height-px"
	hfModelOptWidth             = "width-px"
	hfModelOptNumInferenceSteps = "num-inference-steps"
	hfModelOptGuidanceScale     = "guidance-scale"
	hfModelsFilepath            = "config/hugging-face-models.json"

	dalle2SubCmd    = "dalle2"
	dalle2OptPrompt = "prompt"
	dalle2OptSize   = "size"

	dalle3SubCmd     = "dalle3"
	dalle3OptPrompt  = "prompt"
	dalle3OptSize    = "size"
	dalle3OptQuality = "quality"
	dalle3OptStyle   = "style"
)

func handleRenderCmd(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	}

	switch i.ApplicationCommandData().Options[0].Name {
	case hfSubCmd:
		handleHfSubCmd(s, i, i.ApplicationCommandData().Options[0].Options)
	case dalle2SubCmd:
		handleDalle2SubCmd(s, i, i.ApplicationCommandData().Options[0].Options)
	case dalle3SubCmd:
		handleDalle3SubCmd(s, i, i.ApplicationCommandData().Options[0].Options)
	default:
		err = fmt.Errorf("reached unreachable case")
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
	}
}

var hfModels = func() []*discordgo.ApplicationCommandOptionChoice { return nil }

// A mapping of model names to keywords for that model
var hfModelKeyWords = func() map[string][]string { return nil }

func init() {
	cfg := struct {
		// A map of model names to activation words for the model
		Models map[string][]string `json:"models"`
	}{}

	jsonCfg, err := config.NewJsonConfig(hfModelsFilepath)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(jsonCfg.Raw, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	modelChoices := []*discordgo.ApplicationCommandOptionChoice{}
	for model := range cfg.Models {
		if strings.ToLower(model) == hfModelOptCustom {
			log.Warnf(`Custom model name "%s" conflicts with a builtin model name. It will be ignored.`, hfModelOptCustom)
			continue
		}
		modelChoices = append(modelChoices, &discordgo.ApplicationCommandOptionChoice{
			Name:  model,
			Value: model,
		})
	}

	modelChoices[len(modelChoices)-1] = &discordgo.ApplicationCommandOptionChoice{
		Name:  hfModelOptCustom,
		Value: hfModelOptCustom,
	}

	hfModels = func() []*discordgo.ApplicationCommandOptionChoice { return modelChoices }
	hfModelKeyWords = func() map[string][]string { return cfg.Models }
}

func hfOpts() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        hfModelOpt,
			Description: "The model to use when generating the image.",
			Required:    true,
			Choices:     hfModels(),
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        hfPromptOpt,
			Description: "A prompt to generate an image from. This can be very specific.",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        hfModelOptCustom,
			Description: "Any text-to-image model from huggingface.co",
			Required:    false,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        hfModelOptNegativePrompt,
			Description: "The prompt not to guide the image generation",
			Required:    false,
		},
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        hfModelOptHeight,
			Description: "Specify the height of the image, in pixels.",
			Required:    false,
		},
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        hfModelOptWidth,
			Description: "Specify the width of the image, in pixels.",
			Required:    false,
		},
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        hfModelOptNumInferenceSteps,
			Description: "Denoising steps. Higher number leads to a higher quality at the expense of performance. Default=50",
			Required:    false,
		},
		{
			Type:        discordgo.ApplicationCommandOptionNumber,
			Name:        hfModelOptGuidanceScale,
			Description: "Higher numbers produce images more closely linked to the prompt, but lowers quality. Default=7.5",
			Required:    false,
		},
	}
}

func handleHfSubCmd(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) {
	t2imgRequest := hfapigo.TextToImageRequest{}
	t2imgRequest.Options = *hfapigo.NewOptions().SetUseCache(false).SetWaitForModel(true)
	model := ""
	customModel := ""
	for _, opt := range opts {
		switch opt.Name {
		case hfPromptOpt:
			t2imgRequest.Inputs = opt.StringValue()
		case hfModelOptNegativePrompt:
			t2imgRequest.Parameters.NegativePrompt = opt.StringValue()
		case hfModelOptHeight:
			t2imgRequest.Parameters.Height = opt.IntValue()
		case hfModelOptWidth:
			t2imgRequest.Parameters.Width = opt.IntValue()
		case hfModelOptNumInferenceSteps:
			t2imgRequest.Parameters.NumInferenceSteps = opt.IntValue()
		case hfModelOptGuidanceScale:
			t2imgRequest.Parameters.GuidanceScale = opt.FloatValue()
		case hfModelOpt:
			model = opt.StringValue()
		case hfModelOptCustom:
			customModel = opt.StringValue()
		default:
			log.Warn("Unknown option:", opt.Name)
		}
	}

	if model == hfModelOptCustom {
		if customModel == "" {
			interactionFollowUpEphemeralError(s, i, false, fmt.Errorf(`you must specify a custom model to use when selecting the "%s" model`, hfModelOptCustom))
			return
		}
		model = customModel
	}
	modelKeyWords := hfModelKeyWords()[model]
	if len(modelKeyWords) == 0 {
		modelKeyWords = append(modelKeyWords, "")
	}
	unalteredInput := t2imgRequest.Inputs
	t2imgRequest.Inputs = fmt.Sprintf("%s%s", modelKeyWords[rand.Intn(len(modelKeyWords))], t2imgRequest.Inputs)

	img, imgFmt, err := hfapigo.SendTextToImageRequest(model, &t2imgRequest)
	if err != nil {
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, false, err)
		return
	}

	imgMimeType := mime.TypeByExtension(fmt.Sprintf(".%s", imgFmt))
	buf := new(bytes.Buffer)
	switch imgMimeType {
	case "image/jpeg":
		err = jpeg.Encode(buf, img, &jpeg.Options{
			Quality: 100,
		})
	case "image/png":
		err = png.Encode(buf, img)
	default:
		err = fmt.Errorf("unsupported image type (%s) returned", imgFmt)
	}
	if err != nil {
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
		return
	}

	content := fmt.Sprintf("> %s\n\nImage generated using [%s](<https://huggingface.co/%s>).", unalteredInput, model, model)
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
		Files: []*discordgo.File{
			{
				Name:        fmt.Sprintf("HuggingFaceImg.%s", imgFmt),
				ContentType: imgMimeType,
				Reader:      buf,
			},
		},
	})
	if err != nil {
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
	}
}

func dalle2Opts() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        dalle2OptPrompt,
			Description: "A prompt to generate an image from. This can be very specific.",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        dalle2OptSize,
			Description: "The size of the image to be generated",
			Required:    true,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{
					Name:  images.Dalle2SmallImage,
					Value: images.Dalle2SmallImage,
				},
				{
					Name:  images.Dalle2MediumImage,
					Value: images.Dalle2MediumImage,
				},
				{
					Name:  images.Dalle2LargeImage,
					Value: images.Dalle2LargeImage,
				},
			},
		},
	}
}

func handleDalle2SubCmd(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) {
	mdata, err := getInteractionMetaData(i)
	if err != nil {
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
		return
	}

	prompt := opts[0].StringValue()
	size := opts[1].StringValue()
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
		} else if strings.Contains(err.Error(), "safety system") {
			interactionFollowUpEphemeralError(s, i, false, err)
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

	content := fmt.Sprintf("> %s\n\nImage generated using [DALL·E 2](<https://openai.com/dall-e-2/>).", prompt)
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
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

func dalle3Opts() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        dalle3OptPrompt,
			Description: "A prompt to generate an image from. This can be very detailed.",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        dalle3OptSize,
			Description: "The size of the image to be generated.",
			Required:    true,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{
					Name:  images.Dalle3SquareImage,
					Value: images.Dalle3SquareImage,
				},
				{
					Name:  images.Dalle3LandscapeImage,
					Value: images.Dalle3LandscapeImage,
				},
				{
					Name:  images.Dalle3PortraitImage,
					Value: images.Dalle3PortraitImage,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        dalle3OptQuality,
			Description: "The quality of the image that will be generated.",
			Required:    false,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{
					Name:  images.QualityStandard,
					Value: images.QualityStandard,
				},
				{
					Name:  images.QualityHD,
					Value: images.QualityHD,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        dalle3OptStyle,
			Description: "Vivid produces more hyper-real images, natural produces less hyper-real images.",
			Required:    false,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{
					Name:  images.StyleVivid,
					Value: images.StyleVivid,
				},
				{
					Name:  images.StyleNatural,
					Value: images.StyleNatural,
				},
			},
		},
	}
}

func handleDalle3SubCmd(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) {
	mdata, err := getInteractionMetaData(i)
	if err != nil {
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
		return
	}

	imageCount := uint64(1)
	request := &images.CreationRequest{
		Model:          images.ModelDalle3,
		N:              &imageCount,
		ResponseFormat: images.ResponseFormatB64JSON,
		User:           mdata.AuthorID,
	}
	for _, opt := range opts {
		switch opt.Name {
		case dalle3OptPrompt:
			request.Prompt = opt.StringValue()
		case dalle3OptSize:
			request.Size = opt.StringValue()
		case dalle3OptQuality:
			request.Quality = opt.StringValue()
		case dalle3OptStyle:
			request.Style = opt.StringValue()
		default:
			log.Warn("Unknown option:", opt.Name)
		}
	}

	resp, modr, err := images.MakeModeratedCreationRequest(request, nil)
	if err != nil {
		targetErr := moderations.NewModerationFlagError()
		if errors.As(err, &targetErr) {
			contentFlags, err := json.MarshalIndent(modr.Results[0].Categories, "", "  ")
			if err != nil {
				log.Error(err)
				contentFlags = []byte("Whoops, couldn't retrieve the details of your violation.")
			}
			interactionFollowUpEphemeralError(s, i, false, fmt.Errorf("sorry! Your prompt does not appear to conform to [Open AI's Usage Policies](<https://beta.openai.com/docs/usage-policies>)\n```JSON\n%s\n```", contentFlags))
		} else if strings.Contains(err.Error(), "safety system") {
			interactionFollowUpEphemeralError(s, i, false, err)
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

	content := fmt.Sprintf("Prompt:\n> %s\n\nRevised Prompt:\n> %s\n\nImage generated using [DALL·E 3](<https://openai.com/dall-e-3/>).", request.Prompt, resp.Data[0].RevisedPrompt)
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
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
