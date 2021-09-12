package kardbot

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/bwmarrin/discordgo"
	"github.com/lus/dgc"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Pointer to the single kardbot instance global to the package
var gbot *kardbot = nil

type kardbot struct {
	Session   *discordgo.Session
	Router    *dgc.Router
	Greetings []string `json:"greetings"`
	Farewells []string `json:"farewells"`

	// TODO: add a subcommand to loglevel to set this
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
		Router: dgc.Create(&dgc.Router{
			Commands:    []*dgc.Command{},
			Middlewares: []dgc.Middleware{},
			PingHandler: func(ctx *dgc.Ctx) {
				ctx.RespondText(fmt.Sprintf("%s %s!", bot().randomGreeting(), ctx.Event.Author.Username))
			},
		}),
	}

	configure()
	validateConfig()
	addHandlers()

	err = bot().Session.Open()
	if err != nil {
		log.Fatal(err)
	}
}

func configure() {
	bot().Session.Identify.Intents = Intents
	bot().Session.SyncEvents = false
	bot().Session.ShouldReconnectOnError = true
	bot().Session.StateEnabled = true

	json.Unmarshal(RawJSONConfig(), bot())

	if bot().EnableDGLogging {
		bot().Session.LogLevel = logrusToDiscordGo()[log.GetLevel()]
	}
}

func validateConfig() {
	// Validate Session
	if bot().Session == nil {
		log.Fatal("Session is nil.")
	}
	if !bot().Session.ShouldReconnectOnError {
		log.Warn("discordgo session will not reconnect on error.")
	}
	if bot().Session.SyncEvents {
		log.Warn("Session events are being executed synchronously, which may result in slower responses. Consider disabling this if command can safely be executed asynchronously.")
	}
	if bot().Session.Identify.Intents == discordgo.IntentsNone {
		log.Warn("No intents registered with the discordgo API, which may result in decreased functionality.")
	}

	// Validate Router
	if bot().Router == nil {
		log.Fatal("Command router is nil.")
	}
	if bot().Router.Commands == nil {
		log.Fatal("Router.Commands is nil.")
	}
	if len(bot().Router.Prefixes) == 0 {
		log.Warn("No command prefixes registered. Commands will not work.")
	}
	if bot().Router.BotsAllowed {
		log.Warn("Command router allows other bots to issue commands.")
	}

	// Validate greetings
	if bot().greetingCount() == 0 {
		log.Fatal("No greetings configured.")
	}

	// Validate farewells
	if bot().farewellCount() == 0 {
		log.Fatal("No farewells configured.")
	}
}

func addHandlers() {
	// Command handlers
	bot().Router.RegisterDefaultHelpCommand(bot().Session, nil)
	for _, cmd := range getCommands() {
		log.Debug("Registering cmd:", cmd.Name)
		bot().Router.RegisterCmd(cmd)
	}
	bot().Router.Initialize(bot().Session)

	// OnReady handlers
	for _, h := range onReadyHandlers() {
		bot().Session.AddHandler(h)
	}

	// OnMessageCreate handlers
	for _, h := range onCreateHandlers() {
		bot().Session.AddHandler(h)
	}

	// Add handlers for any other event type here
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
