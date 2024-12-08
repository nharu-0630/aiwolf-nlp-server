package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kano-lab/aiwolf-nlp-server/model"
)

type AnalysisService struct {
	gamesData        map[string]*GameData
	outputDir        string
	templateFilename string
	endGameStatus    map[string]bool
}

type GameData struct {
	id           string
	filename     string
	agents       []interface{}
	winSide      model.Team
	entries      []interface{}
	timestampMap map[string]int64
	requestMap   map[string]interface{}
}

func NewAnalysisService(config model.Config) *AnalysisService {
	return &AnalysisService{
		gamesData:        make(map[string]*GameData),
		outputDir:        config.AnalysisService.OutputDir,
		templateFilename: config.AnalysisService.Filename,
		endGameStatus:    make(map[string]bool),
	}
}

func (a *AnalysisService) TrackStartGame(id string, agents []*model.Agent) {
	gameData := &GameData{
		id:           id,
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
	filename := strings.ReplaceAll(a.templateFilename, "{game_id}", gameData.id)
	filename = strings.ReplaceAll(filename, "{timestamp}", fmt.Sprintf("%d", time.Now().Unix()))
	teams := make(map[string]struct{})
	for _, agent := range gameData.agents {
		team := agent.(map[string]interface{})["team"].(string)
		teams[team] = struct{}{}
	}
	teamStr := ""
	for team := range teams {
		if teamStr != "" {
			teamStr += "_"
		}
		teamStr += team
	}
	filename = strings.ReplaceAll(filename, "{teams}", teamStr)
	gameData.filename = filename

	a.gamesData[id] = gameData
	a.endGameStatus[id] = false
}

func (a *AnalysisService) TrackEndGame(id string, winSide model.Team) {
	if gameData, exists := a.gamesData[id]; exists {
		gameData.winSide = winSide
		a.endGameStatus[id] = true
		a.saveGameData(id)
	}
}

func (a *AnalysisService) TrackStartRequest(id string, agent model.Agent, packet model.Packet) {
	if gameData, exists := a.gamesData[id]; exists {
		gameData.timestampMap[agent.Name] = time.Now().UnixNano()
		gameData.requestMap[agent.Name] = packet
	}
}

func (a *AnalysisService) TrackEndRequest(id string, agent model.Agent, response string, err error) {
	if gameData, exists := a.gamesData[id]; exists {
		timestamp := time.Now().UnixNano()
		entry := map[string]interface{}{
			"agent":              agent.String(),
			"request_timestamp":  gameData.timestampMap[agent.Name] / 1e6,
			"response_timestamp": timestamp / 1e6,
		}
		if request, ok := gameData.requestMap[agent.Name]; ok {
			jsonData, err := json.Marshal(request)
			if err == nil {
				entry["request"] = string(jsonData)
			}
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

func (a *AnalysisService) saveGameData(id string) {
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
		filePath := filepath.Join(a.outputDir, fmt.Sprintf("%s.json", gameData.filename))
		file, err := os.Create(filePath)
		if err != nil {
			return
		}
		defer file.Close()
		file.Write(jsonData)
	}
}
