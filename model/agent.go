package model

import (
	"encoding/json"
	"errors"
	"log/slog"

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
	slog.Info("エージェントを作成しました", "agent", agent.Name, "role", agent.Role, "connection", agent.Connection.RemoteAddr())
	return agent, nil
}

func (a *Agent) SendPacket(packet Packet) (string, error) {
	req, err := json.Marshal(packet)
	if err != nil {
		slog.Error("パケットの作成に失敗しました", "error", err)
		return "", err
	}
	err = a.Connection.WriteMessage(websocket.TextMessage, req)
	if err != nil {
		slog.Error("パケットの送信に失敗しました", "error", err)
		return "", err
	}
	slog.Info("パケットを送信しました", "agent", a.Name, "packet", packet)
	if packet.Request.RequiredResponse {
		_, res, err := a.Connection.ReadMessage()
		if err != nil {
			slog.Error("レスポンスの受信に失敗しました", "error", err)
			return "", err
		}
		if string(res) == "" {
			slog.Error("レスポンスが空です")
			return "", errors.New("レスポンスが空です")
		}
		slog.Info("レスポンスを受信しました", "agent", a.Name, "response", string(res))
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
