package main

import (
	"github.com/nharu-0630/aiwolf-nlp-server/core"
	"github.com/nharu-0630/aiwolf-nlp-server/model"
)

func main() {
	config, err := model.LoadConfigFromFile("./config/default.yml")
	if err != nil {
		panic(err)
	}
	server := core.NewServer(*config)
	server.Run()
}
