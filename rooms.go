package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var reservednames = []string{"🌻 Green Hill Zone", "🍄 Mushroom Kingdom", "🌳 Vegetable Valley", "👑 Hyrule", "🏰 Wily's Castle", "🌌 Final Destination"}
var takenlist []string
var admissionslist = make(map[string]string)
var viplist = make(map[string]string)
var passwordlist = make(map[string]string)
var owners = make(map[string]*discordgo.User)
var expirations = make(map[string]*time.Ticker)
var vacancy = make(map[string]uint8)

func CommandRooms(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Self Check
	u, err := discord.User("@me")
	if err != nil {
		fmt.Println("Could not get self identifiying user information.")
		return
	}

	if m.Author.ID == u.ID {
		return
	}

	// Split the command into an array and define them into concrete types
	argv := strings.Split(m.Content, " ")

	cmd := argv[0]
	t := argv[1]

	// Check if the parameter matches
	if cmd == "!rooms" {
		// Creation of a room
		switch t {

		case "create":
			if len(argv) != 5 {
				s.ChannelMessageSend(m.ChannelID, "Either too little or too many arguments given for reserving a room.")
				return
			}

			// Make sure a room there are rooms left to giveaway
			if len(takenlist) == 6 {
				s.ChannelMessageSend(m.ChannelID, "Whoops looks like we are all out of rooms to giveaway! Sorry for the inconvenience! :sweat:")
				return
			}

			// Make sure the author doesn't already own a room
			for _, usr := range owners {
				if usr.ID == m.Author.ID {
					s.ChannelMessageSend(m.ChannelID, "Due to anti-monopoly laws, we can't allow you to claim ownership of more than 1 room at a time. We may live in a capitalist society, but that doesn't mean we don't have rules!")
					return
				}
			}

			// Do the rest of the argv reassignments
			limit, err := strconv.Atoi(argv[2])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "The given User Limit parameter is either not a number or overflows a 32-bit integer value and cannot be handled safely.")
				return
			}

			if limit == 1 || limit <= 100 {
				s.ChannelMessageSend(m.ChannelID, "The given User Limit parameter is either too large or too small for Discord to handle.")
				return
			}

			bitrate, err := strconv.Atoi(argv[3])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "The given Bitrate parameter is either not a number or overflows a 32-bit integer value and cannot be handled safely.")
				return
			}

			if bitrate <= 31 || bitrate >= 97 {
				s.ChannelMessageSend(m.ChannelID, "The given Bitrate parameter is either too large or too small for Discord to handle.")
				return
			}

			pwd := argv[4]

			// If the password isn't none, then the user can't
			// create a room in public channels
			dm, err := discord.Channel(m.ChannelID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, ":rotating_light: An unexpected error has occurred, and we cannot complete your request as promised. :rotating_light:")
				return
			}

			if pwd != "none" {
				if len(dm.Recipients) != 1 {
					s.ChannelMessageSend(m.ChannelID, "Sorry, I can't handle private room creation in a public text channel. If you'd like to make a private room, please Direct Message me with the full details of your request.")
					return
				}
			}

			// Check if the password isn't already used
			// This is really dumb that I have to check this,
			// but because of how Discord works, there is
			// literally no way around this
			for _, p := range passwordlist {
				if pwd == p {
					s.ChannelMessageSend(m.ChannelID, "Looks like that password is already in use! Try not to tell anyone else...")
					return
				}
			}

			// Determine which room name gets taken
			rand.Seed(time.Now().UnixNano())
			max := len(reservednames)
			min := 0
			rng := rand.Intn(max-min) + min

			// Take the room and remove it from the reserved list
			room, err := s.GuildChannelCreate(bullysquad, reservednames[rng], discordgo.ChannelTypeGuildVoice)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, ":rotating_light: An unexpected error has occurred, and we cannot complete your request as promised. :rotating_light:")
				return
			}
			takenlist = append(takenlist, reservednames[rng])
			copy(reservednames[rng:], reservednames[(rng+1):])
			for k, n := len(reservednames)-(rng+1)+rng, len(reservednames); k < n; k++ {
				reservednames[k] = ""
			}
			reservednames = reservednames[:len(reservednames)-(rng+1)+rng]

			// Check if we should take, generate, or set no password at all
			perms := make([]*discordgo.PermissionOverwrite, 0, 0)

			// This also includes keeping track of important identication
			// FIXME: Remove admissionident and vipident for their
			// struct type counterparts
			admission := new(discordgo.Role)
			vip := new(discordgo.Role)

			if pwd == "none" {
				perms = nil
			} else {
				// Set the permissions bitset
				generalperms := new(discordgo.PermissionOverwrite)
				generalperms.ID = "67092563995136000"
				generalperms.Type = "role"
				generalperms.Allow = 0
				generalperms.Deny = 3146753

				perms = append(perms, generalperms)

				// Create a new vip role
				vip, err = s.GuildRoleCreate(bullysquad)
				if err != nil {
					s.GuildRoleDelete(bullysquad, vip.ID)
					s.ChannelDelete(room.ID)
					s.ChannelMessageSend(m.ChannelID, ":rotating_light: An unexpected error has occurred, and we cannot complete your request as promised. :rotating_light:")
					return
				}
				vip, err = s.GuildRoleEdit(bullysquad, vip.ID, "🎟️ VIP", 0, false, 0, false)
				if err != nil {
					s.GuildRoleDelete(bullysquad, vip.ID)
					s.ChannelDelete(room.ID)
					s.ChannelMessageSend(m.ChannelID, ":rotating_light: An unexpected error has occurred, and we cannot complete your request as promised. :rotating_light:")
					return
				}
				err = s.GuildMemberRoleAdd(bullysquad, m.Author.ID, vip.ID)
				if err != nil {
					s.GuildRoleDelete(bullysquad, vip.ID)
					s.ChannelDelete(room.ID)
					s.ChannelMessageSend(m.ChannelID, ":rotating_light: An unexpected error has occurred, and we cannot complete your request as promised. :rotating_light:")
					return
				}

				// Set the permission bitset for vip role
				vipperms := new(discordgo.PermissionOverwrite)
				vipperms.ID = vip.ID
				vipperms.Type = "role"
				vipperms.Allow = 32506880
				vipperms.Deny = 0

				perms = append(perms, vipperms)

				// Create a new normal role
				admission, err = s.GuildRoleCreate(bullysquad)
				if err != nil {
					s.GuildRoleDelete(bullysquad, vip.ID)
					s.GuildRoleDelete(bullysquad, admission.ID)
					s.ChannelDelete(room.ID)
					s.ChannelMessageSend(m.ChannelID, ":rotating_light: An unexpected error has occurred, and we cannot complete your request as promised. :rotating_light:")
					return
				}
				admission, err = s.GuildRoleEdit(bullysquad, admission.ID, "🎫 RSVP", 0, false, 0, false)
				if err != nil {
					s.GuildRoleDelete(bullysquad, vip.ID)
					s.GuildRoleDelete(bullysquad, admission.ID)
					s.ChannelDelete(room.ID)
					s.ChannelMessageSend(m.ChannelID, ":rotating_light: An unexpected error has occurred, and we cannot complete your request as promised. :rotating_light:")
					return
				}

				// Set the permission bitset for admission role
				admissionperms := new(discordgo.PermissionOverwrite)
				admissionperms.ID = admission.ID
				admissionperms.Type = "role"
				admissionperms.Allow = 3146752
				admissionperms.Deny = 0

				perms = append(perms, admissionperms)
			}

			// Edit the room with the correct parameters
			settings := new(discordgo.ChannelEdit)
			settings.Position = 1
			settings.ParentID = "437397061814714378"
			settings.Bitrate = bitrate * 1000
			settings.UserLimit = limit
			if perms != nil {
				settings.PermissionOverwrites = perms
			}

			room, err = s.ChannelEditComplex(room.ID, settings)
			if err != nil {
				s.GuildRoleDelete(bullysquad, vip.ID)
				s.GuildRoleDelete(bullysquad, admission.ID)
				s.ChannelDelete(room.ID)
				s.ChannelMessageSend(m.ChannelID, ":rotating_light: An unexpected error has occurred, and we cannot complete your request as promised. :rotating_light:")
				return
			}

			// If the room edit is sucessful, record IDs
			admissionslist[room.ID] = admission.ID
			viplist[room.ID] = vip.ID
			passwordlist[pwd] = room.ID
			owners[room.ID] = m.Author

			s.ChannelMessageSend(m.ChannelID, "Room reservation was successful, enjoy your stay!")

			// Start a new thread for checking inactivity
			ic := time.NewTicker(time.Second * 30)
			expirations[room.ID] = ic
			go InactivityCheck(room.ID)

		case "join":
			if len(argv) != 3 {
				s.ChannelMessageSend(m.ChannelID, "Either too little or too many arguments given for joining a room.")
				return
			}

			dm, err := discord.Channel(m.ChannelID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, ":rotating_light: An unexpected error has occurred, and we cannot complete your request as promised. :rotating_light:")
				return
			}

			pwd := argv[2]
			room := *new(string)

			// If the channel of the message has only one Recipients then
			// we can be sure that the Channel type is a Direct Message
			if len(dm.Recipients) == 1 {
				// Run through the list until we find a matching password
				for k, v := range passwordlist {
					if pwd == k {
						room = v
						break
					}
				}

				// Get access to the room we found a password for
				if room != "" {

					err = s.GuildMemberRoleAdd(bullysquad, m.Author.ID, admissionslist[room])
					if err != nil {
						s.ChannelMessageSend(m.ChannelID, ":rotating_light: An unexpected error has occurred, and we cannot complete your request as promised. :rotating_light:")
						return
					}

					s.ChannelMessageSend(dm.ID, "Looks like you are in! Welcome to the club!")
					return
				}

				// If the password didn't match with anything, notify them
				s.ChannelMessageSend(dm.ID, "The provided password was incorrect, sorry, looks like you aren't getting into Mile High Club today.")
			} else { // Otherwise let them know this is not the correct way
				s.ChannelMessageSend(dm.ID, "Sorry, but we can't let you in! For security reasons we only allow rooms with passwords to be joined through Direct Messages with me.")
				return
			}

		}
	}
}

func InactivityCheck(chanid string) {
	for {
		tick := expirations[chanid]
		for range tick.C {
			// Increment if the room is vacant
			room, _ := discord.Channel(chanid)

			if len(room.Recipients) == 0 {
				vacancy[chanid]++

				// Cleanup if the room has been vacant for too long
				if vacancy[chanid] == 10 {
					discord.GuildRoleDelete(bullysquad, admissionslist[chanid])
					delete(admissionslist, chanid)

					discord.GuildRoleDelete(bullysquad, viplist[chanid])
					delete(viplist, chanid)

					for pwd := range passwordlist {
						delete(passwordlist, pwd)
					}

					delete(owners, chanid)

					delete(expirations, chanid)

					delete(vacancy, chanid)

					for i, name := range takenlist {
						if room.Name == name {
							copy(takenlist[i:], takenlist[i+1:])
							takenlist[len(takenlist)-1] = ""
							takenlist = takenlist[:len(takenlist)-1]
							break
						}
					}

					reservednames = append(reservednames, room.Name)

					discord.ChannelDelete(room.ID)
					return // End the thread
				}
			} else {
				vacancy[chanid] = 0
				continue
			}
		}
	}
}
