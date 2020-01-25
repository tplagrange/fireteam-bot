package discord

import (
    "encoding/json"
    "fmt"
    "math/rand"
    "sort"
    "os"
    "os/signal"
    "strings"
    "syscall"
    "time"

    "github.com/bwmarrin/discordgo"
    golog "github.com/apsdehal/go-logger"
    "github.com/go-resty/resty/v2"
)

type Shader struct {
    Hash    string    `bson:"_id"`
    Name    string
    Icon    string
}

// Use a resty http client to make queries to the backend
// TODO: Replace this with the built in http client?
var rc *resty.Client
var log golog.Logger

func Bot() {
    log, _ := golog.New("bot", 1, os.Stdout)

    token := os.Getenv("BOT_TOKEN")

    // Create a new Discord session using the provided bot token
    log.Info("Starting Discord Bot...")
    d, err := discordgo.New("Bot " + token)
    if err != nil {
        fmt.Println("Error creating Discord session: ", err)
		return
    }

    // Open the websocket and begin listening.
    err = d.Open()
    if err != nil {
        log.Error("Error opening Discord session: " + err.Error())
        return
    }

    // Instantiate the REST client
    rc = resty.New()

    // Register messageCreate as a callback for the messageCreate events.
    d.AddHandler(messageCreate)

    // Wait here until CTRL-C o other term signal is received.
    log.Info("Discord Bot Running.")
    sc := make(chan os.Signal, 1)
    signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
    <-sc

    // Cleanly close down the Discord session.
    d.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
// Include logic here that will deal with user interactions, call functions
// to do things like http calls to the backend api.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}


    // Ensure fireteam bot is being referenced at all
    if (!strings.HasPrefix(m.Content, "-fb")) {
        return
    }

    user := m.Author.ID
    words := strings.Split(m.Content, " ")

    userChannel, err := s.UserChannelCreate(user)
    if ( err != nil ) {
        fmt.Println(err)
    }

    // Output help if not enough input
    if ( len(words) <= 1 ) {
        if err != nil {
            fmt.Println(err)
        }
        s.ChannelMessageSend(userChannel.ID, help())
        return
    }

	// check for "loadout" command
	if ( words[1] == "save" ) {
        if ( len(words) < 3) {
            s.ChannelMessageSend(userChannel.ID, help())
            return
        }
        name := words[2]
        code := saveLoadout(user, name)
        if ( code == 401 ) {
            s.ChannelMessageSend(userChannel.ID, "Hello, please register: http://" + os.Getenv("HOSTNAME") + "/api/bungie/auth/?id=" + user)
        } else if ( code == 300 ) {
            s.ChannelMessageSend(userChannel.ID, "User must select active membership")
        } else if ( code != 200 ) {
            s.ChannelMessageSend(userChannel.ID, "Error saving loadout: " + name)
        } else {
            s.ChannelMessageSend(userChannel.ID, "Saved loadout: " + name)
        }
    } else if ( words[1] == "load" ) {
        if ( len(words) < 3) {
            s.ChannelMessageSend(userChannel.ID, help())
            return
        }
        name := words[2]
        code := equipLoadout(user, name)
        if ( code == 401 ) {
            s.ChannelMessageSend(userChannel.ID, "Hello, please register: http://" + os.Getenv("HOSTNAME") + "/api/bungie/auth/?id=" + user)
        } else if ( code == 300 ) {
            s.ChannelMessageSend(userChannel.ID, "User must select active membership")
        } else if ( code != 200 ) {
            s.ChannelMessageSend(userChannel.ID, "Error saving loadout: " + name)
        } else {
            s.ChannelMessageSend(userChannel.ID, "Set loadout: " + name)
        }
    } else if ( words[1] == "shaders" ) {
        var response resty.Response
        getPartyShaders(user, &response)
        code := response.StatusCode()
        var shaders []Shader
        json.Unmarshal(response.Body(), &shaders)

        if ( code == 401 ) {
            s.ChannelMessageSend(userChannel.ID, "Hello, please register: http://" + os.Getenv("HOSTNAME") + "/api/bungie/auth/?id=" + user)
        } else if ( code == 300 ) {
            s.ChannelMessageSend(userChannel.ID, "User must select active membership")
        } else if ( code != 200 ) {
            s.ChannelMessageSend(userChannel.ID, "Error getting shaders")
        } else {

            if len(words) == 3 {

                randomizeShader(shaders, m.ChannelID, s)

            } else {
                shaderList := make([]string, 0)
                for _, s := range shaders {
                    shaderList = append(shaderList, s.Name)
                }

                sort.Strings(shaderList)

                embed := NewEmbed().
                    SetTitle("Fireteam Shaders").
                    SetDescription("This is a list of shaders that all fireteam members have collected.").
                    AddField("Shaders", strings.Join(shaderList[:], "\n")).
                    SetColor(0x00ff00).MessageEmbed

                s.ChannelMessageSendEmbed(m.ChannelID, embed)
            }


        }
    }
}

