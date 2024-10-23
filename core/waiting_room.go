package core

import (
	"log/slog"
	"sync"

	"github.com/nharu-0630/aiwolf-nlp-server/config"
	"github.com/nharu-0630/aiwolf-nlp-server/model"
)

type WaitingRoom struct {
	connections map[string][]model.Connection
	mu          sync.RWMutex
}

func NewWaitingRoom() *WaitingRoom {
	return &WaitingRoom{
		connections: make(map[string][]model.Connection),
	}
}

func (wr *WaitingRoom) AddConnection(team string, connection model.Connection) {
	wr.mu.Lock()
	defer wr.mu.Unlock()
	wr.connections[team] = append(wr.connections[team], connection)
	slog.Info("新しいクライアントが待機部屋に追加されました", "team", team)
}

func (wr *WaitingRoom) IsReady() bool {
	wr.mu.RLock()
	defer wr.mu.RUnlock()
	return len(wr.connections) == config.AGENT_COUNT_PER_GAME
}

func (wr *WaitingRoom) GetConnections() []model.Connection {
	wr.mu.Lock()
	defer wr.mu.Unlock()
	connections := []model.Connection{}
	for team, conns := range wr.connections {
		connections = append(connections, conns[0])
		wr.connections[team] = conns[1:]
		if len(wr.connections[team]) == 0 {
			delete(wr.connections, team)
		}
	}
	return connections
}
