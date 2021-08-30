package onmessage

import (
  "fmt"

  "github.com/bwmarrin/discordgo"
)

func fromSelf(s *discordgo.Session, m *discordgo.MessageCreate) bool {
  return m.Author.ID == s.State.User.ID
}

func Greeting(s *discordgo.Session, m *discordgo.MessageCreate) {
  if fromSelf(s, m) {
    return
  }

  if m.Content == fmt.Sprintf("Hello %s!", s.State.User.Username) {
    s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Hello %s!", m.Author.Username))
  }
}

func Farewell(s *discordgo.Session, m *discordgo.MessageCreate) {
  if fromSelf(s, m) {
    return
  }

  if m.Content == fmt.Sprintf("Goodbye %s!", s.State.User.Username) {
    s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Goodbye %s!", m.Author.Username))
  }
}