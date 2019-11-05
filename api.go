package main

import (
    "context"
    "fmt"
    "os"
    "net/http"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
)

// Handle the redirect URL from Bungie's OAUTH 2.0 Mechanism
func bungieCallback(c *gin.Context) {
    code := c.Param("code")
    fmt.Println(code)
}

// Redirect the discord user to bungie's OAUTH 2.0 Mechanism
func bungieAuth(c *gin.Context) {
    bungieAuthURL := "https://www.bungie.net/en/OAuth/Authorize?client_id=" + os.Getenv("CLIENT_ID") + "&response_type=code"
    params := c.Param("payload")
    fmt.Println(params)
    fmt.Println(bungieAuthURL)
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
    }
}
