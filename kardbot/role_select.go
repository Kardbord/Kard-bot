package kardbot

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/TannerKvarfordt/Kard-bot/kardbot/dg_helpers"
	"github.com/bwmarrin/discordgo"
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

func createRoleSelect(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if s == nil || i == nil {
		log.Errorf("nil Session pointer (%v) and/or InteractionCreate pointer (%v)", s, i)
		return
	}
	wg := bot().updateLastActive()
	defer wg.Wait()

	metadata, err := getInteractionMetaData(i)
	if err != nil {
		interactionRespondEphemeralError(s, i, true, err)
		log.Error(err)
		return
	}

	if metadata.AuthorPermissions&discordgo.PermissionAdministrator == 0 {
		err = fmt.Errorf("you must be a server administrator to run this command")
		interactionRespondEphemeralError(s, i, false, err)
		log.Error(err)
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

	const roleMentionDelim = "<@&"
	if !strings.HasPrefix(i.ApplicationCommandData().Options[roleSelectRolesOptIdx].StringValue(), roleMentionDelim) {
		interactionFollowUpEphemeralError(s, i, false, fmt.Errorf("the first argument to `%s` must be a role mention", roleSelectRolesOpt))
		return
	}
	rolesAndCtx := strings.Split(i.ApplicationCommandData().Options[roleSelectRolesOptIdx].StringValue(), roleMentionDelim)[1:]
	if len(rolesAndCtx) > maxDiscordSelectMenuOpts*maxDiscordActionRowSize {
		interactionFollowUpEphemeralError(s, i, false, fmt.Errorf("you may only specify up to %d roles", maxDiscordSelectMenuOpts*maxDiscordActionRowSize))
		log.Error(err)
		return
	}
	if len(rolesAndCtx) == 0 {
		interactionFollowUpEphemeralError(s, i, false, fmt.Errorf("you must specify at least one role"))
		log.Error(err)
		return
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
		interactionFollowUpEphemeralError(s, i, true, err)
		log.Error(err)
		return
	}

	for idx := range rolesAndCtx {
		rolesAndCtx[idx] = roleMentionDelim + rolesAndCtx[idx]
		roleID := strings.Replace(strings.Replace(roleRegex().FindString(rolesAndCtx[idx]), "<@&", "", 1), ">", "", 1)
		role, ok := roleMap[roleID]
		if !ok {
			err = fmt.Errorf("no role found for ID: %s", roleID)
			interactionFollowUpEphemeralError(s, i, true, err)
			log.Error(err)
			return
		}

		smIdx := idx / 25
		sMenus[smIdx].Options = append(sMenus[smIdx].Options, discordgo.SelectMenuOption{
			Label: role.Name,
			Value: role.ID,
			// TODO: detect emojis and place them in the appropriate field
			Description: roleRegex().ReplaceAllString(rolesAndCtx[idx], ""),
		})
	}

	e := buildRoleSelectMenuEmbed(i)
	if e.URL != "" && !isReachableURL(e.URL) {
		interactionFollowUpEphemeralError(s, i, false, fmt.Errorf("unreachable URL provided: %s", e.URL))
		return
	}
	if e.Thumbnail != nil && e.Thumbnail.URL != "" && !isReachableURL(e.Thumbnail.URL) {
		interactionFollowUpEphemeralError(s, i, false, fmt.Errorf("unreachable thumbnail URL provided: %s", e.Thumbnail.URL))
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
