package model

import "github.com/gorilla/websocket"

type Connection struct {
	Name string
	Conn *websocket.Conn
}
