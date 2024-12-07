package util

import (
	"log/slog"

	"github.com/kano-lab/aiwolf-nlp-server/model"
)

func CountAliveTeams(statusMap map[model.Agent]model.Status) (int, int) {
	var humans, werewolfs int
	for agent, status := range statusMap {
		if status == model.S_ALIVE {
			switch agent.Role.Species {
			case model.S_HUMAN:
				humans++
			case model.S_WEREWOLF:
				werewolfs++
			}
		}
	}
	return humans, werewolfs
}

func CalcWinSideTeam(statusMap map[model.Agent]model.Status) model.Team {
	humans, werewolfs := CountAliveTeams(statusMap)
	if humans <= werewolfs {
		return model.T_WEREWOLF
	}
	if werewolfs == 0 {
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
	rolesCopy := make(map[model.Role]int)
	for role, count := range roles {
		rolesCopy[role] = count
	}
	agents := make([]*model.Agent, 0)
	for i, conn := range conns {
		role := assignRole(rolesCopy)
		agent, err := model.NewAgent(i+1, role, conn)
		if err != nil {
			slog.Error("エージェントの作成に失敗しました", "error", err)
		}
		agents = append(agents, agent)
	}
	return agents
}

func CreateAgentsWithRole(roleMapConns map[model.Role][]model.Connection) []*model.Agent {
	agents := make([]*model.Agent, 0)
	i := 0
	for role, conns := range roleMapConns {
		for _, conn := range conns {
			agent, err := model.NewAgent(i+1, role, conn)
			i++
			if err != nil {
				slog.Error("エージェントの作成に失敗しました", "error", err)
			}
			agents = append(agents, agent)
		}
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

func GetCandidates(votes []model.Vote, condition func(model.Vote) bool) []model.Agent {
	counter := make(map[model.Agent]int)
	for _, vote := range votes {
		if condition(vote) {
			counter[vote.Target]++
		}
	}
	return getMaxCountCandidates(counter)
}

func getMaxCountCandidates(counter map[model.Agent]int) []model.Agent {
	var max int
	for _, count := range counter {
		if count > max {
			max = count
		}
	}
	candidates := make([]model.Agent, 0)
	for agent, count := range counter {
		if count == max {
			candidates = append(candidates, agent)
		}
	}
	return candidates
}

func GetRoleTeamNamesMap(agents []*model.Agent) map[model.Role][]string {
	roleTeamNamesMap := make(map[model.Role][]string)
	for _, a := range agents {
		roleTeamNamesMap[a.Role] = append(roleTeamNamesMap[a.Role], a.Team)
	}
	return roleTeamNamesMap
}
