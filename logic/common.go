package logic

import (
	"errors"
	"log/slog"
	"time"

	"github.com/nharu-0630/aiwolf-nlp-server/model"
	"github.com/nharu-0630/aiwolf-nlp-server/utils"
)

func (g *Game) findTargetByRequest(agent *model.Agent, request model.Request) (*model.Agent, error) {
	name, err := g.requestToAgent(agent, request)
	if err != nil {
		return nil, err
	}
	target := utils.FindAgentByName(g.Agents, name)
	if target == nil {
		return nil, errors.New("対象エージェントが見つかりません")
	}
	slog.Info("対象エージェントを受信しました", "id", g.ID, "agent", agent.String(), "target", target.String())
	return target, nil
}

func (g *Game) getVotedCandidates(votes []model.Vote) []*model.Agent {
	return utils.GetCandidates(votes, func(vote model.Vote) bool {
		return g.GameStatuses[g.CurrentDay].StatusMap[vote.Target] == model.S_ALIVE
	})
}

func (g *Game) getAttackVotedCandidates(votes []model.Vote) []*model.Agent {
	return utils.GetCandidates(votes, func(vote model.Vote) bool {
		return g.GameStatuses[g.CurrentDay].StatusMap[vote.Target] == model.S_ALIVE && vote.Target.Role.Team != model.T_WEREWOLF
	})
}

func (g *Game) closeAllAgents() {
	for _, agent := range g.Agents {
		agent.Close()
	}
}

func (g *Game) requestToEveryone(request model.Request) {
	for _, agent := range g.Agents {
		g.requestToAgent(agent, request)
	}
}

func (g *Game) requestToAgent(agent *model.Agent, request model.Request) (string, error) {
	info := model.NewInfo(agent, g.GameStatuses[g.CurrentDay], g.GameStatuses[g.CurrentDay-1], g.Settings)
	var packet model.Packet
	switch request {
	case model.R_NAME:
		packet = model.Packet{Request: &request}
	case model.R_INITIALIZE, model.R_DAILY_INITIALIZE:
		g.resetLastIdxMaps()
		packet = model.Packet{Request: &request, Info: &info, Settings: &g.Settings}
	case model.R_VOTE, model.R_DIVINE, model.R_GUARD, model.R_ATTACK:
		packet = model.Packet{Request: &request, Info: &info}
	case model.R_DAILY_FINISH, model.R_TALK, model.R_WHISPER:
		talks, whispers := g.minimize(agent, info.TalkList, info.WhisperList)
		packet = model.Packet{Request: &request, TalkHistory: talks, WhisperHistory: whispers}
	case model.R_FINISH:
		info.RoleMap = utils.GetRoleMap(g.Agents)
		packet = model.Packet{Request: &request, Info: &info}
	default:
		return "", errors.New("一致するリクエストがありません")
	}
	return agent.SendPacket(packet, time.Duration(g.Settings.ActionTimeout)*time.Millisecond, time.Duration(g.Settings.ResponseTimeout)*time.Millisecond)
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
