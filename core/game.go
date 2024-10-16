package core

import (
	"errors"
	"log"
	"math/rand"

	"github.com/dgryski/trifles/uuid"
	"github.com/gorilla/websocket"
	"github.com/nharu-0630/aiwolf-nlp-server/model"
	"github.com/nharu-0630/aiwolf-nlp-server/utils"
)

type Game struct {
	ID                string                    // ID
	Settings          model.Settings            // 設定
	Agents            []*model.Agent            // エージェント
	CurrentDay        int                       // 現在の日付
	GameStatuses      map[int]*model.GameStatus // 日ごとのゲーム状態
	LastTalkIdxMap    map[*model.Agent]int      // 最後に送信したトークのインデックス
	LastWhisperIdxMap map[*model.Agent]int      // 最後に送信した囁きのインデックス
}

func NewGame(settings model.Settings, conns []*websocket.Conn) *Game {
	log.Println("新規ゲームを作成します")
	id := uuid.UUIDv4()
	roles := model.Roles(len(conns))
	if len(roles) == 0 {
		log.Panic("接続数に対応する役職がありません")
	}
	settings.RoleNumMap = model.Roles(len(conns))
	agents := createAgents(conns, roles)
	log.Printf("新規ゲームを作成しました: %s", id)
	gameStatus := model.NewInitializeGameStatus(agents)
	gameStatuses := make(map[int]*model.GameStatus)
	gameStatuses[0] = &gameStatus
	log.Printf("ゲーム設定: %+v", settings)
	log.Printf("エージェント数: %d", len(agents))
	return &Game{
		ID:                id,
		Settings:          settings,
		Agents:            agents,
		CurrentDay:        0,
		GameStatuses:      gameStatuses,
		LastTalkIdxMap:    make(map[*model.Agent]int),
		LastWhisperIdxMap: make(map[*model.Agent]int),
	}
}

func createAgents(conns []*websocket.Conn, roles map[model.Role]int) []*model.Agent {
	agents := make([]*model.Agent, 0)
	for i, conn := range conns {
		role := assignRole(roles)
		agent, err := model.NewAgent(i, role, conn)
		if err != nil {
			log.Panic(err)
		}
		log.Printf("エージェントを作成しました: %s", agent.Name)
		log.Printf("エージェントの役職: %s", agent.Role)
		log.Printf("エージェントの接続先: %s", agent.Connection.RemoteAddr())
		agents = append(agents, agent)
	}
	return agents
}

func assignRole(roles map[model.Role]int) model.Role {
	for r, n := range roles {
		if n > 0 {
			roles[r]--
			return r
		}
	}
	return model.R_VILLAGER
}

func (g *Game) sendRequestToEveryone(request model.Request) {
	for _, agent := range g.Agents {
		g.sendRequest(agent, request)
	}
}

func (g *Game) Start() {
	log.Println("ゲームを開始します")
	for utils.CalcWinSideTeam(g.GameStatuses[g.CurrentDay].StatusMap) == model.T_NONE {
		g.doDay()
		g.doNight()
		gameStatus := g.GameStatuses[g.CurrentDay].NextDay()
		g.GameStatuses[g.CurrentDay+1] = &gameStatus
		g.CurrentDay++
	}
	g.sendRequestToEveryone(model.R_FINISH)
	log.Printf("ゲームが終了しました: %s", utils.CalcWinSideTeam(g.GameStatuses[g.CurrentDay].StatusMap))
}

func (g *Game) doDay() {
	log.Printf("昼のターンを開始します: %d日目", g.CurrentDay)
	g.sendRequestToEveryone(model.R_DAILY_INITIALIZE)
	if g.Settings.IsTalkOnFirstDay && g.CurrentDay == 0 {
		g.doWhisper()
	}
	g.doTalk()
}

func (g *Game) doNight() {
	log.Printf("夜のターンを開始します: %d日目", g.CurrentDay)
	g.sendRequestToEveryone(model.R_DAILY_FINISH)
	if g.Settings.IsTalkOnFirstDay && g.CurrentDay == 0 {
		g.doWhisper()
	}
	if g.CurrentDay != 0 {
		g.handleExecution()
	}
	g.doDivine()
	if g.CurrentDay != 0 {
		g.doWhisper()
		g.doGuard()
		g.handleAttack()
	}
}

