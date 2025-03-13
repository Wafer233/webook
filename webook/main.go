package main

import (
	"webook/internal/integration/startup"
)

func main() {
	server := startup.InitWebServer()
	server.Run(":8080")
}
