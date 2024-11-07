package core

import (
	"log/slog"
	"sync"

	"github.com/nharu-0630/aiwolf-nlp-server/model"
)

type WaitingRoom struct {
	agentCount  int
	selfMatch   bool
	connections map[string][]model.Connection
	mu          sync.RWMutex
}

func NewWaitingRoom(config model.Config) *WaitingRoom {
	return &WaitingRoom{
		agentCount:  config.Game.AgentCount,
		selfMatch:   config.Server.SelfMatch,
		connections: make(map[string][]model.Connection),
	}
}

func (wr *WaitingRoom) AddConnection(team string, connection model.Connection) {
	wr.mu.Lock()
	defer wr.mu.Unlock()
	wr.connections[team] = append(wr.connections[team], connection)
	slog.Info("新しいクライアントが待機部屋に追加されました", "team", team, "remote_addr", connection.Conn.RemoteAddr().String())
}

func (wr *WaitingRoom) IsReady() bool {
	wr.mu.RLock()
	defer wr.mu.RUnlock()
	if wr.selfMatch {
		for _, conns := range wr.connections {
			if len(conns) >= wr.agentCount {
				return true
			}
		}
		return false
	}
	return len(wr.connections) == wr.agentCount
}

func (wr *WaitingRoom) GetConnections() []model.Connection {
	wr.mu.Lock()
	defer wr.mu.Unlock()
	connections := []model.Connection{}
	if wr.selfMatch {
		for team, conns := range wr.connections {
			if len(conns) >= wr.agentCount {
				connections = append(connections, conns[:wr.agentCount]...)
				wr.connections[team] = conns[wr.agentCount:]
				if len(wr.connections[team]) == 0 {
					delete(wr.connections, team)
				}
			}
		}
		return connections
	}
	for team, conns := range wr.connections {
		connections = append(connections, conns[0])
		wr.connections[team] = conns[1:]
		if len(wr.connections[team]) == 0 {
			delete(wr.connections, team)
		}
	}
	return connections
}
