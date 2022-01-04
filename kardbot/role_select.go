package kardbot

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/TannerKvarfordt/Kard-bot/kardbot/dg_helpers"
	"github.com/bwmarrin/discordgo"
	"github.com/forPelevin/gomoji"
	log "github.com/sirupsen/logrus"
)

const (
	maxRoleSelectMenus    = 4
	roleSelectMenuCommand = "role-select-menu"
	// TODO: Add a help subcommand
)

const (
	roleSelectMenuSubCmdCreate = "create"

	roleSelectMenuCreateOptTitle        = "title"
	roleSelectMenuCreateOptReqTitle     = true
	roleSelectMenuCreateOptRoles        = "roles"
	roleSelectMenuCreateOptReqRoles     = true
	roleSelectMenuCreateOptDesc         = "description"
	roleSelectMenuCreateOptReqDesc      = false
	roleSelectMenuCreateOptURL          = "url"
	roleSelectMenuCreateOptReqURL       = false
	roleSelectMenuCreateOptImage        = "image-url"
	roleSelectMenuCreateOptReqImage     = false
	roleSelectMenuCreateOptThumbnail    = "thumbnail-url"
	roleSelectMenuCreateOptReqThumbnail = false
	roleSelectMenuCreateOptColor        = "embed-color"
	roleSelectMenuCreateOptReqColor     = false
)
const (
	roleSelectMenuCreateOptIdxTitle = iota
	roleSelectMenuCreateOptIdxRoles
	roleSelectMenuCreateOptIdxDesc      // index only valid when registering the command, since this is an optional argument.
	roleSelectMenuCreateOptIdxURL       // index only valid when registering the command, since this is an optional argument.
	roleSelectMenuCreateOptIdxImage     // index only valid when registering the command, since this is an optional argument.
	roleSelectMenuCreateOptIdxThumbnail // index only valid when registering the command, since this is an optional argument.
	roleSelectMenuCreateOptIdxColor     // index only valid when registering the command, since this is an optional argument.
	roleSelectMenuSubCmdCreateOptCount  // This MUST be the last constant defined in this block
)

const (
	roleSelectMenuSubCmdUpdate = "update"

	roleSelectMenuUpdateOptAction          = "action"
	roleSelectMenuUpdateOptReqAction       = true
	roleSelectMenuUpdateOptActionChoiceAdd = "add-role"
	roleSelectMenuUpdateOptActionChoiceDel = "delete-role"
	roleSelectMenuUpdateOptMsgID           = "message-id"
	roleSelectMenuUpdateOptReqMsgID        = true
	roleSelectMenuUpdateOptRole            = "role"
	roleSelectMenuUpdateOptReqRole         = true
	roleSelectMenuUpdateOptCtx             = "role-context"
	roleSelectMenuUpdateOptReqCtx          = false
)
const (
	roleSelectMenuUpdateOptIdxAction = iota
	roleSelectMenuUpdateOptIdxMsgID
	roleSelectMenuUpdateOptIdxRole
	roleSelectMenuUpdateOptIdxCtx      // index only valid when registering the command, since this is an optional argument.
	roleSelectMenuSubCmdUpdateOptCount // This MUST be the last constant defined in this block
)

// This block should ONLY contain IDs for components that are guaranteed to be present
// at least once in a role select menu.
const (
	roleSelectMenuComponentIDPrefix    = "role-select-menu"
	roleSelectResetButtonID            = "role-select-reset"
	roleSelectMenuMsgMinComponentCount = iota // This MUST be the last constant defined in this block
)
const roleSelectResetButtonLabel = "Reset your role selection"

func roleSelectMenuSubCmdCreateOpts() []*discordgo.ApplicationCommandOption {
	opts := make([]*discordgo.ApplicationCommandOption, roleSelectMenuSubCmdCreateOptCount)
	for i := range opts {
		switch i {
		case roleSelectMenuCreateOptIdxTitle:
			opts[i] = &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        roleSelectMenuCreateOptTitle,
				Description: "Title describing this selection of roles",
				Required:    roleSelectMenuCreateOptReqTitle,
			}
		case roleSelectMenuCreateOptIdxRoles:
			opts[i] = &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        roleSelectMenuCreateOptRoles,
				Description: fmt.Sprintf("Roles (up to %d) and their context. Ex: @SomeRole context 😺 @NextRole next role context", maxDiscordSelectMenuOpts*maxRoleSelectMenus),
				Required:    roleSelectMenuCreateOptReqRoles,
			}
		case roleSelectMenuCreateOptIdxDesc:
			opts[i] = &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        roleSelectMenuCreateOptDesc,
				Description: "Description of this selection of roles",
				Required:    roleSelectMenuCreateOptReqDesc,
			}
		case roleSelectMenuCreateOptIdxURL:
			opts[i] = &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        roleSelectMenuCreateOptURL,
				Description: "URL associated with this selection of roles",
				Required:    roleSelectMenuCreateOptReqURL,
			}
		case roleSelectMenuCreateOptIdxImage:
			opts[i] = &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        roleSelectMenuCreateOptImage,
				Description: "Image URL for this selection of roles",
				Required:    roleSelectMenuCreateOptReqImage,
			}
		case roleSelectMenuCreateOptIdxThumbnail:
			opts[i] = &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        roleSelectMenuCreateOptThumbnail,
				Description: "Thumbnail URL for this selection of roles",
				Required:    roleSelectMenuCreateOptReqThumbnail,
			}
		case roleSelectMenuCreateOptIdxColor:
			opts[i] = &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        roleSelectMenuCreateOptColor,
				Description: "Color to use when creating the message embed",
				Required:    roleSelectMenuCreateOptReqColor,
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
			}
		default:
			log.Fatalf("Unknown index (%d). There is a bug. :(", i)
		}
	}
	return opts
}

