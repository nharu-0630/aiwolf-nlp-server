package test

import (
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/kano-lab/aiwolf-nlp-server/core"
	"github.com/kano-lab/aiwolf-nlp-server/model"
)

func TestGame(t *testing.T) {
	config, err := model.LoadFromPath("../config/debug.yml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if _, exists := os.LookupEnv("GITHUB_ACTIONS"); exists {
		config.Server.WebSocket.Host = model.WebSocketExternalHost
	}
	go func() {
		server := core.NewServer(*config)
		server.Run()
	}()

	time.Sleep(5 * time.Second)

	u := url.URL{Scheme: "ws", Host: config.Server.WebSocket.Host + ":" + strconv.Itoa(config.Server.WebSocket.Port), Path: "/ws"}
	t.Logf("Connecting to %s", u.String())

	clientsNum := config.Game.AgentCount
	clients := make([]*DummyClient, clientsNum)
	for i := 0; i < clientsNum; i++ {
		client, err := NewDummyClient(u, t)
		if err != nil {
			t.Fatalf("Failed to create WebSocket client: %v", err)
		}
		clients[i] = client
		defer clients[i].Close()
	}

	for _, client := range clients {
		select {
		case <-client.done:
			t.Log("Connection closed")
		case <-time.After(30 * time.Second):
			t.Fatalf("Timeout")
		}
	}

	time.Sleep(5 * time.Second)

	t.Log("Test completed successfully")
}
