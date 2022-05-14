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
	tzSubCmdServerClock      = "server-clock"
	tzSubCmdServerClockTZOpt = "timezones"
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
	wg := bot().updateLastActive()
	defer wg.Wait()

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
	flags := InteractionResponseFlagEphemeral
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
			"Response is optionally ephemeral.")

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:  flags,
			Embeds: []*discordgo.MessageEmbed{e.Truncate().SetType(discordgo.EmbedTypeRich).MessageEmbed},
		},
	}, false, nil
}

func handleTZSubCmdInfo(s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, bool, error) {
	flags := InteractionResponseFlagEphemeral
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
				Flags:   InteractionResponseFlagEphemeral,
				Content: fmt.Sprintf(`For privacy reasons, this bot does not track user timezones. Please specify a specific IANA timezone rather than "%s".`, tz),
			},
		}, false, nil
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   InteractionResponseFlagEphemeral,
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

	// Abbreviation of the clock
	Abbrev string `json:"clock-abbreviation"`

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

	fileBytes, err := json.MarshalIndent(serverClocksMap, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(serverClockConfigFilepath, fileBytes, 0664)
}

func handleTZSubCmdServerClock(s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, bool, error) {
	//	unparsedTZs := ""
	//	format := tzSubCmdFmtDflt
	//	for _, opt := range i.ApplicationCommandData().Options[0].Options[0].Options {
	//		switch opt.Name {
	//		case tzSubCmdServerClockTZOpt:
	//			unparsedTZs = opt.StringValue()
	//		case tzSubCmdFmtOpt:
	//			format = opt.StringValue()
	//		default:
	//			log.Warn("Unknown option: ", opt.Name)
	//		}
	//	}
	//
	return nil, true, fmt.Errorf("unimplemented")
}

func (clock *serverClock) update() {
	clock.mutex.RLock()
	defer clock.mutex.RUnlock()
	currTime := time.Now().UTC()
	customTime := "UNIMPLEMENTED" // TODO: calculate current custom time

	var err error = nil
	if _, err = bot().Session.ChannelEdit(clock.ChannelID, customTime); err != nil {
		log.Error(err)
		atomic.AddUint32(&clock.ErrCount, 1)
		return
	}

	color, _ := fastHappyColorInt64()
	e := dg_helpers.NewEmbed().
		SetTitle("Server Clock").
		SetColor(int(color)).
		SetDescription("The top-most timezone in the list below is this server's custom clock. It is an averaged time calculated from the timezones below it, which were selected by a server administrator.").
		AddField(clock.Name, customTime).
		SetFooter("This bot supports [IANA](https://www.iana.org/time-zones) [Timezones](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones)")

	for _, tz := range clock.Timezones {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			log.Error(err)
			continue
		}
		e.AddField(loc.String(), currTime.Format(clock.Format))
	}

	if clock.MessageID != "" {
		_, err = bot().Session.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Embeds:  []*discordgo.MessageEmbed{e.Truncate().MessageEmbed},
			ID:      clock.MessageID,
			Channel: clock.ChannelID,
		})
	} else {
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
}

func updateServerClocks() {
	serverClocksMapMutex.RLock()
	defer serverClocksMapMutex.RUnlock()

	wg := &sync.WaitGroup{}
	for _, clock := range serverClocksMap {
		wg.Add(1)
		go func(c *serverClock) {
			if atomic.LoadUint32(&c.ErrCount) < bot().ServerClockFailureThreshold {
				c.update()
			} else if atomic.LoadUint32(&c.ErrCount) == bot().ServerClockFailureThreshold {
				c.mutex.RLock()
				log.Infof("Won't update defunct server clock for %s, it has failed to update %d times previously.", c.GuildName, c.ErrCount)
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
	wg.Wait()
}
