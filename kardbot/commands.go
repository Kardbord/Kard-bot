package kardbot

import (
	"fmt"

	"github.com/lus/dgc"
)

var mCommands = [...]dgc.Command{
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
}

// Retrieve args from context and validate they are not nil.
func getArgs(ctx *dgc.Ctx) (*dgc.Arguments, error) {
	if ctx == nil {
		return nil, fmt.Errorf("nil context provided")
	}
	if ctx.Arguments == nil {
		return nil, fmt.Errorf("context arguments were nil")
	}
	return ctx.Arguments, nil
}

// Retrieve args from context and validate that they are not nil.
// Also validate that the expected number of arguments are present.
// If the "exact" argument is true, the number of present arguments
// must exactly match the number found. Otherwise, there must be
// at least as many arguments as expected.
func getArgsExpectCount(ctx *dgc.Ctx, expected int, exact bool) (*dgc.Arguments, error) {
	args, err := getArgs(ctx)
	if err != nil {
		return nil, err
	}

	err = fmt.Errorf("unexpected arg count, expected: %d, actual: %d", expected, ctx.Arguments.Amount())
	if exact && ctx.Arguments.Amount() != expected {
		return nil, err
	} else if ctx.Arguments.Amount() < expected {
		return nil, err
	}
	return args, nil
}
