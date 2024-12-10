package logic

import (
	"fmt"
	"log/slog"

	"github.com/kano-lab/aiwolf-nlp-server/model"
	"github.com/kano-lab/aiwolf-nlp-server/service"
	"github.com/kano-lab/aiwolf-nlp-server/util"
	"github.com/oklog/ulid/v2"
)

type Game struct {
	Config               *model.Config
	ID                   string
	Settings             *model.Settings
	Agents               []*model.Agent
	CurrentDay           int
	GameStatuses         map[int]*model.GameStatus
	LastTalkIdxMap       map[*model.Agent]int
	LastWhisperIdxMap    map[*model.Agent]int
	IsFinished           bool
	AnalysisService      *service.AnalysisService
	DeprecatedLogService *service.DeprecatedLogService
}

func NewGame(config *model.Config, settings *model.Settings, conns []model.Connection) *Game {
	id := ulid.Make().String()
	agents := util.CreateAgents(conns, settings.RoleNumMap)
	gameStatus := model.NewInitializeGameStatus(agents)
	gameStatuses := make(map[int]*model.GameStatus)
	gameStatuses[0] = &gameStatus
	slog.Info("ゲームを作成しました", "id", id)
	return &Game{
		Config:            config,
		ID:                id,
		Settings:          settings,
		Agents:            agents,
		CurrentDay:        0,
		GameStatuses:      gameStatuses,
		LastTalkIdxMap:    make(map[*model.Agent]int),
		LastWhisperIdxMap: make(map[*model.Agent]int),
		IsFinished:        false,
	}
}

func NewGameWithRole(config *model.Config, settings *model.Settings, roleMapConns map[model.Role][]model.Connection) *Game {
	id := ulid.Make().String()
	agents := util.CreateAgentsWithRole(roleMapConns)
	gameStatus := model.NewInitializeGameStatus(agents)
	gameStatuses := make(map[int]*model.GameStatus)
	gameStatuses[0] = &gameStatus
	slog.Info("ゲームを作成しました", "id", id)
	return &Game{
		Config:            config,
		ID:                id,
		Settings:          settings,
		Agents:            agents,
		CurrentDay:        0,
		GameStatuses:      gameStatuses,
		LastTalkIdxMap:    make(map[*model.Agent]int),
		LastWhisperIdxMap: make(map[*model.Agent]int),
		IsFinished:        false,
	}
}

func (g *Game) SetAnalysisService(analysisService *service.AnalysisService) {
	g.AnalysisService = analysisService
}

func (g *Game) SetDeprecatedLogService(deprecatedLogService *service.DeprecatedLogService) {
	g.DeprecatedLogService = deprecatedLogService
}

func (g *Game) Start() model.Team {
	slog.Info("ゲームを開始します", "id", g.ID)
	if g.AnalysisService != nil {
		g.AnalysisService.TrackStartGame(g.ID, g.Agents)
	}
	if g.DeprecatedLogService != nil {
		g.DeprecatedLogService.TrackStartGame(g.ID, g.Agents)
	}
	g.requestToEveryone(model.R_INITIALIZE)
	var winSide model.Team = model.T_NONE
	for winSide == model.T_NONE && util.CalcHasErrorAgents(g.Agents) < int(float64(len(g.Agents))*g.Config.Game.MaxContinueErrorRatio) {
		g.progressDay()
		g.progressNight()
		gameStatus := g.GameStatuses[g.CurrentDay].NextDay()
		g.GameStatuses[g.CurrentDay+1] = &gameStatus
		g.CurrentDay++
		slog.Info("日付が進みました", "id", g.ID, "day", g.CurrentDay)
		winSide = util.CalcWinSideTeam(gameStatus.StatusMap)
	}
	if winSide == model.T_NONE {
		slog.Warn("エラーが多発したため、ゲームを終了します", "id", g.ID)
	}
	g.requestToEveryone(model.R_FINISH)
	if g.DeprecatedLogService != nil {
		for _, agent := range g.Agents {
			g.DeprecatedLogService.AppendLog(g.ID, fmt.Sprintf("%d,status,%d,%s,%s,%s", g.CurrentDay, agent.Idx, agent.Role.Name, g.GameStatuses[g.CurrentDay].StatusMap[*agent].String(), agent.Name))
		}
		villagers, werewolves := util.CountAliveTeams(g.GameStatuses[g.CurrentDay].StatusMap)
		g.DeprecatedLogService.AppendLog(g.ID, fmt.Sprintf("%d,result,%d,%d,%s", g.CurrentDay, villagers, werewolves, winSide))
	}
	g.closeAllAgents()
	if g.AnalysisService != nil {
		g.AnalysisService.TrackEndGame(g.ID, winSide)
	}
	if g.DeprecatedLogService != nil {
		g.DeprecatedLogService.TrackEndGame(g.ID)
	}
	slog.Info("ゲームが終了しました", "id", g.ID, "winSide", winSide)
	g.IsFinished = true
	return winSide
}

func (g *Game) progressDay() {
	slog.Info("昼を開始します", "id", g.ID, "day", g.CurrentDay)
	g.requestToEveryone(model.R_DAILY_INITIALIZE)
	if g.DeprecatedLogService != nil {
		for _, agent := range g.Agents {
			g.DeprecatedLogService.AppendLog(g.ID, fmt.Sprintf("%d,status,%d,%s,%s,%s", g.CurrentDay, agent.Idx, agent.Role.Name, g.GameStatuses[g.CurrentDay].StatusMap[*agent].String(), agent.Name))
		}
	}
	if g.Settings.IsTalkOnFirstDay && g.CurrentDay == 0 {
		g.doWhisper()
	}
	g.doTalk()
	slog.Info("昼を終了します", "id", g.ID, "day", g.CurrentDay)
}

func (g *Game) progressNight() {
	slog.Info("夜を開始します", "id", g.ID, "day", g.CurrentDay)
	g.requestToEveryone(model.R_DAILY_FINISH)
	if g.Settings.IsTalkOnFirstDay && g.CurrentDay == 0 {
		g.doWhisper()
	}
	if g.CurrentDay != 0 {
		g.doExecution()
	}
	g.doDivine()
	if g.CurrentDay != 0 {
		g.doWhisper()
		g.doGuard()
		g.doAttack()
	}
	slog.Info("夜を終了します", "id", g.ID, "day", g.CurrentDay)
}
