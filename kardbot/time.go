package kardbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TannerKvarfordt/Kard-bot/kardbot/config"
	"github.com/TannerKvarfordt/Kard-bot/kardbot/dg_helpers"
	"github.com/TannerKvarfordt/ubiquity/sliceutils"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

const (
	timeCmd             = "time"
	timeCmdOptEphemeral = "ephemeral"
)

const (
	timeSubCmdGroupTZ = "zone"

	// Opts common to all timeSubCmdGroup
	tzSubCmdFmtOpt  = "format"
	tzSubCmdFmtDflt = "Monday, 2006-01-02 3:04PM MST"

	// Sub command
	tzSubCmdHelp = "help"

	// Sub command
	tzSubCmdInfo      = "info"
	tzSubCmdInfoTZOpt = "timezone"

	//Sub command
	tzSubCmdServerClock              = "server-clock"
	tzSubCmdServerClockTZOpt         = "timezones"
	tzSubCmdServerClockCustomNameOpt = "clock-name"
)

func tzFormatOpts() []*discordgo.ApplicationCommandOptionChoice {
	return []*discordgo.ApplicationCommandOptionChoice{
		{
			Name:  "Default",
			Value: tzSubCmdFmtDflt,
		},
		{
			Name:  "Layout",
			Value: time.Layout,
		},
		{
			Name:  "ANSIC",
			Value: time.ANSIC,
		},
		{
			Name:  "UnixDate",
			Value: time.UnixDate,
		},
		{
			Name:  "RubyDate",
			Value: time.RubyDate,
		},
		{
			Name:  "RFC822",
			Value: time.RFC822,
		},
		{
			Name:  "RFC822Z",
			Value: time.RFC822Z,
		},
		{
			Name:  "RFC850",
			Value: time.RFC850,
		},
		{
			Name:  "RFC1123",
			Value: time.RFC1123,
		},
		{
			Name:  "RFC1123Z",
			Value: time.RFC1123Z,
		},
		{
			Name:  "RFC3339",
			Value: time.RFC3339,
		},
		{
			Name:  "RFC3339Nano",
			Value: time.RFC3339Nano,
		},
		{
			Name:  "Kitchen",
			Value: time.Kitchen,
		},
		{
			Name:  "Stamp",
			Value: time.Stamp,
		},
		{
			Name:  "StampMilli",
			Value: time.StampMilli,
		},
		{
			Name:  "StampMicro",
			Value: time.StampMicro,
		},
		{
			Name:  "StampNano",
			Value: time.StampNano,
		},
	}
}

func timeCmdOpts() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        timeSubCmdGroupTZ,
			Description: "Timezone related commands",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        tzSubCmdHelp,
					Description: "Get a list of valid time zones.",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionBoolean,
							Name:        timeCmdOptEphemeral,
							Description: "Should the bot's response be ephemeral? Defaults to true.",
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        tzSubCmdInfo,
					Description: "Get information about a given timezone",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        tzSubCmdInfoTZOpt,
							Description: "The IANA timezone to get information for.",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        tzSubCmdFmtOpt,
							Description: "The format in which the date should be displayed.",
							Choices:     tzFormatOpts(),
						},
						{
							Type:        discordgo.ApplicationCommandOptionBoolean,
							Name:        timeCmdOptEphemeral,
							Description: "Should the bot's response be ephemeral? Defaults to true.",
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        tzSubCmdServerClock,
					Description: "As an admin, create a custom clock for your server.",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        tzSubCmdServerClockTZOpt,
							Description: "Space-separated list of IANA timezones to use in creating the server clock.",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        tzSubCmdServerClockCustomNameOpt,
							Description: "A custom name to identify your server clock timezone.",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        tzSubCmdFmtOpt,
							Description: "The format in which the date should be displayed.",
							Choices:     tzFormatOpts(),
						},
					},
				},
			},
		},
	}
}

