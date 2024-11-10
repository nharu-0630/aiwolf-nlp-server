package logic

import (
	"log/slog"
	"math/rand"

	"github.com/kano-lab/aiwolf-nlp-server/model"
	"github.com/kano-lab/aiwolf-nlp-server/util"
)

func (g *Game) doExecution() {
	slog.Info("追放フェーズを開始します", "id", g.ID, "day", g.CurrentDay)
	var executed *model.Agent
	candidates := make([]*model.Agent, 0)
	for i := 0; i < g.Settings.MaxRevote; i++ {
		g.executeVote()
		candidates = g.getVotedCandidates(g.GameStatuses[g.CurrentDay].Votes)
		if len(candidates) == 1 {
			executed = candidates[0]
			break
		}
	}
	if executed == nil {
		executed = util.SelectRandomAgent(candidates)
	}
	if executed != nil {
		g.GameStatuses[g.CurrentDay].StatusMap[*executed] = model.S_DEAD
		g.GameStatuses[g.CurrentDay].ExecutedAgent = executed
		slog.Info("追放結果を設定しました", "id", g.ID, "agent", executed.Name)

		g.GameStatuses[g.CurrentDay].MediumResult = &model.Judge{
			Day:    g.GameStatuses[g.CurrentDay].Day,
			Agent:  *executed,
			Target: *executed,
			Result: executed.Role.Species,
		}
		slog.Info("霊能結果を設定しました", "id", g.ID, "target", executed.String(), "result", executed.Role.Species)
	} else {
		slog.Warn("追放対象がいないため、追放結果を設定しません", "id", g.ID)
	}
	slog.Info("追放フェーズを終了します", "id", g.ID, "day", g.CurrentDay)
}

func (g *Game) doAttack() {
	slog.Info("襲撃フェーズを開始します", "id", g.ID, "day", g.CurrentDay)
	var attacked *model.Agent
	werewolfs := g.getAliveWerewolves()
	if len(werewolfs) > 0 {
		for i := 0; i < g.Settings.MaxAttackRevote; i++ {
			g.executeAttackVote()
			candidates := g.getAttackVotedCandidates(g.GameStatuses[g.CurrentDay].AttackVotes)
			if len(candidates) == 1 {
				attacked = candidates[0]
				break
			}
		}
		if attacked == nil && !g.Settings.IsEnableNoAttack {
			attacked = util.SelectRandomAgent(g.getAttackVotedCandidates(g.GameStatuses[g.CurrentDay].AttackVotes))
		}

		if attacked != nil && !g.isGuarded(attacked) {
			g.GameStatuses[g.CurrentDay].StatusMap[*attacked] = model.S_DEAD
			g.GameStatuses[g.CurrentDay].AttackedAgent = attacked
			slog.Info("襲撃結果を設定しました", "id", g.ID, "agent", attacked.Name)
		} else if attacked != nil {
			slog.Info("護衛されたため、襲撃結果を設定しません", "id", g.ID, "agent", attacked.Name)
		} else {
			slog.Info("襲撃対象がいないため、襲撃結果を設定しません", "id", g.ID)
		}
	}
	slog.Info("襲撃フェーズを終了します", "id", g.ID, "day", g.CurrentDay)
}

func (g *Game) isGuarded(attacked *model.Agent) bool {
	if g.GameStatuses[g.CurrentDay].Guard == nil {
		return false
	}
	return g.GameStatuses[g.CurrentDay].Guard.Target == *attacked && g.isAlive(&g.GameStatuses[g.CurrentDay].Guard.Agent)
}

func (g *Game) doDivine() {
	slog.Info("占いフェーズを開始します", "id", g.ID, "day", g.CurrentDay)
	for _, agent := range g.getAliveAgents() {
		if agent.Role == model.R_SEER {
			g.conductDivination(agent)
			break
		}
	}
	slog.Info("占いフェーズを終了します", "id", g.ID, "day", g.CurrentDay)
}