func (g *Game) handleExecution() {
	log.Printf("追放フェーズを開始します: %d日目", g.CurrentDay)
	var executed *model.Agent
	candidates := make([]*model.Agent, 0)
	for i := 0; i < g.Settings.MaxRevote; i++ {
		g.doVote()
		candidates = g.getVotedCandidates(g.GameStatuses[g.CurrentDay].VoteList)
		if len(candidates) == 1 {
			executed = candidates[0]
			break
		}
	}
	if executed == nil {
		executed = utils.SelectRandomAgent(candidates)
	}
	if executed != nil {
		g.GameStatuses[g.CurrentDay].StatusMap[*executed] = model.S_DEAD
		g.GameStatuses[g.CurrentDay].ExecutedAgent = executed
		log.Printf("追放されたエージェント: %s", executed.Name)
	}
}

func (g *Game) handleAttack() {
	log.Printf("襲撃フェーズを開始します: %d日目", g.CurrentDay)
	var attacked *model.Agent
	werewolfs := g.getAliveWerewolves()
	if len(werewolfs) > 0 {
		attacked = g.performAttackVote()
		if attacked == nil && !g.Settings.IsEnableNoAttack {
			attacked = utils.SelectRandomAgent(g.getAttackVotedCandidates(g.GameStatuses[g.CurrentDay].AttackVoteList))
		}
		g.finalizeAttack(attacked)
	}
}

func (g *Game) performAttackVote() *model.Agent {
	log.Printf("襲撃投票を開始します: %d日目", g.CurrentDay)
	var attacked *model.Agent
	for i := 0; i < g.Settings.MaxAttackRevote; i++ {
		g.doAttackVote()
		candidates := g.getAttackVotedCandidates(g.GameStatuses[g.CurrentDay].AttackVoteList)
		if len(candidates) == 1 {
			attacked = candidates[0]
			break
		}
	}
	return attacked
}

func (g *Game) finalizeAttack(attacked *model.Agent) {
	if attacked != nil && !g.isGuarded(attacked) {
		g.GameStatuses[g.CurrentDay].StatusMap[*attacked] = model.S_DEAD
		g.GameStatuses[g.CurrentDay].AttackedAgent = attacked
		log.Printf("襲撃されたエージェント: %s", attacked.Name)
	} else if attacked != nil {
		log.Printf("襲撃されたエージェント: %s (Guarded)", attacked.Name)
	} else {
		log.Println("襲撃されたエージェント: なし")
	}
}

func (g *Game) isGuarded(attacked *model.Agent) bool {
	return g.GameStatuses[g.CurrentDay].Guard.Target == *attacked && g.GameStatuses[g.CurrentDay].StatusMap[g.GameStatuses[g.CurrentDay].Guard.Agent] == model.S_ALIVE
}

func (g *Game) doDivine() {
	log.Printf("占いフェーズを開始します: %d日目", g.CurrentDay)
	for _, agent := range g.Agents {
		if agent.Role == model.R_SEER {
			g.performDivination(agent)
			break
		}
	}
}

func (g *Game) performDivination(agent *model.Agent) {
	log.Printf("占いアクションを実行します: %s", agent.Name)
	target, err := g.findAgentByRequest(agent, model.R_DIVINE)
	if err == nil {
		g.GameStatuses[g.CurrentDay].DivineResult = &model.Judge{
			Day:    g.GameStatuses[g.CurrentDay].Day,
			Agent:  *agent,
			Target: *target,
			Result: target.Role.Species,
		}
		log.Printf("占い対象: %s", target.Name)
		log.Printf("占い結果: %s", target.Role.Species)
	}
}

func (g *Game) doGuard() {
	log.Printf("護衛フェーズを開始します: %d日目", g.CurrentDay)
	for _, agent := range g.Agents {
		if agent.Role == model.R_BODYGUARD && g.GameStatuses[g.CurrentDay].ExecutedAgent != agent {
			g.performGuard(agent)
			break
		}
	}
}

