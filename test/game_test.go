package test

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nharu-0630/aiwolf-nlp-server/config"
	"github.com/nharu-0630/aiwolf-nlp-server/core"
)

type TemplateResponse struct {
	Template string `json:"template"`
}

func setupWebSocketConnection(u url.URL, t *testing.T) (*websocket.Conn, chan struct{}) {
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Printf("read: %v", err)
				return
			}
			log.Printf("recv: %s", message)

			var received map[string]interface{}
			if err := json.Unmarshal(message, &received); err != nil {
				log.Printf("json unmarshal: %v", err)
				return
			}

			var response TemplateResponse
			if requestType, ok := received["request"].(string); ok {
				switch requestType {
				case "NAME":
					response = TemplateResponse{Template: "This is a NAME template"}
				case "ROLE":
					response = TemplateResponse{Template: "This is a ROLE template"}
				case "TALK":
					response = TemplateResponse{Template: "This is a TALK template"}
				default:
					response = TemplateResponse{Template: "Unknown request type"}
				}
			} else {
				response = TemplateResponse{Template: "Invalid request format"}
			}

			responseBytes, err := json.Marshal(response)
			if err != nil {
				log.Printf("json marshal: %v", err)
				return
			}

			err = c.WriteMessage(websocket.TextMessage, responseBytes)
			if err != nil {
				log.Printf("write: %v", err)
				return
			}
		}
	}()

	return c, done
}

func teardownWebSocketConnection(c *websocket.Conn, done chan struct{}) {
	c.Close()
	select {
	case <-done:
	case <-time.After(time.Second):
	}
}

func TestConnectServer(t *testing.T) {
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
	log.Printf("connecting to %s", u.String())

	c, done := setupWebSocketConnection(u, t)
	defer teardownWebSocketConnection(c, done)

	select {
	case <-done:
		log.Println("Connection closed")
	case <-time.After(3 * time.Second):
		log.Println("Test completed: Connection was successful for 3 seconds")
	}
}
