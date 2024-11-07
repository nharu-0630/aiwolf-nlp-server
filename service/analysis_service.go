package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nharu-0630/aiwolf-nlp-server/model"
)

type AnalysisService interface {
	TrackStartGame(id string, agents []*model.Agent)
	TrackEndGame(id string, winSide model.Team)
	TrackStartRequest(id string, agent model.Agent, request interface{})
	TrackEndRequest(id string, agent model.Agent, response interface{})
}

type AnalysisServiceImpl struct {
	gamesData map[string]*GameData
	outputDir string
}

type GameData struct {
	agents       []interface{}
	winSide      model.Team
	entries      []interface{}
	timestampMap map[string]int64
	requestMap   map[string]interface{}
}

func NewAnalysisService(config model.Config) *AnalysisServiceImpl {
	return &AnalysisServiceImpl{
		gamesData: make(map[string]*GameData),
		outputDir: config.AnalysisServiceOutputDir,
	}
}

func (a *AnalysisServiceImpl) TrackStartGame(id string, agents []*model.Agent) {
	gameData := &GameData{
		agents:       make([]interface{}, 0),
		entries:      make([]interface{}, 0),
		timestampMap: make(map[string]int64),
		requestMap:   make(map[string]interface{}),
		winSide:      model.T_NONE,
	}
	for _, agent := range agents {
		gameData.agents = append(gameData.agents,
			map[string]interface{}{
				"idx":  agent.Idx,
				"team": agent.Team,
				"name": agent.Name,
				"role": agent.Role,
			},
		)
	}
	a.gamesData[id] = gameData
}

func (a *AnalysisServiceImpl) TrackEndGame(id string, winSide model.Team) {
	if gameData, exists := a.gamesData[id]; exists {
		gameData.winSide = winSide
		a.saveGameData(id)
	}
}

func (a *AnalysisServiceImpl) TrackStartRequest(id string, agent model.Agent, packet model.Packet) {
	if gameData, exists := a.gamesData[id]; exists {
		gameData.timestampMap[agent.Name] = time.Now().UnixNano()
		gameData.requestMap[agent.Name] = packet
	}
}

func (a *AnalysisServiceImpl) TrackEndRequest(id string, agent model.Agent, response string, err error) {
	if gameData, exists := a.gamesData[id]; exists {
		timestamp := time.Now().UnixNano()
		entry := map[string]interface{}{
			"agent":              agent.String(),
			"request_timestamp":  gameData.timestampMap[agent.Name] / 1e6,
			"response_timestamp": timestamp / 1e6,
		}
		if request, ok := gameData.requestMap[agent.Name]; ok {
			entry["request"] = request
		}
		if response != "" {
			entry["response"] = response
		}
		if err != nil {
			entry["error"] = err
		}
		gameData.entries = append(gameData.entries, entry)
		delete(gameData.timestampMap, agent.Name)
		delete(gameData.requestMap, agent.Name)

		a.saveGameData(id)
	}
}

func (a *AnalysisServiceImpl) saveGameData(id string) {
	if gameData, exists := a.gamesData[id]; exists {
		game := map[string]interface{}{
			"game_id":  id,
			"win_side": gameData.winSide,
			"agents":   gameData.agents,
			"entries":  gameData.entries,
		}
		jsonData, err := json.Marshal(game)
		if err != nil {
			return
		}
		if _, err := os.Stat(a.outputDir); os.IsNotExist(err) {
			os.Mkdir(a.outputDir, 0755)
		}
		filePath := filepath.Join(a.outputDir, fmt.Sprintf("%s.json", id))
		file, err := os.Create(filePath)
		if err != nil {
			return
		}
		defer file.Close()
		file.Write(jsonData)
	}
}