func roleSelectMenuSubCmdUpdateOpts() []*discordgo.ApplicationCommandOption {
	opts := make([]*discordgo.ApplicationCommandOption, roleSelectMenuSubCmdUpdateOptCount)
	for i := range opts {
		switch i {
		case roleSelectMenuUpdateOptIdxAction:
			opts[i] = &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        roleSelectMenuUpdateOptAction,
				Description: "Are we adding or removing a role from the menu?",
				Required:    roleSelectMenuUpdateOptReqAction,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  roleSelectMenuUpdateOptActionChoiceAdd,
						Value: roleSelectMenuUpdateOptActionChoiceAdd,
					},
					{
						Name:  roleSelectMenuUpdateOptActionChoiceDel,
						Value: roleSelectMenuUpdateOptActionChoiceDel,
					},
				},
			}
		case roleSelectMenuUpdateOptIdxMsgID:
			opts[i] = &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        roleSelectMenuUpdateOptMsgID,
				Description: "ID of the message to update (Enable dev mode -> Right Click Msg -> Copy ID)",
				Required:    roleSelectMenuUpdateOptReqMsgID,
			}
		case roleSelectMenuUpdateOptIdxRole:
			opts[i] = &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionRole,
				Name:        roleSelectMenuUpdateOptRole,
				Description: "Role to add or remove",
				Required:    roleSelectMenuUpdateOptReqRole,
			}
		case roleSelectMenuUpdateOptIdxCtx:
			opts[i] = &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        roleSelectMenuUpdateOptCtx,
				Description: "Context and emojis for this role. This option is ignored if removing the role.",
				Required:    roleSelectMenuUpdateOptReqCtx,
			}
		default:
			log.Fatalf("Unknown index (%d). There is a bug. :(", i)
		}
	}
	return opts
}

func roleSelectCmdOpts() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        roleSelectMenuSubCmdCreate,
			Description: "Create a new role selection menu",
			Options:     roleSelectMenuSubCmdCreateOpts(),
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        roleSelectMenuSubCmdUpdate,
			Description: "Edit an existing role selection menu",
			Options:     roleSelectMenuSubCmdUpdateOpts(),
		},
	}
}

var roleSelectMenuIDRegex = func() *regexp.Regexp { return nil }

func init() {
	r := regexp.MustCompile(fmt.Sprintf(`%s\d*`, roleSelectMenuComponentIDPrefix))
	if r == nil {
		log.Fatal("failed to compile roleSelectMenuComponentIDPrefix regex")
	}
	roleSelectMenuIDRegex = func() *regexp.Regexp { return r }
}

func strMatchesRoleSelectMenuID(str string) bool {
	return roleSelectMenuIDRegex().MatchString(str)
}

func handleRoleSelectMenuCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Options[0].Name {
	case roleSelectMenuSubCmdCreate:
		handleRoleSelectMenuCreate(s, i)
	case roleSelectMenuSubCmdUpdate:
		handleRoleSelectMenuUpdate(s, i)
	default:
		err := fmt.Errorf(`unknown command: "%s"`, i.ApplicationCommandData().Options[0].Name)
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
	}
}

// Given a discordgo Session and guildID, returns a map of roleID's to Roles for that guild
func guildRoleMap(s *discordgo.Session, guildID string) (map[string]*discordgo.Role, error) {
	if s == nil {
		return nil, fmt.Errorf("nil session provided")
	}
	if guildID == "" {
		return nil, fmt.Errorf("no guildID provided")
	}

	roles, err := s.GuildRoles(guildID)
	if err != nil {
		return nil, err
	}

	roleMap := make(map[string]*discordgo.Role, len(roles))
	for _, role := range roles {
		roleMap[role.ID] = role
	}
	return roleMap, nil
}

var roleRegex = func() *regexp.Regexp { return nil }

func init() {
	regStr := `<@&\d+>`
	r := regexp.MustCompile(regStr)
	if r == nil {
		log.Fatal("Failed to compile roleRegex: ", regStr)
	}
	roleRegex = func() *regexp.Regexp { return r }
}

const roleMentionDelimiter = "<@&"

