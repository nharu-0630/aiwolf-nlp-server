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
	conn        *websocket.Conn
	done        chan struct{}
	role        model.Role
	info        map[string]interface{}
	setting     map[string]interface{}
	talkIndex   int
	prevRequest model.Request
}

func NewDummyClient(u url.URL, t *testing.T) (*DummyClient, error) {
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("dial: %v", err)
	}
	client := &DummyClient{
		conn:        c,
		done:        make(chan struct{}),
		role:        model.Role{},
		info:        make(map[string]interface{}),
		setting:     make(map[string]interface{}),
		talkIndex:   0,
		prevRequest: model.Request{},
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

		var recv map[string]interface{}
		if err := json.Unmarshal(message, &recv); err != nil {
			t.Logf("unmarshal: %v", err)
			continue
		}

		request := model.RequestFromString(recv["request"].(string))
		resp, err := dc.handleRequest(request, recv)
		if err != nil {
			t.Error(err)
		}
		dc.prevRequest = request

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

func (dc *DummyClient) setInfo(recv map[string]interface{}) error {
	if info, ok := recv["info"].(map[string]interface{}); ok {
		dc.info = info
		if dc.role.String() == "" {
			if roleMap, ok := info["roleMap"].(map[string]interface{}); ok {
				for _, v := range roleMap {
					dc.role = model.RoleFromString(v.(string))
					break
				}
			}
		}
	} else {
		return errors.New("info not found")
	}
	return nil
}

func (dc *DummyClient) setSetting(recv map[string]interface{}) error {
	if setting, ok := recv["setting"].(map[string]interface{}); ok {
		dc.setting = setting
	} else {
		return errors.New("setting not found")
	}
	return nil
}

func (dc *DummyClient) handleName(_ map[string]interface{}) (string, error) {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, 8)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b), nil
}

func (dc *DummyClient) handleInitialize(recv map[string]interface{}) (string, error) {
	err := dc.setInfo(recv)
	if err != nil {
		return "", err
	}
	err = dc.setSetting(recv)
	if err != nil {
		return "", err
	}
	return "", nil
}

func (dc *DummyClient) handleCommunication(recv map[string]interface{}) (string, error) {
	request := recv["request"].(string)
	if _, ok := recv[strings.ToLower(request)+"History"].([]interface{}); ok {
	} else {
		return "", errors.New("history not found")
	}
	dc.talkIndex++
	if dc.talkIndex < 3 {
		return fmt.Sprintf("%x", md5.Sum([]byte(time.Now().String()))), nil
	}
	return model.T_OVER, nil
}

func (dc *DummyClient) handleTarget(_ map[string]interface{}) (string, error) {
	if statusMap, ok := dc.info["statusMap"].(map[string]interface{}); ok {
		for k, v := range statusMap {
			if v == model.S_ALIVE.String() {
				return k, nil
			}
		}
		return "", errors.New("target not found")
	}
	return "", errors.New("statusMap not found")
}

func (dc *DummyClient) handleDailyFinish(recv map[string]interface{}) (string, error) {
	if _, ok := recv["talkHistory"].([]interface{}); ok {
	} else {
		return "", errors.New("talkHistory not found")
	}
	if dc.role == model.R_WEREWOLF {
		if _, ok := recv["whisperHistory"].([]interface{}); ok {
		} else {
			return "", errors.New("whisperHistory not found")
		}
	} else {
		if _, ok := recv["whisperHistory"]; ok {
			return "", errors.New("whisperHistory found")
		}
	}
	return "", nil
}

func (dc *DummyClient) handleFinish(recv map[string]interface{}) (string, error) {
	err := dc.setInfo(recv)
	if err != nil {
		return "", err
	}
	return "", nil
}

func (dc *DummyClient) handleRequest(request model.Request, recv map[string]interface{}) (string, error) {
	switch request {
	case model.R_NAME:
		return dc.handleName(recv)
	case model.R_INITIALIZE:
		return dc.handleInitialize(recv)
	case model.R_DAILY_INITIALIZE:
		return dc.handleInitialize(recv)
	case model.R_TALK, model.R_WHISPER:
		return dc.handleCommunication(recv)
	case model.R_VOTE, model.R_DIVINE, model.R_GUARD, model.R_ATTACK:
		return dc.handleTarget(recv)
	case model.R_DAILY_FINISH:
		dc.talkIndex = 0
		return dc.handleDailyFinish(recv)
	case model.R_FINISH:
		return dc.handleFinish(recv)
	}
	return "", errors.New("request not found")
}

func (dc *DummyClient) Close() {
	dc.conn.Close()
	select {
	case <-dc.done:
	case <-time.After(time.Second):
	}
}
