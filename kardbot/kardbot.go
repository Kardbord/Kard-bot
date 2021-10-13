package kardbot

import (
	"encoding/json"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/TannerKvarfordt/Kard-bot/kardbot/config"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/atomic"
)

const AssetsDir string = "./assets"

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

	// Initialized in kardbot.Run, used to determine when
	// kardbot.Stop has been called.
	wg *sync.WaitGroup

	lastActive atomic.Time
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
// Returns a WaitGroup
func Run() {
	if gbot != nil {
		log.Warn("Run has already been called")
		return
	}

	log.RegisterExitHandler(Stop)

	wg := sync.WaitGroup{}
	wg.Add(1)
	gbot = &kardbot{wg: &wg}
	go handleInterrupt()
	gbot.initialize()
	log.Print("Bot is now running. Press CTRL-C to exit.")
}

func Block() {
	bot().wg.Wait()
}

func RunAndBlock() {
	Run()
	Block()
}

// listenForInterrupt blocks the current goroutine until a terminatiing signal is received.
// When the signal is received, stop and clean up the bot.
func handleInterrupt() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	Stop()
	os.Exit(1)
}

// Stop and clean up.
func Stop() {
	if gbot == nil {
		log.Info("Bot is not running")
		return
	}

	defer func() {
		if gbot.wg == nil {
			log.Fatal("nil waitgroup, there is a bug. :(")
		}
		gbot.wg.Done()
	}()

	if gbot.Session == nil {
		log.Info("No session to close")
		return
	}

	log.Info("Closing session")
	err := gbot.Session.Close()
	if err != nil {
		log.Errorf("Session closed with error: %v", err)
	} else {
		log.Info("Session closed")
	}
}

// Initialize the single global bot instance
func (kbot *kardbot) initialize() {
	dgs, err := discordgo.New("Bot " + getBotToken())
	if err != nil {
		log.Fatal(err)
	}
	if dgs == nil {
		log.Fatal("failed to create discordgo session")
	}
	log.Info("Session created")
	kbot.Session = dgs

	kbot.configure()
	log.Info("Configuration read")
	kbot.addOnReadyHandlers()
	log.Info("OnReady handlers registered")
	kbot.prepInteractionHandlers()
	log.Info("Interaction handlers prepared")

	err = kbot.Session.Open()
	if err != nil {
		log.Fatal(err)
	}
	kbot.lastActive = *atomic.NewTime(time.Now())
	scheduler().StartAsync()
	kbot.validateInitialization()
	log.Info("Configuration validated")

	kbot.addOnCreateHandlers()
	log.Info("OnCreate handlers registered")
	kbot.addInteractionHandlers(false)
	log.Info("Interaction handlers registered")
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

func (kbot *kardbot) addInteractionHandlers(unregisterAllPrevCmds bool) {
	if unregisterAllPrevCmds {
		kbot.unregisterAllCommands()
		kbot.registerAllCommands()
		return
	}

	kbot.unregisterOldCommands()
	kbot.bulkOverwriteCommands()
}

func (kbot *kardbot) bulkOverwriteGlobalCommands() {
	// nothing to do, no global commands supported
}

func (kbot *kardbot) bulkOverwriteGuildCommands() {
	for _, guildID := range kbot.SlashGuilds {
		if guildID == "" {
			log.Warn("Empty string specified as slash guild implies global command, skipping.")
			continue
		}
		if guild, err := kbot.Session.Guild(guildID); err != nil {
			log.Warn(err)
		} else {
			log.Infof("Bulk overwriting commands for guild: %s", guild.Name)
		}
		kbot.Session.ApplicationCommandBulkOverwrite(kbot.Session.State.User.ID, guildID, getCommands())
	}
}

func (kbot *kardbot) bulkOverwriteCommands() {
	kbot.bulkOverwriteGlobalCommands()
	kbot.bulkOverwriteGuildCommands()
}

func (kbot *kardbot) unregisterOldGlobalCommands() {
	// No global commands supported at this time,
	// so unregister all of them.
	kbot.unregisterGlobalCommands()
}

func (kbot *kardbot) unregisterOldGuildCommands() {
	for _, guildID := range kbot.SlashGuilds {
		if guildID == "" {
			log.Warn("Empty string specified as slash guild implies global command, skipping.")
			continue
		}
		if cmds, err := kbot.Session.ApplicationCommands(kbot.Session.State.User.ID, guildID); err == nil {
			for _, cmd := range cmds {
				if _, ok := getCommandImpls()[cmd.Name]; !ok {
					log.Infof("Unregistering cmd '%s' in guild %s", cmd.Name, guildID)
					err = kbot.Session.ApplicationCommandDelete(bot().Session.State.User.ID, guildID, cmd.ID)
					if err != nil {
						log.Error(err)
					}
				}
			}
		} else {
			log.Warn(err)
		}
	}
}

func (kbot *kardbot) unregisterOldCommands() {
	kbot.unregisterOldGlobalCommands()
	kbot.unregisterOldGuildCommands()
}

func (kbot *kardbot) unregisterGlobalCommands() {
	// Unregister global commands
	if cmds, err := kbot.Session.ApplicationCommands(kbot.Session.State.User.ID, ""); err == nil {
		for _, cmd := range cmds {
			log.Infof("Unregistering global command: %s", cmd.Name)
			err = kbot.Session.ApplicationCommandDelete(bot().Session.State.User.ID, "", cmd.ID)
			if err != nil {
				log.Error(err)
			}
		}
	} else {
		log.Warn(err)
	}
}

func (kbot *kardbot) registerGlobalCommands() {
	// No global commands at this time
	log.Info("No global commands to register")
}

func (kbot *kardbot) registerGuildcommands() {
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
			log.Infof("Registered %s", cmd.Name)
		}
	}
}

func (kbot *kardbot) registerAllCommands() {
	log.Info("Registering all commands for current bot instance")
	kbot.registerGlobalCommands()
	kbot.registerGuildcommands()
}

func (kbot *kardbot) unregisterAllGuildCommands() {
	// Unregister guild commands
	for _, guildID := range kbot.SlashGuilds {
		if guildID == "" {
			log.Warn("Empty string specified as slash guild implies global command, skipping.")
			continue
		}
		if cmds, err := kbot.Session.ApplicationCommands(kbot.Session.State.User.ID, guildID); err == nil {
			for _, cmd := range cmds {
				log.Infof("Unregistering cmd '%s' in guild %s", cmd.Name, guildID)
				err = kbot.Session.ApplicationCommandDelete(kbot.Session.State.User.ID, guildID, cmd.ID)
				if err != nil {
					log.Error(err)
				}
			}
		} else {
			log.Warn(err)
		}
	}
}

func (kbot *kardbot) unregisterAllCommands() {
	log.Info("Unregistering all commands from previous bot instance")
	kbot.unregisterGlobalCommands()
	kbot.unregisterAllGuildCommands()
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

// updateLastActive creates a go routine which sets the
// lastActive field to time.Now() and ensures the bot shows
// as active. It returns a pointer to a WaitGroup so that callers can
// wait to ensure this routine finished.
func (kbot *kardbot) updateLastActive() *sync.WaitGroup {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		kbot.lastActive.Store(time.Now())

		err := kbot.Session.UpdateListeningStatus("you")
		if err != nil {
			log.Error(err)
		}

		_, err = kbot.Session.UserUpdateStatus(discordgo.StatusOnline)
		if err != nil {
			log.Error(err)
		}

		wg.Done()
	}()

	return &wg
}
