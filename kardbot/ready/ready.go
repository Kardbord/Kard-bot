package ready

import (
  "github.com/bwmarrin/discordgo"
)

func OnReady(s *discordgo.Session, r *discordgo.Ready) {
  s.UpdateListeningStatus("you")
}