func parseRoles(rolesOpt string) ([]string, error) {
	if !strings.HasPrefix(rolesOpt, roleMentionDelimiter) {
		return nil, fmt.Errorf("the first argument to `%s` must be a role mention", roleSelectMenuCreateOptRoles)
	}

	rolesAndCtx := strings.Split(rolesOpt, roleMentionDelimiter)[1:]
	if len(rolesAndCtx) > maxDiscordSelectMenuOpts*maxRoleSelectMenus {
		return nil, fmt.Errorf("you may only specify up to %d roles", maxDiscordSelectMenuOpts*maxRoleSelectMenus)
	}

	if len(rolesAndCtx)%maxDiscordSelectMenuOpts == 1 {
		return nil, fmt.Errorf("discord requires at least two options per select menu, so you must specify a number of roles, `x`, where `x modulo %d` does not equal `1`", maxDiscordActionRows)
	}

	if len(rolesAndCtx) == 0 {
		return nil, fmt.Errorf("you must specify at least one role")
	}

	// remove any duplicate roles
	roleSet := map[string]bool{}
	for _, r := range rolesAndCtx {
		roleSet[r] = true // this value is just to ensure the role key exists in the set
	}

	filteredRolesAndCtx := make([]string, len(roleSet))
	idx := 0
	for roleAndCtx := range roleSet {
		filteredRolesAndCtx[idx] = roleAndCtx
		idx++
	}

	return filteredRolesAndCtx, nil
}

func buildRoleSelectMenus(s *discordgo.Session, metadata interactionMetaData, rolesAndCtx []string) ([]discordgo.SelectMenu, error) {
	if s == nil {
		return nil, fmt.Errorf("nil session (%v)", s)
	}

	sMenus := make([]discordgo.SelectMenu, int(math.Ceil(float64(len(rolesAndCtx))/float64(maxDiscordSelectMenuOpts))))
	for i := range sMenus {
		sMenus[i] = discordgo.SelectMenu{
			CustomID:    roleSelectMenuComponentIDPrefix + fmt.Sprint(i),
			Placeholder: "Select any roles you would like to be added to. 🎭",
		}
	}

	roleMap, err := guildRoleMap(s, metadata.GuildID)
	if err != nil {
		return nil, err
	}

	for idx := range rolesAndCtx {
		rolesAndCtx[idx] = roleMentionDelimiter + rolesAndCtx[idx]
		roleID := strings.Replace(strings.Replace(roleRegex().FindString(rolesAndCtx[idx]), roleMentionDelimiter, "", 1), ">", "", 1)
		role, ok := roleMap[roleID]
		if !ok {
			return nil, fmt.Errorf("no role found for ID: %s", roleID)
		}

		emojiComp, scrubbedRoleAndCtx, err := detectAndScrubDiscordEmojis(rolesAndCtx[idx])
		if err != nil {
			return nil, err
		}

		smIdx := idx / 25
		sMenus[smIdx].Options = append(sMenus[smIdx].Options, discordgo.SelectMenuOption{
			Label: role.Name,
			Value: role.ID,
			// Description is whatever is left over after scrubbing Discord emojis and role mentions
			Description: roleRegex().ReplaceAllString(scrubbedRoleAndCtx, ""),
			Emoji:       emojiComp,
		})
	}

	return sMenus, nil
}

// Scans a string for discord emojis and returns the first one it encounters.
// Also scrubs the provided string of any discord emojis and returns the scrubbed string.
func detectAndScrubDiscordEmojis(str string) (discordgo.ComponentEmoji, string, error) {
	emojiStr := discordgo.EmojiRegex.FindString(str)
	if emojiStr == "" {
		// No discord emojis, so check for unicode emojis.
		uEmojis := gomoji.FindAll(str)
		if len(uEmojis) == 0 {
			return discordgo.ComponentEmoji{}, str, nil
		}
		return discordgo.ComponentEmoji{Name: uEmojis[0].Character}, str, nil
	}

	// See https://discord.com/developers/docs/reference#message-formatting-formats
	const emojiDelim = ":"
	emojiToks := strings.Split(emojiStr, emojiDelim)
	if len(emojiToks) != 3 {
		return discordgo.ComponentEmoji{}, str, fmt.Errorf("unexpected number of emoji tokens when delimiting \"%s\" on \"%s\"", emojiStr, emojiDelim)
	}

	emoji := discordgo.ComponentEmoji{
		Name:     emojiToks[1],
		ID:       strings.TrimSuffix(emojiToks[2], ">"),
		Animated: strings.HasPrefix(emojiToks[0], "<a"),
	}

	return emoji, discordgo.EmojiRegex.ReplaceAllString(str, ""), nil
}

func buildRoleSelectMenuEmbed(opts []*discordgo.ApplicationCommandInteractionDataOption) *dg_helpers.Embed {
	embed := dg_helpers.NewEmbed()
	if opts == nil {
		log.Error("nil opts")
		return embed
	}

	for _, opt := range opts {
		switch opt.Name {
		case roleSelectMenuCreateOptTitle:
			embed.SetTitle(opt.StringValue())
		case roleSelectMenuCreateOptDesc:
			embed.SetDescription(opt.StringValue())
		case roleSelectMenuCreateOptImage:
			embed.SetImage(opt.StringValue())
		case roleSelectMenuCreateOptURL:
			embed.SetURL(opt.StringValue())
		case roleSelectMenuCreateOptThumbnail:
			embed.SetThumbnail(opt.StringValue())
		case roleSelectMenuCreateOptColor:
			embed.SetColor(int(opt.IntValue()))
		}
	}
	embed.SetFooter(fmt.Sprintf(`Warning! Selecting the "%s" button will remove you from any roles present in this menu.`, roleSelectResetButtonLabel))

	return embed
}

