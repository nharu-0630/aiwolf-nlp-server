package core

import (
	"log/slog"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/kano-lab/aiwolf-nlp-server/logic"
	"github.com/kano-lab/aiwolf-nlp-server/model"
	"github.com/kano-lab/aiwolf-nlp-server/service"
	"github.com/kano-lab/aiwolf-nlp-server/util"
)

type Server struct {
	config          model.Config
	upgrader        websocket.Upgrader
	waitingRoom     *WaitingRoom
	matchOptimizer  *MatchOptimizer
	gameSettings    *model.Settings
	games           []*logic.Game
	mu              sync.RWMutex
	analysisService *service.AnalysisService
	apiService      *service.ApiService
}

func NewServer(config model.Config) *Server {
	server := &Server{
		config: config,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		waitingRoom:     NewWaitingRoom(config),
		games:           make([]*logic.Game, 0),
		analysisService: service.NewAnalysisService(config),
	}
	gameSettings, err := model.NewSettings(config)
	if err != nil {
		slog.Error("ゲーム設定の作成に失敗しました", "error", err)
		return nil
	}
	server.gameSettings = gameSettings
	if config.MatchOptimizer.Enable {
		matchOptimizer, err := NewMatchOptimizer(config)
		if err != nil {
			slog.Error("マッチオプティマイザの作成に失敗しました", "error", err)
			return nil
		}
		server.matchOptimizer = matchOptimizer
	}
	if config.ApiService.Enable {
		server.apiService = service.NewApiService(server.analysisService, config)
	}
	return server
}

func (s *Server) Run() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/ws", func(c *gin.Context) {
		s.handleConnections(c.Writer, c.Request)
	})

	if s.config.ApiService.Enable {
		s.apiService.RegisterRoutes(router)
	}

	slog.Info("サーバを起動しました", "host", s.config.Server.WebSocket.Host, "port", s.config.Server.WebSocket.Port)
	err := router.Run(s.config.Server.WebSocket.Host + ":" + strconv.Itoa(s.config.Server.WebSocket.Port))
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

	var game *logic.Game
	if s.config.MatchOptimizer.Enable {
		for team := range s.waitingRoom.connections {
			s.matchOptimizer.UpdateTeam(team)
		}
		roleMapConns, err := s.waitingRoom.GetConnectionsWithMatchOptimizer(s.matchOptimizer.GetScheduledMatchesWithTeam())
		if err != nil {
			slog.Error("マッチオプティマイザからの接続情報の取得に失敗しました", "error", err)
			return
		}
		game = logic.NewGameWithRole(&s.config, s.gameSettings, roleMapConns, s.analysisService)
	} else {
		if !s.waitingRoom.IsReady() {
			return
		}

		connections := s.waitingRoom.GetConnections()
		game = logic.NewGame(&s.config, s.gameSettings, connections, s.analysisService)
	}

	s.mu.Lock()
	s.games = append(s.games, game)
	s.mu.Unlock()

	go func() {
		winSide := game.Start()
		if winSide != model.T_NONE && s.config.MatchOptimizer.Enable {
			s.matchOptimizer.addEndedMatch(util.GetRoleTeamNamesMap(game.Agents))
		}
	}()
}
