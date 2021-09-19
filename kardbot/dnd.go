package kardbot

import (
	"fmt"
	"regexp"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/lus/dgc"
)

const (
	MinDieSides uint = 2 // What the hell would a 1-sided die be? A black hole?
	DieStartVal uint = 1 // Dice numbering starts at this value
)

func rollDice(ctx *dgc.Ctx) {
	args, err := getArgsExpectCount(ctx, 2, true)
	if err != nil {
		log.Error(err)
		return
	}

	sides, err := parseDieSides(args.Get(1).Raw())
	if err != nil {
		log.Error(err)
		return
	}

	count, err := args.Get(0).AsInt()
	if err != nil {
		log.Error(err)
		return
	}
	if count < 1 {
		log.Error("cannot roll a die <1 times")
		return
	}

	output := fmt.Sprintf("Rolling %d D%d's...\n", count, sides)
	printIndividualRolls := count*len(strconv.Itoa(int(sides)))+len(output) < int(MaxDiscordMsgLen)
	total := uint(0)
	if printIndividualRolls {
		for i := 0; i < count; i++ {
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
		roll, err := randFromRange(DieStartVal*uint(count), sides*uint(count))
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
		log.Warnf("There is a bug in msg length validation when rolling %d D%d's", count, sides)
		output = fmt.Sprintf("Rolling %d D%d's...\nTotal: %d", count, sides, total)
	}
	ctx.RespondText(output)
}

// Parse number of sides on dice from a string of the form
// D{NUM} or just {NUM}
func parseDieSides(rawDieSides string) (uint, error) {
	// This regex disallows negative numbers
	matched, err := regexp.MatchString("^(?i)d?[0-9]+$", rawDieSides)
	if err != nil {
		return 0, err
	}
	if !matched {
		// Invalid die sides provided
		return 0, fmt.Errorf("invalid argument provided: %s", rawDieSides)
	}
	// Strip non-numeric characters
	reg, err := regexp.Compile("[^0-9]+")
	if err != nil {
		return 0, err
	}
	dieSidesParsed := reg.ReplaceAllString(rawDieSides, "")
	sides, err := strconv.Atoi(dieSidesParsed)
	if err != nil {
		return 0, fmt.Errorf("could not convert %s to int", dieSidesParsed)
	}
	if sides < int(MinDieSides) {
		return 0, fmt.Errorf("%d is not a valid number of sides", sides)
	}

	return uint(sides), nil
}
