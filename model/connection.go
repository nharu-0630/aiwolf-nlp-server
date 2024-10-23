package model

import (
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/gorilla/websocket"
)

type Connection struct {
	Team string
	Name string
	Conn *websocket.Conn
}

func NewConnection(conn *websocket.Conn) (*Connection, error) {
	req, err := json.Marshal(Packet{
		Request: &R_NAME,
	})
	if err != nil {
		slog.Error("NAMEパケットの作成に失敗しました", "error", err)
		return nil, err
	}
	err = conn.WriteMessage(websocket.TextMessage, req)
	if err != nil {
		slog.Error("NAMEパケットの送信に失敗しました", "error", err)
		return nil, err
	}
	slog.Info("NAMEパケットを送信しました", "remote_addr", conn.RemoteAddr().String())
	_, res, err := conn.ReadMessage()
	if err != nil {
		slog.Error("NAMEリクエストの受信に失敗しました", "error", err)
		return nil, err
	}
	name := strings.TrimRight(string(res), "\n")
	team := strings.TrimRight(name, "1234567890")
	connection := Connection{
		Team: team,
		Name: name,
		Conn: conn,
	}
	slog.Info("クライアントが接続しました", "team", team, "name", name, "remote_addr", conn.RemoteAddr().String())
	return &connection, nil
}
