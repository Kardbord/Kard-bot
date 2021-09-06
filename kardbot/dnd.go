package kardbot

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/lus/dgc"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// TODO: add a unit test for this method
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
	total := 0
	for i := 0; i < count; i++ {
		roll := rand.Intn(sides) + 1
		total += roll
		output += fmt.Sprintf("%d\n", roll)
	}
	if count > 1 {
		output += fmt.Sprintf("Total: %d", total)
	}

	ctx.RespondText(output)
}

// Parse number of sides on dice from a string of the form
// D{NUM} or just {NUM}
func parseDieSides(rawDieSides string) (int, error) {
	matched, err := regexp.MatchString("^(?i)d?([2-9]{1}|[0-9]{2,})$", rawDieSides)
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

	return sides, nil
}
