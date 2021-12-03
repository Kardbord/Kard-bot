package kardbot

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/TannerKvarfordt/Kard-bot/kardbot/dg_helpers"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

const InteractionResponseFlagEphemeral = uint64(1 << 6)

const genericErrorString = "an error occurred. :'("

func interactionRespondWithEphemeralError(s *discordgo.Session, i *discordgo.InteractionCreate, respType discordgo.InteractionResponseType, errStr string) {
	if s == nil {
		log.Error("nil session")
		return
	}
	if i == nil {
		log.Error("nil interaction")
		return
	}
	if errStr == "" {
		log.Warn("empty errStr, using generic error: ", genericErrorString)
		errStr = genericErrorString
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: respType,
		Data: &discordgo.InteractionResponseData{
			Content: errStr,
			Flags:   InteractionResponseFlagEphemeral,
		},
	})
	if err != nil {
		log.Error(err)
	}
}

func interactionRespondWithEphemeralErrorAndNotifyOwner(s *discordgo.Session, i *discordgo.InteractionCreate, respType discordgo.InteractionResponseType, errResp error) {
	ownerID := getOwnerID()
	if ownerID == "" {
		ownerID = "The bot owner"
	} else {
		ownerID = fmt.Sprintf("<@%s>", ownerID)
	}
	interactionRespondWithEphemeralError(s, i, respType, fmt.Sprintf("Something went wrong while processing your command. 😔 %s has been notified.", ownerID))

	metadata, err := getInteractionMetaData(i)
	if err != nil {
		log.Error(err)
		return
	}

	uc, err := bot().Session.UserChannelCreate(metadata.AuthorID)
	if err != nil {
		log.Error(err)
	}

	cmdJson, err := json.MarshalIndent(i.ApplicationCommandData(), "", "  ")
	if err != nil {
		log.Error(err)
		cmdJson = []byte(`"error": "Error marshalling application command"`)
	}

	_, err = bot().Session.ChannelMessageSendComplex(uc.ID, &discordgo.MessageSend{
		Embed: dg_helpers.NewEmbed().
			SetTitle("Error Report").
			AddField("Afflicted User", metadata.AuthorMention).
			AddField("Issued Command", fmt.Sprintf("```json\n%s\n```", cmdJson)).
			AddField("Error", fmt.Sprintf("```\n%s\n```", errResp)).
			Truncate().
			MessageEmbed,
	})
	if err != nil {
		log.Error(err)
	}
}

func authorIsSelf(s *discordgo.Session, i *discordgo.InteractionCreate) (bool, error) {
	if s == nil || i == nil {
		return false, fmt.Errorf("interaction or session is nil")
	}
	metadata, err := getInteractionMetaData(i)
	if err != nil {
		return false, err
	}
	return metadata.AuthorID == s.State.User.ID, nil
}

func authorIsOwner(i *discordgo.InteractionCreate) (bool, error) {
	if getOwnerID() == "" {
		return false, errors.New("owner ID is not set")
	}
	if i == nil {
		return false, errors.New("context is nil")
	}

	if i.Member != nil {
		if i.Member.User == nil {
			return false, errors.New("member.user is nil")
		}
		return i.Member.User.ID == getOwnerID(), nil
	} else if i.User != nil {
		return i.User.ID == getOwnerID(), nil
	} else {
		return false, errors.New("member and user are nil")
	}
}

type interactionMetaData struct {
	AuthorID       string
	AuthorUsername string
	AuthorMention  string
	AuthorEmail    string
	GuildID        string
	ChannelID      string
	InteractionID  string
}

func getInteractionMetaData(i *discordgo.InteractionCreate) (*interactionMetaData, error) {
	if i == nil {
		return nil, errors.New("interaction is nil")
	}

	if i.Member != nil {
		if i.Member.User == nil {
			return nil, errors.New("member.user is nil")
		}
		return &interactionMetaData{
			AuthorID:       i.Member.User.ID,
			AuthorUsername: i.Member.User.Username,
			AuthorMention:  i.Member.User.Mention(),
			AuthorEmail:    i.Member.User.Email,
			GuildID:        i.GuildID,
			ChannelID:      i.ChannelID,
			InteractionID:  i.ID,
		}, nil
	}

	if i.User != nil {
		return &interactionMetaData{
			AuthorID:       i.User.ID,
			AuthorUsername: i.User.Username,
			AuthorMention:  i.User.Mention(),
			AuthorEmail:    i.User.Email,
			GuildID:        i.GuildID,
			ChannelID:      i.ChannelID,
			InteractionID:  i.ID,
		}, nil
	}

	return nil, errors.New("no metadata could be found")
}

func channelIsNSFW(s *discordgo.Session, i *discordgo.InteractionCreate) (bool, error) {
	if s == nil {
		return false, fmt.Errorf("session is nil")
	}
	if i == nil {
		return false, fmt.Errorf("interaction is nil")
	}

	ch, err := s.Channel(i.ChannelID)
	if err != nil {
		return false, err
	}
	if ch == nil {
		return false, fmt.Errorf("could not retrieve channel with ID %s", i.ChannelID)
	}

	// DMs are considered nsfw
	if ch.Type == discordgo.ChannelTypeDM {
		return true, nil
	}

	return ch.NSFW, nil
}
