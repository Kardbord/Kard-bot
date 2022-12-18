package kardbot

import (
	"os"
	"time"
	_ "time/tzdata"

	log "github.com/sirupsen/logrus"

	"github.com/TannerKvarfordt/gopenai/authentication"
	"github.com/TannerKvarfordt/hfapigo"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

const (
	BotTokenEnv         = "KARDBOT_TOKEN"
	BotOwnerEnv         = "KARDBOT_OWNER_ID"
	TestbedGuildEnv     = "KARDBOT_TESTBED_GUILD"
	Intents             = discordgo.IntentsAll
	ImgflipUserEnv      = "IMGFLIP_API_USERNAME"
	ImgflipPassEnv      = "IMGFLIP_API_PASSWORD"
	HuggingFaceTokenEnv = "HUGGING_FACE_TOKEN"
	TimezoneEnv         = "TZ"
	OpenAITokenEnv      = "OPENAI_API_KEY"
)

var (
	getBotToken         = func() string { return "" }
	getOwnerID          = func() string { return "" }
	getTestbedGuild     = func() string { return "" }
	getImgflipUser      = func() string { return "" }
	getImgflipPass      = func() string { return "" }
	getHuggingFaceToken = func() string { return "" }
	getTimezone         = func() string { return "" }
	getOpenAIToken      = func() string { return "" }
)

// Retrieves the bot's auth token from the environment
func init() {
	// This will only add new environment variables,
	// and will NOT overwrite existing ones.
	_ = godotenv.Load( /*.env by default*/ )

	token, tokenFound := os.LookupEnv(BotTokenEnv)
	if !tokenFound {
		log.Fatalf("%s not found in environment", BotTokenEnv)
	} else if token == "" {
		log.Fatalf("%s is the empty string", BotTokenEnv)
	}
	getBotToken = func() string { return token }

	owner, ownerFound := os.LookupEnv(BotOwnerEnv)
	if !ownerFound {
		log.Warnf("%s not found in environment. No commands requiring this privilege can be executed.", BotOwnerEnv)
	} else if owner == "" {
		log.Warnf("%s is the empty string. No commands requiring this privilege can be executed.", BotOwnerEnv)
	}
	getOwnerID = func() string { return owner }

	testbed, testbedFound := os.LookupEnv(TestbedGuildEnv)
	if !testbedFound {
		log.Warnf("%s not found in environment", TestbedGuildEnv)
	} else if testbed == "" {
		log.Warnf("%s is the empty string. Commands will not be registered with a testbed guild.", TestbedGuildEnv)
	}
	getTestbedGuild = func() string { return testbed }

	imgflipUser, userFound := os.LookupEnv(ImgflipUserEnv)
	if !userFound {
		log.Warnf("%s not found in environment", ImgflipUserEnv)
	} else if imgflipUser == "" {
		log.Warnf("%s is the empty string. %s will not work", ImgflipUserEnv, memeCommand)
	}
	getImgflipUser = func() string { return imgflipUser }

	imgflipPass, passFound := os.LookupEnv(ImgflipPassEnv)
	if !passFound {
		log.Warnf("%s not found in environment", ImgflipPassEnv)
	} else if imgflipPass == "" {
		log.Warnf("%s is the empty string. %s will not work", ImgflipPassEnv, memeCommand)
	}
	getImgflipPass = func() string { return imgflipPass }

	hfToken, hfTokFound := os.LookupEnv(HuggingFaceTokenEnv)
	if !hfTokFound {
		log.Warnf("%s not found in environment", HuggingFaceTokenEnv)
	} else if hfToken == "" {
		log.Warnf("%s is the empty string.", HuggingFaceTokenEnv)
	}
	getHuggingFaceToken = func() string { return hfToken }
	hfapigo.SetAPIKey(getHuggingFaceToken())

	// TODO: Fix this properly. This has potential to cause race conditions.
	// 			 See https://github.com/TannerKvarfordt/Kard-bot/issues/57.
	tz, tzFound := os.LookupEnv(TimezoneEnv)
	if !tzFound || tz == "" {
		log.Warnf("%s not found in environment", TimezoneEnv)
	} else {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			log.Error(err)
		} else {
			time.Local = loc
		}
	}
	getTimezone = func() string { return time.Local.String() }
	log.Infof("Using timezone: %s", getTimezone())

	openAIToken, openAITokenFound := os.LookupEnv(OpenAITokenEnv)
	if !openAITokenFound {
		log.Warnf("%s not found in environment", OpenAITokenEnv)
	} else if hfToken == "" {
		log.Warnf("%s is the empty string.", OpenAITokenEnv)
	}
	getOpenAIToken = func() string { return openAIToken }
	authentication.SetAPIKey(getOpenAIToken())
}