func (g *Game) conductDivination(agent *model.Agent) {
	slog.Info("占いアクションを開始します", "id", g.ID, "agent", agent.String())
	target, err := g.findTargetByRequest(agent, model.R_DIVINE)
	if err != nil {
		slog.Warn("占い対象が見つからなかったため、占い結果を設定しません", "id", g.ID)
		return
	}
	if !g.isAlive(target) {
		slog.Warn("占い対象が死亡しているため、占い結果を設定しません", "id", g.ID, "target", target.String())
		return
	}
	if agent == target {
		slog.Warn("占い対象が自分自身であるため、占い結果を設定しません", "id", g.ID, "target", target.String())
		return
	}
	g.GameStatuses[g.CurrentDay].DivineResult = &model.Judge{
		Day:    g.GameStatuses[g.CurrentDay].Day,
		Agent:  *agent,
		Target: *target,
		Result: target.Role.Species,
	}
	slog.Info("占い結果を設定しました", "id", g.ID, "target", target.String(), "result", target.Role.Species)
}

func (g *Game) doGuard() {
	slog.Info("護衛フェーズを開始します", "id", g.ID, "day", g.CurrentDay)
	for _, agent := range g.getAliveAgents() {
		if agent.Role == model.R_BODYGUARD {
			g.conductGuard(agent)
			break
		}
	}
}

func (g *Game) conductGuard(agent *model.Agent) {
	slog.Info("護衛アクションを実行します", "id", g.ID, "agent", agent.String())
	target, err := g.findTargetByRequest(agent, model.R_GUARD)
	if err != nil {
		slog.Warn("護衛対象が見つからなかったため、護衛対象を設定しません", "id", g.ID)
		return
	}
	if !g.isAlive(target) {
		slog.Warn("護衛対象が死亡しているため、護衛対象を設定しません", "id", g.ID, "target", target.String())
		return
	}
	if agent == target {
		slog.Warn("護衛対象が自分自身であるため、護衛対象を設定しません", "id", g.ID, "target", target.String())
		return
	}
	g.GameStatuses[g.CurrentDay].Guard = &model.Guard{
		Day:    g.GameStatuses[g.CurrentDay].Day,
		Agent:  *agent,
		Target: *target,
	}
	slog.Info("護衛対象を設定しました", "id", g.ID, "target", target.String())
}

func (g *Game) executeVote() {
	slog.Info("投票アクションを開始します", "id", g.ID, "day", g.CurrentDay)
	g.GameStatuses[g.CurrentDay].Votes = g.collectVotes(model.R_VOTE, g.getAliveAgents())
}

func (g *Game) executeAttackVote() {
	slog.Info("襲撃投票アクションを開始します", "id", g.ID, "day", g.CurrentDay)
	g.GameStatuses[g.CurrentDay].AttackVotes = g.collectVotes(model.R_ATTACK, g.getAliveWerewolves())
}

func (g *Game) collectVotes(request model.Request, agents []*model.Agent) []model.Vote {
	votes := make([]model.Vote, 0)
	for _, agent := range agents {
		target, err := g.findTargetByRequest(agent, request)
		if err != nil {
			continue
		}
		if !g.isAlive(target) {
			slog.Warn("投票対象が死亡しているため、投票を無視します", "id", g.ID, "agent", agent.String(), "target", target.String())
			continue
		}
		votes = append(votes, model.Vote{
			Day:    g.GameStatuses[g.CurrentDay].Day,
			Agent:  *agent,
			Target: *target,
		})
		slog.Info("投票を受信しました", "id", g.ID, "agent", agent.String(), "target", target.String())
	}
	return votes
}

func (g *Game) doWhisper() {
	slog.Info("囁きフェーズを開始します", "id", g.ID, "day", g.CurrentDay)
	g.GameStatuses[g.CurrentDay].ResetRemainWhisperMap(g.Settings.MaxWhisper)
	g.conductCommunication(model.R_WHISPER)
	g.GameStatuses[g.CurrentDay].ClearRemainWhisperMap()
}

