package main

import (
    "github.com/gin-gonic/gin"
)

func initRoutes(router *gin.Engine) {
    router.GET("/", hello)

    router.GET("/api/loadout/:id", getLoadout)
}
