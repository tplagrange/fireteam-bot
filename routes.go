package main

import (
    "github.com/gin-gonic/gin"
)

func initRoutes(router *gin.Engine) {
    router.GET("/", hello)

    router.GET("/api/bungie/callback", bungieCallback)
    router.GET("/api/bungie/auth/", bungieAuth)

    router.GET("/api/loadout/", getCurrentLoadout)
    router.GET("/api/loadout/:name/", setLoadout)
    router.GET("/api/loadouts/", getLoadouts)

    router.GET("/api/shaders/", getPartyShaders)
}