func (g *Game) doTalk() {
	slog.Info("トークフェーズを開始します", "id", g.ID, "day", g.CurrentDay)
	g.GameStatuses[g.CurrentDay].ResetRemainTalkMap(g.Settings.MaxTalk)
	g.conductCommunication(model.R_TALK)
	g.GameStatuses[g.CurrentDay].ClearRemainTalkMap()
}

func (g *Game) conductCommunication(request model.Request) {
	var agents []*model.Agent
	var maxTurn int
	var remainMap map[model.Agent]int
	var talkList *[]model.Talk
	switch request {
	case model.R_TALK:
		agents = g.getAliveAgents()
		maxTurn = g.Settings.MaxTalkTurn
		remainMap = g.GameStatuses[g.CurrentDay].RemainTalkMap
		talkList = &g.GameStatuses[g.CurrentDay].Talks
	case model.R_WHISPER:
		agents = g.getAliveWerewolves()
		maxTurn = g.Settings.MaxWhisperTurn
		remainMap = g.GameStatuses[g.CurrentDay].RemainWhisperMap
		talkList = &g.GameStatuses[g.CurrentDay].Whispers
	}

	if len(agents) < 2 {
		slog.Warn("エージェント数が2未満のため、通信を行いません", "id", g.ID, "agentNum", len(agents))
		return
	}

	rand.Shuffle(len(agents), func(i, j int) {
		agents[i], agents[j] = agents[j], agents[i]
	})
	skipMap := make(map[model.Agent]int)
	idx := 0

	for i := 0; i < maxTurn; i++ {
		cnt := false
		for _, agent := range agents {
			if remainMap[*agent] == 0 {
				continue
			}
			text := g.getTalkWhisperText(agent, request, skipMap, remainMap)
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
			} else {
				remainMap[*agent] = 0
				slog.Info("発言がオーバーであるため、残り発言回数を0にしました", "id", g.ID, "agent", agent.String())
			}
			slog.Info("発言を受信しました", "id", g.ID, "agent", agent.String(), "text", text, "skip", skipMap[*agent], "remain", remainMap[*agent])
		}
		if !cnt {
			break
		}
	}
}

func (g *Game) getTalkWhisperText(agent *model.Agent, request model.Request, skipMap map[model.Agent]int, remainMap map[model.Agent]int) string {
	text, err := g.requestToAgent(agent, request)
	if text == model.T_FORCE_SKIP {
		text = model.T_SKIP
		slog.Warn("クライアントから強制スキップが指定されたため、発言をスキップに置換しました", "id", g.ID, "agent", agent.String())
	}
	if err != nil {
		text = model.T_FORCE_SKIP
		slog.Warn("リクエストの送受信に失敗したため、発言をスキップに置換しました", "id", g.ID, "agent", agent.String())
	}
	remainMap[*agent]--
	if _, exists := skipMap[*agent]; !exists {
		skipMap[*agent] = 0
	}
	if text == model.T_SKIP {
		skipMap[*agent]++
		if skipMap[*agent] >= g.Settings.MaxSkip {
			text = model.T_OVER
			slog.Warn("スキップ回数が上限に達したため、発言をオーバーに置換しました", "id", g.ID, "agent", agent.String())
		} else {
			slog.Info("発言をスキップしました", "id", g.ID, "agent", agent.String())
		}
	} else if text == model.T_FORCE_SKIP {
		text = model.T_SKIP
		slog.Warn("強制スキップが指定されたため、発言をスキップに置換しました", "id", g.ID, "agent", agent.String())
	}
	if text != model.T_OVER && text != model.T_SKIP && text != model.T_FORCE_SKIP {
		skipMap[*agent] = 0
		slog.Info("発言がオーバーもしくはスキップではないため、スキップ回数をリセットしました", "id", g.ID, "agent", agent.String())
	}
	return text
}
