package main

import (
	"github.com/andersfylling/disgord"
)

func CheckIfProperlyTagged(session disgord.Session, msg *disgord.MessageCreate) {
	channel := msg.Message.ChannelID

	if channel.String() != "542672934368444426" {
		return
	}

	if msg.Message.SpoilerTagAllAttachments == false && len(msg.Message.Attachments) != 0 {
		session.DeleteMessage(channel, msg.Message.ID)
	}
}
