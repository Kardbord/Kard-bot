package kardbot

import (
  "fmt"
  "log"
  "os"
  "os/signal"
  "syscall"

  "github.com/TannerKvarfordt/Kard-bot/kardbot/auth"
  "github.com/TannerKvarfordt/Kard-bot/kardbot/ready"
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

  kbot.session.AddHandler(ready.OnReady)
  kbot.session.AddHandler(sayHello)

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

func sayHello(s *discordgo.Session, m *discordgo.MessageCreate) {
  if m.Author.ID == s.State.User.ID {
    return
  }

  if m.Content == fmt.Sprintf("Hello %s!", s.State.User.Username) {
    s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Hello %s!", m.Author.Username))
  }
}
