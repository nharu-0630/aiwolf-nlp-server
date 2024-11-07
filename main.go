package main

import (
	"github.com/nharu-0630/aiwolf-nlp-server/core"
	"github.com/nharu-0630/aiwolf-nlp-server/model"
)

func main() {
	config := model.DefaultConfig
	server := core.NewServer(config)
	server.Run()
}
