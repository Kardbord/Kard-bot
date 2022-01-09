package kardbot

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/TannerKvarfordt/Kard-bot/kardbot/dg_helpers"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

const (
	embedCmd = "embed"

	embedSubCmdCreate           = "create"
	embedSubCmdCreateOptPreview = "preview"

	embedSubCmdUpdate         = "update"
	embedSubCmdUpdateOptMsgID = "message-id"

	embedSubCmdOptURL          = "url"
	embedSubCmdOptTitle        = "title"
	embedSubCmdOptDesc         = "description"
	embedSubCmdOptColor        = "color"
	embedSubCmdOptFooter       = "footer"
	embedSubCmdOptImageURL     = "image-url"
	embedSubCmdOptThumbnailURL = "thumbnail-url"

	embedSubCmdAddField      = "add-field"
	embedSubCmdOptFieldTitle = "title"
	embedSubCmdOptFieldValue = "content"

	embedSubCmdDelField    = "delete-field"
	embedSubCmdOptFieldIdx = "field-index"
)

const authorIDFieldTitle = "Embed Author"

func embedCmdOpts() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        embedSubCmdCreate,
			Description: "Create a new embed",
			Options: append(embedCmdSubCmdOpts(), &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        embedSubCmdCreateOptPreview,
				Description: "If true, the bot will respone ephemerally",
			}),
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        embedSubCmdAddField,
			Description: "Add a field to an updateable embed",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        embedSubCmdUpdateOptMsgID,
					Description: "Message ID containing the embed to update (enable developer options, right click, copy ID)",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        embedSubCmdOptFieldTitle,
					Description: "Title of the field to be added",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        embedSubCmdOptFieldValue,
					Description: "Content of the field to be added",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        embedSubCmdDelField,
			Description: "Delete an embed field by (zero-based) index",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        embedSubCmdUpdateOptMsgID,
					Description: "Message ID containing the embed to update (enable developer options, right click, copy ID)",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        embedSubCmdOptFieldIdx,
					Description: "Zero-based (first field is the 0th field) index of the field to remove",
					Required:    true,
				},
			},
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
		err           error                          = nil
		reportableErr                                = false
		resp          *discordgo.InteractionResponse = nil
	)
	switch i.ApplicationCommandData().Options[0].Name {
	case embedSubCmdCreate:
		resp, reportableErr, err = handleEmbedSubCmdCreate(s, i)
	case embedSubCmdUpdate:
		resp, reportableErr, err = handleEmbedSubCmdUpdate(s, i)
	case embedSubCmdAddField:
		resp, reportableErr, err = handleEmbedSubCmdAddField(s, i)
	case embedSubCmdDelField:
		resp, reportableErr, err = handleEmbedSubCmdDelField(s, i)
	default:
		interactionRespondEphemeralError(s, i, true, fmt.Errorf("unknown subcommand"))
	}

	if err != nil {
		interactionRespondEphemeralError(s, i, reportableErr, err)
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

func handleEmbedSubCmdCreate(s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, bool, error) {
	e := dg_helpers.NewEmbed()
	flags := uint64(0)
	for _, opt := range i.ApplicationCommandData().Options[0].Options {
		switch opt.Name {
		case embedSubCmdCreateOptPreview:
			if opt.BoolValue() {
				flags = InteractionResponseFlagEphemeral
			}
		case embedSubCmdOptURL:
			e.SetURL(opt.StringValue())
			if !isReachableURL(opt.StringValue()) {
				return nil, false, fmt.Errorf("invalid URL: %s", opt.StringValue())
			}
			// Doesn't count as a non-empty embed on its own
		case embedSubCmdOptTitle:
			e.SetTitle(opt.StringValue())
		case embedSubCmdOptDesc:
			e.SetDescription(opt.StringValue())
		case embedSubCmdOptColor:
			e.SetColor(int(opt.IntValue()))
			// Doesn't count as a non-empty embed on its own
		case embedSubCmdOptFooter:
			e.SetFooter(opt.StringValue())
		case embedSubCmdOptImageURL:
			e.SetImage(opt.StringValue())
			if !isReachableURL(opt.StringValue()) {
				return nil, false, fmt.Errorf("invalid URL: %s", opt.StringValue())
			}
		case embedSubCmdOptThumbnailURL:
			e.SetThumbnail(opt.StringValue())
			if !isReachableURL(opt.StringValue()) {
				return nil, false, fmt.Errorf("invalid URL: %s", opt.StringValue())
			}
		default:
			log.Warn("Unknown option: ", opt.Name)
		}
	}

	mdata, err := getInteractionMetaData(i)
	if err != nil {
		return nil, true, err
	}
	e.Fields = append([]*discordgo.MessageEmbedField{{
		Name:  authorIDFieldTitle,
		Value: mdata.AuthorMention,
	}}, e.Fields...)

	if e.IsEmpty() {
		return nil, false, fmt.Errorf("cannot create an empty embed")
	}

	embeds := []*discordgo.MessageEmbed{e.Truncate().SetType(discordgo.EmbedTypeRich).MessageEmbed}
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:  flags,
			Embeds: embeds,
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{
					discordgo.AllowedMentionTypeEveryone,
					discordgo.AllowedMentionTypeRoles,
					discordgo.AllowedMentionTypeUsers,
				},
			},
		},
	}, false, nil
}