func validateRoleSelectEmbedURLs(e *dg_helpers.Embed) error {
	if e.URL != "" && !isReachableURL(e.URL) {
		return fmt.Errorf("unreachable URL provided: %s", e.URL)
	}
	if e.Thumbnail != nil && e.Thumbnail.URL != "" && !isReachableURL(e.Thumbnail.URL) {
		return fmt.Errorf("unreachable thumbnail URL provided: %s", e.Thumbnail.URL)
	}
	return nil
}

func handleRoleSelectMenuUpdate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if s == nil || i == nil {
		log.Errorf("nil Session pointer (%v) and/or InteractionCreate pointer (%v)", s, i)
		return
	}
	wg := bot().updateLastActive()
	defer wg.Wait()

	if isAdmin, err := isInteractionIssuerAdmin(i); err != nil {
		interactionRespondEphemeralError(s, i, true, err)
		log.Error(err)
		return
	} else if !isAdmin {
		interactionRespondEphemeralError(s, i, false, fmt.Errorf("you must run this command from a server you administer"))
		return
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: InteractionResponseFlagEphemeral,
		},
	})
	if err != nil {
		interactionRespondEphemeralError(s, i, true, err)
		log.Error(err)
		return
	}

	metadata, err := getInteractionMetaData(i)
	if err != nil {
		interactionFollowUpEphemeralError(s, i, true, err)
		log.Error(err)
		return
	}

	optData := i.ApplicationCommandData().Options[0].Options
	msgToEditID := optData[roleSelectMenuUpdateOptIdxMsgID].StringValue()
	msg, err := s.ChannelMessage(metadata.ChannelID, msgToEditID)
	if err != nil || msg == nil {
		log.Warn(err)
		interactionFollowUpEphemeralError(s, i, false, fmt.Errorf("could not find any message in this channel with ID: `%s`", msgToEditID))
		return
	}

	err = isMessageARoleSelectMenu(s, msg)
	if err != nil {
		log.Warn(err)
		interactionFollowUpEphemeralError(s, i, false, fmt.Errorf("provided message ID does not appear to contain a role select menu:\n\t%v", err))
		return
	}

	switch choice := optData[roleSelectMenuUpdateOptIdxAction].StringValue(); choice {
	case roleSelectMenuUpdateOptActionChoiceAdd:
		handleRoleSelectMenuUpdateAdd(s, i, *msg)
	case roleSelectMenuUpdateOptActionChoiceDel:
		handleRoleSelectMenuUpdateDel(s, i, msg)
	default:
		err := fmt.Errorf("unknown update choice: %s", choice)
		interactionFollowUpEphemeralError(s, i, true, err)
		log.Error(err)
		return
	}
}

func handleRoleSelectMenuUpdateAdd(s *discordgo.Session, i *discordgo.InteractionCreate, msgToEdit discordgo.Message) {
	optData := i.ApplicationCommandData().Options[0].Options

	metadata, err := getInteractionMetaData(i)
	if err != nil {
		interactionFollowUpEphemeralError(s, i, true, err)
		log.Error(err)
		return
	}

	var roleToAdd *discordgo.Role
	var roleContext string = ""
	for _, opt := range optData {
		switch opt.Name {
		case roleSelectMenuUpdateOptRole:
			roleToAdd = opt.RoleValue(s, metadata.GuildID)
		case roleSelectMenuUpdateOptCtx:
			roleContext = opt.StringValue()
		}
	}

	if roleToAdd == nil || roleToAdd.Name == "" {
		err = fmt.Errorf("could not retreive role name from user args")
		interactionFollowUpEphemeralError(s, i, true, err)
		log.Error(err)
		return
	}

	msgEdit := &discordgo.MessageEdit{
		Embeds: msgToEdit.Embeds,
		AllowedMentions: &discordgo.MessageAllowedMentions{
			Parse: []discordgo.AllowedMentionType{
				discordgo.AllowedMentionTypeEveryone,
				discordgo.AllowedMentionTypeRoles,
				discordgo.AllowedMentionTypeUsers,
			},
		},
		ID:      msgToEdit.ID,
		Channel: metadata.ChannelID,
	}
	if msgToEdit.Content != "" {
		msgEdit.Content = &msgToEdit.Content
	}

	content := ""
	if ok, err := isRoleInSelectMenuMsg(roleToAdd.ID, s, &msgToEdit); err != nil {
		interactionFollowUpEphemeralError(s, i, true, err)
		log.Error(err)
		return
	} else if ok {
		msgEdit.Components, err = updateExistingRoleSelectMenuOption(s, roleToAdd, roleContext, msgToEdit)
		if err != nil {
			interactionFollowUpEphemeralError(s, i, true, err)
			log.Error(err)
			return
		}
		content = fmt.Sprintf("Updated existing menu option for %s", roleToAdd.Mention())
	} else {
		var newOptAdded bool
		msgEdit.Components, newOptAdded, err = addRoleSelectMenuOption(s, roleToAdd, roleContext, msgToEdit)
		if err != nil {
			interactionFollowUpEphemeralError(s, i, true, err)
			log.Error(err)
			return
		}
		if !newOptAdded {
			_, err = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
				Content: fmt.Sprintf(
					"Congratulations, you've bumped into a Discord API limitation! 🎉 Either you are at the max number of menus (%d), or your existing menus are full and we can't add a new one because discord requires at least %d options per menu. Sorry for the inconvenience! 😔",
					maxDiscordActionRows,
					minDiscordSelectMenuOpts,
				),
			})
			if err != nil {
				interactionFollowUpEphemeralError(s, i, true, err)
				log.Error(err)
			}
			return
		}
		content = fmt.Sprintf("Added %s option to the menu", roleToAdd.Mention())
	}

	_, err = s.ChannelMessageEditComplex(msgEdit)
	if err != nil {
		interactionFollowUpEphemeralError(s, i, true, err)
		log.Error(err)
		return
	}

	_, err = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
		Content: content,
		AllowedMentions: &discordgo.MessageAllowedMentions{
			Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeRoles},
		},
	})
	if err != nil {
		interactionFollowUpEphemeralError(s, i, true, err)
		log.Error(err)
	}
}

