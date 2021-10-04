package kardbot

import "github.com/bwmarrin/discordgo"

type onInteractionHandler func(*discordgo.Session, *discordgo.InteractionCreate)

func getCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
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
	}
}

func getCommandImpls() map[string]onInteractionHandler {
	return map[string]onInteractionHandler{
		"roll":            rollDice,
		"loglevel":        updateLogLevel,
		"pasta":           servePasta,
		"reddit-roulette": redditRoulette,
		"uwu":             uwuify,
	}
}
