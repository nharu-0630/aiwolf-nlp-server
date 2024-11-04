package core

import "github.com/nharu-0630/aiwolf-nlp-server/model"

type MatchOptimizer struct {
	matchHistory []map[model.Role]string
}

func NewMatchOptimizer() *MatchOptimizer {
	return &MatchOptimizer{
		matchHistory: make([]map[model.Role]string, 0),
	}
}

func (mo *MatchOptimizer) AddMatchHistory(match map[model.Role]string) {
	mo.matchHistory = append(mo.matchHistory, match)
}