func handleTimeCmd(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if s == nil || i == nil {
		err := fmt.Errorf("nil Session pointer (%v) and/or InteractionCreate pointer (%v)", s, i)
		interactionRespondEphemeralError(s, i, true, err)
		log.Error(err)
		return
	}

	var (
		err           error                          = nil
		reportableErr                                = false
		resp          *discordgo.InteractionResponse = nil
	)
	subCmdOrGroup := i.ApplicationCommandData().Options[0].Name
	switch subCmdOrGroup {
	case timeSubCmdGroupTZ:
		resp, reportableErr, err = handleTZSubCmd(s, i)
	default:
		interactionRespondEphemeralError(s, i, true, fmt.Errorf("unknown subcommand: %s", subCmdOrGroup))
		return
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

func handleTZSubCmd(s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, bool, error) {
	subCmdName := i.ApplicationCommandData().Options[0].Options[0].Name
	switch subCmdName {
	case tzSubCmdHelp:
		return handleTZSubCmdHelp(s, i)
	case tzSubCmdInfo:
		return handleTZSubCmdInfo(s, i)
	case tzSubCmdServerClock:
		return handleTZSubCmdServerClock(s, i)
	default:
		return nil, true, fmt.Errorf("unknown %s sub command: %s", timeSubCmdGroupTZ, subCmdName)
	}
}

func handleTZSubCmdHelp(s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, bool, error) {
	flags := discordgo.MessageFlagsEphemeral
	for _, opt := range i.ApplicationCommandData().Options[0].Options[0].Options {
		switch opt.Name {
		case timeCmdOptEphemeral:
			if !opt.BoolValue() {
				flags = 0
			}
		default:
			log.Warn("Unknown option: ", opt.Name)
		}
	}

	c, _ := fastHappyColorInt64()
	e := dg_helpers.NewEmbed()
	e.SetTitle("Timezones").
		SetURL("https://en.wikipedia.org/wiki/List_of_tz_database_time_zones").
		SetColor(int(c)).
		SetDescription("This bot supports [Internet Assigned Numbers Authority (IANA)](https://www.iana.org/time-zones) governed timezones. "+
			"A convenient list of valid timezones can be found on [Wikipedia](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones#List). "+
			"See the _TZ database name_ or _Time zone abbreviation_ column. Note that inputs are case-sensitive.\n"+
			"\n*Valid Timezone Input Examples*\n"+
			"- America/Boise\n"+
			"- Asia/Hong_Kong\n"+
			"- Europe/Berlin\n"+
			"- EET\n"+
			"- MST\n"+
			"- MDT\n"+
			"\nSubcommands and their usage are documented below.\n").
		AddField(tzSubCmdHelp, "Prints this help message. Response is optionally ephemeral.").
		AddField(tzSubCmdInfo, "Provides general information about a given timezone. "+
			"Requires an [IANA timezone database name or abbreviation](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones#List) as input. "+
			"Optionally takes a date format in which the provided timezone should be displayed. "+
			"Response is optionally ephemeral.").
		AddField(tzSubCmdServerClock, "Creates a server clock channel that displays the current date and time for specified timezones. "+
			"Also creates an averaged \"Server Time\".")

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:  flags,
			Embeds: []*discordgo.MessageEmbed{e.Truncate().SetType(discordgo.EmbedTypeRich).MessageEmbed},
		},
	}, false, nil
}

func handleTZSubCmdInfo(s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, bool, error) {
	flags := discordgo.MessageFlagsEphemeral
	tz := ""
	format := tzSubCmdFmtDflt
	for _, opt := range i.ApplicationCommandData().Options[0].Options[0].Options {
		switch opt.Name {
		case timeCmdOptEphemeral:
			if !opt.BoolValue() {
				flags = 0
			}
		case tzSubCmdInfoTZOpt:
			tz = opt.StringValue()
		case tzSubCmdFmtOpt:
			format = opt.StringValue()
		default:
			log.Warn("Unknown option: ", opt.Name)
		}
	}

	tz = strings.TrimSpace(tz)
	if strings.ToLower(tz) == "local" {
		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: fmt.Sprintf(`For privacy reasons, this bot does not track user timezones. Please specify a specific IANA timezone rather than "%s".`, tz),
			},
		}, false, nil
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: fmt.Sprintf(`"%s" is not a valid [IANA Timezone](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones).`, tz),
			},
		}, false, nil
	}

	c, _ := fastHappyColorInt64()
	e := dg_helpers.NewEmbed()
	t := time.Now().In(loc)
	abbrev, offset := t.Zone()
	e.SetTitle(loc.String()).
		SetDescription(t.Format(format)).
		SetColor(int(c)).
		AddField("Abbreviation", abbrev).
		AddField("Daylight Savings Time in Effect?", fmt.Sprintf("%t", t.IsDST())).
		AddField("UTC/GMT Offset (hh:mm)", fmt.Sprintf(`%+03d:%02d`, offset/3600, func() int {
			seconds := (offset % 3600) / 60
			if seconds < 0 {
				return seconds * -1
			}
			return seconds
		}()))

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:  flags,
			Embeds: []*discordgo.MessageEmbed{e.Truncate().SetType(discordgo.EmbedTypeRich).MessageEmbed},
		},
	}, false, nil
}

