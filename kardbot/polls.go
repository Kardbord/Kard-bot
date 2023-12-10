package kardbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Kardbord/Kard-bot/kardbot/config"
	"github.com/Kardbord/Kard-bot/kardbot/dg_helpers"
	"github.com/Kardbord/ubiquity/mathutils"
	"github.com/bwmarrin/discordgo"
	"github.com/forPelevin/gomoji"
	cmap "github.com/orcaman/concurrent-map/v2"
	log "github.com/sirupsen/logrus"
)

type poll struct {
	// Poll ID
	MessageID string

	// Channel containing the poll
	ChannelID string

	// Maps user IDs to votes cast.
	// Key: Discord User ID
	// Val: []string
	Votes cmap.ConcurrentMap[string, []string]

	// The date the poll was opened
	Open time.Time

	// The date the poll is to close
	Close time.Time
}

// Create a new poll which closes one week from when it is opened.
// TODO: allow a user-specified duration.
func newPoll(messageID, channelID string) poll {
	return poll{
		MessageID: messageID,
		ChannelID: channelID,
		Votes:     cmap.New[[]string](),
		Open:      time.Now().UTC(),
		Close:     time.Now().UTC().AddDate(0, 0, 7),
	}
}

// Tracks existing polls.
// Key: MessageID
// Val: poll
var polls cmap.ConcurrentMap[string, poll] = cmap.New[poll]()

const pollsStorageFilepath = "config/polls.json"

var pollStorageFileMutex sync.RWMutex

func init() {
	pollStorageFileMutex.RLock()
	defer pollStorageFileMutex.RUnlock()

	jsonCfg, err := config.NewJsonConfig(pollsStorageFilepath)
	if err != nil {
		log.Fatal(err)
	}

	type cfg struct {
		poll
		Votes map[string][]string // hacky workaround to the ConcurrentMap from JSON
	}

	tmp := make(map[string]cfg)

	if err := json.Unmarshal(jsonCfg.Raw, &tmp); err != nil {
		log.Fatal(err)
	}

	for key, val := range tmp {
		p := poll{
			MessageID: val.MessageID,
			ChannelID: val.ChannelID,
			Votes:     cmap.New[[]string](),
			Open:      val.Open,
			Close:     val.Close,
		}
		for k, v := range val.Votes {
			p.setVotes(k, v...)
		}
		polls.Set(key, p)
	}
}

func writePollsToDisk() error {
	fileBytes, err := json.MarshalIndent(polls, "", "  ")
	if err != nil {
		return err
	}

	pollStorageFileMutex.Lock()
	defer pollStorageFileMutex.Unlock()
	return ioutil.WriteFile(pollsStorageFilepath, fileBytes, 0644)
}

func purgeFinishedPolls() error {
	for key, p := range polls.Items() {
		if p.Close.Before(time.Now().UTC()) {
			polls.Remove(key)
			if message, err := bot().Session.ChannelMessage(p.ChannelID, p.MessageID); err == nil {
				if len(message.Embeds) > 0 {
					message.Embeds[0].Footer = &discordgo.MessageEmbedFooter{
						Text: "This poll is now closed.",
					}
				}
				_, err = bot().Session.ChannelMessageEditComplex(&discordgo.MessageEdit{
					Content:    &message.Content,
					Components: []discordgo.MessageComponent{},
					Embeds:     message.Embeds,
					AllowedMentions: &discordgo.MessageAllowedMentions{
						Parse: []discordgo.AllowedMentionType{
							discordgo.AllowedMentionTypeEveryone,
							discordgo.AllowedMentionTypeRoles,
							discordgo.AllowedMentionTypeUsers,
						},
					},
					Flags:   message.Flags,
					ID:      message.ID,
					Channel: message.ChannelID,
				})
				if err != nil {
					log.Error(err)
				}
			} else {
				log.Error(err)
			}
		}
	}

	return writePollsToDisk()
}

const (
	pollCmd                 = "poll"
	pollCmdOptMaxSelections = "max-selections"
	pollCmdOptTitle         = "title"
	pollCmdOptContext       = "context"

	pollSelectMenuID = "poll-menu"
)

