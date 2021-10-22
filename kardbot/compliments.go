package kardbot

import (
	"fmt"
	"math/rand"

	"github.com/bwmarrin/discordgo"

	log "github.com/sirupsen/logrus"
)

const (
	complimentsOptIn   = "opt-in"
	complimentsOptOut  = "opt-out"
	complimentsMorning = "morning"
	complimentsEvening = "evening"
	complimentsGet     = "get-compliment"
	complimentInDM     = "dm"
)

func complimentHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	wg := bot().updateLastActive()
	defer wg.Wait()

	if s == nil || i == nil {
		log.Errorf("nil session or interaction; s=%v, i=%v", s, i)
		return
	}

	switch i.ApplicationCommandData().Options[0].Name {
	case complimentsOptIn:
		switch i.ApplicationCommandData().Options[0].Options[0].Name {
		case complimentsMorning:
			morningComplimentOptIn(s, i)
		case complimentsEvening:
			eveningComplimentOptIn(s, i)
		default:
			log.Error("Unknown subcommand")
		}
	case complimentsOptOut:
		switch i.ApplicationCommandData().Options[0].Options[0].Name {
		case complimentsMorning:
			morningComplimentOptOut(s, i)
		case complimentsEvening:
			eveningComplimentOptOut(s, i)
		default:
			log.Error("Unknown subcommand")
		}
	case complimentsGet:
		getCompliment(s, i)
	default:
		log.Error("Unknown subcommand group")
	}
}

func morningComplimentOptIn(s *discordgo.Session, i *discordgo.InteractionCreate) {
	author, authorID, err := getInteractionCreateAuthorNameAndID(i)
	if err != nil {
		log.Error(err)
		return
	}

	bot().complimentSubsAMMutex.Lock()
	defer bot().complimentSubsAMMutex.Unlock()
	bot().ComplimentSubsAM[authorID] = true
	log.Infof("User %s subscribed to morning compliments", author)

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%s has subscribed to receive daily morning compliments. :)", author),
		},
	})
	if err != nil {
		log.Error(err)
	}
}

func morningComplimentOptOut(s *discordgo.Session, i *discordgo.InteractionCreate) {
	author, authorID, err := getInteractionCreateAuthorNameAndID(i)
	if err != nil {
		log.Error(err)
		return
	}

	bot().complimentSubsAMMutex.Lock()
	defer bot().complimentSubsAMMutex.Unlock()
	bot().ComplimentSubsAM[authorID] = false
	log.Infof("User %s un-subscribed to morning compliments", author)

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%s has unsubscribed from daily morning compliments. :(", author),
		},
	})
	if err != nil {
		log.Error(err)
	}
}

func eveningComplimentOptIn(s *discordgo.Session, i *discordgo.InteractionCreate) {
	author, authorID, err := getInteractionCreateAuthorNameAndID(i)
	if err != nil {
		log.Error(err)
		return
	}

	bot().complimentSubsPMMutex.Lock()
	defer bot().complimentSubsPMMutex.Unlock()
	bot().ComplimentSubsPM[authorID] = true
	log.Infof("User %s subscribed to evening compliments", author)

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%s has subscribed to receive daily evening compliments. :)", author),
		},
	})
	if err != nil {
		log.Error(err)
	}
}

func eveningComplimentOptOut(s *discordgo.Session, i *discordgo.InteractionCreate) {
	author, authorID, err := getInteractionCreateAuthorNameAndID(i)
	if err != nil {
		log.Error(err)
		return
	}

	bot().complimentSubsPMMutex.Lock()
	defer bot().complimentSubsPMMutex.Unlock()
	bot().ComplimentSubsPM[authorID] = false
	log.Infof("User %s un-subscribed to evening compliments", author)

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%s has unsubscribed from daily evening compliments. :(", author),
		},
	})
	if err != nil {
		log.Error(err)
	}
}

func getCompliment(s *discordgo.Session, i *discordgo.InteractionCreate) {
	author, authorID, err := getInteractionCreateAuthorNameAndID(i)
	if err != nil {
		log.Error(err)
		return
	}

	compliment := bot().Compliments[rand.Intn(len(bot().Compliments))]

	sendAsDM := false
	if len(i.ApplicationCommandData().Options[0].Options) > 0 {
		sendAsDM = i.ApplicationCommandData().Options[0].Options[0].BoolValue()
	}

	if sendAsDM {
		uc, err := bot().Session.UserChannelCreate(authorID)
		if err != nil {
			log.Error(err)
		}
		_, err = bot().Session.ChannelMessageSend(uc.ID, compliment)
		if err != nil {
			log.Error(err)
		}
		log.Infof("Told %s that '%s'", author, compliment)

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("I DM'd a compliment to %s. :)", author),
			},
		})
		if err != nil {
			log.Error(err)
		}
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: compliment,
		},
	})
	if err != nil {
		log.Error(err)
	}
	log.Infof("To %s: \"%s\"", author, compliment)
}
