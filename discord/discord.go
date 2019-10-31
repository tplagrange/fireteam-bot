package discord

import (
    "fmt"
    "os"
    "os/signal"
    "strings"
    "syscall"

    "github.com/bwmarrin/discordgo"
)

func Bot() {
    token := os.Getenv("BOT_TOKEN")

    // Create a new Discord session using the provided bot token
    fmt.Print("Starting Discord Bot...")
    d, err := discordgo.New("Bot " + token)
    if err != nil {
        fmt.Println("Error creating Discord session: ", err)
		return
    }

    // Open the websocket and begin listening.
    err = d.Open()
    if err != nil {
        fmt.Println("Error opening Discord session: ", err)
    }

    // Register messageCreate as a callback for the messageCreate events.
    d.AddHandler(messageCreate)

    // Wait here until CTRL-C or other term signal is received.
    fmt.Println("Discord Bot Running.")
    sc := make(chan os.Signal, 1)
    signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
    <-sc

    // Cleanly close down the Discord session.
    d.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// check if the message is "-loadout"
	if strings.HasPrefix(m.Content, "-loadout") {

		// Find the channel that the message came from.
		_, err := s.State.Channel(m.ChannelID)
		if err != nil {
			// Could not find channel.
			return
		}

		s.MessageReactionAdd(m.ChannelID,m.ID, "👍")
	}
}