func getPollOpts() []*discordgo.ApplicationCommandOption {
	opts := []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        pollCmdOptTitle,
			Description: fmt.Sprintf("Title of the poll, maximum %d characters.", maxDiscordSelectMenuPlaceholderChars),
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        pollCmdOptMaxSelections,
			Description: "The maximum number of options a user can vote for",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        pollCmdOptContext,
			Description: "Additional context for the poll",
			Required:    false,
		},
	}
	addnlOpts := make([]*discordgo.ApplicationCommandOption, mathutils.Min(maxDiscordCommandOptions-len(opts), maxDiscordSelectMenuOpts, dg_helpers.EmbedLimitField))
	for i := range addnlOpts {
		addnlOpts[i] = &discordgo.ApplicationCommandOption{}
		opt := addnlOpts[i]
		opt.Type = discordgo.ApplicationCommandOptionString
		opt.Name = fmt.Sprintf("option-%d", i)
		opt.Description = fmt.Sprintf("Poll option %d", 1)
		opt.Required = false
	}
	return append(opts, addnlOpts...)
}

func handlePollCmd(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if s == nil || i == nil {
		log.Error(fmt.Errorf("nil Session pointer (%v) and/or InteractionCreate pointer (%v)", s, i))
		return
	}

	minSelections := 0
	maxSelections := 1
	title := ""
	context := ""
	pollOpts := make([]discordgo.SelectMenuOption, 0, len(i.ApplicationCommandData().Options))
	for _, opt := range i.ApplicationCommandData().Options {
		switch opt.Name {
		case pollCmdOptMaxSelections:
			maxSelections = int(opt.IntValue())
		case pollCmdOptTitle:
			title = opt.StringValue()
		case pollCmdOptContext:
			context = opt.StringValue()
		default:
			emoji, trimmedLabel, err := detectAndScrubDiscordEmojis(opt.StringValue())
			if err != nil {
				log.Error(err)
				interactionRespondEphemeralError(s, i, true, err)
				return
			}
			trimmedLabel = strings.TrimSpace(gomoji.RemoveEmojis(trimmedLabel))
			if len(trimmedLabel) == 0 {
				interactionRespondEphemeralError(s, i, false, fmt.Errorf("options must contain at least one non-whitespace, non-emoji character"))
				return
			}
			pollOpts = append(pollOpts, discordgo.SelectMenuOption{
				// We'll trim off any emojis in the label after we've used
				// it to build our Embed.
				Label: opt.StringValue(),
				Value: trimmedLabel,
				Emoji: emoji,
			})
		}
	}

	if maxSelections < 1 {
		interactionRespondEphemeralError(s, i, false, fmt.Errorf("you must allow at least 1 vote to be cast per user"))
		return
	}

	if len(pollOpts) == 0 {
		interactionRespondEphemeralError(s, i, false, fmt.Errorf("you must specify at least one poll option"))
		return
	}

	maxSelections = mathutils.Min(len(pollOpts), maxSelections)

	color, _ := fastHappyColorInt64()
	e := dg_helpers.NewEmbed().
		SetColor(int(color)).
		SetTitle(title).
		SetDescription(context)

	for i := range pollOpts {
		e.AddField(pollOpts[i].Label, "👍 0 votes, 📈 0% of votes cast")
		// Now we can finish trimming our SelectMenu Labels.
		pollOpts[i].Label = pollOpts[i].Value
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    pollSelectMenuID,
							Placeholder: title,
							MinValues:   &minSelections,
							MaxValues:   maxSelections,
							Options:     pollOpts,
						},
					},
				},
			},
			Embeds: []*discordgo.MessageEmbed{e.Truncate().MessageEmbed},
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{
					discordgo.AllowedMentionTypeEveryone,
					discordgo.AllowedMentionTypeRoles,
					discordgo.AllowedMentionTypeUsers,
				},
			},
		},
	})
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
	}

	resp, err := s.InteractionResponse(i.Interaction)
	if err != nil {
		_, filename, line, ok := runtime.Caller(1)
		if !ok {
			log.Error("couldn't obtain stack data")
		}
		followupWithError(s, i, err, filename, line)
		return
	}

	p := newPoll(resp.ID, resp.ChannelID)
	polls.Set(resp.ID, p)

	if err = writePollsToDisk(); err != nil {
		log.Error(err)
	}
}

