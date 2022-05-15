package kardbot

import (
	"fmt"
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

var (
	IsNotNumericRegex = func() *regexp.Regexp { return nil }
	IsNumericRegex    = func() *regexp.Regexp { return nil }
)

func init() {
	r1 := regexp.MustCompile("[^0-9]+")
	IsNotNumericRegex = func() *regexp.Regexp { return r1 }

	r2 := regexp.MustCompile(`^\d+$`)
	IsNumericRegex = func() *regexp.Regexp { return r2 }
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

// See https://discord.com/developers/docs/interactions/application-commands#application-command-object
var validCommandRegex = func() *regexp.Regexp { return nil }

func init() {
	r := regexp.MustCompile(`^[\w-]{1,32}$`)
	if r == nil {
		log.Fatal("Could not compile validCommandRegex")
	}

	validCommandRegex = func() *regexp.Regexp { return r }
}
