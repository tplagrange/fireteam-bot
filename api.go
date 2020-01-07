package main

import (
    "context"
    "fmt"
    "io/ioutil"
    "os"
    "net/http"
    "net/url"
    "strings"

    "github.com/gin-gonic/gin"
    // "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/bson"
)

// Handle the redirect URL from Bungie's OAUTH 2.0 Mechanism
func bungieCallback(c *gin.Context) {
    code := c.Query("code")
    // state := c.Query("state")

    // Now use the code to receive an access token
    client := &http.Client{}
    data := url.Values{}
    data.Set("grant_type", "authorization_code")
    data.Set("code", code)
    req, _ := http.NewRequest("POST", "https://www.bungie.net/platform/app/oauth/token/", strings.NewReader(data.Encode()))
    req.Header.Add("Authorization", "Basic " + os.Getenv("CLIENT_SECRET"))
    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    resp, _ := client.Do(req)

    defer resp.Body.Close()

    if resp.StatusCode == http.StatusOK {
        bodyBytes, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            fmt.Println(err)
        }
        bodyString := string(bodyBytes)
        fmt.Println(bodyString)
    }
    // Update database
    // collection := db.Database("bot").Collection("users")
}

// Redirect the discord user to bungie's OAUTH 2.0 Mechanism
func bungieAuth(c *gin.Context) {
    discordID := c.Param("id")

    bungieAuthURL := "https://www.bungie.net/en/OAuth/Authorize?client_id=" +
                     os.Getenv("CLIENT_ID") +
                     "&response_type=code" +
                     "&state=" + discordID

    // If yes, delete

    c.Redirect(http.StatusMovedPermanently, bungieAuthURL)
}

// Return a json object containing the guardian's loadout
func getLoadout(c *gin.Context) {
    discordID  := c.Param("id")
    filter     := bson.D{{ "DiscordID", discordID}}
    collection := db.Database("bot").Collection("users")

    var result User
    err := collection.FindOne(context.TODO(), filter).Decode(&result)
    if err != nil {
        fmt.Println(err)
    }
    if result.DiscordID == "" {
        c.String(403, "User does not exist")
    } else {
        fmt.Println("Got a user!")
    }
}
