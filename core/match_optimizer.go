package core

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/nharu-0630/aiwolf-nlp-server/config"
	"github.com/nharu-0630/aiwolf-nlp-server/model"
)

type MatchOptimizer struct {
	matchHistory []map[model.Role][]string
	outputPath   string
}

func NewMatchOptimizer() *MatchOptimizer {
	mo := &MatchOptimizer{
		matchHistory: make([]map[model.Role][]string, 0),
		outputPath:   config.MATCH_OPTIMIZER_PATH,
	}
	mo.loadMatchHistory()
	slog.Info("マッチオプティマイザを作成しました", "length", len(mo.matchHistory))
	return mo
}

func (mo *MatchOptimizer) AddMatchHistory(match map[model.Role][]string) {
	mo.matchHistory = append(mo.matchHistory, match)
	slog.Info("マッチ履歴を追加しました", "length", len(mo.matchHistory))
	mo.saveMatchHistory()
}

func (mo *MatchOptimizer) loadMatchHistory() {
	file, err := os.Open(mo.outputPath)
	if err != nil {
		slog.Error("マッチ履歴の読み込みに失敗しました", "error", err)
		return
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	var loadedHistory []map[string][]string
	err = decoder.Decode(&loadedHistory)
	if err != nil {
		slog.Error("マッチ履歴の読み込みに失敗しました", "error", err)
		return
	}
	for _, match := range loadedHistory {
		convertedMatch := make(map[model.Role][]string)
		for roleStr, team := range match {
			role := model.RoleFromString(roleStr)
			convertedMatch[role] = team
		}
		mo.matchHistory = append(mo.matchHistory, convertedMatch)
	}
}

func (mo *MatchOptimizer) saveMatchHistory() {
	convertedHistory := make([]map[string][]string, len(mo.matchHistory))
	for i, match := range mo.matchHistory {
		convertedMatch := make(map[string][]string)
		for role, team := range match {
			convertedMatch[role.String()] = team
		}
		convertedHistory[i] = convertedMatch
	}
	jsonData, err := json.Marshal(convertedHistory)
	if err != nil {
		slog.Error("マッチ履歴の保存に失敗しました", "error", err)
		return
	}
	dir := filepath.Dir(mo.outputPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
	file, err := os.Create(mo.outputPath)
	if err != nil {
		slog.Error("マッチ履歴の保存に失敗しました", "error", err)
		return
	}
	defer file.Close()
	file.Write(jsonData)
}
