package kardbot

import (
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	// Possible greetings the bot will respond to
	// TODO: make these configurable
	greetings = []string{
		"Hello",
		"Hi",
		"Greetings",
		"Salutations",
	}

	// Possible farewells the bot will respond to
	// TODO: make these configurable
	farewells = []string{
		"Goodbye",
		"Farewell",
		"So long",
		"Bye",
		"See you",
		"See ya",
	}
)

func fromSelf(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	return m.Author.ID == s.State.User.ID
}

func greeting(s *discordgo.Session, m *discordgo.MessageCreate) {
	if fromSelf(s, m) {
		return
	}

	greetingGroup := buildRegexAltGroup(greetings)

	matched, err := regexp.MatchString(
		fmt.Sprintf("^(?i)%s %s[!.]*$", greetingGroup, buildBotNameRegexp(s.State.User.Username)),
		m.Content,
	)
	if err != nil {
		log.Println("Regex error: ", err)
		return
	}
	if matched {
		s.ChannelMessageSend(
			m.ChannelID,
			fmt.Sprintf("%s %s!", greetings[rand.Intn(len(greetings))], m.Author.Username),
		)
	}
}

func farewell(s *discordgo.Session, m *discordgo.MessageCreate) {
	if fromSelf(s, m) {
		return
	}

	farewellGroup := buildRegexAltGroup(farewells)

	matched, err := regexp.MatchString(
		fmt.Sprintf("^(?i)%s %s[!.]*$", farewellGroup, buildBotNameRegexp(s.State.User.Username)),
		m.Content,
	)
	if err != nil {
		log.Println("Regex error: ", err)
		return
	}
	if matched {
		s.ChannelMessageSend(
			m.ChannelID,
			fmt.Sprintf("%s %s!", farewells[rand.Intn(len(farewells))], m.Author.Username),
		)
	}
}

// Some characters are optional when matching the bot name.
// This function returns a regexp string to appropriately
// match the bot name, including any optional characters.
func buildBotNameRegexp(botName string) string {
	// TODO: make these configurable
	optionalRunes := []rune{
		'-',
		'_',
	}
	botNameExp := botName
	for _, r := range optionalRunes {
		botNameExp = strings.ReplaceAll(botNameExp, string(r), fmt.Sprintf("%s?", string(r)))
	}
	log.Println("Built bot exp=", botNameExp)
	return botNameExp
}

// Returns a regexp alternate group of the provided
// strings. For example, input of [a ,b] would result
// in a return value of "(a|b)".
func buildRegexAltGroup(alts []string) string {
	altGroup := "("
	for i, alt := range alts {
		altGroup += alt
		if i+1 < len(alts) {
			altGroup += "|"
		}
	}
	altGroup += ")"
	log.Println("Built altgroup=", altGroup)
	return altGroup
}
