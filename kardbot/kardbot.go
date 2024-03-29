package kardbot

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"runtime/debug"
	"runtime/trace"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/Kardbord/Kard-bot/kardbot/config"
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
	Session        *discordgo.Session
	dgLoggingMutex sync.Mutex

	EnableDGLogging            bool `json:"enable-dg-logging"`
	UnregisterAllCmdsOnStartup bool `json:"unregister-all-cmds-on-startup"`

	// Number of times a server clock can fail to update before
	// it is abandoned by the bot.
	ServerClockFailureThreshold uint32 `json:"server-clock-failure-threshold"`

	// Enable trace regions for profiling
	TraceEnabled bool

	// Initialized in kardbot.Run, used to determine when
	// kardbot.Stop has been called.
	wg *sync.WaitGroup

	lastActive atomic.Time
	status     atomic.String
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
// Setting traceEnabled to true enables
// trace regions for profiling.
func Run(traceEnabled bool) {
	if gbot != nil {
		log.Warn("Run has already been called")
		return
	}

	if traceEnabled {
		log.Info("Bot will run with trace regions enabled")
	} else {
		log.Info("Bot will run with trace regions disabled (this is normal)")
	}

	log.RegisterExitHandler(Stop)

	wg := sync.WaitGroup{}
	wg.Add(1)
	gbot = &kardbot{
		wg:           &wg,
		TraceEnabled: traceEnabled,
	}
	go handleInterrupt()
	gbot.initialize()
	log.Print("Bot is now running. Press CTRL-C to exit.")
}

func Block() {
	bot().wg.Wait()
}

func RunAndBlock(traceEnabled bool) {
	Run(traceEnabled)
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
	if err := writeCreepyDmSubscribersToDisk(); err != nil {
		log.Error(err)
	}

	if err := writeComplimentSubscribersToDisk(); err != nil {
		log.Error(err)
	}

	if err := writeServerClocksToDisk(); err != nil {
		log.Error(err)
	}

	if err := purgeFinishedPolls(); err != nil {
		log.Error(err)
	}

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
	kbot.addInteractionHandlers(kbot.UnregisterAllCmdsOnStartup)
	log.Info("Interaction handlers registered")
}

const kardbotConfigFile = "config/setup.json"

