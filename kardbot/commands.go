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
				},
				{
					Name:        redditRouletteSubCmdSFW,
					Description: "Retrieve a random reddit post that is not marked as NSFW",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        redditRouletteSubCmdNSFW,
					Description: "Retrieve a random reddit post that is marked as NSFW.",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
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
	}
}
