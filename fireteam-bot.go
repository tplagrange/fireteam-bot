package main

import (
    "fmt"
    "net/http"
    "os"
    "os/signal"
    "syscall"

    // Internal Libraries
    "github.com/tplagrange/fireteam-bot/api"
    "github.com/tplagrange/fireteam-bot/discord"

    // External Libraries
    "github.com/gin-gonic/gin"
)

func hello(c *gin.Context) {
    c.String(http.StatusOK, "Hello, world!")
}

func main() {
    port := os.Getenv("PORT")

    if port == "" {
        fmt.Print("Using default port...")
        port = "8080"
    }

    // Create router
    router := gin.New()

    // Define plugins
    router.Use(gin.Logger())

    // Define routes
    initRoutes(router)

    // Start mongoDB Client
    go api.ConnectDB()

    // Start Web Server Routine
    go router.Run(":" + port)

    // Start Discord Bot routine
    go discord.Bot()

    // Wait here until CTRL-C or other term signal is received.
    fmt.Println("Server Started.")
    sc := make(chan os.Signal, 1)
    signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
    <-sc
}
