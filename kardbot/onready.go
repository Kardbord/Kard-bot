package kardbot

import (
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

type onReadyHandler = func(*discordgo.Session, *discordgo.Ready)

// Any callbacks that happen onReady belong in this list.
// It is the duty of each individual function to decide whether or not to run.
// These callbacks must be able to safely execute asynchronously.
func onReadyHandlers() []onReadyHandler {
	return []onReadyHandler{
		onReady,
	}
}

func onReady(s *discordgo.Session, r *discordgo.Ready) {
	err := s.UpdateListeningStatus("you")
	if err != nil {
		log.Error(err)
	}
}
