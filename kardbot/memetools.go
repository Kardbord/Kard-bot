package kardbot

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"

	"github.com/Kardbord/Kard-bot/kardbot/dg_helpers"
	"github.com/Kardbord/imgflipgo/v2"
	"github.com/Kardbord/ubiquity/mathutils"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

const (
	memeCommand     = "build-a-meme"
	maxMemeCommands = 4

	previewOpt    = "preview"
	previewOptIdx = 1

	templateOpt    = "template"
	templateOptIdx = 0

	placeholderOpt     = "placeholders"
	placeholderOptReqd = false
	placeholderOptIdx  = 2 // as this is not a required opt, this index is only guaranteed when registering the command, not when handling it

	maxFontSizeOpt     = "max-font-size-px"
	maxFontSizeOptReqd = false
	maxFontSizeOptIdx  = 3 // as this is not a required opt, this index is only guaranteed when registering the command, not when handling it

	fontOpt     = "font"
	fontOptReqd = false
	fontOptIdx  = 4 // as this is not a required opt, this index is only guaranteed when registering the command, not when handling it

	reservedOptCount = 5
)

var memeCommandRegex = func() *regexp.Regexp { return nil }

func init() {
	r := regexp.MustCompile(fmt.Sprintf(`%s\d*$`, memeCommand))
	if r == nil {
		log.Fatal("Could not compile memeCommandRegex")
	}
	memeCommandRegex = func() *regexp.Regexp {
		return r
	}
}

func strMatchesMemeCmdPattern(str string) bool {
	return memeCommandRegex().MatchString(str)
}

var (
	// Meme.ID to Meme mapping
	memeTemplates func() map[string]imgflipgo.Meme

	memeCommands func() []*discordgo.ApplicationCommand
)

func init() {
	memes, err := imgflipgo.GetMemes()
	if err != nil {
		log.Fatal(err)
	}

	memeMap := make(map[string]imgflipgo.Meme, len(memes))
	for _, meme := range memes {
		if _, ok := memeMap[meme.ID]; ok {
			log.Warnf(`Meme name conflict! %s already exists and will be skipped`, meme.Name)
			continue
		}
		memeMap[meme.ID] = meme
	}

	memeTemplates = func() map[string]imgflipgo.Meme {
		return memeMap
	}
	if len(memeTemplates()) == 0 {
		log.Warn("No meme templates initialized")
	}

	memecmds := buildMemeCommands()

	memeCommands = func() []*discordgo.ApplicationCommand {
		return memecmds
	}
}

func buildMemeCommands() []*discordgo.ApplicationCommand {
	allcmds := []*discordgo.ApplicationCommand{}

	newCmd := func() *discordgo.ApplicationCommand {
		newcmd := &discordgo.ApplicationCommand{
			Options: make([]*discordgo.ApplicationCommandOption, maxDiscordCommandOptions),
		}
		for i := reservedOptCount; i < maxDiscordCommandOptions; i++ {
			newcmd.Options[i] = &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        fmt.Sprint(i - reservedOptCount),
				Description: fmt.Sprintf("Text box %d", i-reservedOptCount),
				Required:    false,
			}
		}
		return newcmd
	}

	memecmd := newCmd()
	tCount := 0 // template counter
	exceededCmds := false
	for _, template := range memeTemplates() {
		if template.BoxCount > maxDiscordCommandOptions-reservedOptCount {
			log.Infof("Skipping %s as it has too many text boxes (%d)", template.Name, template.BoxCount)
			continue
		}
		cmdItr := int(math.Ceil(float64(tCount) / float64(maxDiscordCommandOptions)))
		if cmdItr > maxMemeCommands {
			exceededCmds = true
			log.Warnf("We have exceeded the set limit of %d %s commands. Not all templates will be available.", maxMemeCommands, memeCommand)
			break
		}
		if tCount != 0 && tCount%maxDiscordCommandOptions == 0 {
			memecmd.Name = fmt.Sprintf("%s%d", memeCommand, cmdItr)
			memecmd.Description = fmt.Sprintf("Create a meme using templates %d through %d", tCount-maxDiscordCommandOptions+1, tCount)
			allcmds = append(allcmds, memecmd)
			memecmd = newCmd()
		}

		if memecmd.Options[previewOptIdx] == nil {
			memecmd.Options[previewOptIdx] = &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        previewOpt,
				Description: "Preview the meme (only you will be able to see it).",
				Required:    true,
			}
		}

		if memecmd.Options[templateOptIdx] == nil {
			memecmd.Options[templateOptIdx] = &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        templateOpt,
				Description: "Select a meme template",
				Required:    true,
				Choices:     make([]*discordgo.ApplicationCommandOptionChoice, mathutils.Min(maxDiscordOptionChoices, len(memeTemplates())-tCount)),
			}
		}

		if memecmd.Options[placeholderOptIdx] == nil {
			memecmd.Options[placeholderOptIdx] = &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        placeholderOpt,
				Description: "Include placeholder text?",
				Required:    placeholderOptReqd,
			}
		}

		if memecmd.Options[maxFontSizeOptIdx] == nil {
			memecmd.Options[maxFontSizeOptIdx] = &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        maxFontSizeOpt,
				Description: "Max font size to use, in pixels.",
				Required:    maxFontSizeOptReqd,
			}
		}

		if memecmd.Options[fontOptIdx] == nil {
			memecmd.Options[fontOptIdx] = &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        fontOpt,
				Description: "Select a font to use",
				Required:    fontOptReqd,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  string(imgflipgo.FontArial),
						Value: imgflipgo.FontArial,
					},
					{
						Name:  string(imgflipgo.FontImpact),
						Value: imgflipgo.FontImpact,
					},
				},
			}
		}

		choiceIdx := tCount % maxDiscordOptionChoices
		memecmd.Options[templateOptIdx].Choices[choiceIdx] = &discordgo.ApplicationCommandOptionChoice{
			Name:  template.Name,
			Value: template.ID,
		}

		tCount++
	}
	if tCount != 0 && !exceededCmds {
		cmdItr := int(math.Ceil(float64(tCount) / float64(maxDiscordCommandOptions)))
		memecmd.Name = fmt.Sprintf("%s%d", memeCommand, cmdItr)
		memecmd.Description = fmt.Sprintf("Create a meme using templates %d through %d", tCount-maxDiscordCommandOptions+1, tCount)
		allcmds = append(allcmds, memecmd)
	}

	return allcmds
}

