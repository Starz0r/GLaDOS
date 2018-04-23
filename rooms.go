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

func cmdRooms(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Self Check
	u, err := discord.User("@me")
	if err != nil {
		fmt.Println("Could not get self identifiying user information.")
		return
	}

	if m.Author.ID == u.ID {
		return
	}

	//TODO: Add a return to every end of execution to ensure there are no fallthroughs

	//TODO: Make sure the length of argv is 5
	// Split the command into an array and define them into concrete types
	argv := strings.Split(m.Content, " ")

	cmd := argv[0]
	t := argv[1]

	// Check if the parameter matches
	if cmd == "!rooms" {
		// Creation of a room
		if t == "create" {

			// Do the rest of the argv reassignments
			//TODO: Make sure check the limit is valid (2-99 or 0, can't be 1)
			limit, err := strconv.Atoi(argv[2])
			if err != nil {
				return
			}

			//TODO: Make sure bitrate is valid (32-96)
			bitrate, err := strconv.Atoi(argv[3])
			if err != nil {
				return
			}

			pwd := argv[4]

			// If the password isn't none, then the user can't
			// create a room in public channels
			dm, _ := discord.Channel(m.ChannelID)

			if pwd != "none" {
				if len(dm.Recipients) != 1 {
					s.ChannelMessageSend(m.ChannelID, "Sorry, I can't handle private room creation in a public text channel. If you'd like to make a private room, please Direct Message me with the full details of your request.")
					return
				}
			}

			// Determine which room name gets taken
			rand.Seed(time.Now().UnixNano())
			max := len(reservednames)
			min := 0
			rng := rand.Intn(max-min) + min

			// Take the room and remove it from the reserved list
			//TODO: Handle this error
			room, _ := s.GuildChannelCreate(bullysquad, reservednames[rng], discordgo.ChannelTypeGuildVoice)
			takenlist = append(takenlist, reservednames[rng])
			copy(reservednames[rng:], reservednames[(rng+1):])
			for k, n := len(reservednames)-(rng+1)+rng, len(reservednames); k < n; k++ {
				reservednames[k] = ""
			}
			reservednames = reservednames[:len(reservednames)-(rng+1)+rng]

			// Check if we should take, generate, or set no password at all
			perms := make([]*discordgo.PermissionOverwrite, 0, 0)

			// This also includes keeping track of important identication
			admissionident := *new(string)
			vipident := *new(string)

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
				vip, _ := s.GuildRoleCreate(bullysquad)
				vip, _ = s.GuildRoleEdit(bullysquad, vip.ID, "🎟️ Room Admission - VIP", 0, false, 0, false)
				s.GuildMemberRoleAdd(bullysquad, m.Author.ID, vip.ID)

				// Set the permission bitset for vip role
				vipperms := new(discordgo.PermissionOverwrite)
				vipperms.ID = vip.ID
				vipperms.Type = "role"
				vipperms.Allow = 32506880
				vipperms.Deny = 0

				perms = append(perms, vipperms)

				// Create a new normal role
				admission, _ := s.GuildRoleCreate(bullysquad)
				admission, _ = s.GuildRoleEdit(bullysquad, admission.ID, "🎫 Room Admission", 0, false, 0, false)

				// Set the permission bitset for admission role
				admissionperms := new(discordgo.PermissionOverwrite)
				admissionperms.ID = admission.ID
				admissionperms.Type = "role"
				admissionperms.Allow = 3146752
				admissionperms.Deny = 0

				perms = append(perms, admissionperms)

				// Bookkeep important identification
				admissionident = admission.ID
				vipident = vip.ID
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
			//TODO: Handle this error by cleaning up if it fails
			s.ChannelEditComplex(room.ID, settings)

			//TODO: Stop users from picking the same password already in passwordlist

			// If the room edit is sucessful, record IDs
			admissionslist[room.ID] = admissionident
			viplist[room.ID] = vipident
			passwordlist[pwd] = room.ID

			s.ChannelMessageSend(m.ChannelID, "Room reservation was successful, enjoy your stay!")
		} else if t == "join" {
			//TODO: Handle this error
			dm, _ := discord.Channel(m.ChannelID)

			//TODO: Check if argv length is at least 3 here
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
					s.GuildMemberRoleAdd(bullysquad, m.Author.ID, admissionslist[room])
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
