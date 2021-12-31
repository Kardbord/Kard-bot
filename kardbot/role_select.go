package kardbot

import (
	"fmt"
	"math"
	"regexp"
	"strings"

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
)

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
				Description: fmt.Sprintf("Roles (up to %d) and their context. Ex: @SomeRole context 😺 @NextRole next role context", maxDiscordSelectMenuOpts*maxDiscordActionRowSize),
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
	if len(rolesAndCtx) > maxDiscordSelectMenuOpts*maxDiscordActionRowSize {
		return nil, fmt.Errorf("you may only specify up to %d roles", maxDiscordSelectMenuOpts*maxDiscordActionRowSize)
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
	embedArgs := dg_helpers.NewEmbed()
	if i == nil {
		log.Error("nil interaction")
		return embedArgs
	}

	for _, opt := range i.ApplicationCommandData().Options {
		switch opt.Name {
		case roleSelectMenuTitleOpt:
			embedArgs.SetTitle(opt.StringValue())
		case roleSelectMenuDescOpt:
			embedArgs.SetDescription(opt.StringValue())
		case roleSelectMenuImageOpt:
			embedArgs.SetImage(opt.StringValue())
		case roleSelectMenuURLOpt:
			embedArgs.SetURL(opt.StringValue())
		case roleSelectMenuThumbnailOpt:
			embedArgs.SetThumbnail(opt.StringValue())
		case roleSelectMenuColorOpt:
			embedArgs.SetColor(int(opt.IntValue()))
		}
	}

	return embedArgs
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
		iEdit.Components = append(iEdit.Components, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{m},
		})
	}

	_, err = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, iEdit)
	if err != nil {
		interactionFollowUpEphemeralError(s, i, true, err)
		log.Error(err)
		return
	}
}
