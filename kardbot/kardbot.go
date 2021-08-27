package kardbot

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/TannerKvarfordt/Kard-bot/kardbot/auth"
	"github.com/bwmarrin/discordgo"
)

type kardbot struct {
	dgSession *discordgo.Session
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
		dgSession: dg,
	}
}

func (kbot *kardbot) Run() {
	kbot.dgSession.Identify.Intents = auth.Intents()
	err := kbot.dgSession.Open()
	if err != nil {
		log.Fatal("failed to open Discord session: ", err)
	}

	log.Println("Kardbot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	kbot.dgSession.Close()
}
