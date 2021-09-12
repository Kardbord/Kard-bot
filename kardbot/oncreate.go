package kardbot

import (
	"fmt"
	"regexp"

	log "github.com/sirupsen/logrus"

	"github.com/bwmarrin/discordgo"
)

type onCreateHandler = func(*discordgo.Session, *discordgo.MessageCreate)

// Any callbacks that happen onMessageCreate belong in this list.
// It is the duty of each individual function to decide whether or not to run.
// These callbacks must be able to safely execute asynchronously.
func onCreateHandlers() []onCreateHandler {
	return []onCreateHandler{
		greeting,
		farewell,
	}
}

func msgIsFromSelf(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	return m.Author.ID == s.State.User.ID
}

func greeting(s *discordgo.Session, m *discordgo.MessageCreate) {
	if msgIsFromSelf(s, m) {
		log.Trace("Ignoring message from self")
		return
	}
	if m.Author.Bot {
		log.Trace("Ignoring message from bot")
		return
	}

	greetingGroup := buildRegexAltGroup(bot().Greetings)

	matched, err := regexp.MatchString(
		fmt.Sprintf("^(?i)%s %s[!.\\s]*$", greetingGroup, buildBotNameRegexp(s.State.User.Username, s.State.User.ID)),
		m.Content,
	)
	if err != nil {
		log.Error(err)
		return
	}
	if matched {
		s.ChannelMessageSend(
			m.ChannelID,
			fmt.Sprintf("%s %s!", bot().randomGreeting(), m.Author.Username),
		)
	}
}

func farewell(s *discordgo.Session, m *discordgo.MessageCreate) {
	if msgIsFromSelf(s, m) {
		log.Trace("Ignoring message from self")
		return
	}
	if m.Author.Bot {
		log.Trace("Ignoring message from bot")
		return
	}

	farewellGroup := buildRegexAltGroup(bot().Farewells)

	matched, err := regexp.MatchString(
		fmt.Sprintf("^(?i)%s %s[!.\\s]*$", farewellGroup, buildBotNameRegexp(s.State.User.Username, s.State.User.ID)),
		m.Content,
	)
	if err != nil {
		log.Error(err)
		return
	}
	if matched {
		s.ChannelMessageSend(
			m.ChannelID,
			fmt.Sprintf("%s %s!", bot().randomFarewell(), m.Author.Username),
		)
	}
}
