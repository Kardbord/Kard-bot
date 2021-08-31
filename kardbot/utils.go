package kardbot

import (
	"fmt"
	"log"
	"strings"
)

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
