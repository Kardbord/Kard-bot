package kardbot

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

const (
	createRoleSelectCommand = "create-role-select"

	roleSelectContentOpt    = "msg-content"
	roleSelectContentOptIdx = 0

	roleSelectRolesOpt    = "roles"
	roleSelectRolesOptIdx = 1

	roleSelectOptCount = 2

	roleSelectMenuComponentIDPrefix = "role-select-menu"
)

func roleSelectCmdOpts() []*discordgo.ApplicationCommandOption {
	opts := make([]*discordgo.ApplicationCommandOption, roleSelectOptCount)
	for i := range opts {
		switch i {
		// TODO: Add a help option
		case roleSelectContentOptIdx:
			opts[i] = &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        roleSelectContentOpt,
				Description: "Context for this role select menu.",
				Required:    true,
			}
		case roleSelectRolesOptIdx:
			opts[i] = &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        roleSelectRolesOpt,
				Description: fmt.Sprintf("Roles (up to %d) and their context. Ex: @SomeRole context 😺 @NextRole next role context", maxDiscordSelectMenuOpts*maxDiscordActionRowSize),
				Required:    true,
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

	iEdit := &discordgo.WebhookEdit{
		// TODO: Use an embed instead?
		Content: i.ApplicationCommandData().Options[roleSelectContentOptIdx].StringValue(),
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
