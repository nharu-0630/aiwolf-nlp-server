package model

import (
	"encoding/json"
	"errors"
	"log/slog"
	"time"

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
		3*time.Minute,
		15*time.Minute,
	)
	if err != nil {
		return nil, err
	}
	agent.Name = name
	slog.Info("エージェントを作成しました", "agent", agent.Name, "role", agent.Role, "connection", agent.Connection.RemoteAddr())
	return agent, nil
}

func (a *Agent) SendPacket(packet Packet, actionTimeout, responseTimeout time.Duration) (string, error) {
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
		a.Connection.SetReadDeadline(time.Now().Add(actionTimeout))
		_, res, err := a.Connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) || websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				slog.Error("接続が閉じられました", "error", err)
				return "", err
			}
			a.Connection.SetReadDeadline(time.Now().Add(responseTimeout))
			for {
				slog.Warn("レスポンスの受信がタイムアウトしました。ヘルスチェックのリクエストを再送します", "agent", a.Name)
				req, err := json.Marshal(Packet{Request: &R_NAME})
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
				_, res, err = a.Connection.ReadMessage()
				if err != nil {
					slog.Error("レスポンスの再受信がタイムアウトしました", "error", err)
					return "", err
				}
				if len(res) == 0 {
					slog.Error("レスポンスが空です")
					return "", errors.New("レスポンスが空です")
				}
				if string(res) == a.Name {
					slog.Info("ヘルスチェックのレスポンスを受信しました", "agent", a.Name, "response", string(res))
					break
				}
			}
			return "", nil
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
