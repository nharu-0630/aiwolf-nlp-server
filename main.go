package main

import (
	"os"

	"github.com/nharu-0630/aiwolf-nlp-server/core"
	"github.com/nharu-0630/aiwolf-nlp-server/model"
)

func main() {
	configPath := "./config/default.yml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}
	config, err := model.LoadConfigFromFile(configPath)
	if err != nil {
		panic(err)
	}
	server := core.NewServer(*config)
	server.Run()
}
