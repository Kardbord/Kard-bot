package onmessage

import (
  "github.com/bwmarrin/discordgo"
)

// Any callbacks that happen onMessageCreate go here
var mOnCreate = [...]func(*discordgo.Session, *discordgo.MessageCreate) {
  Greeting,
}

func OnCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
  for _, f := range mOnCreate {
    f(s, m)
  }
}