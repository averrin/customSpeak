package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"path"
)

func init() {
	flag.StringVar(&token, "t", "MzE5MjExMzU2MDk0MDcwNzg0.DA9oUQ.yoUVo1u8p-WlBW4z2_x2V3Jnv00", "Bot Token")
	flag.Parse()
}

var token string
var buffer = make([][]byte, 0)

func main() {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	// Register ready as a callback for the ready events.
	dg.AddHandler(ready)

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)

	// Register guildCreate as a callback for the guildCreate events.
	// dg.AddHandler(guildCreate)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("CustomSpeak is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {

	// Set the playing status.
	s.UpdateStatus(0, "!airhorn")
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// check if the message is "!airhorn"
	if strings.HasPrefix(m.Content, "!!cs") {

		// Find the channel that the message came from.
		c, err := s.State.Channel(m.ChannelID)
		if err != nil {
			// Could not find channel.
			return
		}

		// Find the guild for that channel.
		g, err := s.State.Guild(c.GuildID)
		if err != nil {
			// Could not find guild.
			return
		}

		// Look for the message sender in that guild's current voice states.
		for _, vs := range g.VoiceStates {
			if vs.UserID == m.Author.ID {
				vc, err := s.ChannelVoiceJoin(g.ID, vs.ChannelID, false, true)
				if err != nil {
					return
				}
				channel, _ := s.Channel(vs.ChannelID)
				vc.AddHandler(func(vc *discordgo.VoiceConnection, event *discordgo.VoiceSpeakingUpdate) {
					if vs.ChannelID != vc.ChannelID {
						return
					}
					u, _ := s.User(vs.UserID)
					fmt.Printf("[%v] %s speaks: %v\n", channel.Name, u.Username, event.Speaking)
					src := "on.png"
					cp := path.Join("custom", u.Username, src)
					if _, err := os.Stat(cp); err == nil {
						src = cp
					}
					if !event.Speaking {
						src = "off.png"
						cp = path.Join("custom", u.Username, src)
						if _, err := os.Stat(cp); err == nil {
							src = cp
						}
					}
					os.Mkdir(channel.Name, 0666)
					dest := path.Join(channel.Name, u.Username+".png")
					os.Remove(dest)
					os.Link(src, dest)
				})
				return
			}
		}
	}
}