func handleRoleSelectMenuUpdateDel(s *discordgo.Session, i *discordgo.InteractionCreate, msgToEdit *discordgo.Message) {
	optData := i.ApplicationCommandData().Options[0].Options

	metadata, err := getInteractionMetaData(i)
	if err != nil {
		interactionFollowUpEphemeralError(s, i, true, err)
		log.Error(err)
		return
	}

	var roleToDel *discordgo.Role
	for _, opt := range optData {
		switch opt.Name {
		case roleSelectMenuUpdateOptRole:
			roleToDel = opt.RoleValue(s, metadata.GuildID)
		}
	}

	if roleToDel == nil {
		err = fmt.Errorf("could not retreive role name from user args")
		interactionFollowUpEphemeralError(s, i, true, err)
		log.Error(err)
		return
	}

	if ok, err := isRoleInSelectMenuMsg(roleToDel.ID, s, msgToEdit); err != nil {
		log.Warn(err)
		interactionFollowUpEphemeralError(s, i, false, fmt.Errorf("provided message ID does not appear to contain a role select menu:\n\t%v", err))
		return
	} else if !ok {
		_, err := s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
			Content: fmt.Sprintf("%s is not present in the menu, nothing to do.", roleToDel.Mention()),
		})
		if err != nil {
			interactionFollowUpEphemeralError(s, i, true, err)
			log.Error(err)
		}
		return
	}

	msgEdit := &discordgo.MessageEdit{
		Embeds: msgToEdit.Embeds,
		AllowedMentions: &discordgo.MessageAllowedMentions{
			Parse: []discordgo.AllowedMentionType{
				discordgo.AllowedMentionTypeEveryone,
				discordgo.AllowedMentionTypeRoles,
				discordgo.AllowedMentionTypeUsers,
			},
		},
		ID:         msgToEdit.ID,
		Channel:    metadata.ChannelID,
		Components: msgToEdit.Components,
	}
	if msgToEdit.Content != "" {
		msgEdit.Content = &msgToEdit.Content
	}

	removed := false
	for _, ar := range msgEdit.Components {
		actionsRow, ok := ar.(*discordgo.ActionsRow)
		if !ok {
			err = fmt.Errorf("bad cast to actions row, this should never happen")
			log.Error(err)
			interactionFollowUpEphemeralError(s, i, true, err)
			return
		}
		for _, c := range actionsRow.Components {
			if c.Type() != discordgo.SelectMenuComponent {
				continue
			}
			selectMenu, ok := c.(*discordgo.SelectMenu)
			if !ok {
				err = fmt.Errorf("bad cast to select menu, this should never happen")
				log.Error(err)
				interactionFollowUpEphemeralError(s, i, true, err)
				return
			}
			for optIdx := range selectMenu.Options {
				if selectMenu.Options[optIdx].Value == roleToDel.ID {
					selectMenu.Options[optIdx] = selectMenu.Options[len(selectMenu.Options)-1]
					selectMenu.Options = selectMenu.Options[:len(selectMenu.Options)-1]
					removed = true
					break
				}
			}
			if removed {
				if len(selectMenu.Options) < minDiscordSelectMenuOpts {
					_, err := s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
						Content: fmt.Sprintf("Cannot remove %s. Discord requires at least %d options per select menu.", roleToDel.Mention(), minDiscordSelectMenuOpts),
					})
					if err != nil {
						interactionFollowUpEphemeralError(s, i, true, err)
						log.Error(err)
					}
					return
				}
				break
			}
		}
		if removed {
			break
		}
	}

	_, err = s.ChannelMessageEditComplex(msgEdit)
	if err != nil {
		interactionFollowUpEphemeralError(s, i, true, err)
		log.Error(err)
		return
	}

	_, err = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
		Content: fmt.Sprintf("%s was removed from the menu", roleToDel.Mention()),
	})
	if err != nil {
		interactionFollowUpEphemeralError(s, i, true, err)
		log.Error(err)
	}
}

