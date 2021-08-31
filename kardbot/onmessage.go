package kardbot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type onCreateHandlerf = func(*discordgo.Session, *discordgo.MessageCreate)

type onCreateHandler struct {
	handler onCreateHandlerf
	help    string
}

// Any callbacks that happen onMessageCreate belong in this list.
// It is the duty of each individual function to decide whether or not to run.
// These callbacks must be able to safely execute asynchronously.
var onCreateHandlers = [...]onCreateHandler{
	{greeting, fmt.Sprintf("Returns your salutations when you greet the bot with %v", greetings)},
	{farewell, fmt.Sprintf("Returns your valediction when you tell the bot %v", farewells)},
}
