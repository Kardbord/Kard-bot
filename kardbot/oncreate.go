package kardbot

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"

	"github.com/TannerKvarfordt/Kard-bot/kardbot/config"
	"github.com/bwmarrin/discordgo"

	log "github.com/sirupsen/logrus"
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

var (
	greetings []string
	farewells []string
)

const greetingAndFarewellConfigFile = "config/greetings-farewells.json"

func init() {
	cfg := struct {
		Greetings []string `json:"greetings"`
		Farewells []string `json:"farewells"`
	}{}

	jsonCfg, err := config.NewJsonConfig(greetingAndFarewellConfigFile)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(jsonCfg.Raw, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	greetings = cfg.Greetings
	farewells = cfg.Farewells

	if len(greetings) == 0 {
		log.Fatal("No greetings configured.")
	}

	if len(farewells) == 0 {
		log.Fatal("No farewells configured.")
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

	greetingGroup := buildRegexAltGroup(greetings)

	matched, err := regexp.MatchString(
		fmt.Sprintf("^(?i)%s %s[!.\\s]*$", greetingGroup, buildBotNameRegexp(s.State.User.Username, s.State.User.ID)),
		m.Content,
	)
	if err != nil {
		log.Error(err)
		return
	}
	if matched {
		wg := bot().updateLastActive()
		defer wg.Wait()
		_, err := s.ChannelMessageSend(
			m.ChannelID,
			fmt.Sprintf("%s %s!", randomGreeting(), m.Author.Username),
		)
		if err != nil {
			log.Error(err)
		}
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

	farewellGroup := buildRegexAltGroup(farewells)

	matched, err := regexp.MatchString(
		fmt.Sprintf("^(?i)%s %s[!.\\s]*$", farewellGroup, buildBotNameRegexp(s.State.User.Username, s.State.User.ID)),
		m.Content,
	)
	if err != nil {
		log.Error(err)
		return
	}
	if matched {
		wg := bot().updateLastActive()
		defer wg.Wait()
		_, err := s.ChannelMessageSend(
			m.ChannelID,
			fmt.Sprintf("%s %s!", randomFarewell(), m.Author.Username),
		)
		if err != nil {
			log.Error(err)
		}
	}
}

func randomGreeting() string {
	return greetings[rand.Intn(len(greetings))]
}

func randomFarewell() string {
	return farewells[rand.Intn(len(farewells))]
}
