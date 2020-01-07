package main

import (
    "github.com/gin-gonic/gin"
)

func initRoutes(router *gin.Engine) {
    router.GET("/", hello)

    router.GET("/api/bungie/callback", bungieCallback)
    router.GET("/api/bungie/auth/:id", bungieAuth)

    router.GET("/api/loadout/:id", getLoadout)
}