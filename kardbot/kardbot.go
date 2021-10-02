package kardbot

import (
	"encoding/json"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TannerKvarfordt/Kard-bot/kardbot/config"
	log "github.com/sirupsen/logrus"

	"github.com/bwmarrin/discordgo"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Pointer to the single kardbot instance global to the package
var gbot *kardbot = nil

type kardbot struct {
	Session   *discordgo.Session
	Greetings []string `json:"greetings"`
	Farewells []string `json:"farewells"`

	// Guilds with which to explicitly register slash commands.
	// Global commands take up to an hour (read, up to 24 hours)
	// to register. Specifying guilds allows slash commands to register
	// instantly.
	SlashGuilds []string `json:"slash-cmd-guilds"`

	EnableDGLogging bool `json:"enable-dg-logging"`
}

// bot() is a getter for the global kardbot instance
func bot() *kardbot {
	if gbot == nil {
		// The only time gbot should be nil is if Run() has not been called.
		// If Run() has not been called, no bot code should be running and
		// this case should not be hit.
		log.Fatal("No kardbot initialized. This should never happen.")
	}
	return gbot
}

// Run initializes and starts the bot.
func Run() {
	initialize()
	log.Print("Bot is now running. Press CTRL-C to exit.")
}

// Block the current goroutine until a terminatiing signal is received
func Block() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

// BlockThenStop blocks the current goroutine until a terminatiing signal is received.
// When the signal is received, stop and clean up the bot.
func BlockThenStop() {
	Block()
	Stop()
}

// Stop and clean up.
func Stop() {
	_ = bot().Session.Close()
}

// Initialize the single global bot instance
func initialize() {
	dgs, err := discordgo.New("Bot " + getBotToken())
	if err != nil {
		log.Fatal(err)
	}
	if dgs == nil {
		log.Fatal("failed to create discordgo session")
	}

	gbot = &kardbot{
		Session: dgs,
	}

	bot().configure()
	bot().addOnReadyHandlers()
	bot().prepInteractionHandlers()

	err = bot().Session.Open()
	if err != nil {
		log.Fatal(err)
	}
	bot().validateInitialization()

	bot().addOnCreateHandlers()
	bot().addInteractionHandlers(true)
}

func (kbot *kardbot) configure() {
	kbot.Session.Identify.Intents = Intents
	kbot.Session.SyncEvents = false
	kbot.Session.ShouldReconnectOnError = true
	kbot.Session.StateEnabled = true

	err := json.Unmarshal(config.RawJSONConfig(), kbot)
	if err != nil {
		log.Fatal(err)
	}
	if getTestbedGuild() != "" {
		kbot.SlashGuilds = append(kbot.SlashGuilds, getTestbedGuild())
	}

	if kbot.EnableDGLogging {
		kbot.Session.LogLevel = logrusToDiscordGo()[log.GetLevel()]
	}
}

func (kbot *kardbot) validateInitialization() {
	// Validate Session
	if kbot.Session == nil {
		log.Fatal("Session is nil.")
	}
	if kbot.Session.State == nil {
		log.Fatal("State is nil")
	}
	if kbot.Session.State.User == nil {
		log.Fatal("User is nil")
	}
	if !kbot.Session.ShouldReconnectOnError {
		log.Warn("discordgo session will not reconnect on error.")
	}
	if kbot.Session.SyncEvents {
		log.Warn("Session events are being executed synchronously, which may result in slower responses. Consider disabling this if command can safely be executed asynchronously.")
	}
	if kbot.Session.Identify.Intents == discordgo.IntentsNone {
		log.Warn("No intents registered with the discordgo API, which may result in decreased functionality.")
	}

	if kbot.Session.State.User.ID == getOwnerID() {
		// Owner has privilege to run any and all bot commands.
		// Giving a bot the reins to its own destiny is how you get Skynet in the long term.
		// In the short term, it might lead to weird edge cases I didn't think of, so I'm
		// disallowing it for now. :)
		log.Fatal("Bot is listed as its own owner")
	}

	// Validate SlashGuilds
	if len(kbot.SlashGuilds) == 0 {
		log.Warn("No guilds are configured to register slash commands with.")
	}

	// Validate greetings
	if kbot.greetingCount() == 0 {
		log.Fatal("No greetings configured.")
	}

	// Validate farewells
	if kbot.farewellCount() == 0 {
		log.Fatal("No farewells configured.")
	}
}

func (kbot *kardbot) prepInteractionHandlers() {
	kbot.Session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := getCommandImpls()[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func (kbot *kardbot) addInteractionHandlers(unregisterAllPrevCmds bool) {
	if unregisterAllPrevCmds {
		kbot.unregisterAllCommands()
	}

	for _, guildID := range kbot.SlashGuilds {
		if guildID == "" {
			log.Warn("Empty string specified as slash guild implies global command. Kard-bot does not support this at this time.")
			continue
		}
		for _, cmd := range getCommands() {
			_, err := kbot.Session.ApplicationCommandCreate(kbot.Session.State.User.ID, guildID, cmd)
			if err != nil {
				log.Fatalf("Cannot create '%v' command: %v", cmd.Name, err)
			}
		}
	}
}

func (kbot *kardbot) addOnReadyHandlers() {
	for _, h := range onReadyHandlers() {
		kbot.Session.AddHandler(h)
	}
}

func (kbot *kardbot) addOnCreateHandlers() {
	for _, h := range onCreateHandlers() {
		kbot.Session.AddHandler(h)
	}
}

func (kbot *kardbot) greetingCount() int {
	return len(kbot.Greetings)
}

func (kbot *kardbot) farewellCount() int {
	return len(kbot.Farewells)
}

func (kbot *kardbot) randomGreeting() string {
	return kbot.Greetings[rand.Intn(kbot.greetingCount())]
}

func (kbot *kardbot) randomFarewell() string {
	return kbot.Farewells[rand.Intn(kbot.farewellCount())]
}

func (kbot *kardbot) unregisterAllCommands() {
	// Unregister global commands
	if cmds, err := kbot.Session.ApplicationCommands(kbot.Session.State.User.ID, ""); err == nil {
		for _, cmd := range cmds {
			_ = kbot.Session.ApplicationCommandDelete(bot().Session.State.User.ID, "", cmd.ID)
		}
	} else {
		log.Warn(err)
	}

	// Unregister guild commands
	for _, guildID := range kbot.SlashGuilds {
		if cmds, err := kbot.Session.ApplicationCommands(kbot.Session.State.User.ID, guildID); err == nil {
			for _, cmd := range cmds {
				_ = kbot.Session.ApplicationCommandDelete(bot().Session.State.User.ID, guildID, cmd.ID)
			}
		} else {
			log.Warn(err)
		}
	}
}
