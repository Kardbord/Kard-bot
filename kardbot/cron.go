package kardbot

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"time"

	"github.com/TannerKvarfordt/Kard-bot/kardbot/dg_helpers"
	"github.com/bwmarrin/discordgo"
	"github.com/gabriel-vasile/mimetype"
	"github.com/go-co-op/gocron"

	log "github.com/sirupsen/logrus"
)

var scheduler = func() *gocron.Scheduler { return nil }
var genChanRegexp = func() *regexp.Regexp { return nil }

func init() {
	s := gocron.NewScheduler(time.UTC)
	if s == nil {
		log.Fatal("Could not create scheduler")
	}
	scheduler = func() *gocron.Scheduler { return s }

	// https://crontab.guru/#0_9_*_*_3
	scheduler().Cron("0 9 * * 3").Do(itIsWednesdayMyDudes)

	// https://crontab.guru/#*_*_*_*_*
	scheduler().Cron("* * * * *").Do(setStatus)

	// ^The above only initializes the scheduler, it does not start it.

	r := regexp.MustCompile("(?i)^general$")
	if r == nil {
		log.Fatal("nil Regexp")
	}
	genChanRegexp = func() *regexp.Regexp { return r }
}

const WednesdayAssetsDir string = AssetsDir + "/wednesday"

func itIsWednesdayMyDudes() {
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
	_, err = fd.Seek(0, 0)
	if err != nil {
		log.Error(err)
		return
	}

	attachment := &discordgo.File{
		Name:        img.Name(),
		ContentType: mimeType.String(),
		Reader:      fd,
	}

	hexColor, _ := fastHappyColorInt64()
	e := dg_helpers.NewEmbed()
	e.SetTitle("It is Wednesday my dudes").
		SetColor(int(hexColor)).
		SetImage("attachment://" + attachment.Name).
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
	if time.Since(bot().lastActive.Load()) > idleTimeoutMinutes {
		err := bot().Session.UpdateListeningStatus("")
		if err != nil {
			log.Error(err)
		}

		_, err = bot().Session.UserUpdateStatus(discordgo.StatusIdle)
		if err != nil {
			log.Error(err)
		}
	}
}
