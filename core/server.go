package core

import (
	"log/slog"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nharu-0630/aiwolf-nlp-server/logic"
	"github.com/nharu-0630/aiwolf-nlp-server/model"
	"github.com/nharu-0630/aiwolf-nlp-server/service"
)

type Server struct {
	host            string
	port            int
	upgrader        websocket.Upgrader
	waitingRoom     *WaitingRoom
	matchOptimizer  *MatchOptimizer
	games           []*logic.Game
	mu              sync.RWMutex
	analysisService *service.AnalysisServiceImpl
	apiService      *service.ApiService
}

func NewServer(host string, port int) *Server {
	analysisService := service.NewAnalysisService()
	return &Server{
		host: host,
		port: port,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		waitingRoom:     NewWaitingRoom(),
		matchOptimizer:  NewMatchOptimizer(),
		games:           make([]*logic.Game, 0),
		analysisService: analysisService,
		apiService:      service.NewApiService(analysisService),
	}
}

func (s *Server) Run() {
	router := gin.Default()

	router.GET("/ws", func(c *gin.Context) {
		s.handleConnections(c.Writer, c.Request)
	})

	s.apiService.RegisterRoutes(router)

	slog.Info("サーバを起動しました", "host", s.host, "port", s.port)
	err := router.Run(s.host + ":" + strconv.Itoa(s.port))
	if err != nil {
		slog.Error("サーバの起動に失敗しました", "error", err)
		return
	}
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
	game := logic.NewGame(gameSetting, connections, s.analysisService)

	s.mu.Lock()
	s.games = append(s.games, game)
	s.mu.Unlock()

	go func() {
		game.Start()
	}()
}
