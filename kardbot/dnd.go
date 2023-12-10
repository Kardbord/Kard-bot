package kardbot

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/Kardbord/ubiquity/mathutils/random"
	"github.com/bwmarrin/discordgo"
	cmap "github.com/orcaman/concurrent-map/v2"
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
			roll, err := random.Range(DieStartVal, sides)
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
		roll, err := random.Range(DieStartVal*uint64(count), sides*uint64(count))
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

const (
	dndRollButtonID         = "roll"
	dndDieSelectID          = "dndDieSelect"
	dndDiceCountSelectID    = "dndButtonDiceCountSelect"
	dndDiceFacesSelectID    = "dndDiceFacesSelect"
	dndOtherOptionsSelectID = "dndOtherOptsSelect"
	dndDiceRollOptEphemeral = "Ephemeral 🔮"
	dndDiceRollOptDM        = "DM 📫"
	dndDiceRollOptNone      = "None ❌"

	d4   dndDie = 4
	d6   dndDie = 6
	d8   dndDie = 8
	d10  dndDie = 10
	d12  dndDie = 12
	d20  dndDie = 20
	d100 dndDie = 100

	defaultDnDButtonDieFaces dndDie = d20
)

func allDnDDice() []dndDie {
	return []dndDie{d4, d6, d8, d10, d12, d20, d100}
}

func addDnDButtons(s *discordgo.Session, i *discordgo.InteractionCreate) {
	const maxNumDice = maxDiscordSelectMenuOpts
	diceCountSelectMenu := discordgo.SelectMenu{
		CustomID:    dndDiceCountSelectID,
		Placeholder: "How many dice?",
		Options:     make([]discordgo.SelectMenuOption, maxNumDice),
	}
	for i := maxNumDice; i > 0; i-- {
		diceCountSelectMenu.Options[i-1] = discordgo.SelectMenuOption{
			Label:       fmt.Sprintf("%d 🎲", i),
			Description: fmt.Sprintf("Roll %d dice", i),
			Value:       fmt.Sprint(i),
			Default:     false,
		}
	}
	diceCountSelectMenu.Options[0].Default = true

	diceFacesSelectMenu := discordgo.SelectMenu{
		CustomID:    dndDiceFacesSelectID,
		Placeholder: "How many faces?",
		Options:     make([]discordgo.SelectMenuOption, len(allDnDDice())),
	}
	for i, die := range allDnDDice() {
		dflt := false
		if die == d20 {
			dflt = true
		}
		diceFacesSelectMenu.Options[i] = discordgo.SelectMenuOption{
			Label:       fmt.Sprintf("%s 🔢", die),
			Description: fmt.Sprintf("Roll a %s", die),
			Value:       fmt.Sprint(die),
			Default:     dflt,
		}
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Select the desired options, then press the button to roll the dice!",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{diceCountSelectMenu},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{diceFacesSelectMenu},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    dndOtherOptionsSelectID,
							Placeholder: "Additional Options ❔",
							Options: []discordgo.SelectMenuOption{
								{
									Label:       dndDiceRollOptNone,
									Description: "No additional options desired",
									Value:       dndDiceRollOptNone,
								},
								{
									Label:       dndDiceRollOptDM,
									Description: "Get the result as a direct message.",
									Value:       dndDiceRollOptDM,
								},
								{
									Label:       dndDiceRollOptEphemeral,
									Description: "Get the result as a message that only you can see and dismiss.",
									Value:       dndDiceRollOptEphemeral,
								},
							},
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Roll the 🎲!",
							Style:    discordgo.PrimaryButton,
							CustomID: dndRollButtonID,
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
	}
}

