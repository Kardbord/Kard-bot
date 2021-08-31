package onready

import (
	"github.com/bwmarrin/discordgo"
)

type onReadyHandler = func(*discordgo.Session, *discordgo.Ready)

// Any callbacks that happen onReady belong in this list.
// It is the duty of each individual function to decide whether or not to run.
// These callbacks must be able to safely execute asynchronously.
var OnReadyHandlers = [...]onReadyHandler{
	onReady,
}

func onReady(s *discordgo.Session, r *discordgo.Ready) {
	s.UpdateListeningStatus("you")
}
