package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"path"
	"sync"
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
	s.UpdateStatus(0, "!!cs")
}

var userStates map[string]bool
var userHist map[string]bool
var users map[string]string
var channel *discordgo.Channel

type Event struct {
	Username string
	Speak    bool
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	userStates = map[string]bool{}
	userHist = map[string]bool{}
	users = map[string]string{}
	var mutex = &sync.Mutex{}

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
		fmt.Println("Fetching users")
		for _, m := range g.Members {
			// u, _ := s.User(vs.UserID)
			users[m.User.ID] = m.User.Username
			fmt.Println(m.User.ID, "  ", m.User.Username)
		}
		for _, vs := range g.VoiceStates {
			if vs.UserID == m.Author.ID {
				vc, err := s.ChannelVoiceJoin(g.ID, vs.ChannelID, false, true)
				if err != nil {
					return
				}
				channel, _ = s.Channel(vs.ChannelID)
				ch := make(chan Event, 50)
				os.Mkdir(channel.Name, 0666)
				vc.AddHandler(func(conn *discordgo.VoiceConnection, event *discordgo.VoiceSpeakingUpdate) {
					if vs.ChannelID != vc.ChannelID {
						return
					}
					ch <- Event{event.UserID, event.Speaking}
				})

				go func(ch chan Event) {
					for {
						e := <-ch
						mutex.Lock()
						userStates[e.Username] = e.Speak
						mutex.Unlock()
					}
				}(ch)
				go func() {
					for {
						mutex.Lock()
						states := map[string]bool{}
						for k, v := range userStates {
							states[k] = v
						}
						mutex.Unlock()
						for u, speak := range states {
							if userHist[u] == speak {
								continue
							}
							userHist[u] = speak

							u = users[u]
							fmt.Printf("[%v] %s speaks: %v\n", time.Now(), u, speak)
							var cp string
							var src string
							if speak {
								src = "on.png"
								cp = path.Join("custom", u, src)
								if _, err := os.Stat(cp); err == nil {
									src = cp
								}
							} else {
								src = "off.png"
								cp = path.Join("custom", u, src)
								if _, err := os.Stat(cp); err == nil {
									src = cp
								}
							}
							dest := path.Join(channel.Name, u+".png")
							os.Remove(dest)
							os.Link(src, dest)
						}
						// time.Sleep(200)
					}
				}()
			}
		}
		fmt.Println("Ready")
	}
}
