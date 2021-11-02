package kardbot

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/TannerKvarfordt/imgflipgo"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

const (
	memeCommand     = "build-a-meme"
	maxMemeCommands = 4
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
	// template.ID to Meme mapping
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
		meme.Name = strings.ToLower(whiteSpaceRegexp().ReplaceAllString(meme.Name, "-"))
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

		subcmd := memecmd.Options[tCount%maxDiscordCommandOptions]
		subcmd.Type = discordgo.ApplicationCommandOptionSubCommand
		subcmd.Name = template.Name
		subcmd.Description = fmt.Sprintf("Create a meme using the %s template", template.Name)
		subcmd.Options = make([]*discordgo.ApplicationCommandOption, template.BoxCount+1)
		subcmd.Options[0] = &discordgo.ApplicationCommandOption{
			Type:        discordgo.ApplicationCommandOptionBoolean,
			Name:        "preview",
			Description: "DM's you this meme so you can preview it before sending it to the channel.",
		}
		for i, textOpt := range subcmd.Options {
			textOpt.Type = discordgo.ApplicationCommandOptionString
			textOpt.Name = fmt.Sprint(i + 1)
			textOpt.Description = fmt.Sprintf("Text for box %d", i+1)
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

func buildAMeme(*discordgo.Session, *discordgo.InteractionCreate) {

}
