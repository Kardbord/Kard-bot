package kardbot

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	cmap "github.com/orcaman/concurrent-map"
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
	dndDieButtonIDPrefix = "dndDieButton"
	dndButtonDiceCount   = "dndButtonDiceCount"

	d4   dndDie = 4
	d6   dndDie = 6
	d8   dndDie = 8
	d10  dndDie = 10
	d12  dndDie = 12
	d20  dndDie = 20
	d100 dndDie = 100
)

func addDnDButtons(s *discordgo.Session, i *discordgo.InteractionCreate) {
	const maxNumDice = 10
	diceCountMenu := discordgo.SelectMenu{
		CustomID:    dndButtonDiceCount,
		Placeholder: "How many 🎲?",
		Options:     make([]discordgo.SelectMenuOption, maxNumDice),
	}
	for i := maxNumDice; i > 0; i-- {
		diceCountMenu.Options[i-1] = discordgo.SelectMenuOption{
			Label:       fmt.Sprintf("%d 🎲", i),
			Description: fmt.Sprintf("Roll %d dice", i),
			Value:       fmt.Sprint(i),
			Default:     false,
		}
	}
	diceCountMenu.Options[0].Default = true

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Select the desired options, then press the button to roll the dice!",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{diceCountMenu},
				},
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

	metadata, err := getInteractionMetaData(i)
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	}

	iCfg, ok := dndDiceRollMsgConfigs.Get(metadata.MessageID)
	if !ok {
		iCfg = dndButtonsMsgConfig{
			MsgID:     metadata.MessageID,
			DiceCount: 1,
		}
	}
	cfg, ok := iCfg.(dndButtonsMsgConfig)
	if !ok {
		err = fmt.Errorf("bad cast to dndButtonsMsgConfig")
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	}

	content := fmt.Sprintf("Rolled %d%s:\n", cfg.DiceCount, die)

	total := uint64(0)
	for j := uint64(0); j < cfg.DiceCount; j++ {
		rollResult, err := randFromRange(1, uint64(die))
		if err != nil {
			log.Error(err)
			interactionRespondEphemeralError(s, i, true, err)
			return
		}
		total += rollResult
		content += fmt.Sprintf("%d\n", rollResult)
	}
	if cfg.DiceCount > 1 {
		content += fmt.Sprintf("Total: %d", total)
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	}
}

func handleDiceCountMenuSelection(s *discordgo.Session, i *discordgo.InteractionCreate) {
	metadata, err := getInteractionMetaData(i)
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	}

	iCfg, ok := dndDiceRollMsgConfigs.Get(metadata.MessageID)
	if !ok {
		iCfg = dndButtonsMsgConfig{
			MsgID: metadata.MessageID,
		}
	}
	cfg, ok := iCfg.(dndButtonsMsgConfig)
	if !ok {
		err = fmt.Errorf("bad cast to dndButtonsMsgConfig")
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	}

	if len(i.MessageComponentData().Values) == 0 {
		err = fmt.Errorf("no values sent")
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	}

	diceCount, err := strconv.ParseUint(i.MessageComponentData().Values[0], 10, 64)
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	}

	cfg.DiceCount = diceCount
	dndDiceRollMsgConfigs.Set(metadata.MessageID, cfg)

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Dice rolls from this message will now roll %d dice.", cfg.DiceCount),
			Flags:   InteractionResponseFlagEphemeral,
		},
	})
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
	}
}

type dndButtonsMsgConfig struct {
	MsgID     string `json:"msg_id"`
	DiceCount uint64 `json:"dice_count,omitempty"`
}

// Map of message IDs to settings
var dndDiceRollMsgConfigs = cmap.New()
