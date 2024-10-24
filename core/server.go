package core

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/nharu-0630/aiwolf-nlp-server/logic"
	"github.com/nharu-0630/aiwolf-nlp-server/model"
)

type Server struct {
	host        string
	port        int
	upgrader    websocket.Upgrader
	waitingRoom *WaitingRoom
	games       []*logic.Game
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
		waitingRoom: NewWaitingRoom(),
		games:       make([]*logic.Game, 0),
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
	for _, game := range s.games {
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
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("クライアントのアップグレードに失敗しました", "error", err)
		return
	}
	connection, err := model.NewConnection(ws)
	if err != nil {
		slog.Error("クライアントの接続に失敗しました", "error", err)
		return
	}
	s.waitingRoom.AddConnection(connection.Team, *connection)

	if !s.waitingRoom.IsReady() {
		return
	}

	gameSetting, err := model.NewSettings()
	if err != nil {
		slog.Error("ゲーム設定の作成に失敗しました", "error", err)
		return
	}
	connections := s.waitingRoom.GetConnections()
	game := logic.NewGame(gameSetting, connections)

	s.mu.Lock()
	s.games = append(s.games, game)
	s.mu.Unlock()

	go func() {
		slog.Info("ゲームを開始します")
		game.Start()
		slog.Info("ゲームが終了しました", "id", game.ID)
	}()
}
