package main

import (
    "context"
    "log"
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
    out := log.New(os.Stdout, "[bot] ", log.LstdFlags|log.Lshortfile)

    port := os.Getenv("PORT")

    if port == "" {
        out.Println("Using default port...")
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

    // Wait here until CTRL-C or other term signal is received.
    out.Println("Server Started.")

    sc := make(chan os.Signal, 1)
    signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
    <-sc

    // Cleanly close down the Discord session.
    err := db.Disconnect(context.TODO())
    if err != nil {
        out.Println(err)
    } else  {
        out.Println("Connection to MongoDB closed.")
    }
}
