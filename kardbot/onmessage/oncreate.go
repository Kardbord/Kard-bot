package onmessage

import (
  "fmt"

  "github.com/bwmarrin/discordgo"
)


func Greeting(s *discordgo.Session, m *discordgo.MessageCreate) {
  if m.Author.ID == s.State.User.ID {
    return
  }

  if m.Content == fmt.Sprintf("Hello %s!", s.State.User.Username) {
    s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Hello %s!", m.Author.Username))
  }
}