package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type Agent struct {
	Idx        int             `json:"idx"`       // インデックス
	Name       string          `json:"name"`      // 名前
	Role       Role            `json:"role"`      // 役職
	Connection *websocket.Conn `json:"-"`         // 接続
	HasError   bool            `json:"has_error"` // エラーが発生したかどうか
}

func NewAgent(idx int, role Role, conn Connection) (*Agent, error) {
	agent := &Agent{
		Idx:        idx,
		Name:       conn.Name,
		Role:       role,
		Connection: conn.Conn,
		HasError:   false,
	}
	slog.Info("エージェントを作成しました", "idx", agent.Idx, "agent", agent.String(), "role", agent.Role, "connection", agent.Connection.RemoteAddr())
	return agent, nil
}

func (a *Agent) SendPacket(packet Packet, actionTimeout, responseTimeout time.Duration) (string, error) {
	if a.HasError {
		slog.Error("エージェントにエラーが発生しているため、リクエストを送信できません", "agent", a.String())
		return "", errors.New("エージェントにエラーが発生しているため、リクエストを送信できません")
	}
	req, err := json.Marshal(packet)
	if err != nil {
		slog.Error("パケットの作成に失敗しました", "error", err)
		a.HasError = true
		return "", err
	}
	err = a.Connection.WriteMessage(websocket.TextMessage, req)
	if err != nil {
		slog.Error("パケットの送信に失敗しました", "error", err)
		a.HasError = true
		return "", err
	}
	slog.Info("パケットを送信しました", "agent", a.String(), "packet", packet)
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
			slog.Info("レスポンスを受信しました", "agent", a.String(), "response", string(res))
			return strings.TrimRight(string(res), "\n"), nil
		case err := <-errChan:
			if websocket.IsUnexpectedCloseError(err) || websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				slog.Error("接続が閉じられました", "error", err)
				a.HasError = true
				return "", err
			}
			slog.Warn("レスポンスの受信に失敗したため、NAMEリクエストを送信します", "agent", a.String(), "error", err)
		case <-time.After(actionTimeout):
			slog.Warn("レスポンスの受信がタイムアウトしたため、NAMEリクエストを送信します", "agent", a.String())
		}
		nameReq, err := json.Marshal(Packet{Request: &R_NAME})
		if err != nil {
			slog.Error("NAMEパケットの作成に失敗しました", "error", err)
			a.HasError = true
			return "", err
		}
		err = a.Connection.WriteMessage(websocket.TextMessage, nameReq)
		if err != nil {
			slog.Error("NAMEパケットの送信に失敗しました", "error", err)
			a.HasError = true
			return "", err
		}
		slog.Info("NAMEパケットを送信しました", "agent", a.String())
		select {
		case res := <-responseChan:
			if string(res) == a.String() {
				slog.Info("NAMEリクエストのレスポンスを受信しました", "agent", a.String(), "response", string(res))
				return "", errors.New("リクエストのレスポンス受信がタイムアウトしました")
			} else {
				slog.Error("不正なNAMEリクエストのレスポンスを受信しました", "agent", a.String(), "response", string(res))
				a.HasError = true
				return "", errors.New("不正なNAMEリクエストのレスポンスを受信しました")
			}
		case err := <-errChan:
			slog.Error("NAMEリクエストのレスポンス受信に失敗しました", "agent", a.String(), "error", err)
			a.HasError = true
			return "", err
		case <-time.After(responseTimeout):
			slog.Error("NAMEリクエストのレスポンス受信がタイムアウトしました", "agent", a.String())
			a.HasError = true
			return "", errors.New("NAMEリクエストのレスポンス受信がタイムアウトしました")
		}
	}
	return "", nil
}

func (a Agent) String() string {
	return "Agent[" + fmt.Sprintf("%02d", a.Idx) + "]"
}

func (a Agent) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}
