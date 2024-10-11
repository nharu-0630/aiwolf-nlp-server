package core

import (
	"math/rand/v2"

	"github.com/dgryski/trifles/uuid"
	"github.com/gorilla/websocket"
	"github.com/nharu-0630/aiwolf-nlp-server/config"
	"github.com/nharu-0630/aiwolf-nlp-server/model"
)

type Game struct {
	GameID            string
	Settings          model.Settings
	Agents            []*model.Agent
	CurrentGameStatus model.GameStatus
	GameStatusByDay   map[int]*model.GameStatus
	LastTalkIdxMap    map[*model.Agent]int
	LastWhisperIdxMap map[*model.Agent]int
}

func NewGame(settings model.Settings, conns []*websocket.Conn) *Game {
	uuid := uuid.UUIDv4()

	roles := model.Roles(len(conns))
	if len(roles) == 0 {
		panic("Invalid number of agents")
	}
	settings.RoleNumMap = model.Roles(len(conns))

	agents := make([]*model.Agent, 0)
	for i, conn := range conns {
		role := model.R_VILLAGER
		for r, n := range roles {
			if n > 0 {
				role = r
				roles[r]--
				break
			}
		}
		agent, err := model.NewAgent(i, role, conn)
		if err != nil {
			panic(err)
		}
		agents = append(agents, agent)
	}

	game := &Game{
		GameID:            uuid,
		Settings:          settings,
		Agents:            agents,
		CurrentGameStatus: model.NewInitializeGameStatus(agents),
		GameStatusByDay:   make(map[int]*model.GameStatus),
		LastTalkIdxMap:    make(map[*model.Agent]int),
		LastWhisperIdxMap: make(map[*model.Agent]int),
	}
	return game
}

func (g *Game) Start() {
	for g.calcWinTeam() == model.T_NULL {
		g.doDay()
		g.doNight()
		g.GameStatusByDay[g.CurrentGameStatus.Day] = &g.CurrentGameStatus
		g.CurrentGameStatus = g.CurrentGameStatus.NextDay()
	}
	for _, agent := range g.Agents {
		g.SendRequest(agent, model.R_FINISH)
	}
}

func (g *Game) doDay() {
	for _, agent := range g.Agents {
		g.SendRequest(agent, model.R_DAILY_INITIALIZE)
	}
	if config.TALK_ON_FIRST_DAY && g.CurrentGameStatus.Day == 0 {
		g.doWhisper()
		g.doTalk()
	} else {
		g.doTalk()
	}
}

func (g *Game) doNight() {
	for _, agent := range g.Agents {
		g.SendRequest(agent, model.R_DAILY_FINISH)
	}
	if config.TALK_ON_FIRST_DAY && g.CurrentGameStatus.Day == 0 {
		g.doWhisper()
	}

	var executed *model.Agent
	candidates := make([]*model.Agent, 0)
	if g.CurrentGameStatus.Day != 0 {
		for i := 0; i < g.Settings.MaxRevote; i++ {
			g.doVote()
			candidates = g.getVotedCandidates(g.CurrentGameStatus.VoteList)
			if len(candidates) == 1 {
				executed = candidates[0]
				break
			}
		}
		if executed == nil {
			rand.Shuffle(len(candidates), func(i, j int) {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			})
			executed = candidates[0]
		}
		if executed != nil {
			g.CurrentGameStatus.StatusMap[*executed] = model.S_DEAD
			g.CurrentGameStatus.ExecutedAgent = *executed
		}
	}

	g.doDivine()

	if g.CurrentGameStatus.Day != 0 {
		g.doWhisper()
		g.doGuard()

		var attacked *model.Agent
		werewolfs := make([]*model.Agent, 0)
		for agent, status := range g.CurrentGameStatus.StatusMap {
			if status == model.S_ALIVE && agent.Role.Team == model.T_WEREWOLF {
				werewolfs = append(werewolfs, &agent)
			}
		}
		if len(werewolfs) > 0 {
			for i := 0; i < g.Settings.MaxAttackRevote; i++ {
				g.doAttackVote()
				candidates = g.getAttackVotedCandidates(g.CurrentGameStatus.AttackVoteList)
				if len(candidates) == 1 {
					attacked = candidates[0]
					break
				}
			}
			if attacked == nil && !g.Settings.IsEnableNoAttack {
				rand.Shuffle(len(candidates), func(i, j int) {
					candidates[i], candidates[j] = candidates[j], candidates[i]
				})
				attacked = candidates[0]
			}
			isGuarded := false
			if attacked != nil {
				if g.CurrentGameStatus.Guard.Target == *attacked && g.CurrentGameStatus.StatusMap[g.CurrentGameStatus.Guard.Agent] == model.S_ALIVE {
					isGuarded = true
				}
			}
			if !isGuarded {
				g.CurrentGameStatus.StatusMap[*attacked] = model.S_DEAD
				g.CurrentGameStatus.AttackedAgent = *attacked
			}
		}
	}
}

