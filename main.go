package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Starz0r/na1-go/strings/concat"
	"github.com/bwmarrin/discordgo"
)

const bullysquad = "67092563995136000"

var auth = concat.Builder("Bot ", "")
var discord, err = discordgo.New(auth)

func main() {
	// Initialize Commands
	discord.AddHandler(cmdRooms)

	// Open Websocket Connection
	discord.Open()

	// Run the program indefinately
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
