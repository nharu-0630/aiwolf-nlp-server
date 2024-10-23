package main

import (
	"github.com/nharu-0630/aiwolf-nlp-server/config"
	"github.com/nharu-0630/aiwolf-nlp-server/core"
)

func main() {
	server := core.NewServer(config.WEBSOCKET_INTERNAL_HOST, config.WEBSOCKET_PORT)
	server.Run()
}
