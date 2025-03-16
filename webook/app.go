package main

import (
	"github.com/gin-gonic/gin"
	"webook/internal/event"
)

type App struct {
	server    *gin.Engine
	consumers []event.Consumer
}
