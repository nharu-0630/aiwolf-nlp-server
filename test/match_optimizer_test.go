package test

import (
	"testing"

	"github.com/kano-lab/aiwolf-nlp-server/core"
	"github.com/kano-lab/aiwolf-nlp-server/model"
)

func TestInitializeMatchOptimizer(t *testing.T) {
	config, err := model.LoadFromPath("../config/debug.yml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	_, err = core.NewMatchOptimizerFromConfig(*config)
	if err != nil {
		t.Fatalf("Failed to create MatchOptimizer: %v", err)
	}
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
