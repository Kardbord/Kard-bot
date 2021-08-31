package kardbot

import (
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

const (
	BotTokenEnv = "DISCORD_BOT_TOKEN"
	Intents     = discordgo.IntentsAllWithoutPrivileged
)

var (
	mBotToken string
)

// Retrieves the bot's auth token from the environment
func init() {
	// This will only add new environment variables,
	// and will NOT overwrite existing ones.
	_ = godotenv.Load( /*.env by default*/ )

	var tokenFound bool
	mBotToken, tokenFound = os.LookupEnv(BotTokenEnv)

	if !tokenFound {
		log.Fatalf("%s not found in environment", BotTokenEnv)
	} else if mBotToken == "" {
		log.Fatalf("%s is the empty string", BotTokenEnv)
	}
}
