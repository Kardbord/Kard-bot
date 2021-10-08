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
	_, authorID, err := getInteractionCreateAuthorNameAndID(i)
	if err != nil {
		return false, err
	}
	return authorID == s.State.User.ID, nil
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

func getInteractionCreateAuthorName(i *discordgo.InteractionCreate) (string, error) {
	if i == nil {
		return "", errors.New("context is nil")
	}

	if i.Member != nil {
		if i.Member.User == nil {
			return "", errors.New("member.user is nil")
		}
		return i.Member.User.Username, nil
	} else if i.User != nil {
		return i.User.Username, nil
	} else {
		return "", errors.New("member and user are nil")
	}
}

func getInteractionCreateAuthorID(i *discordgo.InteractionCreate) (string, error) {
	if i == nil {
		return "", errors.New("context is nil")
	}

	if i.Member != nil {
		if i.Member.User == nil {
			return "", errors.New("member.user is nil")
		}
		return i.Member.User.ID, nil
	} else if i.User != nil {
		return i.User.ID, nil
	} else {
		return "", errors.New("member and user are nil")
	}
}

func getInteractionCreateAuthorNameAndID(i *discordgo.InteractionCreate) (string, string, error) {
	id, err1 := getInteractionCreateAuthorID(i)
	uname, err2 := getInteractionCreateAuthorName(i)
	if err1 != nil && err2 != nil {
		return uname, id, fmt.Errorf("error 1:%v\n\terror 2:%v", err1, err2)
	} else if err1 != nil {
		return uname, id, err1
	} else if err2 != nil {
		return uname, id, err2
	}
	return uname, id, nil
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

	return ch.NSFW, nil
}
