package core

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/nharu-0630/aiwolf-nlp-server/config"
	"github.com/nharu-0630/aiwolf-nlp-server/model"
)

type Server struct {
	host        string
	port        int
	upgrader    websocket.Upgrader
	connections map[string][]model.Connection
	games       map[*Game]*websocket.Conn
	mu          sync.RWMutex
}

func NewServer(host string, port int) *Server {
	return &Server{
		host: host,
		port: port,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		connections: make(map[string][]model.Connection),
		games:       make(map[*Game]*websocket.Conn),
	}
}

func (s *Server) Run() {
	http.HandleFunc("/ws", s.handleConnections)
	http.HandleFunc("/health", s.handleHealthCheck)
	slog.Info("サーバを起動しました", "host", s.host, "port", s.port)
	err := http.ListenAndServe(s.host+":"+strconv.Itoa(s.port), nil)
	if err != nil {
		slog.Error("サーバの起動に失敗しました", "error", err)
		return
	}
}

func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	s.mu.RLock()
	defer s.mu.RUnlock()

	progresses := []map[string]interface{}{}
	for game := range s.games {
		progress := map[string]interface{}{
			"id":  game.ID,
			"day": game.CurrentDay,
		}
		progresses = append(progresses, progress)
	}
	status := map[string]interface{}{
		"status":   "running",
		"progress": progresses,
	}
	json.NewEncoder(w).Encode(status)
}

func (s *Server) handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Panic(err)
		return
	}

	req, err := json.Marshal(model.Packet{
		Request: &model.R_NAME,
	})
	if err != nil {
		slog.Error("NAMEパケットの作成に失敗しました", "error", err)
		return
	}
	err = conn.WriteMessage(websocket.TextMessage, req)
	if err != nil {
		slog.Error("NAMEパケットの送信に失敗しました", "error", err)
		return
	}
	slog.Info("NAMEパケットを送信しました", "remote_addr", conn.RemoteAddr().String())
	_, res, err := conn.ReadMessage()
	if err != nil {
		slog.Error("NAMEリクエストの受信に失敗しました", "error", err)
		return
	}
	name := strings.TrimRight(string(res), "\n")
	team := strings.TrimRight(name, "1234567890")
	connection := model.Connection{
		Name: name,
		Conn: conn,
	}

	s.mu.Lock()
	s.connections[team] = append(s.connections[team], connection)
	s.mu.Unlock()
	slog.Info("新しいクライアントが接続しました", "remote_addr", conn.RemoteAddr().String(), "team", team)

	if len(s.connections) == config.AGENT_COUNT_PER_GAME {
		slog.Info("ゲームを開始します")
		gameSetting, err := model.NewSettings()
		if err != nil {
			slog.Error("ゲーム設定の作成に失敗しました", "error", err)
			return
		}
		s.mu.Lock()
		connections := []model.Connection{}
		for team, conns := range s.connections {
			connections = append(connections, conns[0])
			s.connections[team] = conns[1:]
		}
		game := NewGame(gameSetting, connections)
		s.games[game] = conn
		s.mu.Unlock()
		go func() {
			game.Start()
			s.mu.Lock()
			delete(s.games, game)
			s.mu.Unlock()
		}()
	}
}
