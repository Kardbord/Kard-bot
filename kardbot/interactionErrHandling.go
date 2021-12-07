package kardbot

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/TannerKvarfordt/Kard-bot/kardbot/dg_helpers"
	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	cmap "github.com/orcaman/concurrent-map"
	log "github.com/sirupsen/logrus"
)

const (
	genericErrorString    = "an error occurred. :'("
	selectMenuErrorReport = "select_error_report"
)

type errReportSelectionValue struct {
	// UUID of the associated errorReport
	ErrUUID uuid.UUID `json:"error-uuid,omitempty"`
	// Should the errorReport be made anonymously?
	Anonymous bool `json:"anonymous,omitempty"`
}

func (ers errReportSelectionValue) MarshalToString() string {
	buf, err := json.Marshal(ers)
	if err != nil {
		log.Error(err)
	}
	return string(buf)
}

func (ers *errReportSelectionValue) UnmarshalFromString(ersStr string) error {
	return json.Unmarshal([]byte(ersStr), &ers)
}

type errorReport struct {
	// UUID of this error report
	UUID uuid.UUID
	// The InteractionCreate event that caused the error
	InteractionCreate discordgo.InteractionCreate
	// The error that arose during the InteractionCreate event
	Err error
}

var (
	// Maps UUIDs to errorReports
	errsToReport = cmap.New()

	errReportMsgComponents = func(errUUID uuid.UUID) []discordgo.MessageComponent {
		ownerMention := ""
		if getOwnerID() == "" {
			ownerMention = "the bot owner"
		} else {
			owner, err := bot().Session.User(getOwnerID())
			if err != nil {
				log.Error(err)
				ownerMention = "the bot owner"
			} else {
				ownerMention = owner.Username
			}
		}
		return []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						CustomID:    selectMenuErrorReport,
						Placeholder: "Would you like to send an error report?",
						Options: []discordgo.SelectMenuOption{
							{
								Label:       "Send Anonymous Error Report",
								Description: fmt.Sprintf("Send an anonymous error report to %s.", ownerMention),
								Value:       errReportSelectionValue{errUUID, true}.MarshalToString(),
								Default:     false,
								Emoji: discordgo.ComponentEmoji{
									Name: "📮",
								},
							},
							{
								Label:       "Send Error Report",
								Description: fmt.Sprintf("Send an error report to %s.", ownerMention),
								Default:     false,
								Value:       errReportSelectionValue{errUUID, false}.MarshalToString(),
								Emoji: discordgo.ComponentEmoji{
									Name: "🗳️",
								},
							},
						},
					},
				},
			},
		}
	}
)

func interactionRespondEphemeralError(s *discordgo.Session, i *discordgo.InteractionCreate, notifyOwner bool, errResp error) {
	if s == nil {
		log.Error("nil session")
		return
	}
	if i == nil {
		log.Error("nil interaction")
		return
	}
	if errResp == nil {
		log.Warn("empty errStr, using generic error: ", genericErrorString)
		errResp = errors.New(genericErrorString)
	}

	if !notifyOwner {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprint(errResp),
				Flags:   InteractionResponseFlagEphemeral,
			},
		})
		if err != nil {
			log.Error(err)
		}
		return
	}

	if getOwnerID() == "" {
		log.Error("No ownerID provided, cannot send error report")
		return
	}

	errUUID := uuid.New()
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    "Something went wrong while processing your command. 😔",
			Flags:      InteractionResponseFlagEphemeral,
			Components: errReportMsgComponents(errUUID),
		},
	})
	if err != nil {
		log.Error(err)
		return
	}
	errsToReport.Set(fmt.Sprint(errUUID), errorReport{
		UUID:              errUUID,
		Err:               errResp,
		InteractionCreate: *i,
	})
}

