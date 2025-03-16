package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	app := InitWebServer()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}

	server := app.server
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello webook！")
	})

	server.Run(":8080")
}
