package logic

import (
	"log/slog"
	"math/rand"

	"github.com/nharu-0630/aiwolf-nlp-server/model"
	"github.com/nharu-0630/aiwolf-nlp-server/utils"
)

func (g *Game) doExecution() {
	slog.Info("追放フェーズを開始します", "id", g.ID, "day", g.CurrentDay)
	var executed *model.Agent
	candidates := make([]*model.Agent, 0)
	for i := 0; i < g.Settings.MaxRevote; i++ {
		g.executeVote()
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
		slog.Info("追放結果を設定しました", "id", g.ID, "agent", executed.Name)

		mediums := utils.FilterAgents(g.Agents, func(agent *model.Agent) bool {
			return agent.Role == model.R_MEDIUM
		})
		if len(mediums) > 0 {
			g.GameStatuses[g.CurrentDay].MediumResult = &model.Judge{
				Day:    g.GameStatuses[g.CurrentDay].Day,
				Agent:  *mediums[0],
				Target: *executed,
				Result: executed.Role.Species,
			}
			slog.Info("霊能結果を設定しました", "id", g.ID, "target", executed.String(), "result", executed.Role.Species)
		}
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
		attacked = g.conductAttackVote()
		if attacked == nil && !g.Settings.IsEnableNoAttack {
			attacked = utils.SelectRandomAgent(g.getAttackVotedCandidates(g.GameStatuses[g.CurrentDay].AttackVoteList))
		}

		if attacked != nil && !g.isGuarded(attacked) {
			g.GameStatuses[g.CurrentDay].StatusMap[*attacked] = model.S_DEAD
			g.GameStatuses[g.CurrentDay].AttackedAgent = attacked
			slog.Info("襲撃結果を設定しました", "id", g.ID, "agent", attacked.Name)
		} else if attacked != nil {
			slog.Info("護衛されたため、襲撃結果を設定しません", "id", g.ID, "agent", attacked.Name)
		} else {
			slog.Warn("襲撃対象がいないため、襲撃結果を設定しません", "id", g.ID)
		}
	}
	slog.Info("襲撃フェーズを終了します", "id", g.ID, "day", g.CurrentDay)
}

func (g *Game) conductAttackVote() *model.Agent {
	slog.Info("襲撃投票を開始します", "id", g.ID, "day", g.CurrentDay)
	var attacked *model.Agent
	for i := 0; i < g.Settings.MaxAttackRevote; i++ {
		g.executeAttackVote()
		candidates := g.getAttackVotedCandidates(g.GameStatuses[g.CurrentDay].AttackVoteList)
		if len(candidates) == 1 {
			attacked = candidates[0]
			break
		}
	}
	return attacked
}

func (g *Game) isGuarded(attacked *model.Agent) bool {
	if g.GameStatuses[g.CurrentDay].Guard == nil {
		return false
	}
	return g.GameStatuses[g.CurrentDay].Guard.Target == *attacked && g.GameStatuses[g.CurrentDay].StatusMap[g.GameStatuses[g.CurrentDay].Guard.Agent] == model.S_ALIVE
}

