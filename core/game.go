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
	log.Println("Creating new game...")
	id := uuid.UUIDv4()
	roles := model.Roles(len(conns))
	if len(roles) == 0 {
		log.Panic("Invalid number of agents")
	}
	settings.RoleNumMap = model.Roles(len(conns))
	agents := createAgents(conns, roles)
	log.Printf("Game created with ID: %s", id)
	gameStatus := model.NewInitializeGameStatus(agents)
	gameStatuses := make(map[int]*model.GameStatus)
	gameStatuses[0] = &gameStatus
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
	log.Println("Creating agents...")
	agents := make([]*model.Agent, 0)
	for i, conn := range conns {
		role := assignRole(roles)
		agent, err := model.NewAgent(i, role, conn)
		if err != nil {
			log.Panic(err)
		}
		log.Printf("Agent %d created with role %s", i, role)
		agents = append(agents, agent)
	}
	return agents
}

func assignRole(roles map[model.Role]int) model.Role {
	for r, n := range roles {
		if n > 0 {
			roles[r]--
			log.Printf("Assigned role: %s", r)
			return r
		}
	}
	log.Println("Assigned default role: Villager")
	return model.R_VILLAGER
}

func (g *Game) sendRequestToEveryone(request model.Request) {
	log.Printf("Sending request %s to all agents", request)
	for _, agent := range g.Agents {
		g.sendRequest(agent, request)
	}
}

func (g *Game) Start() {
	log.Println("Starting game...")
	for utils.CalcWinSideTeam(g.GameStatuses[g.CurrentDay].StatusMap) == model.T_NONE {
		g.doDay()
		g.doNight()
		gameStatus := g.GameStatuses[g.CurrentDay].NextDay()
		g.GameStatuses[g.CurrentDay+1] = &gameStatus
		g.CurrentDay++
	}
	g.sendRequestToEveryone(model.R_FINISH)
	log.Println("Game finished")
}

func (g *Game) doDay() {
	log.Printf("Day %d: Starting day phase", g.CurrentDay)
	g.sendRequestToEveryone(model.R_DAILY_INITIALIZE)
	if g.Settings.IsTalkOnFirstDay && g.CurrentDay == 0 {
		g.doWhisper()
	}
	g.doTalk()
}

func (g *Game) doNight() {
	log.Printf("Day %d: Starting night phase", g.CurrentDay)
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
	log.Println("Handling execution phase")
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
		log.Printf("Agent %s executed", executed.Name)
	}
}

func (g *Game) handleAttack() {
	log.Println("Handling attack phase")
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
	log.Println("Performing attack vote")
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
		log.Printf("Agent %s attacked and killed", attacked.Name)
	} else if attacked != nil {
		log.Printf("Agent %s was attacked but guarded", attacked.Name)
	}
}

func (g *Game) isGuarded(attacked *model.Agent) bool {
	return g.GameStatuses[g.CurrentDay].Guard.Target == *attacked && g.GameStatuses[g.CurrentDay].StatusMap[g.GameStatuses[g.CurrentDay].Guard.Agent] == model.S_ALIVE
}

func (g *Game) doDivine() {
	log.Println("Performing divination")
	for _, agent := range g.Agents {
		if agent.Role == model.R_SEER {
			g.performDivination(agent)
			break
		}
	}
}

func (g *Game) performDivination(agent *model.Agent) {
	target, err := g.findAgentByRequest(agent, model.R_DIVINE)
	if err == nil {
		g.GameStatuses[g.CurrentDay].DivineResult = &model.Judge{
			Day:    g.GameStatuses[g.CurrentDay].Day,
			Agent:  *agent,
			Target: *target,
			Result: target.Role.Species,
		}
		log.Printf("Divination result: Agent %s is %s", target.Name, target.Role.Species)
	}
}

func (g *Game) doGuard() {
	log.Println("Performing guard action")
	for _, agent := range g.Agents {
		if agent.Role == model.R_BODYGUARD && g.GameStatuses[g.CurrentDay].ExecutedAgent != agent {
			g.performGuard(agent)
			break
		}
	}
}

func (g *Game) performGuard(agent *model.Agent) {
	target, err := g.findAgentByRequest(agent, model.R_GUARD)
	if err == nil {
		g.GameStatuses[g.CurrentDay].Guard = &model.Guard{
			Day:    g.GameStatuses[g.CurrentDay].Day,
			Agent:  *agent,
			Target: *target,
		}
		log.Printf("Agent %s is guarding %s", agent.Name, target.Name)
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
	return nil, errors.New("Agent not found")
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
	log.Println("Collecting votes")
	g.GameStatuses[g.CurrentDay].VoteList = g.collectVotes(model.R_VOTE, g.getAliveAgents())
}

func (g *Game) doAttackVote() {
	log.Println("Collecting attack votes")
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
		log.Printf("Agent %s voted for %s", agent.Name, target.Name)
	}
	return votes
}

func (g *Game) doWhisper() {
	log.Println("Performing whisper phase")
	g.GameStatuses[g.CurrentDay].ResetRemainWhisperMap(g.Settings.MaxWhisper)
	werewolfs := g.getAliveWerewolves()
	if len(werewolfs) < 2 {
		return
	}
	g.performCommunication(werewolfs, model.R_WHISPER, &g.GameStatuses[g.CurrentDay].WhisperList, g.Settings.MaxWhisperTurn)
}

func (g *Game) doTalk() {
	log.Println("Performing talk phase")
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
			log.Printf("Agent %s: %s", agent.Name, text)
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
	return "", errors.New("Invalid request")
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