func handlePollSubmission(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if s == nil || i == nil {
		log.Error(fmt.Errorf("nil Session pointer (%v) and/or InteractionCreate pointer (%v)", s, i))
		return
	}
	mdata, err := getInteractionMetaData(i)
	if err != nil {
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
		return
	}

	if !polls.Has(mdata.MessageID) {
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Sorry, this poll is closed!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			log.Error(err)
			interactionFollowUpEphemeralError(s, i, true, err)
		}
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Error(err)
		interactionRespondEphemeralError(s, i, true, err)
		return
	}

	p, _ := polls.Get(mdata.MessageID)
	p.setVotes(mdata.AuthorID, i.MessageComponentData().Values...)
	if err = p.updateMessage(s); err != nil {
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
		return
	}

	responseRecordedMsg := "Your response has been recorded! 🗳️"
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &responseRecordedMsg,
	})
	if err != nil {
		log.Error(err)
		interactionFollowUpEphemeralError(s, i, true, err)
	}
	writePollsToDisk()
}

func (p *poll) setVotes(userID string, votes ...string) {
	p.Votes.Set(userID, votes)
}

// Tablulates poll results and updates the discord message
func (p *poll) updateMessage(s *discordgo.Session) error {

	// TODO: This function can probably be a lot cleaner.
	//       I should know better than to try and bang out
	//       a feature when I'm short on time, but here we are.

	message, err := s.ChannelMessage(p.ChannelID, p.MessageID)
	if err != nil {
		return err
	}

	if len(message.Embeds) != 1 {
		return fmt.Errorf("expected a single embed, found %d", len(message.Embeds))
	}
	e := message.Embeds[0]

	// maps candidate names to vote count
	results := make(map[string]uint, len(e.Fields))

	totalVotesCast := 0
	for _, field := range e.Fields {
		_, trimmedName, err := detectAndScrubDiscordEmojis(field.Name)
		if err != nil {
			return err
		}
		trimmedName = gomoji.RemoveEmojis(trimmedName)
		trimmedName = strings.TrimSpace(trimmedName)
		results[field.Name] = 0
		for _, userVotes := range p.Votes.Items() {
			for _, vote := range userVotes {
				if trimmedName == vote {
					results[field.Name]++
					totalVotesCast++
				}
			}
		}
	}

	type resultsTuple struct {
		Name  string
		Value uint
	}
	sortedResults := make([]resultsTuple, 0, len(results))

	for candidate, votes := range results {
		sortedResults = append(sortedResults, resultsTuple{Name: candidate, Value: votes})
	}
	sort.Slice(sortedResults, func(i, j int) bool {
		return sortedResults[i].Value > sortedResults[j].Value
	})

	e.Fields = make([]*discordgo.MessageEmbedField, 0, len(results))
	for _, result := range sortedResults {
		percentOfVotes := float64(0)
		if totalVotesCast > 0 {
			percentOfVotes = math.Round((float64(result.Value) / float64(totalVotesCast)) * 100)
		}
		e.Fields = append(e.Fields, &discordgo.MessageEmbedField{
			Name:  result.Name,
			Value: fmt.Sprintf("👍 %d votes, 📈 %d%% of votes cast", result.Value, uint(percentOfVotes)),
		})
	}

	_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Content:    &message.Content,
		Components: message.Components,
		Embeds:     []*discordgo.MessageEmbed{e},
		AllowedMentions: &discordgo.MessageAllowedMentions{
			Parse: []discordgo.AllowedMentionType{
				discordgo.AllowedMentionTypeEveryone,
				discordgo.AllowedMentionTypeRoles,
				discordgo.AllowedMentionTypeUsers,
			},
		},
		Flags:   message.Flags,
		ID:      message.ID,
		Channel: message.ChannelID,
	})
	return err
}
