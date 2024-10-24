package service

import (
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
	id             string
	agentTimestamp map[string]int64
	agentRequest   map[string]interface{}
	analysis       []interface{}
}

func NewAnalysisService() *AnalysisServiceImpl {
	return &AnalysisServiceImpl{
		agentTimestamp: make(map[string]int64),
		agentRequest:   make(map[string]interface{}),
		analysis:       make([]interface{}, 0),
	}
}

func (a *AnalysisServiceImpl) TrackStartGame(id string) {
	a.id = id
}

func (a *AnalysisServiceImpl) TrackStartRequest(agent model.Agent, packet model.Packet) {
	a.agentTimestamp[agent.Name] = time.Now().UnixNano()
	a.agentRequest[agent.Name] = packet
}

func (a *AnalysisServiceImpl) TrackEndRequest(agent model.Agent, response string, err error) {
	timestamp := time.Now().UnixNano()
	a.analysis = append(a.analysis, map[string]interface{}{
		"agent":    agent.Name,
		"request":  a.agentTimestamp[agent.Name],
		"duration": timestamp - a.agentTimestamp[agent.Name],
		"response": response,
		"error":    err,
	})
	delete(a.agentTimestamp, agent.Name)
	delete(a.agentRequest, agent.Name)
}
