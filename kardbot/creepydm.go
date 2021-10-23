package kardbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"sync"
	"time"

	"github.com/TannerKvarfordt/Kard-bot/kardbot/config"
	"github.com/bwmarrin/discordgo"

	log "github.com/sirupsen/logrus"
)

const (
	creepyDMGet     = "get-creepy-dm"
	creepyChannelDM = "to-channel"
	creepyDMOptIn   = "opt-in"
	creepyDMOptOut  = "opt-out"

	defaultCreepyDMOdds float32 = 0.5
)

var (
	// Creepy DM subscribers
	// key: discord user ID
	// val: is currently subscribed
	creepyDMSubs      map[string]bool
	creepyDMSubsMutex sync.RWMutex

	// Odds [0.0, 1.0] that the user will receive
	// a creepy DM on any given day.
	creepyDMOdds float32

	// List of creepy DMs
	creepyDMs []string
)

type creepyDmSubscribersConfig struct {
	Subs map[string]bool `json:"creepy-dm-subscribers"`
	Odds float32         `json:"creepy-dm-odds"`
}

const creepyDmSubscribersFilepath = "config/creepy-dm-subscribers.json"

var creepyDmSubscribersFileMutex sync.RWMutex

func init() {
	creepyDmSubscribersFileMutex.RLock()
	defer creepyDmSubscribersFileMutex.RUnlock()
	creepyDMSubsMutex.Lock()
	defer creepyDMSubsMutex.Unlock()

	cfg := creepyDmSubscribersConfig{}

	cfg.Odds = defaultCreepyDMOdds

	jsonCfg, err := config.NewJsonConfig(creepyDmSubscribersFilepath)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(jsonCfg.Raw, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	creepyDMOdds = cfg.Odds
	creepyDMSubs = cfg.Subs

	if creepyDMOdds < 0.0 || creepyDMOdds > 1.0 {
		log.Fatalf("creepyDMOdds configuration value (%f) is out of range. Valid values are [0.0, 1.0]", creepyDMOdds)
	}

	if creepyDMOdds < 0.01 {
		log.Warn("creepyDMOdds set at less than 1%")
	}
}

const creepyDmListFilepath = "config/creepy-dms.json"

func init() {
	cfg := struct {
		CreepyDMs []string `json:"creepy-dms"`
	}{}

	jsonCfg, err := config.NewJsonConfig(creepyDmListFilepath)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(jsonCfg.Raw, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	creepyDMs = cfg.CreepyDMs
}

func creepyDMHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	wg := bot().updateLastActive()
	defer wg.Wait()

	if s == nil || i == nil {
		log.Errorf("nil session or interaction; s=%v, i=%v", s, i)
		return
	}

	switch i.ApplicationCommandData().Options[0].Name {
	case creepyDMGet:
		getCreepyDM(s, i)
	case creepyDMOptIn:
		creepyDMsOptIn(s, i)
	case creepyDMOptOut:
		creepyDMsOptOut(s, i)
	default:
		log.Error("Unknown subcommand")
	}
}

func creepyDMsOptIn(s *discordgo.Session, i *discordgo.InteractionCreate) {
	author, authorID, err := getInteractionCreateAuthorNameAndID(i)
	if err != nil {
		log.Error(err)
		return
	}

	creepyDMSubsMutex.Lock()
	creepyDMSubs[authorID] = true
	creepyDMSubsMutex.Unlock()

	err = writeCreepyDmSubscribersToConfig()
	if err != nil {
		log.Errorf("Error persisting user %s's subscription: %v", author, err)
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("%s, you are subscribed to creepy DMs as long as the bot remains up, but there was an error persisting your subscription. Please try to opt-in again.", author),
			},
		})
		if err != nil {
			log.Error(err)
		}
		return
	}

	log.Infof("User %s subscribed to creepy DMs", author)
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%s subscribed to creepy DMs 😈", author),
		},
	})
	if err != nil {
		log.Error(err)
	}
}

