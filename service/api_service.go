package service

import (
	"sync"

	"github.com/gin-gonic/gin"
)

type ApiService struct {
	analysisService *AnalysisServiceImpl
	mu              sync.Mutex
}

func NewApiService(analysisService *AnalysisServiceImpl) *ApiService {
	return &ApiService{
		analysisService: analysisService,
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
