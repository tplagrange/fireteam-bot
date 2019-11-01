package main

import (
    "github.com/tplagrange/fireteam-bot/api"

    "github.com/gin-gonic/gin"
)

func initRoutes(router *gin.Engine) {
    router.GET("/", hello)

    router.GET("/api/loadout/:id", api.GetLoadout)
}
