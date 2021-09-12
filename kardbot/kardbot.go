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

// Stop and clean up. Not necessary to call if
// Run was called with block=true.
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

	json.Unmarshal(rawJSONConfig(), bot())
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
