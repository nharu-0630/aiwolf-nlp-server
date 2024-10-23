package test

import (
	"log"
	"net/url"
	"strconv"
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

func TestConnectServer(t *testing.T) {
	go func() {
		server := core.NewServer()
		server.Run()
	}()

	u := url.URL{Scheme: "ws", Host: config.WEBSOCKET_HOST + ":" + strconv.Itoa(config.WEBSOCKET_PORT), Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	c, done := setupWebSocketConnection(u, t)
	defer teardownWebSocketConnection(c, done)

	select {
	case <-done:
		log.Println("Connection closed")
	case <-time.After(5 * time.Second):
		log.Println("Test completed: Connection was successful for 5 seconds")
	}
}

// func TestGame(t *testing.T) {
// 	go func() {
// 		server := core.NewServer()
// 		server.Run()
// 	}()
// 	time.Sleep(1 * time.Second)

// 	interrupt := make(chan os.Signal, 1)
// 	signal.Notify(interrupt, os.Interrupt)

// 	u := url.URL{Scheme: "ws", Host: config.WEBSOCKET_HOST + ":" + strconv.Itoa(config.WEBSOCKET_PORT), Path: "/ws"}
// 	log.Printf("connecting to %s", u.String())

// 	var wg sync.WaitGroup

// 	for i := 0; i < config.AGENT_COUNT_PER_GAME; i++ {
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			c, done := setupWebSocketConnection(u, t)
// 			defer teardownWebSocketConnection(c, done)

// 			ticker := time.NewTicker(time.Second)
// 			defer ticker.Stop()

// 			for {
// 				select {
// 				case <-done:
// 					return
// 				case t := <-ticker.C:
// 					err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
// 					if err != nil {
// 						log.Printf("write: %v", err)
// 						return
// 					}
// 				case <-interrupt:
// 					log.Println("interrupt")

// 					err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
// 					if err != nil {
// 						log.Printf("write close: %v", err)
// 						return
// 					}
// 					teardownWebSocketConnection(c, done)
// 					return
// 				}
// 			}
// 		}()
// 	}

// 	wg.Wait()
// }
