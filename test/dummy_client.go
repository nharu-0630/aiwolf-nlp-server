package test

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/exp/rand"
)

type DummyClient struct {
	conn *websocket.Conn
	done chan struct{}
}

func NewDummyClient(u url.URL, t *testing.T) (*DummyClient, error) {
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("dial: %v", err)
	}

	client := &DummyClient{
		conn: c,
		done: make(chan struct{}),
	}

	go client.listen(t)

	return client, nil
}

func (dc *DummyClient) listen(t *testing.T) {
	defer close(dc.done)
	var index int
	var statusMap = make(map[string]string)
	for {
		_, message, err := dc.conn.ReadMessage()
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

		resp := dc.handleRequest(received, &index, &statusMap)
		if resp != "" {
			err = dc.conn.WriteMessage(websocket.TextMessage, []byte(resp))
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

func (dc *DummyClient) handleRequest(received map[string]interface{}, index *int, statusMap *map[string]string) string {
	if request, ok := received["request"].(string); ok {
		if info, ok := received["info"].(map[string]interface{}); ok {
			if sm, ok := info["statusMap"].(map[string]interface{}); ok {
				for k, v := range sm {
					if strVal, ok := v.(string); ok {
						(*statusMap)[k] = strVal
					}
				}
			}
		}
		switch request {
		case "NAME":
			const letterBytes = "abcdefghijklmnopqrstuvwxyz"
			b := make([]byte, 8)
			for i := range b {
				b[i] = letterBytes[rand.Intn(len(letterBytes))]
			}
			return string(b)
		case "TALK", "WHISPER":
			*index++
			if *index < 3 {
				return fmt.Sprintf("%x", md5.Sum([]byte(time.Now().String())))
			}
			return "Over"
		case "VOTE", "DIVINE", "GUARD", "ATTACK":
			for k, v := range *statusMap {
				if v == "ALIVE" {
					return k
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

func (dc *DummyClient) Close() {
	dc.conn.Close()
	select {
	case <-dc.done:
	case <-time.After(time.Second):
	}
}
