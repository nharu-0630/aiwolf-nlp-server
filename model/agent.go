package model

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type Agent struct {
	Idx        int             `json:"idx"`  // インデックス
	Name       string          `json:"name"` // 名前
	Role       Role            `json:"role"` // 役職
	Connection *websocket.Conn `json:"-"`    // 接続
}

func NewAgent(idx int, role Role, conn *websocket.Conn) (*Agent, error) {
	agent := &Agent{
		Idx:        idx,
		Role:       role,
		Connection: conn,
	}
	name, err := agent.SendPacket(
		Packet{
			Request: &R_NAME,
		},
	)
	if err != nil {
		return nil, err
	}
	agent.Name = name
	log.Println("New agent created:", agent)
	return agent, nil
}

func (a *Agent) SendPacket(packet Packet) (string, error) {
	req, err := json.Marshal(packet)
	if err != nil {
		return "", err
	}
	err = a.Connection.WriteMessage(websocket.TextMessage, req)
	if err != nil {
		return "", err
	}
	if packet.Request.RequiredResponse {
		_, res, err := a.Connection.ReadMessage()
		if err != nil {
			return "", err
		}
		return string(res), nil
	}
	return "", nil
}

func (a Agent) String() string {
	return a.Name
}

func (a Agent) MarshalJSON() ([]byte, error) {
	if a == (Agent{}) {
		return json.Marshal(nil)
	}
	return json.Marshal(a.String())
}