func creepyDMsOptOut(s *discordgo.Session, i *discordgo.InteractionCreate) {
	author, authorID, err := getInteractionCreateAuthorNameAndID(i)
	if err != nil {
		log.Error(err)
		return
	}

	creepyDMSubsMutex.Lock()
	creepyDMSubs[authorID] = false
	creepyDMSubsMutex.Unlock()

	err = writeCreepyDmSubscribersToConfig()
	if err != nil {
		log.Errorf("Error persisting user %s's opt-out: %v", author, err)
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("%s, you are unsubscribed to creepy DMs as long as the bot remains up, but there was an error persisting your opt-out. Please try to opt-out again.", author),
			},
		})
		if err != nil {
			log.Error(err)
		}
		return
	}

	log.Infof("User %s unsubscribed from creepy DMs", author)

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%s unsubscribed from creepy DMs 👿", author),
		},
	})
	if err != nil {
		log.Error(err)
	}
}

func getCreepyDM(s *discordgo.Session, i *discordgo.InteractionCreate) {
	msg := creepyDMs[rand.Intn(len(creepyDMs))]

	sendToChannel := false
	if len(i.ApplicationCommandData().Options[0].Options) > 0 {
		sendToChannel = i.ApplicationCommandData().Options[0].Options[0].BoolValue()
	}

	if sendToChannel {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: msg,
			},
		})
		if err != nil {
			log.Error(err)
		}
		return
	}

	author, authorID, err := getInteractionCreateAuthorNameAndID(i)
	if err != nil {
		log.Error(err)
		return
	}

	uc, err := s.UserChannelCreate(authorID)
	if err != nil {
		log.Error(err)
		return
	}

	_, err = s.ChannelMessageSend(uc.ID, msg)
	if err != nil {
		log.Error(err)
		return
	}
	log.Tracef("Sent %s a creepy DM", author)

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%s was sent a creepy DM", author),
		},
	})
	if err != nil {
		log.Error(err)
	}
}

// sendCreepyDMs is run every day. It spawns a goroutine for each
// subscriber and randomly decides whether or not said subscriber
// will receive creepy-PM that day. If so the goroutine sleeps
// for a random amount of time, not exceeding 24 hours, before
// sending the DM.
func sendCreepyDMs() {
	wg := sync.WaitGroup{}
	defer wg.Wait()
	creepyDMSubsMutex.RLock()
	// Defers are LIFO; this MUST happen prior to wg.Wait
	// or any subscribe/unsubscribe commands will be deadlocked.
	defer creepyDMSubsMutex.RUnlock()

	// Made this an anonymous inner function so that I wouldn't
	// accidentally use an uncopied value from bot().CreepyDMSubs
	// after the mutex is released.
	sendDM := func(subID string) error {
		const minutesPerDay = 1440
		time.Sleep(time.Minute * time.Duration(rand.Intn(minutesPerDay)))

		activeWG := bot().updateLastActive()
		defer activeWG.Wait()

		user, err := bot().Session.User(subID)
		if err != nil {
			return err
		}
		if !isSubbedToCreepyDMs(subID, user.Username) {
			log.Infof("%s has unsubbed from creepy DMs since this routine started", user.Username)
			return nil
		}
		if rand.Float32() > creepyDMOdds {
			log.Infof("%s escaped a creepy DM this time...", user.Username)
			return nil
		}
		log.Infof("%s will get a creepy DM today >:)", user.Username)

		dm := creepyDMs[rand.Intn(len(creepyDMs))]
		uc, err := bot().Session.UserChannelCreate(subID)
		if err != nil {
			return err
		}

		_, err = bot().Session.ChannelMessageSend(uc.ID, dm)
		return err
	}

	for subscriberID := range creepyDMSubs {
		wg.Add(1)
		go func(subID string) {
			defer wg.Done()
			if err := sendDM(subID); err != nil {
				log.Error(err)
			}
		}(subscriberID)
	}
}

func isSubbedToCreepyDMs(subscriberID, subscriberName string) bool {
	creepyDMSubsMutex.RLock()
	defer creepyDMSubsMutex.RUnlock()

	if ok, isSubbed := creepyDMSubs[subscriberID]; !ok || !isSubbed {
		return false
	}
	return true
}

func writeCreepyDmSubscribersToConfig() error {
	creepyDmSubscribersFileMutex.Lock()
	defer creepyDmSubscribersFileMutex.Unlock()
	creepyDMSubsMutex.RLock()
	defer creepyDMSubsMutex.RUnlock()

	cfg := creepyDmSubscribersConfig{
		Subs: creepyDMSubs,
		Odds: creepyDMOdds,
	}

	fileBytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(creepyDmSubscribersFilepath, fileBytes, 0664)
}
