package kardbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"sync"

	"github.com/TannerKvarfordt/Kard-bot/kardbot/config"
	"github.com/bwmarrin/discordgo"

	log "github.com/sirupsen/logrus"
)

const (
	complimentsOptIn   = "opt-in"
	complimentsOptOut  = "opt-out"
	complimentsMorning = "morning"
	complimentsEvening = "evening"
	complimentsGet     = "get-compliment"
	complimentInDM     = "dm"
)

var (
	// Morning compliment subscribers.
	// key: discord user ID
	// val: is currently subscribed
	complimentSubsAM      map[string]bool
	complimentSubsAMMutex sync.RWMutex

	// Evening compliment subscribers.
	// key: discord user ID
	// val: is currently subscribed
	complimentSubsPM      map[string]bool
	complimentSubsPMMutex sync.RWMutex

	// List of compliments
	compliments []string
)

const complimentSubscribersFilepath = "config/compliment-subscribers.json"

var complimentSubscribersFileMutex sync.RWMutex

type complimentSubscribersConfig struct {
	SubsAM map[string]bool `json:"compliment-subscribers-morning"`
	SubsPM map[string]bool `json:"compliment-subscribers-evening"`
}

func init() {
	complimentSubscribersFileMutex.RLock()
	defer complimentSubscribersFileMutex.RUnlock()
	complimentSubsAMMutex.Lock()
	defer complimentSubsAMMutex.Unlock()
	complimentSubsPMMutex.Lock()
	defer complimentSubsPMMutex.Unlock()

	cfg := complimentSubscribersConfig{}

	jsonCfg, err := config.NewJsonConfig(complimentSubscribersFilepath)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(jsonCfg.Raw, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	complimentSubsAM = cfg.SubsAM
	complimentSubsPM = cfg.SubsPM
}

const complimentListFilepath = "config/compliments.json"

func init() {
	cfg := struct {
		Compliments []string `json:"compliments"`
	}{}

	jsonCfg, err := config.NewJsonConfig(complimentListFilepath)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(jsonCfg.Raw, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	compliments = cfg.Compliments

	// Validate compliments
	if len(compliments) == 0 {
		log.Fatal("No compliments configured.")
	}
}

func complimentHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	wg := bot().updateLastActive()
	defer wg.Wait()

	if s == nil || i == nil {
		log.Errorf("nil session or interaction; s=%v, i=%v", s, i)
		return
	}

	switch i.ApplicationCommandData().Options[0].Name {
	case complimentsOptIn:
		switch i.ApplicationCommandData().Options[0].Options[0].Name {
		case complimentsMorning:
			morningComplimentOptIn(s, i)
		case complimentsEvening:
			eveningComplimentOptIn(s, i)
		default:
			log.Error("Unknown subcommand")
		}
	case complimentsOptOut:
		switch i.ApplicationCommandData().Options[0].Options[0].Name {
		case complimentsMorning:
			morningComplimentOptOut(s, i)
		case complimentsEvening:
			eveningComplimentOptOut(s, i)
		default:
			log.Error("Unknown subcommand")
		}
	case complimentsGet:
		getCompliment(s, i)
	default:
		log.Error("Unknown subcommand group")
	}
}

func morningComplimentOptIn(s *discordgo.Session, i *discordgo.InteractionCreate) {
	author, authorID, err := getInteractionCreateAuthorNameAndID(i)
	if err != nil {
		log.Error(err)
		return
	}

	complimentSubsAMMutex.Lock()
	complimentSubsAM[authorID] = true
	complimentSubsAMMutex.Unlock()

	err = writeComplimentSubscribersToConfig()
	if err != nil {
		log.Errorf("Error persisting user %s's subscription: %v", author, err)
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("%s, you are subscribed to receive morning compliments as long as the bot is up, but there was an error persisting your subscription. Please try to opt-in again.", author),
			},
		})
		if err != nil {
			log.Error(err)
		}
		return
	}

	log.Infof("User %s subscribed to morning compliments", author)
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%s has subscribed to receive daily morning compliments. :)", author),
		},
	})
	if err != nil {
		log.Error(err)
	}
}

func morningComplimentOptOut(s *discordgo.Session, i *discordgo.InteractionCreate) {
	author, authorID, err := getInteractionCreateAuthorNameAndID(i)
	if err != nil {
		log.Error(err)
		return
	}

	complimentSubsAMMutex.Lock()
	complimentSubsAM[authorID] = false
	complimentSubsAMMutex.Unlock()

	err = writeComplimentSubscribersToConfig()
	if err != nil {
		log.Errorf("Error persisting user %s's opt-out: %v", author, err)
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("%s, you are unsubscribed from morning compliments as long as the bot is up, but there was an error persisting your opt-out. Please try to opt-out again.", author),
			},
		})
		if err != nil {
			log.Error(err)
		}
		return
	}

	log.Infof("User %s un-subscribed to morning compliments", author)

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%s has unsubscribed from daily morning compliments. :(", author),
		},
	})
	if err != nil {
		log.Error(err)
	}
}