func buildAMeme(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var flags discordgo.MessageFlags = 0
	isPreview := i.ApplicationCommandData().Options[previewOptIdx].BoolValue()
	if isPreview {
		flags = discordgo.MessageFlagsEphemeral
	}
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: flags,
		},
	})
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	}

	template, ok := memeTemplates()[i.ApplicationCommandData().Options[templateOptIdx].StringValue()]
	if !ok {
		errmsg := fmt.Sprintf("Error! No template found with name %s", i.ApplicationCommandData().Options[0].Name)
		log.Error(errmsg)
		interactionFollowUpEphemeralError(s, i, true, errors.New(errmsg))
		return
	}

	boxes := make([]imgflipgo.TextBox, template.BoxCount)
	includePlaceholders := false
	var maxFontSize *uint = nil
	var font *imgflipgo.Font = nil
	for _, arg := range i.ApplicationCommandData().Options {
		switch arg.Name {
		case templateOpt:
			continue
		case previewOpt:
			continue
		case placeholderOpt:
			includePlaceholders = arg.BoolValue()
		case maxFontSizeOpt:
			mfs := int(arg.IntValue())
			if mfs > 0 {
				tmp := uint(mfs)
				maxFontSize = &tmp
			} else {
				log.Warn("font size cannot be <= 0")
				maxFontSize = nil
			}
		case fontOpt:
			f := imgflipgo.Font(arg.StringValue())
			font = &f
		default:
			boxIdx, err := strconv.Atoi(arg.Name)
			if err != nil {
				log.Error(err)
				interactionFollowUpEphemeralError(s, i, true, err)
				return
			}
			if boxIdx >= len(boxes) {
				log.Debugf("This template has less than %d boxes, skipping...", boxIdx)
				continue
			}
			boxes[boxIdx].Text = arg.StringValue()
		}
	}

	for i := range boxes {
		if boxes[i].Text == "" {
			if includePlaceholders {
				boxes[i].Text = fmt.Sprintf("Placeholder %d", i)
			} else {
				boxes[i].Text = " "
			}
		}
	}

	if font != nil {
		log.Debugf("Using %s", *font)
	}
	resp, err := imgflipgo.CaptionImage(&imgflipgo.CaptionRequest{
		TemplateID:    template.ID,
		Username:      getImgflipUser(),
		Password:      getImgflipPass(),
		MaxFontSizePx: maxFontSize,
		Font:          font,
		TextBoxes:     boxes,
	})
	if err != nil {
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
		return
	}

	hexColor, _ := fastHappyColorInt64()
	embed := dg_helpers.NewEmbed().
		SetColor(int(hexColor)).
		SetImage(resp.Data.URL)

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed.MessageEmbed},
	})
	if err != nil {
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
	}
}
