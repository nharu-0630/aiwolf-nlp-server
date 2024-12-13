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
		configPath    = flag.String("c", "./default.yml", "設定ファイルのパス")
		analyzerMode  = flag.Bool("a", false, "解析モード")
		reductionMode = flag.Bool("r", false, "縮約モード")
		srcConfigPath = flag.String("s", "", "ソース設定ファイルのパス")
		dstConfigPath = flag.String("d", "", "デスティネーション設定ファイルのパス")
		showVersion   = flag.Bool("v", false, "バージョンを表示")
		showHelp      = flag.Bool("h", false, "ヘルプを表示")
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

	if *analyzerMode {
		core.Analyzer(*config)
		return
	}

	if *reductionMode {
		srcConfig, err := model.LoadFromPath(*srcConfigPath)
		if err != nil {
			panic(err)
		}
		dstConfig, err := model.LoadFromPath(*dstConfigPath)
		if err != nil {
			panic(err)
		}
		core.Reduction(*srcConfig, *dstConfig)
		return
	}

	server := core.NewServer(*config)
	server.Run()
}
