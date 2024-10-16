package utils

import (
	"math/rand"

	"github.com/nharu-0630/aiwolf-nlp-server/model"
)

func SelectRandomAgent(agents []*model.Agent) *model.Agent {
	return agents[rand.Intn(len(agents))]
}

func FilterAgents(agents []*model.Agent, filter func(*model.Agent) bool) []*model.Agent {
	filtered := make([]*model.Agent, 0)
	for _, agent := range agents {
		if filter(agent) {
			filtered = append(filtered, agent)
		}
	}
	return filtered
}

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

func GetRoleMap(agents []*model.Agent) map[model.Agent]model.Role {
	roleMap := make(map[model.Agent]model.Role)
	for _, a := range agents {
		roleMap[*a] = a.Role
	}
	return roleMap
}
