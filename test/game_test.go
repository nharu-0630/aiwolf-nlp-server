package test

import (
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/kano-lab/aiwolf-nlp-server/core"
	"github.com/kano-lab/aiwolf-nlp-server/model"
	"golang.org/x/exp/rand"
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

	names := make([]string, config.Game.AgentCount)
	for i := 0; i < config.Game.AgentCount; i++ {
		const letterBytes = "abcdefghijklmnopqrstuvwxyz"
		b := make([]byte, 8)
		for i := range b {
			b[i] = letterBytes[rand.Intn(len(letterBytes))]
		}
		names[i] = string(b)
	}

	clients := make([]*DummyClient, config.Game.AgentCount)
	for i := 0; i < config.Game.AgentCount; i++ {
		client, err := NewDummyClient(u, names[i], t)
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

func TestInfiniteGame(t *testing.T) {
	config, err := model.LoadFromPath("../config/infinite_debug.yml")
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

	names := make([]string, config.Game.AgentCount)
	for i := 0; i < config.Game.AgentCount; i++ {
		const letterBytes = "abcdefghijklmnopqrstuvwxyz"
		b := make([]byte, 8)
		for i := range b {
			b[i] = letterBytes[rand.Intn(len(letterBytes))]
		}
		names[i] = string(b)
	}

	for {
		clients := make([]*DummyClient, config.Game.AgentCount)
		for i := 0; i < config.Game.AgentCount; i++ {
			client, err := NewDummyClient(u, names[i], t)
			if err != nil {
				t.Fatalf("Failed to create WebSocket client: %v", err)
				break
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
	}

	time.Sleep(5 * time.Second)
	t.Log("Test completed successfully")
}