func (g *Game) performGuard(agent *model.Agent) {
	log.Printf("護衛アクションを実行します: %s", agent.Name)
	target, err := g.findAgentByRequest(agent, model.R_GUARD)
	if err == nil {
		g.GameStatuses[g.CurrentDay].Guard = &model.Guard{
			Day:    g.GameStatuses[g.CurrentDay].Day,
			Agent:  *agent,
			Target: *target,
		}
		log.Printf("護衛対象: %s", target.Name)
	}
}

func (g *Game) findAgentByName(name string, err error) (*model.Agent, error) {
	if err != nil {
		return nil, err
	}
	for _, agent := range g.Agents {
		if agent.Name == name {
			return agent, nil
		}
	}
	return nil, errors.New("一致するエージェントが見つかりません")
}

func (g *Game) findAgentByRequest(agent *model.Agent, request model.Request) (*model.Agent, error) {
	name, err := g.sendRequest(agent, request)
	return g.findAgentByName(name, err)
}

func (g *Game) getVotedCandidates(voteList []model.Vote) []*model.Agent {
	return g.getCandidates(voteList, func(vote model.Vote) bool {
		return g.GameStatuses[g.CurrentDay].StatusMap[vote.Target] == model.S_ALIVE
	})
}

func (g *Game) getAttackVotedCandidates(voteList []model.Vote) []*model.Agent {
	return g.getCandidates(voteList, func(vote model.Vote) bool {
		return g.GameStatuses[g.CurrentDay].StatusMap[vote.Target] == model.S_ALIVE && vote.Target.Role.Team != model.T_WEREWOLF
	})
}

func (g *Game) getCandidates(voteList []model.Vote, condition func(model.Vote) bool) []*model.Agent {
	counter := make(map[*model.Agent]int)
	for _, vote := range voteList {
		if condition(vote) {
			counter[&vote.Target]++
		}
	}
	return g.getMaxCountCandidates(counter)
}

func (g *Game) getMaxCountCandidates(counter map[*model.Agent]int) []*model.Agent {
	var max int
	for _, count := range counter {
		if count > max {
			max = count
		}
	}
	candidates := make([]*model.Agent, 0)
	for agent, count := range counter {
		if count == max {
			candidates = append(candidates, agent)
		}
	}
	return candidates
}

func (g *Game) doVote() {
	log.Printf("投票フェーズを開始します: %d日目", g.CurrentDay)
	g.GameStatuses[g.CurrentDay].VoteList = g.collectVotes(model.R_VOTE, g.getAliveAgents())
}

func (g *Game) doAttackVote() {
	log.Printf("襲撃投票フェーズを開始します: %d日目", g.CurrentDay)
	g.GameStatuses[g.CurrentDay].AttackVoteList = g.collectVotes(model.R_ATTACK, g.getAliveWerewolves())
}

func (g *Game) collectVotes(request model.Request, agents []*model.Agent) []model.Vote {
	votes := make([]model.Vote, 0)
	for _, agent := range agents {
		target, err := g.findAgentByRequest(agent, request)
		if err != nil {
			continue
		}
		votes = append(votes, model.Vote{
			Day:    g.GameStatuses[g.CurrentDay].Day,
			Agent:  *agent,
			Target: *target,
		})
		log.Printf("投票者: %s", agent.Name)
		log.Printf("投票対象: %s", target.Name)
	}
	return votes
}

func (g *Game) doWhisper() {
	log.Printf("囁きフェーズを開始します: %d日目", g.CurrentDay)
	g.GameStatuses[g.CurrentDay].ResetRemainWhisperMap(g.Settings.MaxWhisper)
	werewolfs := g.getAliveWerewolves()
	if len(werewolfs) < 2 {
		return
	}
	g.performCommunication(werewolfs, model.R_WHISPER, &g.GameStatuses[g.CurrentDay].WhisperList, g.Settings.MaxWhisperTurn)
}

func (g *Game) doTalk() {
	log.Printf("発言フェーズを開始します: %d日目", g.CurrentDay)
	g.GameStatuses[g.CurrentDay].ResetRemainTalkMap(g.Settings.MaxTalk)
	agents := g.getAliveAgents()
	g.performCommunication(agents, model.R_TALK, &g.GameStatuses[g.CurrentDay].TalkList, g.Settings.MaxTalkTurn)
}

