package test

import (
	"testing"

	"github.com/kano-lab/aiwolf-nlp-server/core"
	"github.com/kano-lab/aiwolf-nlp-server/model"
)

func TestMatchOptimizer(t *testing.T) {
	config, err := model.LoadConfigFromPath("../config/debug.yml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	_, err = core.NewMatchOptimizerFromConfig(*config)
	if err != nil {
		t.Fatalf("Failed to create MatchOptimizer: %v", err)
	}
}
