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
	createRoleSelectCommand = "create-role-menu"

	roleSelectMenuTitleOpt    = "title"
	roleSelectMenuTitleOptReq = true

	roleSelectRolesOpt    = "roles"
	roleSelectRolesOptReq = true
	maxRoleSelectMenus    = 4

	roleSelectMenuDescOpt    = "description"
	roleSelectMenuDescOptReq = false

	roleSelectMenuURLOpt    = "url"
	roleSelectMenuURLOptReq = false

	roleSelectMenuImageOpt    = "image-url"
	roleSelectMenuImageOptReq = false

	roleSelectMenuThumbnailOpt    = "thumbnail-url"
	roleSelectMenuThumbnailOptReq = false

	roleSelectMenuColorOpt    = "embed-color"
	roleSelectMenuColorOptReq = false

	roleSelectMenuComponentIDPrefix = "role-select-menu"
	roleSelectResetButtonID         = "role-select-reset"
	roleSelectResetButtonLabel      = "Reset your role selection"
)

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

const (
	roleSelectMenuTitleOptIdx = iota
	roleSelectRolesOptIdx
	roleSelectMenuDescOptIdx      // index only valid when registering the command, since this is an optional argument.
	roleSelectMenuURLOptIdx       // index only valid when registering the command, since this is an optional argument.
	roleSelectMenuFooterOptIdx    // index only valid when registering the command, since this is an optional argument.
	roleSelectMenuThumbnailOptIdx // index only valid when registering the command, since this is an optional argument.
	roleSelectMenuColorOptIdx     // index only valid when registering the command, since this is an optional argument.

	// This MUST be the last constant defined in this block
	roleSelectOptCount
)

func roleSelectCmdOpts() []*discordgo.ApplicationCommandOption {
	opts := make([]*discordgo.ApplicationCommandOption, roleSelectOptCount)
	for i := range opts {
		switch i {
		// TODO: Add a help option
		case roleSelectMenuTitleOptIdx:
			opts[i] = &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        roleSelectMenuTitleOpt,
				Description: "Title describing this selection of roles",
				Required:    roleSelectMenuTitleOptReq,
			}
		case roleSelectRolesOptIdx:
			opts[i] = &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        roleSelectRolesOpt,
				Description: fmt.Sprintf("Roles (up to %d) and their context. Ex: @SomeRole context 😺 @NextRole next role context", maxDiscordSelectMenuOpts*maxRoleSelectMenus),
				Required:    roleSelectRolesOptReq,
			}
		case roleSelectMenuDescOptIdx:
			opts[i] = &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        roleSelectMenuDescOpt,
				Description: "Description of this selection of roles",
				Required:    roleSelectMenuDescOptReq,
			}
		case roleSelectMenuURLOptIdx:
			opts[i] = &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        roleSelectMenuURLOpt,
				Description: "URL associated with this selection of roles",
				Required:    roleSelectMenuURLOptReq,
			}
		case roleSelectMenuFooterOptIdx:
			opts[i] = &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        roleSelectMenuImageOpt,
				Description: "Image URL for this selection of roles",
				Required:    roleSelectMenuImageOptReq,
			}
		case roleSelectMenuThumbnailOptIdx:
			opts[i] = &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        roleSelectMenuThumbnailOpt,
				Description: "Thumbnail URL for this selection of roles",
				Required:    roleSelectMenuThumbnailOptReq,
			}
		case roleSelectMenuColorOptIdx:
			opts[i] = &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        roleSelectMenuColorOpt,
				Description: "Color to use when creating the message embed",
				Required:    roleSelectMenuColorOptReq,
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
		}
	}

	return opts
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

