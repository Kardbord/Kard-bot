package kardbot

import (
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/lucasb-eyer/go-colorful"
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

func RandomBoolean() bool {
	return rand.Int31()&0x01 == 0
}

func isHTTPS(url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		log.Warn(err)
		return false
	}
	// final URL resolved
	return strings.HasPrefix(resp.Request.URL.String(), "https://")
}

var isNotNumericRegex = func() *regexp.Regexp { return nil }

func init() {
	r := regexp.MustCompile("[^0-9]+")
	isNotNumericRegex = func() *regexp.Regexp { return r }
}

// Taken from https://stackoverflow.com/questions/41602230
func firstN(s string, n int) string {
	i := 0
	for j := range s {
		if i == n {
			return s[:j]
		}
		i++
	}
	return s
}

func fastHappyColorInt64() (int64, error) {
	i, err := strconv.ParseInt(strings.Replace(colorful.FastHappyColor().Hex(), "#", "", -1), 16, 32)
	return i, err
}

var sentenceEndPunctRegex = func() *regexp.Regexp { return nil }

func init() {
	r := regexp.MustCompile(`\s*[^\d\w>]+\s*$`)
	if r == nil {
		log.Fatal("Could not init sentenceEndPunctRegex")
	}

	sentenceEndPunctRegex = func() *regexp.Regexp { return r }
}
