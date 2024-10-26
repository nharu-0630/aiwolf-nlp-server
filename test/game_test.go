package test

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nharu-0630/aiwolf-nlp-server/config"
	"github.com/nharu-0630/aiwolf-nlp-server/core"
)

type WebSocketClient struct {
	conn *websocket.Conn
	done chan struct{}
}

func NewWebSocketClient(u url.URL, t *testing.T) (*WebSocketClient, error) {
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("dial: %v", err)
	}

	client := &WebSocketClient{
		conn: c,
		done: make(chan struct{}),
	}

	go client.listen(t)

	return client, nil
}

func (client *WebSocketClient) listen(t *testing.T) {
	defer close(client.done)
	var index int
	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) || websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				t.Logf("connection closed: %v", err)
				return
			}
			t.Logf("read: %v", err)
			return
		}
		t.Logf("recv: %s", message)

		var received map[string]interface{}
		if err := json.Unmarshal(message, &received); err != nil {
			t.Logf("unmarshal: %v", err)
			continue
		}

		resp := client.handleRequest(received, &index)
		if resp != "" {
			err = client.conn.WriteMessage(websocket.TextMessage, []byte(resp))
			if err != nil {
				if websocket.IsUnexpectedCloseError(err) || websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					t.Logf("connection closed: %v", err)
					return
				}
				t.Logf("write: %v", err)
				continue
			}
			t.Logf("send: %s", resp)
		}
	}
}

func (client *WebSocketClient) handleRequest(received map[string]interface{}, index *int) string {
	if request, ok := received["request"].(string); ok {
		switch request {
		case "NAME":
			return fmt.Sprintf("%x", md5.Sum([]byte(client.conn.LocalAddr().String())))
		case "TALK", "WHISPER":
			*index++
			if *index < 3 {
				return fmt.Sprintf("%x", md5.Sum([]byte(time.Now().String())))
			}
			return "Over"
		case "VOTE", "DIVINE", "GUARD", "ATTACK":
			if info, ok := received["info"].(map[string]interface{}); ok {
				if statusMap, ok := info["statusMap"].(map[string]interface{}); ok {
					for agent, status := range statusMap {
						if status == "ALIVE" {
							return agent
						}
					}
				}
			}
		case "INITIALIZE", "DAILY_INITIALIZE", "DAILY_FINISH":
			*index = 0
			return ""
		case "FINISH":
			return ""
		default:
			return "Invalid request"
		}
	}
	return "Invalid request"
}

func (client *WebSocketClient) Close() {
	client.conn.Close()
	select {
	case <-client.done:
	case <-time.After(time.Second):
	}
}

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

	clients := make([]*WebSocketClient, config.AGENT_COUNT_PER_GAME)
	for i := 0; i < 5; i++ {
		client, err := NewWebSocketClient(u, t)
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