func parseRoles(i *discordgo.InteractionCreate) ([]string, error) {
	if !strings.HasPrefix(i.ApplicationCommandData().Options[roleSelectRolesOptIdx].StringValue(), roleMentionDelimiter) {
		return nil, fmt.Errorf("the first argument to `%s` must be a role mention", roleSelectRolesOpt)
	}

	rolesAndCtx := strings.Split(i.ApplicationCommandData().Options[roleSelectRolesOptIdx].StringValue(), roleMentionDelimiter)[1:]
	if len(rolesAndCtx) > maxDiscordSelectMenuOpts*maxRoleSelectMenus {
		return nil, fmt.Errorf("you may only specify up to %d roles", maxDiscordSelectMenuOpts*maxRoleSelectMenus)
	}

	if len(rolesAndCtx) == 0 {
		return nil, fmt.Errorf("you must specify at least one role")
	}

	return rolesAndCtx, nil
}

func buildRoleSelectMenus(s *discordgo.Session, i *discordgo.InteractionCreate, rolesAndCtx []string) ([]discordgo.SelectMenu, error) {
	if s == nil || i == nil {
		return nil, fmt.Errorf("nil session (%v) or interaction (%v)", s, i)
	}

	sMenus := make([]discordgo.SelectMenu, int(math.Ceil(float64(len(rolesAndCtx))/float64(maxDiscordSelectMenuOpts))))
	for i := range sMenus {
		sMenus[i] = discordgo.SelectMenu{
			CustomID:    roleSelectMenuComponentIDPrefix + fmt.Sprint(i),
			Placeholder: "Select any roles you would like to be added to. 🎭",
		}
	}

	metadata, err := getInteractionMetaData(i)
	if err != nil {
		return nil, err
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
		// TODO: Add a "None of these" option to deselect all roles
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

func buildRoleSelectMenuEmbed(i *discordgo.InteractionCreate) *dg_helpers.Embed {
	embed := dg_helpers.NewEmbed()
	if i == nil {
		log.Error("nil interaction")
		return embed
	}

	for _, opt := range i.ApplicationCommandData().Options {
		switch opt.Name {
		case roleSelectMenuTitleOpt:
			embed.SetTitle(opt.StringValue())
		case roleSelectMenuDescOpt:
			embed.SetDescription(opt.StringValue())
		case roleSelectMenuImageOpt:
			embed.SetImage(opt.StringValue())
		case roleSelectMenuURLOpt:
			embed.SetURL(opt.StringValue())
		case roleSelectMenuThumbnailOpt:
			embed.SetThumbnail(opt.StringValue())
		case roleSelectMenuColorOpt:
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

func createRoleSelect(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		interactionRespondEphemeralError(s, i, true, err)
		log.Error(err)
		return
	}

	rolesAndCtx, err := parseRoles(i)
	if err != nil {
		interactionFollowUpEphemeralError(s, i, false, err)
		return
	}

	sMenus, err := buildRoleSelectMenus(s, i, rolesAndCtx)
	if err != nil {
		interactionFollowUpEphemeralError(s, i, true, err)
		return
	}

	e := buildRoleSelectMenuEmbed(i)
	err = validateRoleSelectEmbedURLs(e)
	if err != nil {
		interactionFollowUpEphemeralError(s, i, false, err)
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

	var possibleRoleIDs map[string]bool
	for _, ar := range i.Message.Components {
		if ar == nil {
			continue
		}
		actionRow, ok := ar.(*discordgo.ActionsRow)
		if !ok {
			return nil, fmt.Errorf("bad cast to *discordgo.ActionsRow, type is %T", ar)
		}
		for _, sm := range actionRow.Components {
			if sm == nil {
				continue
			}
			if sm.Type() != discordgo.SelectMenuComponent {
				continue
			}

			selectMenu, ok := sm.(*discordgo.SelectMenu)
			if !ok {
				return nil, fmt.Errorf("bad cast to *discordgo.SelectMenu, type is %T", sm)
			}

			if selectMenu.CustomID == buttonID || buttonID == "" {
				possibleRoleIDs = make(map[string]bool, len(selectMenu.Options))
				for _, opt := range selectMenu.Options {
					possibleRoleIDs[opt.Value] = false
				}
				break
			}
		}
		if len(possibleRoleIDs) > 0 {
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
