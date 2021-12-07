package kardbot

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

const delBotDMCmd = "del-bot-dm"

// API-defined limit of messages that can be retrieved or deleted at once
const msgLimit = 100

func deleteBotDMs(s *discordgo.Session, i *discordgo.InteractionCreate) {
	wg := bot().updateLastActive()
	defer wg.Wait()

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
			Flags: InteractionResponseFlagEphemeral,
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
		s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
			Content: fmt.Sprintf("looks like you tried to use `/%s` outside of our DMs. Run it from there instead! :)", delBotDMCmd),
		})
		return
	}

	msgsToDelete := int(i.ApplicationCommandData().Options[0].IntValue())
	if msgsToDelete <= 0 {
		interactionFollowUpEphemeralError(s, i, false, fmt.Errorf("you must specify a positive, non-zero number of messages to delete"))
		return
	}

	msgs, err := s.ChannelMessages(ch.ID, MinOf(msgsToDelete, msgLimit), ch.LastMessageID, "", "")
	if err != nil {
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
		return
	} else {
		lastMsg, err := s.ChannelMessage(ch.ID, ch.LastMessageID)
		if err != nil {
			log.Error(err)
			interactionFollowUpEphemeralError(s, i, true, err)
			return
		}
		msgs = append(msgs, lastMsg)
	}

	// No way to bulk delete messages in a DM channel
	for _, msg := range msgs {
		msgAuthorID := ""
		if msg.Author != nil {
			msgAuthorID = msg.Author.ID
		} else if msg.Member != nil && msg.Member.User != nil {
			msgAuthorID = msg.Member.User.ID
		}
		if msgAuthorID == "" {
			log.Error("Could not ascertain msg author, skipping")
			continue
		}

		if msgAuthorID != s.State.User.ID {
			log.Debug("Not deleting message not from self")
			continue
		}
		err = s.ChannelMessageDelete(ch.ID, msg.ID)
		if err != nil {
			log.Error(err)
		}
	}

	_, err = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
		Content: fmt.Sprintf("Deleted last %d bot DMs", MinOf(msgsToDelete, msgLimit)),
	})
	if err != nil {
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
	}
}