func handleEmbedSubCmdUpdate(s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, bool, error) {
	e := dg_helpers.NewEmbed()

	metadata, err := getInteractionMetaData(i)
	if err != nil {
		return nil, true, err
	}

	msgToUpdate, reportableErr, err := getUpdateableEmbed(metadata.ChannelID, i.ApplicationCommandData().Options[0].Options[0].StringValue(), metadata.AuthorMention, s)
	if err != nil {
		return nil, reportableErr, err
	}

	// The above call to getUpdateableEmbed verifies that there is exactly one embed
	e.MessageEmbed = msgToUpdate.Embeds[0]

	for _, opt := range i.ApplicationCommandData().Options[0].Options {
		switch opt.Name {
		case embedSubCmdUpdateOptMsgID:
			continue
		case embedSubCmdOptURL:
			e.SetURL(opt.StringValue())
			if !isReachableURL(opt.StringValue()) {
				return nil, false, fmt.Errorf("invalid URL: %s", opt.StringValue())
			}
			// Doesn't count as a non-empty embed on its own
		case embedSubCmdOptTitle:
			e.SetTitle(opt.StringValue())
		case embedSubCmdOptDesc:
			e.SetDescription(opt.StringValue())
		case embedSubCmdOptColor:
			e.SetColor(int(opt.IntValue()))
			// Doesn't count as a non-empty embed on its own
		case embedSubCmdOptFooter:
			e.SetFooter(opt.StringValue())
		case embedSubCmdOptImageURL:
			e.SetImage(opt.StringValue())
			if !isReachableURL(opt.StringValue()) {
				return nil, false, fmt.Errorf("invalid URL: %s", opt.StringValue())
			}
		case embedSubCmdOptThumbnailURL:
			e.SetThumbnail(opt.StringValue())
			if !isReachableURL(opt.StringValue()) {
				return nil, false, fmt.Errorf("invalid URL: %s", opt.StringValue())
			}
		default:
			log.Warn("Unknown option: ", opt.Name)
		}
	}

	if e.IsEmpty() {
		return nil, false, fmt.Errorf("cannot create an empty embed")
	}

	_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Content:    &msgToUpdate.Content,
		Components: msgToUpdate.Components,
		Embeds:     []*discordgo.MessageEmbed{e.Truncate().MessageEmbed},
		AllowedMentions: &discordgo.MessageAllowedMentions{
			Parse: []discordgo.AllowedMentionType{
				discordgo.AllowedMentionTypeEveryone,
				discordgo.AllowedMentionTypeUsers,
				discordgo.AllowedMentionTypeRoles,
			},
		},
		ID:      msgToUpdate.ID,
		Channel: metadata.ChannelID,
	})

	if err != nil {
		return nil, true, err
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   InteractionResponseFlagEphemeral,
			Content: "The embed was successfully updated!",
		},
	}, false, nil
}

func handleEmbedSubCmdAddField(s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, bool, error) {
	e := dg_helpers.NewEmbed()

	metadata, err := getInteractionMetaData(i)
	if err != nil {
		return nil, true, err
	}

	msgToUpdate, reportableErr, err := getUpdateableEmbed(metadata.ChannelID, i.ApplicationCommandData().Options[0].Options[0].StringValue(), metadata.AuthorMention, s)
	if err != nil {
		return nil, reportableErr, err
	}

	e.MessageEmbed = msgToUpdate.Embeds[0]

	var (
		title string = ""
		value string = ""
	)
	for _, opt := range i.ApplicationCommandData().Options[0].Options {
		switch opt.Name {
		case embedSubCmdUpdateOptMsgID:
			continue
		case embedSubCmdOptFieldTitle:
			title = opt.StringValue()
		case embedSubCmdOptFieldValue:
			value = opt.StringValue()
		default:
			log.Warn("Unknown option: ", opt.Name)
		}
	}

	if title == "" {
		return nil, false, fmt.Errorf("you must provide a title for the field")
	}
	if value == "" {
		return nil, false, fmt.Errorf("you must provide content for the field")
	}

	e.AddField(title, value)

	_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Content:    &msgToUpdate.Content,
		Components: msgToUpdate.Components,
		Embeds:     []*discordgo.MessageEmbed{e.Truncate().MessageEmbed},
		AllowedMentions: &discordgo.MessageAllowedMentions{
			Parse: []discordgo.AllowedMentionType{
				discordgo.AllowedMentionTypeEveryone,
				discordgo.AllowedMentionTypeUsers,
				discordgo.AllowedMentionTypeRoles,
			},
		},
		ID:      msgToUpdate.ID,
		Channel: metadata.ChannelID,
	})

	if err != nil {
		return nil, true, err
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   InteractionResponseFlagEphemeral,
			Content: "The embed was successfully updated!",
		},
	}, false, nil
}

