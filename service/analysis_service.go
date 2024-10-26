package service

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/nharu-0630/aiwolf-nlp-server/model"
)

type AnalysisService interface {
	TrackStartGame(id string)
	TrackEndGame()
	TrackStartRequest(agent model.Agent, request interface{})
	TrackEndRequest(agent model.Agent, response interface{})
}

type AnalysisServiceImpl struct {
	id           string
	agents       []interface{}
	winSide      model.Team
	entries      []interface{}
	timestampMap map[string]int64
	requestMap   map[string]interface{}
}

func NewAnalysisService() *AnalysisServiceImpl {
	return &AnalysisServiceImpl{
		agents:       make([]interface{}, 0),
		entries:      make([]interface{}, 0),
		timestampMap: make(map[string]int64),
		requestMap:   make(map[string]interface{}),
	}
}

func (a *AnalysisServiceImpl) TrackStartGame(id string, agents []*model.Agent) {
	a.id = id
	for _, agent := range agents {
		a.agents = append(a.agents,
			map[string]interface{}{
				"idx":  agent.Idx,
				"team": agent.Team,
				"name": agent.Name,
				"role": agent.Role,
			},
		)
	}
}

func (a *AnalysisServiceImpl) TrackEndGame(winSide model.Team) {
	a.winSide = winSide
	entry := map[string]interface{}{
		"game_id":  a.id,
		"win_side": a.winSide,
		"agents":   a.agents,
		"entries":  a.entries,
	}
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return
	}
	file, err := os.Create(fmt.Sprintf("game_%s.json", a.id))
	if err != nil {
		return
	}
	defer file.Close()
	file.Write(jsonData)
}

func (a *AnalysisServiceImpl) TrackStartRequest(agent model.Agent, packet model.Packet) {
	a.timestampMap[agent.Name] = time.Now().UnixNano()
	a.requestMap[agent.Name] = packet
}

func (a *AnalysisServiceImpl) TrackEndRequest(agent model.Agent, response string, err error) {
	timestamp := time.Now().UnixNano()
	entry := map[string]interface{}{
		"agent":              agent.String(),
		"request_timestamp":  a.timestampMap[agent.Name] / 1e6,
		"response_timestamp": timestamp / 1e6,
	}
	if request, ok := a.requestMap[agent.Name]; ok {
		entry["request"] = request
	}
	if response != "" {
		entry["response"] = response
	}
	if err != nil {
		entry["error"] = err
	}
	a.entries = append(a.entries, entry)
	delete(a.timestampMap, agent.Name)
	delete(a.requestMap, agent.Name)
}
