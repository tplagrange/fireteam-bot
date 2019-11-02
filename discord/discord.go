package discord

import (
    "fmt"
    "os"
    "os/signal"
    "strings"
    "syscall"

    "github.com/bwmarrin/discordgo"
    "github.com/go-resty/resty/v2"
)

// Use a resty http client to make queries to the backend
var rc *resty.Client

func Bot() {
    token := os.Getenv("BOT_TOKEN")

    // Create a new Discord session using the provided bot token
    fmt.Println("Starting Discord Bot...")
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

    // Instantiate the REST client
    rc = resty.New()

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

        user := m.Author.ID

        // Debug to acknowledge the message in discord
		s.MessageReactionAdd(m.ChannelID,m.ID, "👍")
        // s.ChannelMessageSend(m.ChannelID, "Here's your loadout")

        // make api call to backend, request loadout for discord user id
        res, err := rc.R().EnableTrace().Get("http://localhost:" + os.Getenv("PORT") + "/api/loadout/" + user)
        if err != nil {
            fmt.Println(err)
        }

        if res.StatusCode() == 403 {
            userChannel, err := s.UserChannelCreate(user)
            if err != nil {
                fmt.Println(err)
            }

            // User must authenticate with bungie
            // Direct message the user with a registration link
            s.ChannelMessageSend(userChannel.ID, "[Hello, " + user.Username + ". Please register](https://www.bungie.net/en/OAuth/Authorize?client_id=30633&response_type=code)")
        }

    }
}
