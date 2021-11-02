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

	previewOpt = "preview"
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
	memecmd := &discordgo.ApplicationCommand{
		Options: make([]*discordgo.ApplicationCommandOption, MinOf(maxDiscordCommandOptions, len(memeTemplates()))),
	}
	tCount := 0 // template counter
	exceededCmds := false
	for _, template := range memeTemplates() {
		if template.BoxCount > maxDiscordCommandOptions {
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
			memecmd = &discordgo.ApplicationCommand{
				Options: make([]*discordgo.ApplicationCommandOption, MinOf(maxDiscordCommandOptions, len(memeTemplates())-tCount)),
			}
		}

		subcmdIdx := tCount % maxDiscordCommandOptions
		if subcmdIdx >= len(memecmd.Options) {
			log.Fatalf("Attempted to index out of range. Valid range was %d for memecmd %s. Tried to index %d", len(memecmd.Options), memeCommand, subcmdIdx)
		}

		memecmd.Options[subcmdIdx] = &discordgo.ApplicationCommandOption{}
		subcmd := memecmd.Options[subcmdIdx]
		subcmd.Type = discordgo.ApplicationCommandOptionSubCommand
		subcmd.Name = template.ID
		subcmd.Description = fmt.Sprint(template.Name)
		subcmd.Options = make([]*discordgo.ApplicationCommandOption, template.BoxCount+1)
		subcmd.Options[0] = &discordgo.ApplicationCommandOption{
			Type:        discordgo.ApplicationCommandOptionBoolean,
			Name:        previewOpt,
			Description: "DM's you this meme so you can preview it before sending it to the channel.",
			Required:    true,
		}
		for i := 1; i < len(subcmd.Options); i++ {
			subcmd.Options[i] = &discordgo.ApplicationCommandOption{}
			subcmd.Options[i].Type = discordgo.ApplicationCommandOptionString
			subcmd.Options[i].Name = fmt.Sprint(i)
			subcmd.Options[i].Description = fmt.Sprintf("Text for box %d", i)
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

	template, ok := memeTemplates()[i.ApplicationCommandData().Options[0].Name]
	if !ok {
		log.Errorf("No template found with name %s", i.ApplicationCommandData().Options[0].Name)
		return
	}

	boxes := make([]imgflipgo.TextBox, template.BoxCount)
	isPreview := false
	for _, arg := range i.ApplicationCommandData().Options[0].Options {
		if isNumericRegex().MatchString(arg.Name) {
			boxIdx, err := strconv.Atoi(arg.Name)
			if err != nil {
				log.Error(err)
				return
			}
			if boxIdx >= len(boxes) {
				log.Errorf("Index %d exceeds number of expected boxes", boxIdx)
			}
			boxes[boxIdx].Text = arg.StringValue()
		} else if arg.Name == previewOpt {
			isPreview = true
		} else {
			log.Errorf("Unknown argument: %s", arg.Name)
			return
		}
	}

	switch len(i.ApplicationCommandData().Options[0].Options) {
	case 0:
		fallthrough
	case 1:
		for i := range boxes {
			boxes[i].Text = fmt.Sprintf("Placeholder %d", i)
		}
	default:
		for i := range boxes {
			if i != 0 && boxes[i].Text == "" {
				boxes[i].Text = fmt.Sprintf("Placeholder %d", i)
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

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("%s is cooking up a meme! :D", mention),
				AllowedMentions: &discordgo.MessageAllowedMentions{
					Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeUsers},
				},
			},
		})
		if err != nil {
			log.Error(err)
		}
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed.MessageEmbed},
		},
	})
	if err != nil {
		log.Error(err)
	}
}
