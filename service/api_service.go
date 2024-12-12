package service

import (
	"maps"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/kano-lab/aiwolf-nlp-server/model"
)

type ApiService struct {
	analysisService    *AnalysisService
	publishRunningGame bool
}

func NewApiService(analysisService *AnalysisService, config model.Config) *ApiService {
	return &ApiService{
		analysisService:    analysisService,
		publishRunningGame: config.ApiService.PublishRunningGame,
	}
}

func (api *ApiService) RegisterRoutes(router *gin.Engine) {
	router.GET("/api/games", api.handleGameIDs)
	router.GET("/api/game", api.handleGameData)
	router.GET("/api/teams", api.handleTeams)
}

func (api *ApiService) handleGameIDs(c *gin.Context) {
	if len(api.analysisService.gamesData) == 0 {
		c.JSON(200, gin.H{"games": []string{}})
		return
	}
	c.JSON(200, gin.H{
		"games": slices.Collect(maps.Keys(api.analysisService.gamesData)),
	})
}

func (api *ApiService) handleGameData(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "id is required"})
		return
	}
	data, exists := api.analysisService.gamesData[id]
	if !exists {
		c.JSON(404, gin.H{"error": "game not found"})
		return
	}
	if !api.publishRunningGame && !api.analysisService.endGameStatus[id] {
		c.JSON(403, gin.H{"error": "game is running"})
		return
	}
	resp := gin.H{
		"game_id":  id,
		"win_side": data.winSide,
		"agents":   data.agents,
		"entries":  data.entries,
	}
	c.JSON(200, resp)
}

func (api *ApiService) handleTeams(c *gin.Context) {
	c.JSON(200, "Not implemented")
	return

	resp := gin.H{}
	for id, data := range api.analysisService.gamesData {
		if api.analysisService.endGameStatus[id] {
			continue
		}
		teams := []string{}
		for _, agent := range data.agents {
			if agentMap, ok := agent.(map[string]interface{}); ok {
				team := agentMap["team"].(string)
				teams = append(teams, team)
			}
		}
		for _, team := range teams {
			if _, exists := resp[team]; !exists {
				resp[team] = []string{}
			}
			resp[team] = append(resp[team].([]string), id)
		}
	}
	c.JSON(200, resp)
}
