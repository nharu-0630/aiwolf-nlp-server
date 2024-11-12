package core

import (
	"errors"
	"log/slog"
	"math/rand/v2"
	"sync"

	"github.com/kano-lab/aiwolf-nlp-server/model"
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

func (wr *WaitingRoom) GetConnectionsWithMatchOptimizer(scheduledMatches []map[model.Role][]string) (map[model.Role][]model.Connection, error) {
	wr.mu.Lock()
	defer wr.mu.Unlock()
	var roleMapConns = make(map[model.Role][]model.Connection)

	if len(scheduledMatches) == 0 {
		return nil, errors.New("スケジュールされたマッチがありません")
	}

	var readyMatch = map[model.Role][]string{}
	ready := true
	for _, match := range scheduledMatches {
		for _, teams := range match {
			for _, team := range teams {
				if len(wr.connections[team]) == 0 {
					ready = false
					break
				}
			}
			if !ready {
				break
			}
		}
		if ready {
			readyMatch = match
			break
		}
	}
	if !ready {
		return nil, errors.New("スケジュールされたマッチ内に不足しているチームがあります")
	}
	slog.Info("スケジュールされたマッチの接続を取得しました")

	for role, teams := range readyMatch {
		for _, team := range teams {
			roleMapConns[role] = append(roleMapConns[role], wr.connections[team][0])
			wr.connections[team] = wr.connections[team][1:]
			if len(wr.connections[team]) == 0 {
				delete(wr.connections, team)
			}
		}
	}
	return roleMapConns, nil
}

func (wr *WaitingRoom) GetConnections() ([]model.Connection, error) {
	wr.mu.Lock()
	defer wr.mu.Unlock()

	connections := []model.Connection{}
	ready := false
	if wr.selfMatch {
		for team, conns := range wr.connections {
			if len(conns) >= wr.agentCount {
				connections = append(connections, conns[:wr.agentCount]...)
				wr.connections[team] = conns[wr.agentCount:]
				if len(wr.connections[team]) == 0 {
					delete(wr.connections, team)
				}
				ready = true
				break
			}
		}
	} else {
		if len(wr.connections) >= wr.agentCount {
			var teams []string
			for team := range wr.connections {
				teams = append(teams, team)
			}
			rand.Shuffle(len(teams), func(i, j int) {
				teams[i], teams[j] = teams[j], teams[i]
			})
			for _, team := range teams[:wr.agentCount] {
				conns := wr.connections[team]
				connections = append(connections, conns[0])
				wr.connections[team] = conns[1:]
				if len(wr.connections[team]) == 0 {
					delete(wr.connections, team)
				}
			}
			ready = true
		}
	}

	if !ready {
		return nil, errors.New("待機部屋内の接続が不足しています")
	}
	slog.Info("マッチの接続を取得しました")
	return connections, nil
}
