package main

import (
	"flag"
	"os"

	"github.com/kano-lab/aiwolf-nlp-server/core"
	"github.com/kano-lab/aiwolf-nlp-server/model"
)

func main() {
	var (
		configPath = flag.String("c", "./default.yml", "設定ファイルのパス")
		help       = flag.Bool("h", false, "ヘルプを表示")
	)
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	config, err := model.LoadConfigFromPath(*configPath)
	if err != nil {
		panic(err)
	}
	server := core.NewServer(*config)
	server.Run()
}