func (g *Game) performCommunication(agents []*model.Agent, request model.Request, talkList *[]model.Talk, turnCount int) {
	rand.Shuffle(len(agents), func(i, j int) {
		agents[i], agents[j] = agents[j], agents[i]
	})
	skipCountMap := make(map[*model.Agent]int)
	idx := 0
	for i := 0; i < turnCount; i++ {
		cnt := false
		for _, agent := range agents {
			if g.GameStatuses[g.CurrentDay].RemainTalkMap[*agent] == 0 {
				continue
			}
			text := g.getTalkWhisperText(agent, request, skipCountMap)
			talk := model.Talk{
				Idx:   idx,
				Day:   g.GameStatuses[g.CurrentDay].Day,
				Turn:  i,
				Agent: *agent,
				Text:  text,
			}
			idx++
			*talkList = append(*talkList, talk)
			if text != model.T_OVER {
				cnt = true
			}
			log.Printf("発言者: %s", agent.Name)
			log.Printf("発言内容: %s", text)
		}
		if !cnt {
			break
		}
	}
}

func (g *Game) getTalkWhisperText(agent *model.Agent, request model.Request, skipCountMap map[*model.Agent]int) string {
	text, err := g.sendRequest(agent, request)
	if err != nil {
		text = model.T_SKIP
	}
	g.GameStatuses[g.CurrentDay].RemainTalkMap[*agent]--
	if text == model.T_SKIP {
		skipCountMap[agent]++
		if skipCountMap[agent] >= g.Settings.MaxSkip {
			text = model.T_OVER
		}
	}
	if text != model.T_OVER && text != model.T_SKIP {
		skipCountMap[agent] = 0
	}
	return text
}

func (g *Game) sendRequest(agent *model.Agent, request model.Request) (string, error) {
	info := model.NewInfo(agent, g.GameStatuses[g.CurrentDay], g.GameStatuses[g.CurrentDay-1], g.Settings)
	switch request {
	case model.R_NAME, model.R_ROLE:
		return agent.SendPacket(model.Packet{Request: &request})
	case model.R_INITIALIZE, model.R_DAILY_INITIALIZE:
		g.resetLastIdxMaps()
		return agent.SendPacket(model.Packet{Request: &request, Info: &info, Settings: &g.Settings})
	case model.R_VOTE, model.R_DIVINE, model.R_GUARD, model.R_ATTACK:
		return agent.SendPacket(model.Packet{Request: &request, Info: &info})
	case model.R_DAILY_FINISH, model.R_TALK, model.R_WHISPER:
		talks, whispers := g.minimize(agent, info.TalkList, info.WhisperList)
		return agent.SendPacket(model.Packet{Request: &request, TalkHistory: talks, WhisperHistory: whispers})
	case model.R_FINISH:
		info.RoleMap = utils.GetRoleMap(g.Agents)
		return agent.SendPacket(model.Packet{Request: &request, Info: &info})
	}
	return "", errors.New("不明なリクエスト")
}

func (g *Game) resetLastIdxMaps() {
	g.LastTalkIdxMap = make(map[*model.Agent]int)
	g.LastWhisperIdxMap = make(map[*model.Agent]int)
}

func (g *Game) minimize(agent *model.Agent, talks []model.Talk, whispers []model.Talk) ([]model.Talk, []model.Talk) {
	lastTalkIdx := g.LastTalkIdxMap[agent]
	lastWhisperIdx := g.LastWhisperIdxMap[agent]
	g.LastTalkIdxMap[agent] = len(talks)
	g.LastWhisperIdxMap[agent] = len(whispers)
	return talks[lastTalkIdx:], whispers[lastWhisperIdx:]
}

func (g *Game) getAliveAgents() []*model.Agent {
	return utils.FilterAgents(g.Agents, func(agent *model.Agent) bool {
		return g.GameStatuses[g.CurrentDay].StatusMap[*agent] == model.S_ALIVE
	})
}

func (g *Game) getAliveWerewolves() []*model.Agent {
	return utils.FilterAgents(g.Agents, func(agent *model.Agent) bool {
		return g.GameStatuses[g.CurrentDay].StatusMap[*agent] == model.S_ALIVE && agent.Role.Team == model.T_WEREWOLF
	})
}
