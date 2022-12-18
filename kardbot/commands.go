package kardbot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

type onInteractionHandler func(*discordgo.Session, *discordgo.InteractionCreate)

// The max number of command options any single command is allowed to have.
const (
	maxDiscordCommandOptions             = 25
	maxDiscordOptionChoices              = 25
	maxDiscordSelectMenuOpts             = 25
	maxDiscordActionRows                 = 5
	maxDiscordSelectMenuPlaceholderChars = 150
)

func getCommands() []*discordgo.ApplicationCommand {
	allcmds := []*discordgo.ApplicationCommand{
		{
			Name:        RollCmd,
			Description: "Dice rolling fun!",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        RollSubCmdDnD,
					Description: "Get a set of buttons for rolling DnD dice.",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        RollSubCmdCustom,
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
			},
		},
		{
			Name:        oddsCmd,
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
			Name:        logLevelCmd,
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
			Name:        pastaCmd,
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
			Name:        uwuCmd,
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
					Name:        uwuSubCmdPasta,
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
							Name:        uwuSubCmdPasta,
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
			Name:        redditRouletteCmd,
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
			Name:        complimentsCmd,
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
			Name:        creepyDMCmd,
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
			Name:        madlibCmd,
			Description: "The bot will fill in any blanks indicated with " + madlibBlank,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "prompt",
					Description: fmt.Sprintf("An input prompt containing blanks. Example: 'The %s jumps over the %s.'", madlibBlank, madlibBlank),
					Required:    true,
				},
			},
		},
		{
			Name:        pollCmd,
			Description: "Create a poll",
			Options:     getPollOpts(),
		},
		{
			Name:        storyTimeCmd,
			Description: "The bot will tell you a short story (but not a good or sensical one) based on a given prompt.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        storyTimePromptOpt,
					Description: "A prompt to generate a story from.",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        storyTimeModelOpt,
					Description: "The text generation AI model to use. For example, gpt2 or EleutherAI/gpt-neo-125M.",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        storyTimeHelpOpt,
					Description: "Print a helpful message about how to use this command.",
					Required:    false,
				},
			},
		},
		{
			Name:        roleSelectMenuCommand,
			Description: "Allow users to select roles for themselves",
			Options:     roleSelectCmdOpts(),
		},
		{
			Name:        embedCmd,
			Description: "Build a custom embed",
			Options:     embedCmdOpts(),
		},
		{
			Name:        timeCmd,
			Description: "Time related commands",
			Options:     timeCmdOpts(),
		},
		{
			Name:        dalle2Cmd,
			Description: "Ask an AI to generate an image from a prompt. Uses Open AI's DALL·E 2.",
			Options:     dalle2Opts(),
		},
		{
			Name:        helpCmd,
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
		RollCmd:               roll,
		logLevelCmd:           updateLogLevel,
		pastaCmd:              servePasta,
		redditRouletteCmd:     redditRoulette,
		uwuCmd:                uwuify,
		complimentsCmd:        complimentHandler,
		creepyDMCmd:           creepyDMHandler,
		helpCmd:               botInfo,
		oddsCmd:               whatAreTheOdds,
		memeCommand:           buildAMeme,
		delBotDMCmd:           deleteBotDMs,
		storyTimeCmd:          storyTime,
		roleSelectMenuCommand: handleRoleSelectMenuCommand,
		embedCmd:              handleEmbedCmd,
		madlibCmd:             handleMadLibCmd,
		timeCmd:               handleTimeCmd,
		pollCmd:               handlePollCmd,
		dalle2Cmd:             handleDalle2Cmd,
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

func getComponentImpls() map[string]onInteractionHandler {
	return map[string]onInteractionHandler{
		selectMenuErrorReport:           handleErrorReportSelection,
		dndRollButtonID:                 handleDnDButtonPress,
		dndDiceCountSelectID:            handleDiceCountMenuSelection,
		dndDiceFacesSelectID:            handleDiceFacesMenuSelection,
		dndOtherOptionsSelectID:         handleDnDOtherOptionsSelection,
		roleSelectMenuComponentIDPrefix: handleRoleSelection,
		roleSelectResetButtonID:         handleRoleSelectReset,
		pollSelectMenuID:                handlePollSubmission,
	}
}
