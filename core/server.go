package core

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/nharu-0630/aiwolf-nlp-server/config"
	"github.com/nharu-0630/aiwolf-nlp-server/model"
)

type Server struct {
	upgrader    websocket.Upgrader
	connections []*websocket.Conn
	games       map[*Game]*websocket.Conn
	mu          sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		connections: make([]*websocket.Conn, 0),
		games:       make(map[*Game]*websocket.Conn),
	}
}

func (s *Server) Run() {
	http.HandleFunc("/ws", s.handleConnections)
	http.HandleFunc("/health", s.handleHealthCheck)
	slog.Info("サーバを起動しました", "host", config.WEBSOCKET_HOST, "port", config.WEBSOCKET_PORT)
	err := http.ListenAndServe(config.WEBSOCKET_HOST+":"+strconv.Itoa(config.WEBSOCKET_PORT), nil)
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

	s.mu.Lock()
	s.connections = append(s.connections, conn)
	s.mu.Unlock()
	slog.Info("新しいクライアントが接続しました", "remote_addr", conn.RemoteAddr().String())

	if len(s.connections) == config.AGENT_COUNT_PER_GAME {
		slog.Info("ゲームを開始します")
		gameSetting, err := model.NewSettings()
		if err != nil {
			slog.Error("ゲーム設定の作成に失敗しました", "error", err)
			return
		}
		s.mu.Lock()
		game := NewGame(gameSetting, s.connections)
		s.games[game] = conn
		s.mu.Unlock()
		go func() {
			game.Start()
			s.mu.Lock()
			s.connections = append(s.connections, s.games[game])
			delete(s.games, game)
			s.mu.Unlock()
		}()
		s.mu.Lock()
		s.connections = make([]*websocket.Conn, 0)
		s.mu.Unlock()
	}
}
