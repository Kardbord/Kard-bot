package kardbot

import (
	"github.com/bwmarrin/discordgo"
)

type onReadyHandlerf = func(*discordgo.Session, *discordgo.Ready)

type onReadyHandler struct {
	handler onReadyHandlerf
	help    string
}

// Any callbacks that happen onReady belong in this list.
// It is the duty of each individual function to decide whether or not to run.
// These callbacks must be able to safely execute asynchronously.
var onReadyHandlers = [...]onReadyHandler{
	{onReady, "Updates \"Listening Status\""},
}

func onReady(s *discordgo.Session, r *discordgo.Ready) {
	s.UpdateListeningStatus("you")
}
