package kardbot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

type onInteractionHandler func(*discordgo.Session, *discordgo.InteractionCreate)

// The max number of command options any single command is allowed to have.
const (
	maxDiscordCommandOptions = 25
	maxDiscordOptionChoices  = 25
)

func getCommands() []*discordgo.ApplicationCommand {
	allcmds := []*discordgo.ApplicationCommand{
		{
			Name:        "roll",
			Description: "Rolls a D{X} die {Y} times, where X and Y are provided by the user.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "dice-count",
					Description: "How many dice should be rolled?",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "dice-sides",
					Description: "How many sides on the dice? Can optionally be prefixed with 'D' or 'd'.",
					Required:    true,
				},
			},
		},
		{
			Name:        "what-are-the-odds",
			Description: "What are the odds that an event will occur? Use third-person voice for best results.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "event",
					Description: "something that you want to know the odds of",
					Required:    true,
				},
			},
		},
		{
			Name:        "loglevel",
			Description: "Update the log level of the bot. Only works for whitelisted users.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "level",
					Description: "Level at which to log.",
					Required:    true,
					Choices:     logLevelChoices(),
				},
			},
		},
		{
			Name:        "pasta",
			Description: "Serves you a delicious pasta!",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "selection",
					Description: "Your choice of delicious pasta!",
					Required:    true,
					Choices:     pastaChoices(),
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        pastaOptionTTS,
					Description: "Read the copy-pasta aloud to everyone viewing the channel using text-to-speech",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        pastaOptionUwu,
					Description: "Sprinkle a healthy dose of UwU on your pasta!",
					Required:    false,
					Choices:     uwuChoices(),
				},
			},
		},
		{
			Name:        "uwu",
			Description: "UwU-ifies messages, because why wouldn't it?",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "custom",
					Description: "UwU-ify your own custom message!",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "uwulevel",
							Description: "How badly should your message be uwu-ed?",
							Required:    true,
							Choices:     uwuChoices(),
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "message",
							Description: "Your message, soon to be uwuified!",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionBoolean,
							Name:        "tts",
							Description: "Read the uwuified copy-pasta aloud to everyone viewing the channel using text-to-speech",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "pasta",
					Description: "UwU-ify a copy-pasta!",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "uwulevel",
							Description: "How badly should your message be uwu-ed?",
							Required:    true,
							Choices:     uwuChoices(),
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "pasta",
							Description: "A pasta, but served with a healthy topping of UwU!",
							Required:    true,
							Choices:     pastaChoices(),
						},
						{
							Type:        discordgo.ApplicationCommandOptionBoolean,
							Name:        "tts",
							Description: "Read the uwuified copy-pasta aloud to everyone viewing the channel using text-to-speech",
							Required:    false,
						},
					},
				},
			},
		},
		{
			Name:        "reddit-roulette",
			Description: "Retrieve a random reddit post",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        redditRouletteSubCmdAny,
					Description: "Retrieve a random reddit post. May be NSFW.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "subreddits",
							Description: "Subreddits from which you want the random post to be retrieved, separated by spaces",
							Required:    false,
						},
					},
				},
				{
					Name:        redditRouletteSubCmdSFW,
					Description: "Retrieve a random reddit post that is not marked as NSFW",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "subreddits",
							Description: "Subreddits from which you want the random post to be retrieved, separated by spaces",
							Required:    false,
						},
					},
				},
				{
					Name:        redditRouletteSubCmdNSFW,
					Description: "Retrieve a random reddit post that is marked as NSFW.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "subreddits",
							Description: "Subreddits from which you want the random post to be retrieved, separated by spaces",
							Required:    false,
						},
					},
				},
			},
		},
		{
			Name:        "compliments",
			Description: "Receive a daily compliment!",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        complimentsGet,
					Description: "Get a compliment!",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionBoolean,
							Name:        complimentInDM,
							Description: "Get the compliment as a DM",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Name:        complimentsOptIn,
					Description: "Opt-in to daily DM compliments",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Name:        complimentsMorning,
							Description: "Opt in to daily DM compliments in the morning",
						},
						{
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Name:        complimentsEvening,
							Description: "Opt in to daily DM compliments in the evening",
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Name:        complimentsOptOut,
					Description: "Opt-out of daily DM compliments",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Name:        complimentsMorning,
							Description: "Opt out of daily DM compliments in the morning",
						},
						{
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Name:        complimentsEvening,
							Description: "Opt out of daily DM compliments in the evening",
						},
					},
				},
			},
		},
		{
			Name:        "creepy-dms",
			Description: "What do you mean you don't want to receive creepy DMs?",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        creepyDMGet,
					Description: "Get a creepy DM ASAP",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionBoolean,
							Name:        creepyChannelDM,
							Description: "Send the message to the channel instead of as a DM",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        creepyDMOptIn,
					Description: "Opt in to random creepy DMs (1 per day max)",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        creepyDMOptOut,
					Description: "Opt out of random creepy DMs",
				},
			},
		},
		{
			Name:        delBotDMCmd,
			Description: "Clear your DM history with the bot. Only works when issued from DMs.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "msg-count",
					Description: fmt.Sprintf("How many messages back should we delete? Up to %d.", msgLimit),
					Required:    true,
				},
			},
		},
		{
			Name:        storyTimeCmd,
			Description: "The bot will tell you a short story based on a given prompt.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "prompt",
					Description: "A prompt to generate a story from.",
					Required:    true,
				},
			},
		},
		{
			Name:        "help",
			Description: "Get helpful information about the bot.",
		},
	}

	// Have to get creative when adding meme template commands
	// since there is a limit to the number of options a command
	// can have.
	allcmds = append(allcmds, memeCommands()...)
	return allcmds
}

