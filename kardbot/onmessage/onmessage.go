package onmessage

import "github.com/bwmarrin/discordgo"

type onCreateHandler = func(*discordgo.Session, *discordgo.MessageCreate)

// Any callbacks that happen onMessageCreate belong in this list.
// It is the duty of each individual function to decide whether or not to run.
// These callbacks must be able to safely execute asynchronously.
var OnCreateHandlers = [...]onCreateHandler{
  Greeting,
}
