package main

import (
    "fmt"
    "net/http"
    "os"

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
    router.GET("/", hello)

    // Start Listening
    router.Run(":" + port)
}