func isRoleInSelectMenuMsg(roleID string, s *discordgo.Session, msgToEdit *discordgo.Message) (bool, error) {
	if err := isMessageARoleSelectMenu(s, msgToEdit); err != nil {
		return false, fmt.Errorf("message is not a role select menu, this should never happen. err: %v", err)
	}

	for _, ar := range msgToEdit.Components {
		actionsRow, ok := ar.(*discordgo.ActionsRow)
		if !ok {
			return false, fmt.Errorf("bad cast to actions row, this should never happen")
		}
		for _, c := range actionsRow.Components {
			if c.Type() != discordgo.SelectMenuComponent {
				continue
			}
			selectMenu, ok := c.(*discordgo.SelectMenu)
			if !ok {
				return false, fmt.Errorf("bad cast to SelectMenu, this should never happen")
			}
			if len(selectMenu.Options) == 0 {
				log.Warn("Empty select menu")
			}
			for _, choice := range selectMenu.Options {
				if choice.Value == roleID {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

// Returns the updated message components, a bool indicating whether or not there was room for the new option, and an error indicating if something went wrong.
func addRoleSelectMenuOption(s *discordgo.Session, roleToAdd *discordgo.Role, roleCtx string, msgToEdit discordgo.Message) ([]discordgo.MessageComponent, bool, error) {
	if err := isMessageARoleSelectMenu(s, &msgToEdit); err != nil {
		return msgToEdit.Components, false, fmt.Errorf("message is not a role select menu, this should never happen. err: %v", err)
	}

	for _, ar := range msgToEdit.Components {
		actionsRow, ok := ar.(*discordgo.ActionsRow)
		if !ok {
			return msgToEdit.Components, false, fmt.Errorf("bad cast to actions row, this should never happen")
		}
		for _, c := range actionsRow.Components {
			if c.Type() != discordgo.SelectMenuComponent {
				continue
			}
			selectMenu, ok := c.(*discordgo.SelectMenu)
			if !ok {
				return msgToEdit.Components, false, fmt.Errorf("bad cast to SelectMenu, this should never happen")
			}
			if len(selectMenu.Options) != maxDiscordSelectMenuOpts {
				emoji, sanitizedCtx, err := detectAndScrubDiscordEmojis(roleCtx)
				if err != nil {
					return msgToEdit.Components, false, err
				}
				selectMenu.Options = append(selectMenu.Options, discordgo.SelectMenuOption{
					Label: roleToAdd.Name,
					Value: roleToAdd.ID,
					// Description is whatever is left over after scrubbing Discord emojis and role mentions
					Description: roleRegex().ReplaceAllString(sanitizedCtx, ""),
					Emoji:       emoji,
				})
				return msgToEdit.Components, true, nil
			}
		}
	}

	return msgToEdit.Components, false, nil
}

// Returns the updated message components and an error indicating success or failure
func updateExistingRoleSelectMenuOption(s *discordgo.Session, roleToAdd *discordgo.Role, roleCtx string, msgToEdit discordgo.Message) ([]discordgo.MessageComponent, error) {
	if err := isMessageARoleSelectMenu(s, &msgToEdit); err != nil {
		return msgToEdit.Components, fmt.Errorf("message is not a role select menu, this should never happen. err: %v", err)
	}

	for _, ar := range msgToEdit.Components {
		actionsRow, ok := ar.(*discordgo.ActionsRow)
		if !ok {
			return msgToEdit.Components, fmt.Errorf("bad cast to actions row, this should never happen")
		}
		for _, c := range actionsRow.Components {
			if c.Type() != discordgo.SelectMenuComponent {
				continue
			}
			selectMenu, ok := c.(*discordgo.SelectMenu)
			if !ok {
				return msgToEdit.Components, fmt.Errorf("bad cast to SelectMenu, this should never happen")
			}
			for idx := range selectMenu.Options {
				choice := &selectMenu.Options[idx]
				if choice.Value == roleToAdd.ID {
					emoji, sanitizedCtx, err := detectAndScrubDiscordEmojis(roleCtx)
					if err != nil {
						return nil, err
					}
					choice.Description = roleRegex().ReplaceAllString(sanitizedCtx, "")
					choice.Emoji = emoji
					return msgToEdit.Components, nil
				}
			}
		}
	}
	return msgToEdit.Components, fmt.Errorf("did not find existing role to update")
}

func isMessageARoleSelectMenu(s *discordgo.Session, m *discordgo.Message) error {
	if m == nil {
		return fmt.Errorf("message is nil")
	}

	if (m.Author == nil && m.Member == nil) || (m.Author == nil && m.Member.User == nil) {
		return fmt.Errorf("cannot verify message author is %s", s.State.User.Mention())
	}
	if m.Author != nil && m.Author.ID != s.State.User.ID {
		return fmt.Errorf("message not authored by %s", s.State.User.Mention())
	}
	if m.Member != nil && m.Member.User != nil && m.Member.User.ID != s.State.User.ID {
		return fmt.Errorf("message not authored by %s", s.State.User.Mention())
	}

	if len(m.Components) < roleSelectMenuMsgMinComponentCount {
		return fmt.Errorf("message contains unexpected number of components")
	}

	for arIdx, ar := range m.Components {
		if ar == nil {
			return fmt.Errorf("encountered a nil component")
		}

		actionsRow, ok := ar.(*discordgo.ActionsRow)
		if !ok {
			return fmt.Errorf("encountered unexpected component type, expected an ActionsRow")
		}

		if len(actionsRow.Components) != 1 {
			// all role SelectMenu message ActionsRows have exactly one component
			return fmt.Errorf("encountered unexpectedly empty ActionsRow")
		}

		c := actionsRow.Components[0]
		if c == nil {
			return fmt.Errorf("encountered a nil component")
		}
		if arIdx == len(m.Components)-1 {
			// last component is always the reset button
			button, ok := c.(*discordgo.Button)
			if !ok {
				return fmt.Errorf("encountered unexpected component type, expected a Button")
			}
			if button.CustomID != roleSelectResetButtonID {
				return fmt.Errorf("button ID does not belong to a role select menu")
			}
		} else {
			selectMenu, ok := c.(*discordgo.SelectMenu)
			if !ok {
				return fmt.Errorf("encountered unexpected component type, expected a SelectMenu")
			}
			if !strMatchesRoleSelectMenuID(selectMenu.CustomID) {
				return fmt.Errorf("encountered select menu ID that does not belong to a role select menu")
			}
		}
	}
	return nil
}

func handleRoleSelectMenuCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if s == nil || i == nil {
		log.Errorf("nil Session pointer (%v) and/or InteractionCreate pointer (%v)", s, i)
		return
	}
	wg := bot().updateLastActive()
	defer wg.Wait()

	if isAdmin, err := isInteractionIssuerAdmin(i); err != nil {
		interactionRespondEphemeralError(s, i, true, err)
		log.Error(err)
		return
	} else if !isAdmin {
		interactionRespondEphemeralError(s, i, false, fmt.Errorf("you must be a server administrator to run this command"))
		return
	}

	rolesAndCtx, err := parseRoles(i.ApplicationCommandData().Options[0].Options[roleSelectMenuCreateOptIdxRoles].StringValue())
	if err != nil {
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprint(err),
				Flags:   InteractionResponseFlagEphemeral,
			},
		})
		if err != nil {
			interactionFollowUpEphemeralError(s, i, false, err)
			log.Error(err)
		}
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		interactionRespondEphemeralError(s, i, true, err)
		log.Error(err)
		return
	}

	metadata, err := getInteractionMetaData(i)
	if err != nil {
		interactionFollowUpEphemeralError(s, i, true, err)
		log.Error(err)
		return
	}
	sMenus, err := buildRoleSelectMenus(s, *metadata, rolesAndCtx)
	if err != nil {
		interactionFollowUpEphemeralError(s, i, true, err)
		return
	}

	e := buildRoleSelectMenuEmbed(i.ApplicationCommandData().Options[0].Options)
	err = validateRoleSelectEmbedURLs(e)
	if err != nil {
		interactionFollowUpEphemeralError(s, i, false, err)
		return
	}

	iEdit := &discordgo.WebhookEdit{
		Embeds: []*discordgo.MessageEmbed{e.Truncate().MessageEmbed},
		AllowedMentions: &discordgo.MessageAllowedMentions{
			Parse: []discordgo.AllowedMentionType{
				discordgo.AllowedMentionTypeEveryone,
				discordgo.AllowedMentionTypeRoles,
				discordgo.AllowedMentionTypeUsers,
			},
		},
	}
	for _, m := range sMenus {
		m.MaxValues = len(m.Options)
		m.MinValues = 0
		iEdit.Components = append(iEdit.Components, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{m},
		})
	}
	iEdit.Components = append(iEdit.Components, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label: roleSelectResetButtonLabel,
				Style: discordgo.DangerButton,
				Emoji: discordgo.ComponentEmoji{
					Name: "💣",
				},
				CustomID: roleSelectResetButtonID,
			},
		},
	})

	_, err = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, iEdit)
	if err != nil {
		interactionFollowUpEphemeralError(s, i, true, err)
		log.Error(err)
		return
	}
}