func (g *Game) doDivine() {
	for _, agent := range g.Agents {
		if agent.Role == model.R_SEER {
			name, err := g.SendRequest(agent, model.R_DIVINE)
			if err != nil {
				name = ""
			}
			target := model.Agent{}
			for _, a := range g.Agents {
				if a.Name == name {
					target = *a
					break
				}
			}
			if target.Name == "" {
				continue
			}
			judge := model.Judge{
				Day:    g.CurrentGameStatus.Day,
				Agent:  *agent,
				Target: target,
				Result: target.Role.Species,
			}
			g.CurrentGameStatus.DivineResult = judge
			break
		}
	}
}

func (g *Game) doGuard() {
	for _, agent := range g.Agents {
		if agent.Role == model.R_BODYGUARD {
			if g.CurrentGameStatus.ExecutedAgent == *agent {
				continue
			}
			name, err := g.SendRequest(agent, model.R_GUARD)
			if err != nil {
				name = ""
			}
			target := model.Agent{}
			for _, a := range g.Agents {
				if a.Name == name {
					target = *a
					break
				}
			}
			if target.Name == "" || target == *agent {
				continue
			}
			guard := model.Guard{
				Day:    g.CurrentGameStatus.Day,
				Agent:  *agent,
				Target: target,
			}
			g.CurrentGameStatus.Guard = guard
			break
		}
	}
}