func handleEmbedSubCmdDelField(s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, bool, error) {
	e := dg_helpers.NewEmbed()

	metadata, err := getInteractionMetaData(i)
	if err != nil {
		return nil, true, err
	}

	msgToUpdate, reportableErr, err := getUpdateableEmbed(metadata.ChannelID, i.ApplicationCommandData().Options[0].Options[0].StringValue(), metadata.AuthorMention, s)
	if err != nil {
		return nil, reportableErr, err
	}

	e.MessageEmbed = msgToUpdate.Embeds[0]

	idxToDel := i.ApplicationCommandData().Options[0].Options[1].IntValue()

	if idxToDel == 0 {
		return nil, false, fmt.Errorf(`cannot remove the "%s" field`, authorIDFieldTitle)
	}
	if idxToDel < 0 {
		return nil, false, fmt.Errorf("invalid field index (%d), must be a non-negative integer", idxToDel)
	}
	if idxToDel > int64(len(e.Fields)-1) {
		return nil, false, fmt.Errorf("invalid field index (%d), must not be greater than %d for this message", idxToDel, len(e.Fields)-1)
	}

	e.Fields = append(e.Fields[:idxToDel], e.Fields[idxToDel+1:]...)

	_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Content:    &msgToUpdate.Content,
		Components: msgToUpdate.Components,
		Embeds:     []*discordgo.MessageEmbed{e.Truncate().MessageEmbed},
		AllowedMentions: &discordgo.MessageAllowedMentions{
			Parse: []discordgo.AllowedMentionType{
				discordgo.AllowedMentionTypeEveryone,
				discordgo.AllowedMentionTypeUsers,
				discordgo.AllowedMentionTypeRoles,
			},
		},
		ID:      msgToUpdate.ID,
		Channel: metadata.ChannelID,
	})

	if err != nil {
		return nil, true, err
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   InteractionResponseFlagEphemeral,
			Content: "The embed was successfully updated!",
		},
	}, false, nil
}

var embedAuthorMentionRegex = regexp.MustCompile(`^<@\d+>$`)

// Checks that the provided MessageID refers to a message that was authored by the bot,
// contains a single embed, and that the embed was created by the specified authorID
func getUpdateableEmbed(channelID, messageID, authorMention string, s *discordgo.Session) (*discordgo.Message, bool, error) {
	msgToUpdate, err := s.ChannelMessage(channelID, messageID)
	if err != nil {
		return nil, true, err
	}

	const notUpdateableMsg = "Message does not appear to be an updateable embed: "
	if s.State.User.ID != msgToUpdate.Author.ID {
		return nil, false, fmt.Errorf("%snot authored by %s", notUpdateableMsg, s.State.User.Mention())
	}

	if len(msgToUpdate.Embeds) != 1 {
		return nil, false, fmt.Errorf("%sexpected only 1 embed, got %d", notUpdateableMsg, len(msgToUpdate.Embeds))
	}

	// First field should always be the author of the embed
	if len(msgToUpdate.Embeds[0].Fields) < 1 {
		return nil, false, fmt.Errorf("%sexpected at least one field", notUpdateableMsg)
	}

	if msgToUpdate.Embeds[0].Fields[0].Name != authorIDFieldTitle {
		return nil, false, fmt.Errorf("%scould not determine embed author, embed has unexpected format", notUpdateableMsg)
	}

	var authorIDField string = ""
	{
		authorFieldIDs := embedAuthorMentionRegex.FindAllString(msgToUpdate.Embeds[0].Fields[0].Value, -1)
		switch len(authorFieldIDs) {
		case 0:
			break
		case 1:
			authorIDField = authorFieldIDs[0]
		default:
			authorIDField = authorFieldIDs[len(authorFieldIDs)-1]
		}
	}

	if !strings.Contains(authorIDField, authorMention) {
		return nil, false, fmt.Errorf("%syou do not appear to be the author of the embed you wish to edit", notUpdateableMsg)
	}

	return msgToUpdate, false, nil
}
