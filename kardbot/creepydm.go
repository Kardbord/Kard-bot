package kardbot

import (
	"fmt"
	"math/rand"

	"github.com/bwmarrin/discordgo"

	log "github.com/sirupsen/logrus"
)

const (
	creepyDMGet     = "get-creepy-dm"
	creepyChannelDM = "to-channel"
	creepyDMOptIn   = "opt-in"
	creepyDMOptOut  = "opt-out"
)

func creepyDMHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	wg := bot().updateLastActive()
	defer wg.Wait()

	if s == nil || i == nil {
		log.Errorf("nil session or interaction; s=%v, i=%v", s, i)
		return
	}

	switch i.ApplicationCommandData().Options[0].Name {
	case creepyDMGet:
		getCreepyDM(s, i)
	case creepyDMOptIn:
		creepyDMsOptIn(s, i)
	case creepyDMOptOut:
		creepyDMsOptOut(s, i)
	default:
		log.Error("Unknown subcommand")
	}
}

func creepyDMsOptIn(s *discordgo.Session, i *discordgo.InteractionCreate) {
	author, authorID, err := getInteractionCreateAuthorNameAndID(i)
	if err != nil {
		log.Error(err)
		return
	}

	bot().creepyDMSubsMutex.Lock()
	defer bot().creepyDMSubsMutex.Unlock()
	bot().CreepyDMSubs[authorID] = true
	log.Infof("User %s subscribed to creepy DMs", author)

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%s subscribed to creepy DMs 😈", author),
		},
	})
	if err != nil {
		log.Error(err)
	}
}

func creepyDMsOptOut(s *discordgo.Session, i *discordgo.InteractionCreate) {
	author, authorID, err := getInteractionCreateAuthorNameAndID(i)
	if err != nil {
		log.Error(err)
		return
	}

	bot().creepyDMSubsMutex.Lock()
	defer bot().creepyDMSubsMutex.Unlock()
	bot().CreepyDMSubs[authorID] = false
	log.Infof("User %s unsubscribed from creepy DMs", author)

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%s unsubscribed from creepy DMs", author),
		},
	})
	if err != nil {
		log.Error(err)
	}
}

func getCreepyDM(s *discordgo.Session, i *discordgo.InteractionCreate) {
	msg := bot().CreepyDMs[rand.Intn(len(bot().CreepyDMs))]

	sendToChannel := false
	if len(i.ApplicationCommandData().Options[0].Options) > 0 {
		sendToChannel = i.ApplicationCommandData().Options[0].Options[0].BoolValue()
	}

	if sendToChannel {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: msg,
			},
		})
		if err != nil {
			log.Error(err)
		}
		return
	}

	author, authorID, err := getInteractionCreateAuthorNameAndID(i)
	if err != nil {
		log.Error(err)
		return
	}

	uc, err := s.UserChannelCreate(authorID)
	if err != nil {
		log.Error(err)
		return
	}

	_, err = s.ChannelMessageSend(uc.ID, msg)
	if err != nil {
		log.Error(err)
		return
	}
	log.Tracef("Sent %s a creepy DM", author)

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%s was sent a creepy DM", author),
		},
	})
	if err != nil {
		log.Error(err)
	}
}
