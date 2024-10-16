package core

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/nharu-0630/aiwolf-nlp-server/config"
	"github.com/nharu-0630/aiwolf-nlp-server/model"
)

type Server struct {
	upgrader    websocket.Upgrader
	connections []*websocket.Conn
}

func NewServer() *Server {
	return &Server{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		connections: make([]*websocket.Conn, 0),
	}
}

func (s *Server) Run() {
	http.HandleFunc("/ws", s.handleConnections)
	log.Println("WebSocket server started on :" + strconv.Itoa(config.WEBSOCKET_PORT))
	err := http.ListenAndServe(":"+strconv.Itoa(config.WEBSOCKET_PORT), nil)
	if err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}

func (s *Server) handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	s.connections = append(s.connections, conn)
	log.Printf("New client connected. Total clients: %d", len(s.connections))

	if len(s.connections) == config.GAME_AGENT_COUNT {
		log.Println(strconv.Itoa(config.GAME_AGENT_COUNT) + " clients connected, starting game...")
		gameSetting := model.Settings{
			MaxTalk:          config.MAX_TALK_COUNT_PER_AGENT,
			MaxTalkTurn:      config.MAX_TALK_COUNT,
			MaxWhisper:       config.MAX_WHISPER_COUNT_PER_AGENT,
			MaxWhisperTurn:   config.MAX_WHISPER_COUNT,
			MaxSkip:          config.MAX_SKIP_COUNT,
			IsEnableNoAttack: config.IS_ENABLE_NO_ATTACK,
			IsVoteVisible:    config.IS_VOTE_VISIBLE,
			IsTalkOnFirstDay: config.IS_TALK_ON_FIRST_DAY,
			ResponseTimeout:  config.RESPONSE_TIMEOUT,
			ActionTimeout:    config.ACTION_TIMEOUT,
			MaxRevote:        config.MAX_REVOTE_COUNT,
			MaxAttackRevote:  config.MAX_ATTACK_REVOTE_COUNT,
		}
		game := NewGame(gameSetting, s.connections)
		game.Start()

		log.Printf("Game created with ID: %s", game.ID)
		s.connections = make([]*websocket.Conn, 0)
	}
}
