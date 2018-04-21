package main

import (
	"github.com/Starz0r/na1-go/strings/concat"
	"github.com/bwmarrin/discordgo"
)

var auth = concat.Builder("Bot ", "")
var discord, _ = discordgo.New(auth)

func main() {
	// Run the program indefinately
	<-make(chan struct{})
	return
}
