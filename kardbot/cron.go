package kardbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/TannerKvarfordt/Kard-bot/kardbot/config"
	"github.com/TannerKvarfordt/Kard-bot/kardbot/dg_helpers"
	"github.com/bwmarrin/discordgo"
	"github.com/gabriel-vasile/mimetype"
	"github.com/go-co-op/gocron"

	log "github.com/sirupsen/logrus"
)

var scheduler = func() *gocron.Scheduler { return nil }

func init() {
	s := gocron.NewScheduler(time.Local)
	if s == nil {
		log.Fatal("Could not create scheduler")
	}
	scheduler = func() *gocron.Scheduler { return s }

	// https://crontab.guru/#0_9_*_*_3
	scheduler().Cron("0 9 * * 3").Do(itIsWednesdayMyDudes)

	// https://crontab.guru/#*_*_*_*_*
	scheduler().Cron("* * * * *").Do(setStatus)

	// https://crontab.guru/#30_7_*_*_*
	scheduler().Cron("30 7 * * *").Do(sendMorningCompliments)

	// https://crontab.guru/#30_20_*_*_*
	scheduler().Cron("30 20 * * *").Do(sendEveningCompliments)

	// https://crontab.guru/#0_0_*_*_*
	scheduler().Cron("0 0 * * *").Do(sendCreepyDMs)

	// ^The above only initializes the scheduler, it does not start it.
}

const WednesdayAssetsDir string = AssetsDir + "/wednesday"

var genChanRegexp = func() *regexp.Regexp { return nil }

func init() {
	r := regexp.MustCompile("(?i)^general$")
	if r == nil {
		log.Fatal("nil Regexp")
	}
	genChanRegexp = func() *regexp.Regexp { return r }
}

func itIsWednesdayMyDudes() {
	wg := bot().updateLastActive()
	defer wg.Wait()

	log.Info("It is wednesday my dudes")
	session := bot().Session
	if session == nil {
		log.Error("nil session")
		return
	}

	guilds, err := session.UserGuilds(100, "", "")
	if err != nil {
		log.Error(err)
	}

	// Prepare the message contents
	imgCandidates, err := ioutil.ReadDir(WednesdayAssetsDir)
	if err != nil {
		log.Error(err)
		return
	}
	if len(imgCandidates) < 1 {
		log.Error("No wednesday images")
		return
	}

	img := imgCandidates[rand.Intn(len(imgCandidates))]
	if !isImageRegex().MatchString(img.Name()) {
		log.Errorf("%s is not an image", img.Name())
		return
	}

	log.Debugf("Opening %s/%s", WednesdayAssetsDir, img.Name())
	fd, err := os.Open(fmt.Sprintf("%s/%s", WednesdayAssetsDir, img.Name()))
	if err != nil {
		log.Error(err)
		return
	}
	defer fd.Close()

	mimeType, err := mimetype.DetectReader(fd)
	if err != nil {
		log.Error(err)
		return
	}
	hexColor, _ := fastHappyColorInt64()
	e := dg_helpers.NewEmbed()
	e.SetTitle("It is Wednesday my dudes").
		SetColor(int(hexColor)).
		SetImage("attachment://" + img.Name()).
		Truncate()

	for _, g := range guilds {
		if g == nil {
			log.Warn("nil guild encountered")
			continue
		}

		chans, err := session.GuildChannels(g.ID)
		if err != nil {
			log.Error(err)
			continue
		}

		for _, c := range chans {
			if c.Type != discordgo.ChannelTypeGuildText {
				continue
			}
			if genChanRegexp().MatchString(c.Name) {
				_, err = fd.Seek(0, 0)
				if err != nil {
					log.Error(err)
					break
				}
				attachment := &discordgo.File{
					Name:        img.Name(),
					ContentType: mimeType.String(),
					Reader:      fd,
				}
				_, err := session.ChannelMessageSendComplex(c.ID, &discordgo.MessageSend{
					Embed: e.MessageEmbed,
					Files: []*discordgo.File{attachment},
				})
				if err != nil {
					log.Error(err)
				}
				break
			}
		}
	}
}

const idleTimeoutMinutes time.Duration = time.Minute * 5

func setStatus() {
	if bot().status.Load() != string(discordgo.StatusIdle) && time.Since(bot().lastActive.Load()) > idleTimeoutMinutes {
		err := bot().Session.UpdateListeningStatus("")
		if err != nil {
			log.Error(err)
		}

		idleSince := int(time.Now().Local().UnixMilli())
		err = bot().Session.UpdateStatusComplex(discordgo.UpdateStatusData{
			IdleSince: &idleSince,
			AFK:       true,
			Status:    string(discordgo.StatusIdle),
		})
		if err != nil {
			log.Error(err)
		} else {
			bot().status.Store(string(discordgo.StatusIdle))
			log.Infof("Set bot status to %s", bot().status.Load())
		}
	}
}

func sendMorningCompliments() {
	wg := bot().updateLastActive()
	defer wg.Wait()

	bot().complimentSubsAMMutex.RLock()
	defer bot().complimentSubsAMMutex.RUnlock()
	sendCompliments(bot().ComplimentSubsAM)
}

func sendEveningCompliments() {
	wg := bot().updateLastActive()
	defer wg.Wait()

	bot().complimentSubsPMMutex.RLock()
	defer bot().complimentSubsPMMutex.RUnlock()
	sendCompliments(bot().ComplimentSubsPM)
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

		compliment := bot().Compliments[rand.Intn(len(bot().Compliments))]
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

const defaultCreepyDMOdds float32 = 0.5

var creepyDMOdds = func() float32 { return defaultCreepyDMOdds }

func init() {
	var cfg = struct {
		CreepyDMOdds float32 `json:"creepy-dm-odds"`
	}{defaultCreepyDMOdds}
	json.Unmarshal(config.RawJSONConfig(), &cfg)

	if cfg.CreepyDMOdds != 0.0 {
		creepyDMOdds = func() float32 { return cfg.CreepyDMOdds }
	}

	if creepyDMOdds() < 0.0 || creepyDMOdds() > 1.0 {
		log.Fatalf("creepyDMOdds configuration value (%f) is out of range. Valid values are [0.0, 1.0]", creepyDMOdds())
	}

	if creepyDMOdds() < 0.01 {
		log.Warn("creepyDMOdds set at less than 1%")
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
	bot().creepyDMSubsMutex.RLock()
	// Defers are LIFO; this MUST happen prior to wg.Wait
	// or any subscribe/unsubscribe commands will be deadlocked.
	defer bot().creepyDMSubsMutex.RUnlock()

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
		if rand.Float32() > creepyDMOdds() {
			log.Infof("%s escaped a creepy DM this time...", user.Username)
			return nil
		}
		log.Infof("%s will get a creepy DM today >:)", user.Username)

		dm := bot().CreepyDMs[rand.Intn(len(bot().CreepyDMs))]
		uc, err := bot().Session.UserChannelCreate(subID)
		if err != nil {
			return err
		}

		_, err = bot().Session.ChannelMessageSend(uc.ID, dm)
		return err
	}

	for subscriberID := range bot().CreepyDMSubs {
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
	bot().creepyDMSubsMutex.RLock()
	defer bot().creepyDMSubsMutex.RUnlock()

	if ok, isSubbed := bot().CreepyDMSubs[subscriberID]; !ok || !isSubbed {
		return false
	}
	return true
}
