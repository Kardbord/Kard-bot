package kardbot

import (
	"fmt"
	"regexp"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/lus/dgc"
)

const (
	MinDieSides = 2 // What the hell would a 1-sided die be? A black hole?
	DieStartVal = 1 // Dice numbering starts at this value
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
		log.Errorf("could not get arg[0]=%s as int - %v", args.Get(0).Raw(), err)
		return
	}
	if count < 1 {
		log.Error("cannot roll a die <1 times")
		return
	}

	output := fmt.Sprintf("Rolling %d D%d's...\n", count, sides)
	total := uint(0)
	for i := 0; i < count; i++ {
		roll, err := randFromRange(DieStartVal, sides)
		if err != nil {
			log.Error(err)
			return
		}
		total += roll
		output += fmt.Sprintf("%d\n", roll)
	}
	if count > 1 {
		output += fmt.Sprintf("Total: %d", total)
	}

	// TODO: limit response text to 2000 characters (Discord imposed limit)
	ctx.RespondText(output)
}

// Parse number of sides on dice from a string of the form
// D{NUM} or just {NUM}
func parseDieSides(rawDieSides string) (uint, error) {
	// This regex disallows negative numbers
	matched, err := regexp.MatchString("^(?i)d?[0-9]+$", rawDieSides)
	if err != nil {
		return 0, fmt.Errorf("regexp err: %v", err)
	}
	if !matched {
		// Invalid die sides provided
		return 0, fmt.Errorf("invalid argument provided: %s", rawDieSides)
	}
	// Strip non-numeric characters
	reg, err := regexp.Compile("[^0-9]+")
	if err != nil {
		return 0, fmt.Errorf("regexp err: %v", err)
	}
	dieSidesParsed := reg.ReplaceAllString(rawDieSides, "")
	sides, err := strconv.Atoi(dieSidesParsed)
	if err != nil {
		return 0, fmt.Errorf("could not convert %s to int", dieSidesParsed)
	}
	if sides < MinDieSides {
		return 0, fmt.Errorf("%d is not a valid number of sides", sides)
	}

	return uint(sides), nil
}