type serverClock struct {
	// Guild owning the server clock.
	// Required.
	GuildID string `json:"guild-id"`

	// Name of the guild owning the server clock.
	GuildName string `json:"guild-name"`

	// Dedicated Channel for the server clock.
	// Required.
	ChannelID string `json:"channel-id"`

	// Message to update with the server clock time.
	// Optional. Will be added if it does not exist.
	MessageID string `json:"message-id"`

	// Timezones to be included in the server clock
	Timezones []string `json:"timezones"`

	// Name of the clock
	Name string `json:"clock-name"`

	// The format in which to display timezones
	Format string `json:"format"`

	// Number of consecutive times the clock has failed to update.
	// If it exceeds bot().ServerClockErrorThreshold, the bot will
	// no longer attempt to update this clock.
	ErrCount uint32 `json:"error-count"`

	mutex sync.RWMutex
}

const serverClockConfigFilepath = "config/server-clocks.json"

var serverClockConfigFilepathMutex sync.RWMutex

var (
	// Map of Guild IDs to serverClock objects
	serverClocksMap      map[string]*serverClock
	serverClocksMapMutex sync.RWMutex
)

func init() {
	serverClockConfigFilepathMutex.RLock()
	defer serverClockConfigFilepathMutex.RUnlock()
	serverClocksMapMutex.Lock()
	defer serverClocksMapMutex.Unlock()

	jsonCfg, err := config.NewJsonConfig(serverClockConfigFilepath)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(jsonCfg.Raw, &serverClocksMap)
	if err != nil {
		log.Fatal(err)
	}
}

func writeServerClocksToDisk() error {
	serverClockConfigFilepathMutex.Lock()
	defer serverClockConfigFilepathMutex.Unlock()
	serverClocksMapMutex.RLock()
	defer serverClocksMapMutex.RUnlock()
	for _, clock := range serverClocksMap {
		clock.mutex.RLock()
		defer clock.mutex.RUnlock()
	}

	fileBytes, err := json.MarshalIndent(serverClocksMap, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(serverClockConfigFilepath, fileBytes, 0664)
}

func handleTZSubCmdServerClock(s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, bool, error) {
	mdata, err := getInteractionMetaData(i)
	if err != nil {
		log.Error(err)
		return nil, true, err
	}

	if !hasPermissions(mdata.AuthorPermissions, discordgo.PermissionManageChannels) {
		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "You must run this command from a server where you have the Manage Channels permission.",
			},
		}, false, nil
	}

	serverClocksMapMutex.RLock()
	if clock, ok := serverClocksMap[mdata.GuildID]; ok {
		clock.mutex.RLock()
		chID := clock.ChannelID
		clock.mutex.RUnlock()
		if _, err := s.Channel(chID); err == nil {
			serverClocksMapMutex.RUnlock()
			return &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   discordgo.MessageFlagsEphemeral,
					Content: fmt.Sprintf("This server already has a clock. To replace it, delete the <#%s> channel and re-issue this command.", clock.ChannelID),
				},
			}, false, nil
		}
	}
	serverClocksMapMutex.RUnlock()

	tzs, clockname, format := []string{}, "", tzSubCmdFmtDflt
	for _, opt := range i.ApplicationCommandData().Options[0].Options[0].Options {
		switch opt.Name {
		case tzSubCmdServerClockTZOpt:
			tzs = sliceutils.RemoveDuplicates(strings.Split(opt.StringValue(), " ")...)
		case tzSubCmdFmtOpt:
			format = opt.StringValue()
		case tzSubCmdServerClockCustomNameOpt:
			clockname = opt.StringValue()
		default:
			log.Warn("Unknown option: ", opt.Name)
		}
	}

	invalidTZs := make([]string, 0, len(tzs))
	for _, tz := range tzs {
		if _, err := time.LoadLocation(tz); err != nil {
			invalidTZs = append(invalidTZs, tz)
		}
	}
	if len(invalidTZs) > 0 {
		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: fmt.Sprintf("The following time zones are not valid: `%v`", invalidTZs),
			},
		}, false, nil
	}

	g, err := s.Guild(mdata.GuildID)
	if err != nil {
		log.Error(err)
		return nil, true, err
	}

	tzChan, err := s.GuildChannelCreateComplex(g.ID, discordgo.GuildChannelCreateData{
		Name:                 clockname,
		Type:                 discordgo.ChannelTypeGuildText,
		Topic:                fmt.Sprintf("Server clock provided by %s.", s.State.User.Mention()),
		PermissionOverwrites: []*discordgo.PermissionOverwrite{}, // TODO: restrict channel so it is read-only?
	})
	if err != nil {
		log.Error(err)
		return nil, true, err
	}

	newClock := &serverClock{
		GuildID:   mdata.GuildID,
		GuildName: g.Name,
		ChannelID: tzChan.ID,
		Timezones: tzs,
		Name:      clockname,
		Format:    format,
	}
	serverClocksMapMutex.Lock()
	serverClocksMap[mdata.GuildID] = newClock
	serverClocksMapMutex.Unlock()
	newClock.update()

	if err := writeServerClocksToDisk(); err != nil {
		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: fmt.Sprintf("There was an error persisting your server clock. Please delete the %s channel if it was created, and reissue the command.", tzChan.Name),
			},
		}, false, nil
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: fmt.Sprintf("Your server clock has been created! Check it out at %s. You may want to pin the clock message in that channel, or make it read-only.", tzChan.Mention()),
		},
	}, false, nil
}

