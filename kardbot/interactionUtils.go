package kardbot

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

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
