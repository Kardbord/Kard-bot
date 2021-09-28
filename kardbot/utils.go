package kardbot

import (
	"fmt"
	"math/rand"
	"strings"

	log "github.com/sirupsen/logrus"
)

const MaxDiscordMsgLen uint64 = 2000

// Some characters are optional when matching the bot name.
// This function returns a regexp string to appropriately
// match the bot name, including any optional characters.
func buildBotNameRegexp(botName, botID string) string {
	optionalRunes := []rune{'-'}

	botNameExp := botName
	for _, r := range optionalRunes {
		botNameExp = strings.ReplaceAll(botNameExp, string(r), fmt.Sprintf("%s?", string(r)))
	}
	// Checks for possible '@' mentions
	botNameExp = "(" + botNameExp + "|<@!" + botID + ">" + "|<@" + botID + ">" + ")"
	log.Trace("Built bot name regex:", botNameExp)
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
	log.Trace("Built alt group regex:", altGroup)
	return altGroup
}

// Return a non-negative random number in the inclusive range [min, max].
// If max <= min, returns the maximum uint value and an error.
func randFromRange(min, max uint64) (uint64, error) {
	if max <= min {
		return ^uint64(0), fmt.Errorf("max (%d) cannot be less than or equal to min (%d)", max, min)
	}
	return uint64(rand.Intn(int((max-min)+1))) + min, nil
}

func randomBoolean() bool {
	return rand.Int31()&0x01 == 0
}
