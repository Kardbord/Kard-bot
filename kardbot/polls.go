package kardbot

import (
	"fmt"

	"github.com/TannerKvarfordt/Kard-bot/kardbot/dg_helpers"
	"github.com/TannerKvarfordt/ubiquity/mathutils"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

const (
	pollCmd                 = "poll"
	pollCmdOptMaxSelections = "max-selections"
	pollCmdOptTitle         = "title"
	pollCmdOptContext       = "context"

	pollSelectMenuID = "poll-menu"
)

func getPollOpts() []*discordgo.ApplicationCommandOption {
	opts := []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        pollCmdOptMaxSelections,
			Description: "The maximum number of options a user can vote for",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        pollCmdOptTitle,
			Description: fmt.Sprintf("Title of the poll, maximum %d characters.", maxDiscordSelectMenuPlaceholderChars),
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        pollCmdOptContext,
			Description: "Additional context for the poll",
			Required:    false,
		},
	}
	addnlOpts := make([]*discordgo.ApplicationCommandOption, mathutils.Min(maxDiscordCommandOptions-len(opts), maxDiscordSelectMenuOpts, dg_helpers.EmbedLimitField))
	for i := range addnlOpts {
		addnlOpts[i] = &discordgo.ApplicationCommandOption{}
		opt := addnlOpts[i]
		opt.Type = discordgo.ApplicationCommandOptionString
		opt.Name = fmt.Sprintf("option-%d", i)
		opt.Description = fmt.Sprintf("Poll option %d", 1)
		opt.Required = false
	}
	return append(opts, addnlOpts...)
}

func handlePollCmd(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if s == nil || i == nil {
		log.Error(fmt.Errorf("nil Session pointer (%v) and/or InteractionCreate pointer (%v)", s, i))
		return
	}
	wg := bot().updateLastActive()
	defer wg.Wait()

	minSelections := 0
	maxSelections := 1
	title := ""
	context := ""
	pollOpts := make([]discordgo.SelectMenuOption, 0, len(i.ApplicationCommandData().Options))
	for _, opt := range i.ApplicationCommandData().Options {
		switch opt.Name {
		case pollCmdOptMaxSelections:
			maxSelections = int(opt.IntValue())
		case pollCmdOptTitle:
			title = opt.StringValue()
		case pollCmdOptContext:
			context = opt.StringValue()
		default:
			emoji, trimmedLabel, err := detectAndScrubDiscordEmojis(opt.StringValue())
			if err != nil {
				log.Error(err)
				interactionRespondEphemeralError(s, i, true, err)
				return
			}
			pollOpts = append(pollOpts, discordgo.SelectMenuOption{
				Label: trimmedLabel,
				Value: trimmedLabel,
				Emoji: emoji,
			})
		}
	}

	if maxSelections < 1 {
		interactionRespondEphemeralError(s, i, false, fmt.Errorf("you must allow at least 1 vote to be cast per user"))
		return
	}

	if len(pollOpts) == 0 {
		interactionRespondEphemeralError(s, i, false, fmt.Errorf("you must specify at least one poll option"))
		return
	}

	maxSelections = mathutils.Min(len(pollOpts), maxSelections)

	color, _ := fastHappyColorInt64()
	e := dg_helpers.NewEmbed().
		SetColor(int(color)).
		SetTitle(title).
		SetDescription(context)

	for _, opt := range pollOpts {
		// IMPORTANT: if you change the structure of this field value,
		// you may also need to update the implementation of handlePollSubmission.
		e.AddField(opt.Label, "👍 0 votes, 🗠 0% of votes cast")
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    pollSelectMenuID,
							Placeholder: title,
							MinValues:   &minSelections,
							MaxValues:   maxSelections,
							Options:     pollOpts,
						},
					},
				},
			},
			Embeds: []*discordgo.MessageEmbed{e.Truncate().MessageEmbed},
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{
					discordgo.AllowedMentionTypeEveryone,
					discordgo.AllowedMentionTypeRoles,
					discordgo.AllowedMentionTypeUsers,
				},
			},
		},
	})
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
	}
}

func handlePollSubmission(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if s == nil || i == nil {
		log.Error(fmt.Errorf("nil Session pointer (%v) and/or InteractionCreate pointer (%v)", s, i))
		return
	}
	wg := bot().updateLastActive()
	defer wg.Wait()

	//metadata, err := getInteractionMetaData(i)
	//if err != nil {
	//	log.Error(err)
	//	interactionRespondEphemeralError(s, i, true, err)
	//	return
	//}

	// TODO: Handle the response.
	// TODO: Protect with each poll msgID with a mutex.
	//       Can probably be done with a thread-safe map.
	//       No need to write anything to disk, just check
	//       that no other routines are updating the same
	//       message ID.

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Your response was submitted! 🗳️",
			Flags:   InteractionResponseFlagEphemeral,
		},
	})
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	}
}
