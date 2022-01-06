package kardbot

import (
	"fmt"

	"github.com/TannerKvarfordt/Kard-bot/kardbot/dg_helpers"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

const (
	embedCmd = "embed"

	embedSubCmdCreate = "create"

	embedSubCmdUpdate         = "update"
	embedSubCmdUpdateOptMsgID = "message-id"

	embedSubCmdOptURL          = "url"
	embedSubCmdOptTitle        = "title"
	embedSubCmdOptDesc         = "description"
	embedSubCmdOptColor        = "color"
	embedSubCmdOptFooter       = "footer"
	embedSubCmdOptImageURL     = "image-url"
	embedSubCmdOptThumbnailURL = "thumbnail-url"
	// TODO: support fields
)

func embedCmdOpts() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        embedSubCmdCreate,
			Description: "Create a new embed",
			Options:     embedCmdSubCmdOpts(),
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        embedSubCmdUpdate,
			Description: "Update an existing embed",
			Options: append([]*discordgo.ApplicationCommandOption{{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        embedSubCmdUpdateOptMsgID,
				Description: "Message ID containing the embed to update (enable developer options, right click, copy ID)",
				Required:    true,
			}}, embedCmdSubCmdOpts()...),
		},
	}
}

func embedCmdSubCmdOpts() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        embedSubCmdOptTitle,
			Description: "Title for the embed",
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        embedSubCmdOptURL,
			Description: "URL to link to from the embed's title",
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        embedSubCmdOptDesc,
			Description: "Description portion of the embed",
		},
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        embedSubCmdOptColor,
			Description: "Color to use for the embed sidebar",
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{
					Name:  "Red",
					Value: 0xFF0000,
				},
				{
					Name:  "Orange",
					Value: 0xFFA500,
				},
				{
					Name:  "Yellow",
					Value: 0xFFFF00,
				},
				{
					Name:  "Green",
					Value: 0x008000,
				},
				{
					Name:  "Blue",
					Value: 0x0000FF,
				},
				{
					Name:  "Purple",
					Value: 0x800080,
				},
				{
					Name:  "Brown",
					Value: 0x964B00,
				},
				{
					Name:  "Black",
					Value: 0x000000,
				},
				{
					Name:  "White",
					Value: 0xFFFFFF,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        embedSubCmdOptFooter,
			Description: "Footer for the embed",
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        embedSubCmdOptImageURL,
			Description: "URL of the image to display in the embed",
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        embedSubCmdOptThumbnailURL,
			Description: "URL of the thumbnail to display in the embed",
		},
	}
}

func handleEmbedCmd(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if s == nil || i == nil {
		log.Errorf("nil Session pointer (%v) and/or InteractionCreate pointer (%v)", s, i)
		return
	}

	var (
		err  error                          = nil
		resp *discordgo.InteractionResponse = nil
	)
	switch i.ApplicationCommandData().Options[0].Name {
	case embedSubCmdCreate:
		resp, err = handleEmbedSubCmdCreate(s, i)
	case embedSubCmdUpdate:
		resp, err = handleEmbedSubCmdUpdate(s, i)
	default:
		interactionRespondEphemeralError(s, i, true, fmt.Errorf("unknown subcommand"))
	}

	if err != nil {
		interactionRespondEphemeralError(s, i, false, err)
		return
	}
	if resp == nil {
		interactionRespondEphemeralError(s, i, true, fmt.Errorf("nil response returned"))
		log.Error(err)
		return
	}

	err = s.InteractionRespond(i.Interaction, resp)
	if err != nil {
		interactionRespondEphemeralError(s, i, true, err)
		log.Error(err)
		return
	}
}

func handleEmbedSubCmdCreate(s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error) {
	e := dg_helpers.NewEmbed()
	embedEmpty := true
	for _, opt := range i.ApplicationCommandData().Options[0].Options {
		switch opt.Name {
		case embedSubCmdOptURL:
			e.SetURL(opt.StringValue())
			if !isReachableURL(opt.StringValue()) {
				return nil, fmt.Errorf("invalid URL: %s", opt.StringValue())
			}
			// Doesn't count as a non-empty embed on its own
		case embedSubCmdOptTitle:
			e.SetTitle(opt.StringValue())
			embedEmpty = false
		case embedSubCmdOptDesc:
			e.SetDescription(opt.StringValue())
			embedEmpty = false
		case embedSubCmdOptColor:
			e.SetColor(int(opt.IntValue()))
			// Doesn't count as a non-empty embed on its own
		case embedSubCmdOptFooter:
			e.SetFooter(opt.StringValue())
			embedEmpty = false
		case embedSubCmdOptImageURL:
			e.SetImage(opt.StringValue())
			if !isReachableURL(opt.StringValue()) {
				return nil, fmt.Errorf("invalid URL: %s", opt.StringValue())
			}
			embedEmpty = false
		case embedSubCmdOptThumbnailURL:
			e.SetThumbnail(opt.StringValue())
			if !isReachableURL(opt.StringValue()) {
				return nil, fmt.Errorf("invalid URL: %s", opt.StringValue())
			}
			embedEmpty = false
		default:
			log.Warn("Unknown option: ", opt.Name)
		}
	}

	var (
		embeds  []*discordgo.MessageEmbed = nil
		content string                    = "Cannot create an empty embed"
		flags   uint64                    = InteractionResponseFlagEphemeral
	)
	if !embedEmpty {
		flags = 0
		content = ""
		embeds = []*discordgo.MessageEmbed{e.Truncate().SetType(discordgo.EmbedTypeRich).MessageEmbed}
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   flags,
			Embeds:  embeds,
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{
					discordgo.AllowedMentionTypeEveryone,
					discordgo.AllowedMentionTypeRoles,
					discordgo.AllowedMentionTypeUsers,
				},
			},
		},
	}, nil
}

func handleEmbedSubCmdUpdate(s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error) {
	return nil, fmt.Errorf("unimplemented")
}
