package kardbot

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"

	"github.com/TannerKvarfordt/Kard-bot/kardbot/dg_helpers"
	"github.com/bwmarrin/discordgo"
	"github.com/gabriel-vasile/mimetype"

	log "github.com/sirupsen/logrus"
)

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

	msg, err := chooseWednesdayImage()
	if err != nil {
		log.Error(err)
		return
	}

	guilds, err := session.UserGuilds(100, "", "")
	if err != nil {
		log.Error(err)
	}

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
				_, err := session.ChannelMessageSendComplex(c.ID, msg)
				if err != nil {
					log.Error(err)
				}
				break
			}
		}
	}
}

func chooseWednesdayImage() (*discordgo.MessageSend, error) {
	imgCandidates, err := ioutil.ReadDir(WednesdayAssetsDir)
	if err != nil {
		return nil, err
	}
	// TODO: add a chance of using dalle2 to generate the Wednesday image
	if len(imgCandidates) < 1 {
		return nil, fmt.Errorf("no wednesday images")
	}

	img := imgCandidates[rand.Intn(len(imgCandidates))]
	if !isImageRegex().MatchString(img.Name()) {
		return nil, fmt.Errorf("%s is not an image", img.Name())
	}

	log.Debugf("Opening %s/%s", WednesdayAssetsDir, img.Name())
	fd, err := os.Open(fmt.Sprintf("%s/%s", WednesdayAssetsDir, img.Name()))
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	mimeType, err := mimetype.DetectReader(fd)
	if err != nil {
		return nil, err
	}
	_, err = fd.Seek(0, 0)
	if err != nil {
		return nil, err
	}
	hexColor, _ := fastHappyColorInt64()
	e := dg_helpers.NewEmbed()
	e.SetTitle("It is Wednesday my dudes").
		SetColor(int(hexColor)).
		SetImage("attachment://" + img.Name()).
		Truncate()

	return &discordgo.MessageSend{
		Embed: e.MessageEmbed,
		Files: []*discordgo.File{
			{
				Name:        img.Name(),
				ContentType: mimeType.String(),
				Reader:      fd,
			},
		},
	}, nil
}