func (g *Game) getVotedCandidates(voteList []model.Vote) []*model.Agent {
	counter := make(map[*model.Agent]int)
	for _, vote := range voteList {
		if g.CurrentGameStatus.StatusMap[vote.Target] == model.S_ALIVE {
			counter[&vote.Target]++
		}
	}
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

func (g *Game) getAttackVotedCandidates(voteList []model.Vote) []*model.Agent {
	counter := make(map[*model.Agent]int)
	for _, vote := range voteList {
		if g.CurrentGameStatus.StatusMap[vote.Target] == model.S_ALIVE && vote.Target.Role.Team != model.T_WEREWOLF {
			counter[&vote.Target]++
		}
	}
	if !g.Settings.IsEnableNoAttack {
		for agent, status := range g.CurrentGameStatus.StatusMap {
			if status == model.S_ALIVE && agent.Role.Team != model.T_WEREWOLF {
				counter[&agent]++
			}
		}
	}
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
	g.CurrentGameStatus.VoteList = make([]model.Vote, 0)
	agents := make([]*model.Agent, 0)
	for agent, status := range g.CurrentGameStatus.StatusMap {
		if status == model.S_ALIVE {
			agents = append(agents, &agent)
		}
	}
	votes := make([]model.Vote, 0)
	for _, agent := range agents {
		name, err := g.SendRequest(agent, model.R_VOTE)
		if err != nil {
			name = ""
		}
		target := model.Agent{}
		for _, a := range agents {
			if a.Name == name {
				target = *a
				break
			}
		}
		if target.Name == "" {
			for target.Name == "" || target == *agent {
				target = *agents[rand.IntN(len(agents))]
			}
		}
		vote := model.Vote{
			Day:    g.CurrentGameStatus.Day,
			Agent:  *agent,
			Target: target,
		}
		votes = append(votes, vote)
	}
	g.CurrentGameStatus.VoteList = votes
}

func (g *Game) doAttackVote() {
	g.CurrentGameStatus.AttackVoteList = make([]model.Vote, 0)
	werewolfs := make([]*model.Agent, 0)
	for agent, status := range g.CurrentGameStatus.StatusMap {
		if status == model.S_ALIVE && agent.Role.Team == model.T_WEREWOLF {
			werewolfs = append(werewolfs, &agent)
		}
	}
	agents := make([]*model.Agent, 0)
	for agent, status := range g.CurrentGameStatus.StatusMap {
		if status == model.S_ALIVE {
			agents = append(agents, &agent)
		}
	}
	votes := make([]model.Vote, 0)
	for _, agent := range werewolfs {
		name, err := g.SendRequest(agent, model.R_ATTACK)
		if err != nil {
			name = ""
		}
		target := model.Agent{}
		for _, a := range agents {
			if a.Name == name {
				target = *a
				break
			}
		}
		if target.Name == "" || target == *agent || g.CurrentGameStatus.StatusMap[target] != model.S_ALIVE || target.Role.Team == model.T_WEREWOLF {
			continue
		}
		vote := model.Vote{
			Day:    g.CurrentGameStatus.Day,
			Agent:  *agent,
			Target: target,
		}
		votes = append(votes, vote)
	}
	g.CurrentGameStatus.AttackVoteList = votes
}

func (g *Game) doWhisper() {
	g.CurrentGameStatus.ResetRemainWhisperMap()
	werewolfs := make([]*model.Agent, 0)
	for agent, status := range g.CurrentGameStatus.StatusMap {
		if status == model.S_ALIVE && agent.Role.Team == model.T_WEREWOLF {
			werewolfs = append(werewolfs, &agent)
		}
	}
	if len(werewolfs) < 2 {
		return
	}
	rand.Shuffle(len(werewolfs), func(i, j int) {
		werewolfs[i], werewolfs[j] = werewolfs[j], werewolfs[i]
	})
	skipCountMap := make(map[*model.Agent]int)
	idx := 0
	for i := 0; i < config.MAX_WHISPER_COUNT; i++ {
		cnt := false
		for _, agent := range werewolfs {
			if g.CurrentGameStatus.RemainWhisperMap[*agent] == 0 {
				continue
			}
			var text string = model.T_OVER
			text, err := g.SendRequest(agent, model.R_WHISPER)
			if err != nil {
				text = model.T_SKIP
			}
			g.CurrentGameStatus.RemainWhisperMap[*agent]--
			if text == model.T_SKIP {
				skipCountMap[agent]++
				if skipCountMap[agent] >= g.Settings.MaxSkip {
					text = model.T_OVER
				}
			}
			talk := model.Talk{
				Idx:   idx,
				Day:   g.CurrentGameStatus.Day,
				Turn:  i,
				Agent: *agent,
				Text:  text,
			}
			idx++
			g.CurrentGameStatus.WhisperList = append(g.CurrentGameStatus.WhisperList, talk)
			if text != model.T_OVER && text != model.T_SKIP {
				skipCountMap[agent] = 0
			}
			if text != model.T_OVER {
				cnt = true
			}
		}
		if !cnt {
			break
		}
	}
}

func (g *Game) doTalk() {
	g.CurrentGameStatus.ResetRemainTalkMap()
	agents := make([]*model.Agent, 0)
	for agent, status := range g.CurrentGameStatus.StatusMap {
		if status == model.S_ALIVE {
			agents = append(agents, &agent)
		}
	}
	rand.Shuffle(len(agents), func(i, j int) {
		agents[i], agents[j] = agents[j], agents[i]
	})
	skipCountMap := make(map[*model.Agent]int)
	idx := 0
	for i := 0; i < config.MAX_TALK_COUNT; i++ {
		cnt := false
		for _, agent := range agents {
			if g.CurrentGameStatus.RemainTalkMap[*agent] == 0 {
				continue
			}
			var text string = model.T_OVER
			text, err := g.SendRequest(agent, model.R_TALK)
			if err != nil {
				text = model.T_SKIP
			}
			g.CurrentGameStatus.RemainTalkMap[*agent]--
			if text == model.T_SKIP {
				skipCountMap[agent]++
				if skipCountMap[agent] >= g.Settings.MaxSkip {
					text = model.T_OVER
				}
			}
			talk := model.Talk{
				Idx:   idx,
				Day:   g.CurrentGameStatus.Day,
				Turn:  i,
				Agent: *agent,
				Text:  text,
			}
			idx++
			g.CurrentGameStatus.TalkList = append(g.CurrentGameStatus.TalkList, talk)
			if text != model.T_OVER && text != model.T_SKIP {
				skipCountMap[agent] = 0
			}
			if text != model.T_OVER {
				cnt = true
			}
		}
		if !cnt {
			break
		}
	}
}
func (g *Game) SendRequest(agent *model.Agent, request model.Request) (string, error) {
	switch request {
	case model.R_NAME, model.R_ROLE:
		return agent.SendPacket(model.Packet{
			Request: &request,
		})
	case model.R_INITIALIZE, model.R_DAILY_INITIALIZE:
		g.LastTalkIdxMap = make(map[*model.Agent]int)
		g.LastWhisperIdxMap = make(map[*model.Agent]int)
		info := g.CurrentGameStatus.ConvertToInfo(agent, g.Settings)
		return agent.SendPacket(model.Packet{
			Request:  &request,
			Info:     &info,
			Settings: &g.Settings,
		})
	case model.R_VOTE, model.R_DIVINE, model.R_GUARD, model.R_ATTACK:
		info := g.CurrentGameStatus.ConvertToInfo(agent, g.Settings)
		return agent.SendPacket(model.Packet{
			Request: &request,
			Info:    &info,
		})
	case model.R_DAILY_FINISH, model.R_TALK, model.R_WHISPER:
		info := g.CurrentGameStatus.ConvertToInfo(agent, g.Settings)
		talks, whispers := g.minimize(agent, info.TalkList, info.WhisperList)
		return agent.SendPacket(model.Packet{
			Request:        &request,
			TalkHistory:    talks,
			WhisperHistory: whispers,
		})
	case model.R_FINISH:
		info := g.CurrentGameStatus.ConvertToInfo(agent, g.Settings)
		roleMap := make(map[model.Agent]model.Role)
		for _, a := range g.Agents {
			roleMap[*a] = a.Role
		}
		info.RoleMap = roleMap
		return agent.SendPacket(model.Packet{
			Request: &request,
			Info:    &info,
		})
	}
	return "", nil
}

func (g *Game) minimize(agent *model.Agent, talks []model.Talk, whispers []model.Talk) ([]model.Talk, []model.Talk) {
	lastTalkIdx := 0
	lastWhisperIdx := 0
	if idx, ok := g.LastTalkIdxMap[agent]; ok {
		lastTalkIdx = idx
	}
	if idx, ok := g.LastWhisperIdxMap[agent]; ok {
		lastWhisperIdx = idx
	}
	g.LastTalkIdxMap[agent] = len(talks)
	g.LastWhisperIdxMap[agent] = len(whispers)
	return talks[lastTalkIdx:], whispers[lastWhisperIdx:]
}

func (g *Game) calcWinTeam() model.Team {
	var villager, werewolf int
	for agent, status := range g.CurrentGameStatus.StatusMap {
		if status == model.S_ALIVE {
			if agent.Role.Team == model.T_VILLAGER {
				villager++
			} else {
				werewolf++
			}
		}
	}
	if werewolf == 0 {
		return model.T_VILLAGER
	}
	if villager <= werewolf {
		return model.T_WEREWOLF
	}
	return model.T_NULL
}