func (clock *serverClock) update() {
	log.Trace("Waiting for clock mutex.")
	clock.mutex.RLock()
	defer clock.mutex.RUnlock()
	currTime := time.Now().UTC()

	log.Trace("Calculating custom time")
	customTime := "00:00AM"
	customOffsetMinutes := 0
	for _, tz := range clock.Timezones {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			log.Error(err)
			continue
		}
		_, offset := currTime.In(loc).Zone()
		customOffsetMinutes += offset / 60
	}
	customOffsetMinutes = customOffsetMinutes / len(clock.Timezones)
	customTime = currTime.UTC().Add(time.Minute * time.Duration(customOffsetMinutes)).Format(time.Kitchen)

	log.Trace("Building embed")
	e := dg_helpers.NewEmbed().
		SetTitle("Server Clock").
		SetDescription("The top-most timezone in the list below is this server's custom clock. It is an averaged time calculated from the timezones below it, which were selected by a server administrator. This bot supports [IANA](https://www.iana.org/time-zones) [Timezones](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones).").
		AddField(clock.Name, customTime)

	for _, tz := range clock.Timezones {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			log.Error(err)
			continue
		}
		e.AddField(loc.String(), currTime.In(loc).Format(clock.Format))
	}

	var err error
	if clock.MessageID != "" {
		log.Trace("Creating new server clock message")
		_, err = bot().Session.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Embeds:  []*discordgo.MessageEmbed{e.Truncate().MessageEmbed},
			ID:      clock.MessageID,
			Channel: clock.ChannelID,
		})
	} else {
		log.Trace("Updating existing server clock message")
		var m *discordgo.Message = nil
		m, err = bot().Session.ChannelMessageSendComplex(clock.ChannelID, &discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{e.Truncate().MessageEmbed},
		})
		if err == nil {
			clock.mutex.RUnlock()
			clock.mutex.Lock()
			clock.MessageID = m.ID
			clock.mutex.Unlock()
			clock.mutex.RLock()
		}
	}

	if err != nil {
		log.Error(err)
		atomic.AddUint32(&clock.ErrCount, 1)
		return
	}
	atomic.StoreUint32(&clock.ErrCount, 0)
	log.Trace("Done with clock update.")
}

func updateServerClocks() {
	log.Trace("Waiting for serverClocksMap mutex")
	serverClocksMapMutex.RLock()
	defer serverClocksMapMutex.RUnlock()

	wg := &sync.WaitGroup{}
	log.Trace("Starting clock updates")
	for _, clock := range serverClocksMap {
		wg.Add(1)
		go func(c *serverClock) {
			if atomic.LoadUint32(&c.ErrCount) < bot().ServerClockFailureThreshold {
				c.update()
			} else if atomic.LoadUint32(&c.ErrCount) == bot().ServerClockFailureThreshold {
				c.mutex.RLock()
				log.Warnf("Won't update defunct server clock for %s, it has failed to update %d times previously.", c.GuildName, c.ErrCount)
				bot().Session.ChannelMessageSend(c.ChannelID, fmt.Sprintf(
					"This clock has failed to update %d consecutive times, and is now considered defunct. Ensure that Kard-bot has appropriate permissions, then delete this channel and reissue the `/%s %s %s` command.",
					c.ErrCount, timeCmd, timeSubCmdGroupTZ, tzSubCmdServerClock,
				))
				c.mutex.RUnlock()
				// Only report defunct once.
				atomic.AddUint32(&c.ErrCount, 1)
			}
			wg.Done()
		}(clock)
	}
	log.Trace("Waiting for clock updates to complete.")
	wg.Wait()
	log.Trace("Done with all clock updates for this minute.")
}