func handleDnDButtonPress(s *discordgo.Session, i *discordgo.InteractionCreate) {
	metadata, err := getInteractionMetaData(i)
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	}

	cfg, ok := dndDiceRollMsgConfigs.Get(string(getDndDiceRollMsgKey(*metadata)))
	if !ok {
		cfg = newDnDRollButtonConfig(metadata.MessageID, metadata.AuthorID)
	}

	faces := cfg.Faces
	content := fmt.Sprintf("%s rolled %d%s:\n", metadata.AuthorMention, cfg.DiceCount, faces)

	total := uint64(0)
	for j := uint64(0); j < cfg.DiceCount; j++ {
		rollResult, err := random.Range(1, uint64(faces))
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

	var flags discordgo.MessageFlags = 0
	if cfg.Ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	if cfg.DM {
		uc, err := s.UserChannelCreate(metadata.AuthorID)
		if err != nil {
			log.Error(err)
			interactionRespondEphemeralError(s, i, true, err)
			return
		}

		_, err = s.ChannelMessageSend(uc.ID, content)
		if err != nil {
			log.Error(err)
			interactionRespondEphemeralError(s, i, true, err)
			return
		}
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})
		if err != nil {
			log.Error(err)
			interactionRespondEphemeralError(s, i, true, err)
		}
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   flags,
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

	cfg, ok := dndDiceRollMsgConfigs.Get(string(getDndDiceRollMsgKey(*metadata)))
	if !ok {
		cfg = newDnDRollButtonConfig(metadata.MessageID, metadata.AuthorID)
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
	dndDiceRollMsgConfigs.Set(string(getDndDiceRollMsgKey(*metadata)), cfg)

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
	}
}

func handleDiceFacesMenuSelection(s *discordgo.Session, i *discordgo.InteractionCreate) {
	metadata, err := getInteractionMetaData(i)
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	}

	cfg, ok := dndDiceRollMsgConfigs.Get(string(getDndDiceRollMsgKey(*metadata)))
	if !ok {
		cfg = newDnDRollButtonConfig(metadata.MessageID, metadata.AuthorID)
	}

	if len(i.MessageComponentData().Values) == 0 {
		err = fmt.Errorf("no values sent")
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	}

	var die dndDie
	err = die.parseFromString(i.MessageComponentData().Values[0])
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	}

	cfg.Faces = die
	dndDiceRollMsgConfigs.Set(string(getDndDiceRollMsgKey(*metadata)), cfg)

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
	}
}

func handleDnDOtherOptionsSelection(s *discordgo.Session, i *discordgo.InteractionCreate) {
	metadata, err := getInteractionMetaData(i)
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	}

	cfg, ok := dndDiceRollMsgConfigs.Get(string(getDndDiceRollMsgKey(*metadata)))
	if !ok {
		cfg = newDnDRollButtonConfig(metadata.MessageID, metadata.AuthorID)
	}

	dm := false
	ephemeral := false
	for _, val := range i.MessageComponentData().Values {
		if val == dndDiceRollOptNone {
			dm = false
			ephemeral = false
			break
		} else if val == dndDiceRollOptEphemeral {
			ephemeral = true
		} else if val == dndDiceRollOptDM {
			dm = true
		}
	}

	cfg.DM = dm
	cfg.Ephemeral = ephemeral
	dndDiceRollMsgConfigs.Set(string(getDndDiceRollMsgKey(*metadata)), cfg)
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
	}
}

type dndRollButtonConfig struct {
	UserID    string
	MsgID     string
	DiceCount uint64
	Faces     dndDie
	Ephemeral bool
	DM        bool
}

func newDnDRollButtonConfig(msgID, userID string) dndRollButtonConfig {
	return dndRollButtonConfig{
		UserID:    userID,
		MsgID:     msgID,
		DiceCount: 1,
		Faces:     defaultDnDButtonDieFaces,
		Ephemeral: false,
		DM:        false,
	}
}

type dndDiceRollMsgKey string

// Map of dndDiceRollMsgKey to dndRollButtonConfigs
var dndDiceRollMsgConfigs = cmap.New[dndRollButtonConfig]()

func getDndDiceRollMsgKey(mdata interactionMetaData) dndDiceRollMsgKey {
	return dndDiceRollMsgKey(mdata.MessageID + mdata.AuthorID)
}
