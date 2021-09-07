package kardbot

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

const (
	BotTokenEnv = "KARDBOT_TOKEN"
	BotOwnerEnv = "KARDBOT_OWNER_ID"
	Intents     = discordgo.IntentsAllWithoutPrivileged
)

var (
	mBotToken string
	mOwnerID  string
)

// Retrieves the bot's auth token from the environment
func init() {
	// This will only add new environment variables,
	// and will NOT overwrite existing ones.
	_ = godotenv.Load( /*.env by default*/ )

	botToken, tokenFound := os.LookupEnv(BotTokenEnv)
	if !tokenFound {
		log.Fatalf("%s not found in environment", BotTokenEnv)
	} else if botToken == "" {
		log.Fatalf("%s is the empty string", BotTokenEnv)
	}
	mBotToken = botToken

	ownerID, ownerFound := os.LookupEnv(BotOwnerEnv)
	if !ownerFound {
		log.Warnf("%s not found in environment. No commands requiring this privilege can be executed.", BotOwnerEnv)
	} else if ownerID == "" {
		log.Warnf("%s is the empty string. No commands requiring this privilege can be executed.", BotOwnerEnv)
	}
	mOwnerID = ownerID
}
