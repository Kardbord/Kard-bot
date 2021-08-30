package kardbot

import (
  "log"
  "os"
  "os/signal"
  "syscall"

  "github.com/TannerKvarfordt/Kard-bot/kardbot/auth"
  "github.com/TannerKvarfordt/Kard-bot/kardbot/onmessage"
  "github.com/TannerKvarfordt/Kard-bot/kardbot/onready"
  "github.com/bwmarrin/discordgo"
)

type kardbot struct {
  session *discordgo.Session
}

func NewKardbot() kardbot {
  dg, err := discordgo.New("Bot " + auth.BotToken())
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

func (kbot *kardbot) Run(block bool) {
  kbot.session.Identify.Intents = auth.Intents()

  kbot.addHandlers()

  err := kbot.session.Open()
  log.Printf("Bot is now running. Press CTRL-C to exit.")
  if err != nil {
    log.Fatal("failed to open Discord session: ", err)
  }

  if block {
    sc := make(chan os.Signal, 1)
    signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
    <-sc
  }

  kbot.session.Close()
}

func (kbot *kardbot) addHandlers() {
  kbot.session.AddHandler(onready.OnReady)
  kbot.session.AddHandler(onmessage.OnCreate)
}