package main

import (
	"github.com/andersfylling/disgord"
)

const bullysquad = "67092563995136000"

var discord disgord.Session

func main() {
	err := *new(error)

	discord, err = disgord.NewSession(&disgord.Config{
		BotToken: "",
	})
	if err != nil {
		panic(err)
	}

	// Initialize Commands
	discord.On(disgord.EventMessageCreate, CheckIfProperlyTagged)

	// Open Websocket Connection
	err = discord.Connect()
	if err != nil {
		panic(err)
	}

	// Run the program indefinately
	discord.DisconnectOnInterrupt()
}
