package main

import (
	"github.com/nharu-0630/aiwolf-nlp-server/config"
	"github.com/nharu-0630/aiwolf-nlp-server/core"
)

func main() {
	host := config.WEBSOCKET_INTERNAL_HOST
	port := config.WEBSOCKET_PORT
	server := core.NewServer(host, port)
	server.Run()
}