func getCommandImpls() map[string]onInteractionHandler {
	return map[string]onInteractionHandler{
		"roll":              rollDice,
		"loglevel":          updateLogLevel,
		"pasta":             servePasta,
		"reddit-roulette":   redditRoulette,
		"uwu":               uwuify,
		"compliments":       complimentHandler,
		"creepy-dms":        creepyDMHandler,
		"help":              botInfo,
		"what-are-the-odds": whatAreTheOdds,
		memeCommand:         buildAMeme,
		delBotDMCmd:         deleteBotDMs,
		storyTimeCmd:        storyTime,
	}
}

// TODO: expand on this for all fields
func validateCmdRegex() bool {
	conforms := true
	for _, cmd := range getCommands() {
		if !validCommandRegex().MatchString(cmd.Name) {
			log.Errorf(`Command %s does not conform to Discord's command naming requirements`, cmd.Name)
			conforms = false
		}
		for _, opt1 := range cmd.Options {
			if !validCommandRegex().MatchString(opt1.Name) {
				log.Errorf(`Option %s.%s does not conform to Discord's command naming requirements`, cmd.Name, opt1.Name)
				conforms = false
			}

			switch opt1.Type {
			case discordgo.ApplicationCommandOptionSubCommand:
				for _, opt2 := range opt1.Options {
					if !validCommandRegex().MatchString(opt2.Name) {
						log.Errorf(`Option %s.%s.%s does not conform to Discord's command naming requirements`, cmd.Name, opt1.Name, opt2.Name)
						conforms = false
					}
				}

			case discordgo.ApplicationCommandOptionSubCommandGroup:
				for _, opt2 := range opt1.Options {
					if !validCommandRegex().MatchString(opt2.Name) {
						log.Errorf(`Option %s.%s.%s does not conform to Discord's command naming requirements`, cmd.Name, opt1.Name, opt2.Name)
						conforms = false
					}

					for _, opt3 := range opt2.Options {
						if !validCommandRegex().MatchString(opt3.Name) {
							log.Errorf(`Option %s.%s.%s.%s does not conform to Discord's command naming requirements`, cmd.Name, opt1.Name, opt2.Name, opt3.Name)
							conforms = false
						}
					}
				}
			}
		}
	}

	return conforms
}
