package kardbot

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

const (
	RollCmd          = "roll"
	RollSubCmdDnD    = "dnd"
	RollSubCmdCustom = "custom"

	MinDieSides uint64 = 2 // What the hell would a 1-sided die be? A black hole?
	DieStartVal uint64 = 1 // Dice numbering starts at this value
)

func roll(s *discordgo.Session, i *discordgo.InteractionCreate) {
	wg := bot().updateLastActive()
	defer wg.Wait()

	if isSelf, err := authorIsSelf(s, i); err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	} else if isSelf {
		log.Trace("Ignoring message from self")
		return
	}

	switch i.ApplicationCommandData().Options[0].Name {
	case RollSubCmdCustom:
		rollCustomDice(s, i)
	case RollSubCmdDnD:
		addDnDButtons(s, i)
	default:
		err := fmt.Errorf("unknown subcommand: %s", i.ApplicationCommandData().Options[0].Name)
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
	}
}

type dndDie uint

func (d dndDie) String() string {
	return fmt.Sprintf("D%d", d)
}

func (d *dndDie) parseFromString(s string) error {
	var tmp dndDie
	switch s {
	case d4.String():
		tmp = d4
	case d6.String():
		tmp = d6
	case d8.String():
		tmp = d8
	case d10.String():
		tmp = d10
	case d12.String():
		tmp = d12
	case d20.String():
		tmp = d20
	case d100.String():
		tmp = d100
	default:
		return fmt.Errorf("unknown dndDie type: %s", s)
	}
	*d = tmp
	return nil
}

func (d dndDie) buttonID() string {
	return dndDieButtonIDPrefix + d.String()
}

func (d *dndDie) parseFromButtonID(buttonID string) error {
	return d.parseFromString(strings.TrimPrefix(buttonID, dndDieButtonIDPrefix))
}

const (
	dndDieButtonIDPrefix = "dndButton"

	d4   dndDie = 4
	d6   dndDie = 6
	d8   dndDie = 8
	d10  dndDie = 10
	d12  dndDie = 12
	d20  dndDie = 20
	d100 dndDie = 100
)

func addDnDButtons(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Roll some dice!",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    d4.String(),
							Style:    discordgo.PrimaryButton,
							CustomID: d4.buttonID(),
						},
						discordgo.Button{
							Label:    d6.String(),
							Style:    discordgo.PrimaryButton,
							CustomID: d6.buttonID(),
						},
						discordgo.Button{
							Label:    d8.String(),
							Style:    discordgo.PrimaryButton,
							CustomID: d8.buttonID(),
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    d10.String(),
							Style:    discordgo.PrimaryButton,
							CustomID: d10.buttonID(),
						},
						discordgo.Button{
							Label:    d12.String(),
							Style:    discordgo.PrimaryButton,
							CustomID: d12.buttonID(),
						},
						discordgo.Button{
							Label:    d20.String(),
							Style:    discordgo.PrimaryButton,
							CustomID: d20.buttonID(),
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    d100.String(),
							Style:    discordgo.PrimaryButton,
							CustomID: d100.buttonID(),
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
	}
}

func handleDnDButtonPress(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var die dndDie
	err := die.parseFromButtonID(i.MessageComponentData().CustomID)
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	}

	result, err := randFromRange(1, uint64(die))
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Rolled %d%s:\n%d", 1, die, result),
		},
	})
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	}
}

func rollCustomDice(s *discordgo.Session, i *discordgo.InteractionCreate) {
	count := i.ApplicationCommandData().Options[0].Options[0].IntValue()
	if count < 1 {
		err := fmt.Errorf("cannot roll a die <1 times")
		log.Error(err)
		interactionRespondEphemeralError(s, i, false, err)
		return
	}

	sides, err := parseDieSides(i.ApplicationCommandData().Options[0].Options[1].StringValue())
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, false, err)
		return
	}

	output := fmt.Sprintf("Rolling %d D%d's...\n", count, sides)
	printIndividualRolls := uint64(count)*uint64(len(strconv.FormatUint(sides, 10)))+uint64(len(output)) < MaxDiscordMsgLen
	total := uint64(0)
	if printIndividualRolls {
		for j := int64(0); j < count; j++ {
			roll, err := randFromRange(DieStartVal, sides)
			if err != nil {
				log.Error(err)
				interactionRespondEphemeralError(s, i, true, err)
				return
			}
			total += roll
			if printIndividualRolls {
				output += fmt.Sprintf("%d\n", roll)
			}
		}
	} else {
		// No need to track individual dice rolls if we are only printing a total
		roll, err := randFromRange(DieStartVal*uint64(count), sides*uint64(count))
		if err != nil {
			log.Error(err)
			interactionRespondEphemeralError(s, i, true, err)
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
		interactionRespondEphemeralError(s, i, true, err)
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
	dieSidesParsed := IsNotNumericRegex().ReplaceAllString(rawDieSides, "")
	sides, err := strconv.ParseUint(dieSidesParsed, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("could not convert %s to int", dieSidesParsed)
	}
	if sides < MinDieSides {
		return 0, fmt.Errorf("%d is not a valid number of sides", sides)
	}

	return sides, nil
}
