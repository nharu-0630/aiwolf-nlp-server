package main

import (
	"flag"
	"os"

	"github.com/kano-lab/aiwolf-nlp-server/core"
	"github.com/kano-lab/aiwolf-nlp-server/model"
)

var (
	version  string
	revision string
	build    string
)

func main() {
	var (
		configPath  = flag.String("c", "./default.yml", "設定ファイルのパス")
		showVersion = flag.Bool("v", false, "バージョンを表示")
		showHelp    = flag.Bool("h", false, "ヘルプを表示")
	)
	flag.Parse()

	if *showVersion {
		println("version:", version)
		println("revision:", revision)
		println("build:", build)
		os.Exit(0)
	}

	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	config, err := model.LoadFromPath(*configPath)
	if err != nil {
		panic(err)
	}
	server := core.NewServer(*config)
	server.Run()
}
