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
	upgrader websocket.Upgrader
	clients  []*websocket.Conn
}

func NewServer() *Server {
	return &Server{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients: make([]*websocket.Conn, 0),
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

	s.clients = append(s.clients, conn)
	log.Printf("New client connected. Total clients: %d", len(s.clients))

	if len(s.clients) == config.GAME_AGENT_COUNT {
		log.Println(strconv.Itoa(config.GAME_AGENT_COUNT) + " clients connected, starting game...")
		gameSetting := model.Settings{}
		game := NewGame(gameSetting, s.clients)
		game.Start()

		log.Printf("Game created with ID: %s", game.GameID)
		s.clients = make([]*websocket.Conn, 0)
	}
}
