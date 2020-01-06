package main

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "os/signal"
    "syscall"

    // Internal Libraries
    "github.com/tplagrange/fireteam-bot/discord"

    // External Libraries
    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/mongo"
    _ "github.com/joho/godotenv/autoload"
)

var db *mongo.Client

func hello(c *gin.Context) {
    c.String(http.StatusOK, "Hello, world!")
}

func main() {
    port := os.Getenv("APP_PORT")

    if port == "" {
        fmt.Print("Using default port...")
        port = "8080"
    }

    // Create router
    router := gin.New()

    // Define plugins
    router.Use(gin.Logger())

    // Define routes from './routes.go'
    initRoutes(router)

    // Start mongoDB Client
    db = connectClient()

    // Start Web Server Routine
    go router.Run(":" + port)

    // Start Discord Bot routine
    go discord.Bot()

    fmt.Println("Server Started.")
}
