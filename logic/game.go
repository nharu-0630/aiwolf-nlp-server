package logic

import (
	"log/slog"

	"github.com/nharu-0630/aiwolf-nlp-server/config"
	"github.com/nharu-0630/aiwolf-nlp-server/model"
	"github.com/nharu-0630/aiwolf-nlp-server/service"
	"github.com/nharu-0630/aiwolf-nlp-server/util"
	"github.com/oklog/ulid/v2"
)

type Game struct {
	ID                string                       // ID
	Settings          model.Settings               // 設定
	Agents            []*model.Agent               // エージェント
	CurrentDay        int                          // 現在の日付
	GameStatuses      map[int]*model.GameStatus    // 日ごとのゲーム状態
	LastTalkIdxMap    map[*model.Agent]int         // 最後に送信したトークのインデックス
	LastWhisperIdxMap map[*model.Agent]int         // 最後に送信した囁きのインデックス
	AnalysisService   *service.AnalysisServiceImpl // 分析サービス
}

func NewGame(settings model.Settings, conns []model.Connection, analysisService *service.AnalysisServiceImpl) *Game {
	id := ulid.Make().String()
	agents := util.CreateAgents(conns, settings.RoleNumMap)
	gameStatus := model.NewInitializeGameStatus(agents)
	gameStatuses := make(map[int]*model.GameStatus)
	gameStatuses[0] = &gameStatus
	slog.Info("ゲームを作成しました", "id", id)
	return &Game{
		ID:                id,
		Settings:          settings,
		Agents:            agents,
		CurrentDay:        0,
		GameStatuses:      gameStatuses,
		LastTalkIdxMap:    make(map[*model.Agent]int),
		LastWhisperIdxMap: make(map[*model.Agent]int),
		AnalysisService:   analysisService,
	}
}

func NewGameWithRole(settings model.Settings, roleMapConns map[model.Role][]model.Connection, analysisService *service.AnalysisServiceImpl) *Game {
	id := ulid.Make().String()
	agents := util.CreateAgentsWithRole(roleMapConns)
	gameStatus := model.NewInitializeGameStatus(agents)
	gameStatuses := make(map[int]*model.GameStatus)
	gameStatuses[0] = &gameStatus
	slog.Info("ゲームを作成しました", "id", id)
	return &Game{
		ID:                id,
		Settings:          settings,
		Agents:            agents,
		CurrentDay:        0,
		GameStatuses:      gameStatuses,
		LastTalkIdxMap:    make(map[*model.Agent]int),
		LastWhisperIdxMap: make(map[*model.Agent]int),
		AnalysisService:   analysisService,
	}
}

func (g *Game) Start() {
	slog.Info("ゲームを開始します", "id", g.ID)
	g.AnalysisService.TrackStartGame(g.ID, g.Agents)
	var winSide model.Team = model.T_NONE
	for winSide == model.T_NONE && util.CalcHasErrorAgents(g.Agents) < int(float64(len(g.Agents))*config.MAX_HAS_ERROR_AGENTS_RATIO) {
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
	g.closeAllAgents()
	g.AnalysisService.TrackEndGame(g.ID, winSide)
	slog.Info("ゲームが終了しました", "id", g.ID, "winSide", winSide)	
}

func (g *Game) progressDay() {
	slog.Info("昼を開始します", "id", g.ID, "day", g.CurrentDay)
	g.requestToEveryone(model.R_DAILY_INITIALIZE)
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