// Assumes that a deferred response has already been sent.
// Will delete the deferred response and send an ephemeral follow up response.
func interactionFollowUpEphemeralError(s *discordgo.Session, i *discordgo.InteractionCreate, notifyOwner bool, errResp error) {
	if s == nil {
		log.Error("nil session")
		return
	}
	if i == nil {
		log.Error("nil interaction")
		return
	}
	if errResp == nil {
		log.Warn("empty errStr, using generic error: ", genericErrorString)
		errResp = errors.New(genericErrorString)
	}

	err := s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
	if err != nil {
		log.Warn(err)
	}

	if !notifyOwner {
		_, err = s.FollowupMessageCreate(s.State.User.ID, i.Interaction, false, &discordgo.WebhookParams{
			Content: fmt.Sprint(errResp),
			Flags:   InteractionResponseFlagEphemeral,
		})
		if err != nil {
			log.Error(err)
		}
		return
	}

	errUUID := uuid.New()
	_, err = s.FollowupMessageCreate(s.State.User.ID, i.Interaction, false, &discordgo.WebhookParams{
		Content:    "Something went wrong while processing your command. 😔",
		Flags:      InteractionResponseFlagEphemeral,
		Components: errReportMsgComponents(errUUID),
	})
	if err != nil {
		log.Error(err)
		return
	}
	errsToReport.Set(fmt.Sprint(errUUID), errorReport{
		UUID:              errUUID,
		Err:               errResp,
		InteractionCreate: *i,
	})
}

func handleErrorReportSelection(s *discordgo.Session, i *discordgo.InteractionCreate) {
	wg := bot().updateLastActive()
	defer wg.Wait()

	data := i.MessageComponentData()
	if len(data.Values) == 0 {
		log.Error("No values returned with component interaction data")
		return
	}

	selection := errReportSelectionValue{}
	err := selection.UnmarshalFromString(data.Values[0])
	if err != nil {
		log.Error(err)
		return
	}

	iErrReport, ok := errsToReport.Get(selection.ErrUUID.String())
	if !ok {
		log.Errorf("No error report found with UUID=%s", selection.ErrUUID)
		return
	}

	errReport, ok := iErrReport.(errorReport)
	if !ok {
		log.Error("Could not convert interface to errorReport")
		return
	}

	err = dmOwnerErrorReport(s, errReport, selection.Anonymous)
	if err != nil {
		log.Error(err)
		return
	}

	ownerMention := ""
	if getOwnerID() == "" {
		ownerMention = "The bot owner"
	} else {
		ownerMention = fmt.Sprintf("<@%s>", getOwnerID())
	}

	buttonPrefix := ""
	if selection.Anonymous {
		buttonPrefix = "Anonymous "
	}
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%s\nThanks for submitting an error report! %s has been notified of the problem.", i.Message.Content, ownerMention),
			Flags:   InteractionResponseFlagEphemeral,
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Users: []string{getOwnerID()},
			},
			// No good way to delete components from a message, so this will have to do for now.
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							CustomID: "no_handler",
							Label:    fmt.Sprintf("%sError Report Submitted", buttonPrefix),
							Style:    discordgo.SecondaryButton,
							Disabled: true,
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Error(err)
		return
	}
	errsToReport.Remove(selection.ErrUUID.String())
}

func dmOwnerErrorReport(s *discordgo.Session, errReport errorReport, anonymous bool) error {
	if s == nil {
		return errors.New("nil session provided")
	}
	metadata, err := getInteractionMetaData(&errReport.InteractionCreate)
	if err != nil {
		return err
	}
	uc, err := bot().Session.UserChannelCreate(metadata.AuthorID)
	if err != nil {
		return err
	}
	cmdJson, err := json.MarshalIndent(errReport.InteractionCreate.ApplicationCommandData(), "", "  ")
	if err != nil {
		return err
	}
	embed := dg_helpers.NewEmbed()
	if anonymous {
		embed.AddField("Afflicted User", "anonymous")
	} else {
		embed.AddField("Afflicted User", metadata.AuthorMention)
	}
	_, err = s.ChannelMessageSendComplex(uc.ID, &discordgo.MessageSend{
		Embed: embed.SetTitle("Error Report").
			AddField("Issued Command", fmt.Sprintf("```json\n%s\n```", cmdJson)).
			AddField("Error", fmt.Sprintf("```\n%s\n```", errReport.Err)).
			Truncate().
			MessageEmbed,
	})
	return err
}
