package kardbot

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/bwmarrin/discordgo"
	"github.com/lus/dgc"
)

type kardbot struct {
	session *discordgo.Session
	router  *dgc.Router
}

// NewKardbot retuns a new bot instance.
func NewKardbot() kardbot {
	dg, err := discordgo.New("Bot " + mBotToken)
	if err != nil {
		log.Fatal("discordgo error: ", err)
	}
	if dg == nil {
		log.Fatal("failed to create discordgo session")
	}
	return kardbot{
		session: dg,
	}
}

// Start the bot.
//
// The "block" argument indicates whether or
// not this call should block the current
// goroutine until a terminating signal is
// received. If set to true, Stop will be
// called on the bot object when the signal
// is received.
func (kbot *kardbot) Run(block bool) {
	kbot.configure()
	kbot.addHandlers()

	err := kbot.session.Open()
	log.Print("Bot is now running. Press CTRL-C to exit.")
	if err != nil {
		log.Fatal("failed to open Discord session: ", err)
	}
	defer func() {
		if block {
			sc := make(chan os.Signal, 1)
			signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
			<-sc
			kbot.Stop()
		}
	}()

	kbot.registerCommands()
}

// Stop and clean up. Not necessary to call if
// Run was called with block=true.
func (kbot *kardbot) Stop() {
	_ = kbot.session.Close()
}

func (kbot *kardbot) configure() {
	kbot.session.Identify.Intents = Intents
	kbot.session.SyncEvents = false
	kbot.session.ShouldReconnectOnError = true
	kbot.session.StateEnabled = true
}

func (kbot *kardbot) addHandlers() {

	// OnReady handlers
	for _, h := range onReadyHandlers {
		kbot.session.AddHandler(h)
	}

	// OnMessageCreate handlers
	for _, h := range onCreateHandlers {
		kbot.session.AddHandler(h)
	}

	// Add handlers for any other event type here
}

func (kbot *kardbot) registerCommands() {
	kbot.router = dgc.Create(&dgc.Router{
		// TODO: make these configurable
		Prefixes:         []string{"!"},
		IgnorePrefixCase: true,
		BotsAllowed:      false,
		Commands:         []*dgc.Command{},
		Middlewares:      []dgc.Middleware{},
		PingHandler:      func(ctx *dgc.Ctx) { ctx.RespondText("Pong!") },
	})

	kbot.router.RegisterDefaultHelpCommand(kbot.session, nil)

	for _, cmd := range mCommands {
		log.Debug("Registering cmd:", cmd.Name)
		kbot.router.RegisterCmd(cmd)
	}

	kbot.router.Initialize(kbot.session)

}
