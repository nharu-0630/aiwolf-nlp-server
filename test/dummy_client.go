package test

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kano-lab/aiwolf-nlp-server/model"
	"golang.org/x/exp/rand"
)

type DummyClient struct {
	conn      *websocket.Conn
	done      chan struct{}
	role      *model.Role
	info      map[string]interface{}
	setting   map[string]interface{}
	talkIndex int
}

func NewDummyClient(u url.URL, t *testing.T) (*DummyClient, error) {
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("dial: %v", err)
	}
	client := &DummyClient{
		conn:      c,
		done:      make(chan struct{}),
		role:      nil,
		info:      make(map[string]interface{}),
		setting:   make(map[string]interface{}),
		talkIndex: 0,
	}
	go client.listen(t)
	return client, nil
}

func (dc *DummyClient) listen(t *testing.T) {
	defer close(dc.done)
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

		resp := dc.handleRequest(received)
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

func (dc *DummyClient) setInfo(recv map[string]interface{}) {
	if info, ok := recv["info"].(map[string]interface{}); ok {
		dc.info = info
		if dc.role == nil {
			if roleMap, ok := info["roleMap"].(map[string]interface{}); ok {
				for _, v := range roleMap {
					role := model.RoleFromString(v.(string))
					dc.role = &role
					break
				}
			}
		}
	} else {
		panic(errors.New("info not found"))
	}
}

func (dc *DummyClient) setSetting(recv map[string]interface{}) {
	if setting, ok := recv["setting"].(map[string]interface{}); ok {
		dc.setting = setting
	} else {
		panic(errors.New("setting not found"))
	}
}

func (dc *DummyClient) handleName(_ map[string]interface{}) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, 8)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (dc *DummyClient) handleInitialize(recv map[string]interface{}) string {
	dc.setInfo(recv)
	dc.setSetting(recv)
	return ""
}

func (dc *DummyClient) handleCommunication(recv map[string]interface{}) string {
	request := recv["request"].(string)
	if _, ok := recv[strings.ToLower(request)+"History"].([]interface{}); ok {
	} else {
		panic(errors.New("history not found"))
	}
	dc.talkIndex++
	if dc.talkIndex < 3 {
		return fmt.Sprintf("%x", md5.Sum([]byte(time.Now().String())))
	}
	return model.T_OVER
}

func (dc *DummyClient) handleDailyFinish(recv map[string]interface{}) string {
	if _, ok := recv["talkHistory"].([]interface{}); ok {
	} else {
		panic(errors.New("talkHistory not found"))
	}
	if dc.role == &model.R_WEREWOLF {
		if _, ok := recv["whisperHistory"].([]interface{}); ok {
		} else {
			panic(errors.New("whisperHistory not found"))
		}
	}
	return ""
}

func (dc *DummyClient) handleFinish(recv map[string]interface{}) string {
	dc.setInfo(recv)
	return ""
}

func (dc *DummyClient) handleRequest(recv map[string]interface{}) string {
	if request, ok := recv["request"].(string); ok {
		switch request {
		case "NAME":
			return dc.handleName(recv)
		case "INITIALIZE", "DAILY_INITIALIZE":
			return dc.handleInitialize(recv)
		case "TALK", "WHISPER":
			return dc.handleCommunication(recv)
		case "VOTE", "DIVINE", "GUARD", "ATTACK":
			if statusMap, ok := dc.info["statusMap"].(map[string]interface{}); ok {
				for k, v := range statusMap {
					if v == "ALIVE" {
						return k
					}
				}
			} else {
				panic(errors.New("statusMap not found"))
			}
		case "DAILY_FINISH":
			dc.talkIndex = 0
			return dc.handleDailyFinish(recv)
		case "FINISH":
			return dc.handleFinish(recv)
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