// Retrieves the selection options from a roleSelectMenu. If a buttonID is provided,
// only options from that SelectMenu will be included in the returned result.
func getPossibleRoleIDs(i *discordgo.InteractionCreate, buttonID string) (map[string]bool, error) {
	if i == nil {
		return nil, fmt.Errorf("nil interaction")
	}

	var possibleRoleIDs map[string]bool = nil
	for _, ar := range i.Message.Components {
		if ar == nil || ar.Type() != discordgo.ActionsRowComponent {
			continue
		}
		actionRow, ok := ar.(*discordgo.ActionsRow)
		if !ok {
			return nil, fmt.Errorf("bad cast to *discordgo.ActionsRow, type is %T", ar)
		}
		for _, sm := range actionRow.Components {
			if sm == nil || sm.Type() != discordgo.SelectMenuComponent {
				continue
			}

			selectMenu, ok := sm.(*discordgo.SelectMenu)
			if !ok {
				return nil, fmt.Errorf("bad cast to *discordgo.SelectMenu, type is %T", sm)
			}

			if selectMenu.CustomID == buttonID || buttonID == "" {
				if possibleRoleIDs == nil {
					possibleRoleIDs = make(map[string]bool, len(selectMenu.Options))
				}
				for _, opt := range selectMenu.Options {
					possibleRoleIDs[opt.Value] = false
				}
				break
			}
		}
		if len(possibleRoleIDs) > 0 && buttonID != "" {
			break
		}
	}
	if len(possibleRoleIDs) == 0 {
		return nil, fmt.Errorf("no possible role IDs found, is the select menu empty?")
	}
	return possibleRoleIDs, nil
}