// func handleCode(int code, )

//     // Debug to acknowledge the message in discord
//     // s.MessageReactionAdd(m.ChannelID,m.ID, "👍")
//     // s.ChannelMessageSend(m.ChannelID, "Here's your loadout")
// }

func getShaders() {

}

func randomizeShader(shaders []Shader, channelID string, s *discordgo.Session) {
    rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
    rIndex := rand.Intn(len(shaders))
    shader := shaders[rIndex]

    embed := NewEmbed().
        SetTitle("Random Shader").
        SetDescription("🎲: Randomize\n👎: Blacklist (not implemented)").
        AddField("Shader", "**" + shader.Name + "**").
        SetImage("https://bungie.net" + shader.Icon).
        SetColor(0x00ff00).MessageEmbed
    msg, _ := s.ChannelMessageSendEmbed(channelID, embed)

    s.MessageReactionAdd(channelID, msg.ID, "🎲")
    s.MessageReactionAdd(channelID, msg.ID, "👎")

    c := make(chan bool, 1)
    go func() {
        m, _ := s.ChannelMessage(channelID, msg.ID)
        for {
            for i, reaction := range m.Reactions {
                if reaction.Count > 1 {
                    if i == 0 {
                        fmt.Println("Randomizing shader again")
                    } else {
                        fmt.Println("Blacklist not implemented")
                    }
                    c <- true
                    return
                }
            }
            time.Sleep(1 * time.Second)
            m, _ = s.ChannelMessage(channelID, msg.ID)
        }
    }()

    select {
    case <- c:
        if len(shaders) == 0 {
            log.Info("No more shaders to randomize")
            return
        }
        newShaders := append(shaders[:rIndex], shaders[rIndex+1:]...)
        s.ChannelMessageDelete(channelID, msg.ID)
        randomizeShader(newShaders, channelID, s)
    case <- time.After(5 * time.Minute):
        s.MessageReactionsRemoveAll(channelID, msg.ID)
        log.Info("Timeout on reaction")
    }
}

// Save a loadout for a user
func saveLoadout(user string, loadoutName string) int {
    // make api call to backend, request loadout for discord user id
    res, err := rc.R().EnableTrace().Get("http://localhost:" + os.Getenv("PORT") + "/api/loadout/" +
                "?id=" + user +
                "&name=" + loadoutName)
    if err != nil {
        fmt.Println(err)
    }

    return res.StatusCode()
}

// Equip a loadout for a user
func equipLoadout(user string, loadoutName string) int {
    res, err := rc.R().EnableTrace().Get("http://localhost:" + os.Getenv("PORT") + "/api/loadout/" +
                loadoutName + "/" +
                "?id=" + user)
    if err != nil {
        fmt.Println(err)
    }

    return res.StatusCode()
}

func getPartyShaders(user string, res *resty.Response) {
    response, err := rc.R().EnableTrace().Get("http://localhost:" + os.Getenv("PORT") +
            "/api/shaders/" +
            "?id=" + user)
    if err != nil {
        fmt.Println(err)
    }

    *res = *response

    return
}

// Outputs the how-to
func help() string {
    return "Please format commands to fireteam-bot as `-fb command parameter`"
}
