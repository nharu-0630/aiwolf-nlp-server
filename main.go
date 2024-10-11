package main

import (
	"github.com/nharu-0630/aiwolf-nlp-server/core"
)

func main() {
	server := core.NewServer()
	server.Run()
}
