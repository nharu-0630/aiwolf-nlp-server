package service

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kano-lab/aiwolf-nlp-server/model"
)

type ApiService struct {
	analysisService    *AnalysisService
	publishRunningGame bool
	mu                 sync.Mutex
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
}

func (api *ApiService) handleGameIDs(c *gin.Context) {
	api.mu.Lock()
	defer api.mu.Unlock()

	gameIDs := api.getGameIDs()

	c.JSON(200, gin.H{
		"games": gameIDs,
	})
}

func (api *ApiService) handleGameData(c *gin.Context) {
	api.mu.Lock()
	defer api.mu.Unlock()

	gameID := c.Query("id")
	if gameID == "" {
		c.JSON(400, gin.H{"error": "id is required"})
		return
	}

	gameData, exists := api.analysisService.gamesData[gameID]
	if !exists {
		c.JSON(404, gin.H{"error": "game not found"})
		return
	}

	if !api.publishRunningGame && !api.analysisService.endGameStatus[gameID] {
		c.JSON(403, gin.H{"error": "game is running"})
		return
	}

	responseData := gin.H{
		"game_id":  gameID,
		"win_side": gameData.winSide,
		"agents":   gameData.agents,
		"entries":  gameData.entries,
	}

	c.JSON(200, responseData)
}

func (api *ApiService) getGameIDs() []string {
	var gameIDs []string
	for gameID := range api.analysisService.gamesData {
		gameIDs = append(gameIDs, gameID)
	}
	return gameIDs
}
