package utils

import (
	"math/rand"

	"github.com/nharu-0630/aiwolf-nlp-server/model"
)

func SelectRandomAgent(agents []*model.Agent) *model.Agent {
	if len(agents) == 0 {
		return nil
	}
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

func FindAgentByName(agents []*model.Agent, name string) *model.Agent {
	for _, agent := range agents {
		if agent.Name == name {
			return agent
		}
	}
	return nil
}