func (g *Game) doDivine() {
	slog.Info("占いフェーズを開始します", "id", g.ID, "day", g.CurrentDay)
	for _, agent := range g.Agents {
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
	if err == nil {
		g.GameStatuses[g.CurrentDay].DivineResult = &model.Judge{
			Day:    g.GameStatuses[g.CurrentDay].Day,
			Agent:  *agent,
			Target: *target,
			Result: target.Role.Species,
		}
		slog.Info("占い結果を設定しました", "id", g.ID, "target", target.String(), "result", target.Role.Species)
	} else {
		slog.Warn("占い対象が見つからなかったため、占い結果を設定しません", "id", g.ID)
	}
}

func (g *Game) doGuard() {
	slog.Info("護衛フェーズを開始します", "id", g.ID, "day", g.CurrentDay)
	for _, agent := range g.Agents {
		if agent.Role == model.R_BODYGUARD && g.GameStatuses[g.CurrentDay].ExecutedAgent != agent {
			g.conductGuard(agent)
			break
		}
	}
}

func (g *Game) conductGuard(agent *model.Agent) {
	slog.Info("護衛アクションを実行します", "id", g.ID, "agent", agent.String())
	target, err := g.findTargetByRequest(agent, model.R_GUARD)
	if err == nil {
		g.GameStatuses[g.CurrentDay].Guard = &model.Guard{
			Day:    g.GameStatuses[g.CurrentDay].Day,
			Agent:  *agent,
			Target: *target,
		}
		slog.Info("護衛対象を設定しました", "id", g.ID, "target", target.String())
	} else {
		slog.Warn("護衛対象が見つからなかったため、護衛対象を設定しません", "id", g.ID)
	}
}

func (g *Game) executeVote() {
	slog.Info("投票アクションを開始します", "id", g.ID, "day", g.CurrentDay)
	g.GameStatuses[g.CurrentDay].VoteList = g.collectVotes(model.R_VOTE, g.getAliveAgents())
}

func (g *Game) executeAttackVote() {
	slog.Info("襲撃投票アクションを開始します", "id", g.ID, "day", g.CurrentDay)
	g.GameStatuses[g.CurrentDay].AttackVoteList = g.collectVotes(model.R_ATTACK, g.getAliveWerewolves())
}

func (g *Game) collectVotes(request model.Request, agents []*model.Agent) []model.Vote {
	votes := make([]model.Vote, 0)
	for _, agent := range agents {
		target, err := g.findTargetByRequest(agent, request)
		if err != nil {
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
	g.conductCommunication(model.R_WHISPER)
}

func (g *Game) doTalk() {
	slog.Info("発言フェーズを開始します", "id", g.ID, "day", g.CurrentDay)
	g.conductCommunication(model.R_TALK)
}

func (g *Game) conductCommunication(request model.Request) {
	var agents []*model.Agent
	var maxTurn int
	var remainMap map[model.Agent]int
	var talkList *[]model.Talk
	switch request {
	case model.R_TALK:
		agents = g.getAliveAgents()
		g.GameStatuses[g.CurrentDay].ResetRemainTalkMap(g.Settings.MaxTalkTurn)
		maxTurn = g.Settings.MaxTalk
		remainMap = g.GameStatuses[g.CurrentDay].RemainTalkMap
		talkList = &g.GameStatuses[g.CurrentDay].TalkList
	case model.R_WHISPER:
		agents = g.getAliveWerewolves()
		g.GameStatuses[g.CurrentDay].ResetRemainWhisperMap(g.Settings.MaxWhisperTurn)
		maxTurn = g.Settings.MaxWhisper
		remainMap = g.GameStatuses[g.CurrentDay].RemainWhisperMap
		talkList = &g.GameStatuses[g.CurrentDay].WhisperList
	}

	if len(agents) < 2 {
		slog.Warn("エージェント数が2未満のため、通信を行いません", "id", g.ID, "agentNum", len(agents))
	}

	rand.Shuffle(len(agents), func(i, j int) {
		agents[i], agents[j] = agents[j], agents[i]
	})
	skipCountMap := make(map[*model.Agent]int)
	idx := 0

	for i := 0; i < maxTurn; i++ {
		cnt := false
		for _, agent := range agents {
			if remainMap[*agent] == 0 {
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
			} else {
				remainMap[*agent] = 0
				slog.Info("発言がオーバーであるため、残り発言回数を0にしました", "id", g.ID, "agent", agent.String())
			}
			slog.Info("発言を受信しました", "id", g.ID, "agent", agent.String(), "text", text)
		}
		if !cnt {
			break
		}
	}
}

func (g *Game) getTalkWhisperText(agent *model.Agent, request model.Request, skipCountMap map[*model.Agent]int) string {
	text, err := g.requestToAgent(agent, request)
	if text == model.T_FORCE_SKIP {
		text = model.T_SKIP
		slog.Warn("クライアントから強制スキップが指定されたため、発言をスキップに置換しました", "id", g.ID, "agent", agent.String())
	}
	if err != nil {
		text = model.T_FORCE_SKIP
		slog.Warn("リクエストの送受信に失敗したため、発言をスキップに置換しました", "id", g.ID, "agent", agent.String())
	}
	g.GameStatuses[g.CurrentDay].RemainTalkMap[*agent]--
	if text == model.T_SKIP {
		skipCountMap[agent]++
		if skipCountMap[agent] >= g.Settings.MaxSkip {
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
		skipCountMap[agent] = 0
		slog.Info("発言がオーバーもしくはスキップではないため、スキップ回数をリセットしました", "id", g.ID, "agent", agent.String())
	}
	return text
}
