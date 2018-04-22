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
var admissionslist map[string]string
var viplist map[string]string

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

	// TODO: Check if the array is at least correctly sized before assignments
	// Split the command into an array and define them into concrete types
	argv := strings.Split(m.Content, " ")

	cmd := argv[0]
	t := argv[1]

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

	// Check if the parameter matches
	if cmd == "!rooms" {
		// Creation of a room
		if t == "create" {
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

			// If the room edit is sucessful, record IDs
			admissionslist[room.ID] = admissionident
			viplist[room.ID] = vipident
		}
	}

}