func eveningComplimentOptIn(s *discordgo.Session, i *discordgo.InteractionCreate) {
	author, authorID, err := getInteractionCreateAuthorNameAndID(i)
	if err != nil {
		log.Error(err)
		return
	}

	complimentSubsPMMutex.Lock()
	complimentSubsPM[authorID] = true
	complimentSubsPMMutex.Unlock()

	err = writeComplimentSubscribersToConfig()
	if err != nil {
		log.Errorf("Error persisting user %s's subscription: %v", author, err)
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("%s, you are subscribed to receive evening compliments as long as the bot is up, but there was an error persisting your subscription. Please try to opt-in again.", author),
			},
		})
		if err != nil {
			log.Error(err)
		}
		return
	}

	log.Infof("User %s subscribed to evening compliments", author)
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%s has subscribed to receive daily evening compliments. :)", author),
		},
	})
	if err != nil {
		log.Error(err)
	}
}

func eveningComplimentOptOut(s *discordgo.Session, i *discordgo.InteractionCreate) {
	author, authorID, err := getInteractionCreateAuthorNameAndID(i)
	if err != nil {
		log.Error(err)
		return
	}

	complimentSubsPMMutex.Lock()
	complimentSubsPM[authorID] = false
	complimentSubsPMMutex.Unlock()

	err = writeComplimentSubscribersToConfig()
	if err != nil {
		log.Errorf("Error persisting user %s's opt-out: %v", author, err)
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("%s, you are unsubscribed from evening compliments as long as the bot is up, but there was an error persisting your opt-out. Please try to opt-out again.", author),
			},
		})
		if err != nil {
			log.Error(err)
		}
		return
	}

	log.Infof("User %s un-subscribed to evening compliments", author)
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%s has unsubscribed from daily evening compliments. :(", author),
		},
	})
	if err != nil {
		log.Error(err)
	}
}

func getCompliment(s *discordgo.Session, i *discordgo.InteractionCreate) {
	author, authorID, err := getInteractionCreateAuthorNameAndID(i)
	if err != nil {
		log.Error(err)
		return
	}

	compliment := compliments[rand.Intn(len(compliments))]

	sendAsDM := false
	if len(i.ApplicationCommandData().Options[0].Options) > 0 {
		sendAsDM = i.ApplicationCommandData().Options[0].Options[0].BoolValue()
	}

	if sendAsDM {
		uc, err := bot().Session.UserChannelCreate(authorID)
		if err != nil {
			log.Error(err)
		}
		_, err = bot().Session.ChannelMessageSend(uc.ID, compliment)
		if err != nil {
			log.Error(err)
		}
		log.Infof("Told %s that '%s'", author, compliment)

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("I DM'd a compliment to %s. :)", author),
			},
		})
		if err != nil {
			log.Error(err)
		}
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: compliment,
		},
	})
	if err != nil {
		log.Error(err)
	}
	log.Infof("To %s: \"%s\"", author, compliment)
}

func sendMorningCompliments() {
	wg := bot().updateLastActive()
	defer wg.Wait()

	complimentSubsAMMutex.RLock()
	defer complimentSubsAMMutex.RUnlock()
	sendCompliments(complimentSubsAM)
}

func sendEveningCompliments() {
	wg := bot().updateLastActive()
	defer wg.Wait()

	complimentSubsPMMutex.RLock()
	defer complimentSubsPMMutex.RUnlock()
	sendCompliments(complimentSubsPM)
}

func sendCompliments(subscribers map[string]bool) error {
	var sendCompliment = func(subscriberID string, wg *sync.WaitGroup) {
		if wg == nil {
			log.Error("nil waitgroup provided")
			return
		}
		wg.Add(1)
		defer wg.Done()

		user, err := bot().Session.User(subscriberID)
		if err != nil {
			log.Error(err)
			return
		}

		uc, err := bot().Session.UserChannelCreate(subscriberID)
		if err != nil {
			log.Error(err)
		}

		compliment := compliments[rand.Intn(len(compliments))]
		_, err = bot().Session.ChannelMessageSend(uc.ID, compliment)
		if err != nil {
			log.Error(err)
		}
		log.Infof("Told %s that '%s'", user.Username, compliment)
	}

	wg := sync.WaitGroup{}
	for sid, isSubbed := range subscribers {
		if isSubbed {
			go sendCompliment(sid, &wg)
		}
	}

	wg.Wait()
	return nil
}

func writeComplimentSubscribersToConfig() error {
	complimentSubscribersFileMutex.Lock()
	defer complimentSubscribersFileMutex.Unlock()
	complimentSubsAMMutex.RLock()
	defer complimentSubsAMMutex.RUnlock()
	complimentSubsPMMutex.RLock()
	defer complimentSubsPMMutex.RUnlock()

	cfg := complimentSubscribersConfig{
		SubsAM: complimentSubsAM,
		SubsPM: complimentSubsPM,
	}

	fileBytes, err := json.MarshalIndent(cfg, "", "")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(complimentSubscribersFilepath, fileBytes, 0664)
}
