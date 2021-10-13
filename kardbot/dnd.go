package kardbot

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

const (
	MinDieSides uint64 = 2 // What the hell would a 1-sided die be? A black hole?
	DieStartVal uint64 = 1 // Dice numbering starts at this value
)

func rollDice(s *discordgo.Session, i *discordgo.InteractionCreate) {
	wg := bot().updateLastActive()
	defer wg.Wait()

	if isSelf, err := authorIsSelf(s, i); err != nil {
		log.Error(err)
		return
	} else if isSelf {
		log.Trace("Ignoring message from self")
		return
	}

	count := i.ApplicationCommandData().Options[0].UintValue()
	if count < 1 {
		log.Error("cannot roll a die <1 times")
		return
	}

	sides, err := parseDieSides(i.ApplicationCommandData().Options[1].StringValue())
	if err != nil {
		log.Error(err)
		return
	}

	output := fmt.Sprintf("Rolling %d D%d's...\n", count, sides)
	printIndividualRolls := count*uint64(len(strconv.FormatUint(sides, 10)))+uint64(len(output)) < MaxDiscordMsgLen
	total := uint64(0)
	if printIndividualRolls {
		for i := uint64(0); i < count; i++ {
			roll, err := randFromRange(DieStartVal, sides)
			if err != nil {
				log.Error(err)
				return
			}
			total += roll
			if printIndividualRolls {
				output += fmt.Sprintf("%d\n", roll)
			}
		}
	} else {
		// No need to track individual dice rolls if we are only printing a total
		roll, err := randFromRange(DieStartVal*count, sides*count)
		if err != nil {
			log.Error(err)
			return
		}
		total = roll
	}
	if count > 1 {
		output += fmt.Sprintf("Total: %d", total)
	}

	// Cheap sanity check in case the above logic did not catch a message that is too large
	if len(output) > int(MaxDiscordMsgLen) {
		log.Warnf("There is a bug in msg length validation when rolling %d D%d's. Possible overflow?", count, sides)
		output = fmt.Sprintf("Rolling %d D%d's...\nTotal: %d", count, sides, total)
	}
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: output,
		},
	})
	if err != nil {
		log.Error(err)
	}
}

var dieSidesRegex = func() *regexp.Regexp { return nil }

func init() {
	r := regexp.MustCompile("^(?i)d?[0-9]+$")
	dieSidesRegex = func() *regexp.Regexp {
		return r
	}
}

// Parse number of sides on dice from a string of the form
// D{NUM} or just {NUM}
func parseDieSides(rawDieSides string) (uint64, error) {
	// This regex disallows negative numbers
	matched := dieSidesRegex().MatchString(rawDieSides)
	if !matched {
		// Invalid die sides provided
		return 0, fmt.Errorf("invalid argument provided: %s", rawDieSides)
	}
	// Strip non-numeric characters
	dieSidesParsed := isNotNumericRegex().ReplaceAllString(rawDieSides, "")
	sides, err := strconv.ParseUint(dieSidesParsed, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("could not convert %s to int", dieSidesParsed)
	}
	if sides < MinDieSides {
		return 0, fmt.Errorf("%d is not a valid number of sides", sides)
	}

	return sides, nil
}
