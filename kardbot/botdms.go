package kardbot

import (
	"fmt"
	"sort"
	"time"

	"github.com/TannerKvarfordt/ubiquity/mathutils"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

const delBotDMCmd = "del-bot-dm"

// API-defined limit of messages that can be retrieved or deleted at once
const msgLimit = 100

func deleteBotDMs(s *discordgo.Session, i *discordgo.InteractionCreate) {
	fromSelf, err := authorIsSelf(s, i)
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	}
	if fromSelf {
		log.Warn("Ignoring deleteDM request from self")
		return
	}

	imeta, err := getInteractionMetaData(i)
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
		return
	}

	ch, err := s.Channel(imeta.ChannelID)
	if err != nil {
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
		return
	}
	if ch.Type != discordgo.ChannelTypeDM {
		uc, err := s.UserChannelCreate(imeta.AuthorID)
		if err != nil {
			log.Error(err)
			interactionFollowUpEphemeralError(s, i, true, err)
			return
		}

		_, err = s.ChannelMessageSend(uc.ID, fmt.Sprintf("Looks like you tried to use `/%s` outside of our DMs. Run it from here instead! :)", delBotDMCmd))
		if err != nil {
			log.Error(err)
			interactionFollowUpEphemeralError(s, i, true, err)
			return
		}

		time.Sleep(time.Millisecond * 100)
		errMsg := fmt.Sprintf("looks like you tried to use `/%s` outside of our DMs. Run it from there instead! :)", delBotDMCmd)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &errMsg,
		})
		return
	}

	msgsToDelete := int(i.ApplicationCommandData().Options[0].IntValue())
	if msgsToDelete <= 0 {
		interactionFollowUpEphemeralError(s, i, false, fmt.Errorf("you must specify a positive, non-zero number of messages to delete"))
		return
	}

	msgs, err := s.ChannelMessages(ch.ID, mathutils.Min(msgsToDelete, msgLimit), "", "", "")
	if err != nil {
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
		return
	}

	// Ensure messages are sorted
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].Timestamp.Before(msgs[j].Timestamp)
	})

	// No way to bulk delete messages in a DM channel
	deletedCount := 0
	for i := 0; i < len(msgs); i++ {
		msg := msgs[i]
		msgAuthorID := ""
		if msg.Author != nil {
			msgAuthorID = msg.Author.ID
		} else if msg.Member != nil && msg.Member.User != nil {
			msgAuthorID = msg.Member.User.ID
		}
		if msgAuthorID == "" {
			log.Errorf("Could not ascertain msg author, skipping msg:\n> %s", msg.Content)
			continue
		}

		if msgAuthorID != s.State.User.ID {
			log.Tracef("Not deleting user message:\n> %s", msg.Content)
			continue
		}
		err = s.ChannelMessageDelete(ch.ID, msg.ID)
		if err != nil {
			log.Error(err)
		}
		deletedCount++
	}

	errMsg := fmt.Sprintf("Skipped %d user DMs. Deleted last %d bot DMs.", mathutils.Min(msgsToDelete, msgLimit)-deletedCount, deletedCount)
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &errMsg,
	})
	if err != nil {
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
	}
}
