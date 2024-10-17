package test

import (
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nharu-0630/aiwolf-nlp-server/config"
	"github.com/nharu-0630/aiwolf-nlp-server/core"
)

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

func TestGame(t *testing.T) {
	// Start the server
	go func() {
		server := core.NewServer()
		server.Run()
	}()
	// Wait for the server to start
	time.Sleep(1 * time.Second)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: config.WEBSOCKET_HOST + ":" + strconv.Itoa(config.WEBSOCKET_PORT), Path: "/"}
	log.Printf("connecting to %s", u.String())

	var wg sync.WaitGroup

	// Create the required number of websocket clients in parallel
	for i := 0; i < config.GAME_AGENT_COUNT; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, done := setupWebSocketConnection(u, t)
			defer teardownWebSocketConnection(c, done)

			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-done:
					return
				case t := <-ticker.C:
					err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
					if err != nil {
						log.Printf("write: %v", err)
						return
					}
				case <-interrupt:
					log.Println("interrupt")

					// Cleanly close the connection by sending a close message and then
					// waiting (with timeout) for the server to close the connection.
					err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
					if err != nil {
						log.Printf("write close: %v", err)
						return
					}
					teardownWebSocketConnection(c, done)
					return
				}
			}
		}()
	}

	wg.Wait()
}
