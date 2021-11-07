package kardbot

import (
	"fmt"
	"math"
	"regexp"
	"strconv"

	"github.com/TannerKvarfordt/Kard-bot/kardbot/dg_helpers"
	"github.com/TannerKvarfordt/imgflipgo"
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

	reservedOptCount = 3
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
				Description: "Preview the meme via DM.",
				Required:    true,
			}
		}

		if memecmd.Options[templateOptIdx] == nil {
			memecmd.Options[templateOptIdx] = &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        templateOpt,
				Description: "Select a meme template",
				Required:    true,
				Choices:     make([]*discordgo.ApplicationCommandOptionChoice, MinOf(maxDiscordOptionChoices, len(memeTemplates())-tCount)),
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
	wg := bot().updateLastActive()
	defer wg.Wait()

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Error(err)
		return
	}

	template, ok := memeTemplates()[i.ApplicationCommandData().Options[templateOptIdx].StringValue()]
	if !ok {
		log.Errorf("No template found with name %s", i.ApplicationCommandData().Options[0].Name)
		return
	}

	boxes := make([]imgflipgo.TextBox, template.BoxCount)
	includePlaceholders := false
	for argidx, arg := range i.ApplicationCommandData().Options {
		if argidx == templateOptIdx || argidx == previewOptIdx {
			continue
		}
		if arg.Name == placeholderOpt {
			includePlaceholders = arg.BoolValue()
			continue
		}
		boxIdx, err := strconv.Atoi(arg.Name)
		if err != nil {
			log.Error(err)
			return
		}
		if boxIdx >= len(boxes) {
			log.Debugf("This template has less than %d boxes, skipping...", boxIdx)
			continue
		}
		boxes[boxIdx].Text = arg.StringValue()
	}

	if len(i.ApplicationCommandData().Options) == reservedOptCount {
		for i := range boxes {
			if includePlaceholders {
				boxes[i].Text = fmt.Sprintf("Placeholder %d", i)
			} else {
				boxes[i].Text = " "
			}
		}
	} else {
		for i := range boxes {
			if boxes[i].Text == "" {
				if includePlaceholders {
					boxes[i].Text = fmt.Sprintf("Placeholder %d", i)
				} else {
					boxes[i].Text = " "
				}
			}
		}
	}

	resp, err := imgflipgo.CaptionImage(&imgflipgo.CaptionRequest{
		TemplateID: template.ID,
		Username:   getImgflipUser(),
		Password:   getImgflipPass(),
		TextBoxes:  boxes,
	})
	if err != nil {
		log.Error(err)
		return
	}

	hexColor, _ := fastHappyColorInt64()
	embed := dg_helpers.NewEmbed().
		SetColor(int(hexColor)).
		SetImage(resp.Data.URL)

	isPreview := i.ApplicationCommandData().Options[previewOptIdx].BoolValue()
	if isPreview {
		mention, err := getInteractionCreateAuthorMention(i)
		if err != nil {
			mention = "Someone"
		}

		authorID, err := getInteractionCreateAuthorID(i)
		if err != nil {
			log.Error(err)
			return
		}

		uc, err := s.UserChannelCreate(authorID)
		if err != nil {
			log.Error(err)
			return
		}

		_, err = s.ChannelMessageSendEmbed(uc.ID, embed.MessageEmbed)
		if err != nil {
			log.Error(err)
			return
		}

		_, err = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
			Content: fmt.Sprintf("%s is cooking up a meme! :D", mention),
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeUsers},
			},
		})
		if err != nil {
			log.Error(err)
		}
		return
	}

	_, err = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
		Embeds: []*discordgo.MessageEmbed{embed.MessageEmbed},
	})
	if err != nil {
		log.Error(err)
	}
}
