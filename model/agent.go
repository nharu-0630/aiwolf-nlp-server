package model

import (
	"encoding/json"
	"errors"
	"fmt"
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
		responseChan := make(chan []byte)
		errChan := make(chan error)
		go func() {
			_, res, err := a.Connection.ReadMessage()
			if err != nil {
				errChan <- err
				return
			}
			responseChan <- res
		}()
		select {
		case res := <-responseChan:
			slog.Info("レスポンスを受信しました", "agent", a.Name, "response", string(res))
			return string(res), nil
		case err := <-errChan:
			if websocket.IsUnexpectedCloseError(err) || websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				slog.Error("接続が閉じられました", "error", err)
				return "", err
			}
			slog.Warn("レスポンスの受信に失敗しました。NAMEリクエストを送信します", "agent", a.Name, "error", err)
		case <-time.After(actionTimeout):
			slog.Warn("レスポンスの受信がタイムアウトしました。NAMEリクエストを送信します", "agent", a.Name)
		}
		nameReq, err := json.Marshal(Packet{Request: &R_NAME})
		if err != nil {
			slog.Error("NAMEパケットの作成に失敗しました", "error", err)
			return "", err
		}
		err = a.Connection.WriteMessage(websocket.TextMessage, nameReq)
		if err != nil {
			slog.Error("NAMEパケットの送信に失敗しました", "error", err)
			return "", err
		}
		slog.Info("NAMEパケットを送信しました", "agent", a.Name)
		select {
		case res := <-responseChan:
			if string(res) == a.Name {
				slog.Info("NAMEリクエストのレスポンスを受信しました", "agent", a.Name, "response", string(res))
				return "", nil
			} else {
				slog.Error("不正なNAMEリクエストのレスポンスを受信しました", "agent", a.Name, "response", string(res))
				return "", errors.New("不正なNAMEリクエストのレスポンスを受信しました")
			}
		case err := <-errChan:
			slog.Error("NAMEリクエストのレスポンス受信に失敗しました", "agent", a.Name, "error", err)
			return "", err
		case <-time.After(responseTimeout):
			slog.Error("NAMEリクエストのレスポンス受信がタイムアウトしました", "agent", a.Name)
			return "", errors.New("NAMEリクエストのレスポンス受信がタイムアウトしました")
		}
	}
	return "", nil
}

func (a Agent) String() string {
	return "Agent[" + fmt.Sprintf("%02d", a.Idx) + "]"
}

func (a Agent) MarshalJSON() ([]byte, error) {
	if a == (Agent{}) {
		return json.Marshal(nil)
	}
	return json.Marshal(a.String())
}