func handleRoleSelection(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if s == nil || i == nil {
		log.Errorf("nil Session pointer (%v) and/or InteractionCreate pointer (%v)", s, i)
		return
	}
	wg := bot().updateLastActive()
	defer wg.Wait()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: InteractionResponseFlagEphemeral,
		},
	})

	selectedRoleIDMap, err := getPossibleRoleIDs(i, i.MessageComponentData().CustomID)
	if err != nil {
		time.Sleep(time.Millisecond * 200) // wait a bit for the deferred response to be received
		interactionFollowUpEphemeralError(s, i, true, err)
		log.Error(err)
		return
	}

	for _, roleID := range i.MessageComponentData().Values {
		if _, ok := selectedRoleIDMap[roleID]; !ok {
			time.Sleep(time.Millisecond * 200) // wait a bit for the deferred response to be received
			err = fmt.Errorf("%s is not a valid role ID, there is a bug. :(", roleID)
			interactionFollowUpEphemeralError(s, i, true, err)
			log.Error(err)
			return
		}
		selectedRoleIDMap[roleID] = true
	}

	metadata, err := getInteractionMetaData(i)
	if err != nil {
		interactionFollowUpEphemeralError(s, i, true, err)
		log.Error(err)
		return
	}

	addedRoles := ""
	removedRoles := ""
	for roleID, isSelected := range selectedRoleIDMap {
		if isSelected {
			err = s.GuildMemberRoleAdd(metadata.GuildID, metadata.AuthorID, roleID)
			if err == nil {
				userHadRole := false
				for _, r := range metadata.AuthorGuildRoles {
					if r == roleID {
						userHadRole = true
						break
					}
				}
				if !userHadRole {
					addedRoles += fmt.Sprintf("<@&%s>\n", roleID)
				}
			}
		} else {
			err = s.GuildMemberRoleRemove(metadata.GuildID, metadata.AuthorID, roleID)
			if err == nil {
				for _, r := range metadata.AuthorGuildRoles {
					if r == roleID {
						removedRoles += fmt.Sprintf("<@&%s>\n", roleID)
						break
					}
				}
			}
		}
		if err != nil {
			interactionFollowUpEphemeralError(s, i, true, err)
			log.Error(err)
			return
		}
	}

	embedColor, err := fastHappyColorInt64()
	if err != nil {
		log.Warn(err)
		embedColor = 0
	}

	embed := dg_helpers.NewEmbed().SetTitle("Your Roles Have Been Updated 🎭").SetColor(int(embedColor))
	if addedRoles != "" {
		embed.AddField("Added Roles ✅", addedRoles)
	}
	if removedRoles != "" {
		embed.AddField("Removed Roles ❌", removedRoles)
	}
	_, err = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
		Embeds: []*discordgo.MessageEmbed{embed.Truncate().MessageEmbed},
		AllowedMentions: &discordgo.MessageAllowedMentions{
			Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeRoles},
		},
	})
	if err != nil {
		interactionFollowUpEphemeralError(s, i, true, err)
		log.Error(err)
		return
	}
}

func handleRoleSelectReset(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if s == nil || i == nil {
		log.Errorf("nil Session pointer (%v) and/or InteractionCreate pointer (%v)", s, i)
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: InteractionResponseFlagEphemeral,
		},
	})

	possibleRoleIDs, err := getPossibleRoleIDs(i, "")
	if err != nil {
		time.Sleep(time.Millisecond * 200)
		interactionFollowUpEphemeralError(s, i, true, err)
		log.Error(err)
		return
	}

	metadata, err := getInteractionMetaData(i)
	if err != nil {
		interactionFollowUpEphemeralError(s, i, true, err)
		log.Error(err)
		return
	}

	removedRoles := ""
	for roleID := range possibleRoleIDs {
		err = s.GuildMemberRoleRemove(metadata.GuildID, metadata.AuthorID, roleID)
		if err != nil {
			interactionFollowUpEphemeralError(s, i, true, err)
			log.Error(err)
			return
		}
		for _, r := range metadata.AuthorGuildRoles {
			if r == roleID {
				removedRoles += fmt.Sprintf("<@&%s>\n", roleID)
				break
			}
		}
	}

	embedColor, err := fastHappyColorInt64()
	if err != nil {
		log.Warn(err)
		embedColor = 0
	}

	embed := dg_helpers.NewEmbed().SetTitle("Your Roles Have Been Updated 🎭").SetColor(int(embedColor))
	if removedRoles != "" {
		embed.AddField("Removed Roles ❌", removedRoles)
	} else {
		embed.AddField("Removed Roles ❌", "No roles to remove")
	}

	s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
		Embeds: []*discordgo.MessageEmbed{embed.Truncate().MessageEmbed},
		AllowedMentions: &discordgo.MessageAllowedMentions{
			Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeRoles},
		},
	})
}
