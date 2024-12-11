package test

import (
	"log/slog"
	"testing"

	"github.com/kano-lab/aiwolf-nlp-server/core"
	"github.com/kano-lab/aiwolf-nlp-server/model"
)

func TestInitializeMatchOptimizer(t *testing.T) {
	config, err := model.LoadFromPath("../config/debug.yml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	mo, err := core.NewMatchOptimizerFromConfig(*config)
	if err != nil {
		t.Fatalf("Failed to create MatchOptimizer: %v", err)
	}

	roleCounts := make(map[int]map[model.Role]int)
	for i := 0; i < mo.TeamCount; i++ {
		roleCounts[i] = make(map[model.Role]int)
	}
	for _, match := range mo.ScheduledMatches {
		for role, idxs := range match.RoleIdxs {
			for _, idx := range idxs {
				roleCounts[idx][role]++
			}
		}
	}
	t.Log(roleCounts)
	slog.Info("チームの役職統計を取得しました", "role_counts", roleCounts)
}

func TestLoadMatchOptimizer(t *testing.T) {
	config, err := model.LoadFromPath("../config/debug.yml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	mo, err := core.NewMatchOptimizer(*config)
	if err != nil {
		t.Fatalf("Failed to create MatchOptimizer: %v", err)
	}
	t.Log(mo)
}
