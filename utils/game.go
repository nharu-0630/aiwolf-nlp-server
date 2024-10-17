package utils

import (
	"log"

	"github.com/gorilla/websocket"
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

func GetRoleMap(agents []*model.Agent) map[model.Agent]model.Role {
	roleMap := make(map[model.Agent]model.Role)
	for _, a := range agents {
		roleMap[*a] = a.Role
	}
	return roleMap
}

func CreateAgents(conns []*websocket.Conn, roles map[model.Role]int) []*model.Agent {
	agents := make([]*model.Agent, 0)
	for i, conn := range conns {
		role := assignRole(roles)
		agent, err := model.NewAgent(i, role, conn)
		if err != nil {
			log.Panic(err)
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
