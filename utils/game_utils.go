package utils

import (
	"log/slog"

	"github.com/nharu-0630/aiwolf-nlp-server/model"
)

func CountAliveTeams(statusMap map[model.Agent]model.Status) (int, int) {
	var villagers, werewolves int
	for agent, status := range statusMap {
		if status == model.S_ALIVE {
			switch agent.Role.Team {
			case model.T_VILLAGER:
				villagers++
			case model.T_WEREWOLF:
				werewolves++
			}
		}
	}
	return villagers, werewolves
}

func CalcWinSideTeam(statusMap map[model.Agent]model.Status) model.Team {
	villagers, werewolves := CountAliveTeams(statusMap)
	if villagers == 0 {
		return model.T_WEREWOLF
	} else if werewolves == 0 {
		return model.T_VILLAGER
	}
	return model.T_NONE
}

func CalcHasErrorAgents(agents []*model.Agent) int {
	var count int
	for _, a := range agents {
		if a.HasError {
			count++
		}
	}
	return count
}

func GetRoleMap(agents []*model.Agent) map[model.Agent]model.Role {
	roleMap := make(map[model.Agent]model.Role)
	for _, a := range agents {
		roleMap[*a] = a.Role
	}
	return roleMap
}

func CreateAgents(conns []model.Connection, roles map[model.Role]int) []*model.Agent {
	agents := make([]*model.Agent, 0)
	for i, conn := range conns {
		role := assignRole(roles)
		agent, err := model.NewAgent(i+1, role, conn)
		if err != nil {
			slog.Error("エージェントの作成に失敗しました", "error", err)
		}
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

func GetCandidates(votes []model.Vote, condition func(model.Vote) bool) []*model.Agent {
	counter := make(map[*model.Agent]int)
	for _, vote := range votes {
		if condition(vote) {
			counter[&vote.Target]++
		}
	}
	return getMaxCountCandidates(counter)
}

func getMaxCountCandidates(counter map[*model.Agent]int) []*model.Agent {
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
