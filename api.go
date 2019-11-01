package main

import (
    "context"
    "fmt"
    // "net/http"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
)

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
