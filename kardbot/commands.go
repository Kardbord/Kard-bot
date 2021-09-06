package kardbot

import (
	"fmt"

	"github.com/lus/dgc"
)

var mCommands = [...]*dgc.Command{
	{
		Name: "roll",
		Aliases: []string{
			"dice",
			"rollDice",
		},
		Description: "Rolls a D{X} die {Y} times, where X and Y are provided by the user.",
		Usage:       "roll Y DX",
		Example:     "roll 1 D20",
		Flags:       []string{},
		IgnoreCase:  true,
		SubCommands: []*dgc.Command{},
		RateLimiter: nil,
		Handler:     rollDice,
	},
	{
		Name:        "loglevel",
		Aliases:     []string{},
		Description: fmt.Sprintf("Update the log level of the bot.\nOnly works for whitelisted users.\nValid log levels: %v", getLogLevelKeys()),
		Usage:       "loglevel LEVEL",
		Example:     "loglevel trace",
		Flags:       []string{},
		IgnoreCase:  true,
		SubCommands: []*dgc.Command{},
		RateLimiter: nil,
		Handler:     updateLogLevel,
	},
}
