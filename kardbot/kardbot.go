package kardbot

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/bwmarrin/discordgo"
	"github.com/lus/dgc"
)

// Pointer to the single kardbot instance global to the package
var gbot *kardbot = nil

type kardbot struct {
	session *discordgo.Session
	router  *dgc.Router
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

// BlockAndStop blocks the current goroutine until a terminatiing signal is received.
// When the signal is received, stop and clean up the bot.
func BlockAndStop() {
	Block()
	Stop()
}

// Stop and clean up. Not necessary to call if
// Run was called with block=true.
func Stop() {
	_ = bot().session.Close()
}

// Initialize the single global bot instance
func initialize() {
	dgs, err := discordgo.New("Bot " + getBotToken())
	if err != nil {
		log.Fatal("discordgo error: ", err)
	}
	if dgs == nil {
		log.Fatal("failed to create discordgo session")
	}

	gbot = &kardbot{
		session: dgs,
		router: dgc.Create(&dgc.Router{
			// TODO: make these configurable
			Prefixes:         []string{"!"},
			IgnorePrefixCase: true,
			BotsAllowed:      false,
			Commands:         []*dgc.Command{},
			Middlewares:      []dgc.Middleware{},
			PingHandler:      func(ctx *dgc.Ctx) { ctx.RespondText("Pong!") },
		}),
	}

	configure()
	addHandlers()

	err = bot().session.Open()
	if err != nil {
		log.Fatal("failed to open Discord session: ", err)
	}
}

func configure() {
	bot().session.Identify.Intents = Intents
	bot().session.SyncEvents = false
	bot().session.ShouldReconnectOnError = true
	bot().session.StateEnabled = true
}

func addHandlers() {
	// Command handlers
	bot().router.RegisterDefaultHelpCommand(bot().session, nil)
	for _, cmd := range getCommands() {
		log.Debug("Registering cmd:", cmd.Name)
		bot().router.RegisterCmd(cmd)
	}
	bot().router.Initialize(bot().session)

	// OnReady handlers
	for _, h := range onReadyHandlers() {
		bot().session.AddHandler(h)
	}

	// OnMessageCreate handlers
	for _, h := range onCreateHandlers() {
		bot().session.AddHandler(h)
	}

	// Add handlers for any other event type here
}
