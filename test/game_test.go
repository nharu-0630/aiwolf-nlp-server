package test

import (
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/nharu-0630/aiwolf-nlp-server/config"
	"github.com/nharu-0630/aiwolf-nlp-server/core"
)

func TestGame(t *testing.T) {
	host := config.WEBSOCKET_INTERNAL_HOST
	if _, exists := os.LookupEnv("GITHUB_ACTIONS"); exists {
		host = config.WEBSOCKET_EXTERNAL_HOST
	}
	port := config.WEBSOCKET_PORT
	go func() {
		server := core.NewServer(host, port)
		server.Run()
	}()

	time.Sleep(5 * time.Second)

	u := url.URL{Scheme: "ws", Host: host + ":" + strconv.Itoa(port), Path: "/ws"}
	t.Logf("Connecting to %s", u.String())

	clientsNum := config.AGENT_COUNT_PER_GAME
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