func (kbot *kardbot) configure() {
	kbot.Session.Identify.Intents = Intents
	kbot.Session.SyncEvents = false
	kbot.Session.ShouldReconnectOnError = true
	kbot.Session.StateEnabled = true

	jsonCfg, err := config.NewJsonConfig(kardbotConfigFile)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(jsonCfg.Raw, kbot)
	if err != nil {
		log.Fatal(err)
	}

	if kbot.EnableDGLogging {
		kbot.dgLoggingMutex.Lock()
		kbot.Session.LogLevel = logrusToDiscordGo()[log.GetLevel()]
		kbot.dgLoggingMutex.Unlock()
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
		log.Warn("Session events are being executed synchronously, which may result in slower responses. Consider disabling this if commands can safely be executed asynchronously.")
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
}

func (kbot *kardbot) prepInteractionHandlers() {
	kbot.Session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		defer func() {
			// Attempt to quit gracefully in the event of a panic.
			// Won't do any good if the panic came from another goroutine.
			if r := recover(); r != nil {
				if panicErr, ok := r.(error); ok && i.Type == discordgo.InteractionApplicationCommand {
					interactionRespondEphemeralError(s, i, true, panicErr)
				}
				log.Fatalf("Panicked!\n%v\nStack Trace:%s\n", r, debug.Stack())
			}
		}()

		var handler func(*discordgo.Session, *discordgo.InteractionCreate) = nil
		command := "unknown interaction type"
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			command = i.ApplicationCommandData().Name
			if strMatchesMemeCmdPattern(command) {
				command = memeCommand
			}
			if h, ok := getCommandImpls()[command]; ok {
				handler = h
			}
		case discordgo.InteractionMessageComponent:
			command = i.MessageComponentData().CustomID
			if roleSelectMenuIDRegex.MatchString(command) {
				command = roleSelectMenuComponentIDPrefix
			}
			if h, ok := getComponentImpls()[command]; ok {
				handler = h
			}
		}

		if handler == nil {
			err := fmt.Errorf("interaction failed: %s", command)
			log.Error(err)
			interactionRespondEphemeralError(s, i, true, err)
			return
		}

		wg := kbot.updateLastActive()
		defer wg.Wait()

		if kbot.TraceEnabled {
			ctx, task := trace.NewTask(context.Background(), command)
			defer task.End()
			r := trace.StartRegion(ctx, command)
			handler(s, i)
			r.End()
		} else {
			handler(s, i)
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
	if !validateCmdRegex() {
		log.Fatal("One or more commands is invalid.")
	}

	if unregisterAllPrevCmds {
		kbot.unregisterAllCommands()
		kbot.bulkOverwriteAllCommands()
		return
	}

	kbot.unregisterOldCommands()
	kbot.bulkOverwriteAllCommands()
}

func (kbot *kardbot) unregisterOldCommands() {
	kbot.unregisterOldGlobalCommands()
	kbot.unregisterOldGuildCommands()
}

func (kbot *kardbot) unregisterOldGlobalCommands() {
	cmds, err := kbot.Session.ApplicationCommands(kbot.Session.State.User.ID, "")
	if err != nil {
		log.Fatal(err)
	}
	kbot.unregisterOldCommandsFromList(cmds, "")
}

func (kbot *kardbot) unregisterOldGuildCommands() {
	guilds, err := kbot.GetAllGuilds()
	if err != nil {
		log.Fatal(err)
	}

	for _, g := range guilds {
		if g.ID == "" {
			log.Warn("Empty string specified as slash guild implies global command, skipping.")
			continue
		}
		cmds, err := kbot.Session.ApplicationCommands(kbot.Session.State.User.ID, g.ID)
		if err != nil {
			log.Fatal(err)
		}

		kbot.unregisterOldCommandsFromList(cmds, g.ID)
	}
}

func (kbot *kardbot) unregisterOldCommandsFromList(cmds []*discordgo.ApplicationCommand, guildID string) {
	for _, cmd := range cmds {
		cmdname := cmd.Name
		if strMatchesMemeCmdPattern(cmdname) {
			cmdname = memeCommand
		}
		if _, ok := getCommandImpls()[cmdname]; !ok {
			if guildID == "" {
				log.Infof("Unregistering global command %s", cmd.Name)
			} else {
				log.Infof("Unregistering cmd '%s' from guild %s", cmd.Name, guildID)
			}
			err := kbot.Session.ApplicationCommandDelete(kbot.Session.State.User.ID, guildID, cmd.ID)
			if err != nil {
				log.Error(err)
			}
		}
	}
}

func (kbot *kardbot) bulkOverwriteGlobalCommands() {
	log.Infof("Bulk overwriting global commands")
	_, err := kbot.Session.ApplicationCommandBulkOverwrite(kbot.Session.State.User.ID, "", getCommands())
	if err != nil {
		log.Fatal(err)
	}
}

func (kbot *kardbot) unregisterAllGlobalCommands() {
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

func (kbot *kardbot) bulkOverwriteTestGuildcommands() {
	if getTestbedGuild() == "" {
		log.Info("No testbed guild specified")
		return
	}
	log.Info("Bulk overwriting testbed guild commands")
	_, err := kbot.Session.ApplicationCommandBulkOverwrite(kbot.Session.State.User.ID, getTestbedGuild(), getCommands())
	if err != nil {
		log.Fatalf("Failed to register commands in guild %s: %v", getTestbedGuild(), err)
	}
}

func (kbot *kardbot) bulkOverwriteAllCommands() {
	log.Info("Registering all commands for current bot instance")
	kbot.bulkOverwriteGlobalCommands()
	kbot.bulkOverwriteTestGuildcommands()
}

func (kbot *kardbot) unregisterAllGuildCommands() {
	guilds, err := kbot.GetAllGuilds()
	if err != nil {
		log.Fatal(err)
	}

	for _, g := range guilds {
		if g.ID == "" {
			log.Warn("Empty string specified as slash guild implies global command, skipping.")
			continue
		}
		if cmds, err := kbot.Session.ApplicationCommands(kbot.Session.State.User.ID, g.ID); err == nil {
			for _, cmd := range cmds {
				log.Infof("Unregistering cmd '%s' in guild %s", cmd.Name, g.Name)
				err = kbot.Session.ApplicationCommandDelete(kbot.Session.State.User.ID, g.ID, cmd.ID)
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
	kbot.unregisterAllGlobalCommands()
	kbot.unregisterAllGuildCommands()
}

func (kbot *kardbot) GetAllGuilds() ([]*discordgo.UserGuild, error) {
	botGuilds := []*discordgo.UserGuild{}

	rGuilds, err := kbot.Session.UserGuilds(100, "", "")
	if err != nil {
		return nil, err
	}
	botGuilds = append(botGuilds, rGuilds...)

	for len(rGuilds) > 0 {
		rGuilds, err = kbot.Session.UserGuilds(100, "", rGuilds[len(rGuilds)-1].ID)
		if err != nil {
			return nil, err
		}
		botGuilds = append(botGuilds, rGuilds...)
	}

	return botGuilds, nil
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

		err = kbot.Session.UpdateStatusComplex(discordgo.UpdateStatusData{
			AFK:    false,
			Status: string(discordgo.StatusOnline),
		})
		if err != nil {
			log.Error(err)
		} else {
			kbot.status.Store(string(discordgo.StatusOnline))
			log.Tracef("Set bot status to %s", kbot.status.Load())
		}

		err = kbot.Session.UpdateListeningStatus("you")
		if err != nil {
			log.Error(err)
		}

		wg.Done()
	}()

	return &wg
}
