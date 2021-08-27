package auth

import (
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

const BotTokenEnv = "DISCORD_BOT_TOKEN"
var (
	mBotToken string
	mIntents = map[discordgo.Intent]bool{
		// TODO: add intents
	}
)

// Retrieves the bot's auth token from the environment
func init() {
	// This will only add new environment variables,
	// and will NOT overwrite existing ones.
	_ = godotenv.Load(/*.env by default*/)

	var tokenFound bool
	mBotToken, tokenFound = os.LookupEnv(BotTokenEnv)

	if !tokenFound {
		log.Fatalf("%s not found in environment", BotTokenEnv)
	} else if mBotToken == "" {
		log.Fatalf("%s is the empty string", BotTokenEnv)
	}
}

// Returns the bot's auth token
func BotToken() string {
	return mBotToken
}

// Returns a list of all intents
func IntentList() []discordgo.Intent {
	intentList := make([]discordgo.Intent, len(mIntents))

	i := 0
	for intent := range mIntents {
		intentList[i] = intent
		i++
	}
	return intentList
}

// Bitwise ORs together all intents
func Intents() discordgo.Intent {
	intents := discordgo.IntentsNone
	for i := range mIntents {
		intents |= i
	}
	return intents
}

