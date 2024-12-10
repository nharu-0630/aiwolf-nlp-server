package core

import (
	"bufio"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/kano-lab/aiwolf-nlp-server/model"
)

func Analyzer(config model.Config) {
	data, err := os.ReadFile(config.MatchOptimizer.OutputPath)
	if err != nil {
		slog.Warn("マッチオプティマイザの読み込みに失敗しました", "error", err)
		return
	}
	var mo MatchOptimizer
	if err := json.Unmarshal(data, &mo); err != nil {
		slog.Error("マッチオプティマイザのパースに失敗しました", "error", err)
		return
	}

	slog.Info("MatchOptimizer")
	for idx, team := range mo.IdxTeamMap {
		slog.Info("IdxTeamMap", "idx", idx, "team", team)

		scheduledRoles := make(map[model.Role]int)
		for _, match := range mo.ScheduledMatches {
			for role, idxs := range match.RoleIdxs {
				for _, i := range idxs {
					if idx == i {
						scheduledRoles[role]++
					}
				}
			}
		}
		slog.Info("ScheduledRoles", "idx", idx, "roles", scheduledRoles)

		endedRoles := make(map[model.Role]int)
		for _, match := range mo.EndedMatches {
			for role, idxs := range match {
				for _, i := range idxs {
					if idx == i {
						endedRoles[role]++
					}
				}
			}
		}
		slog.Info("EndedRoles", "idx", idx, "roles", endedRoles)
	}

	if config.DeprecatedLogService.Enable {
		slog.Info("DeprecatedLogService")

		filePaths, err := filepath.Glob(filepath.Join(config.DeprecatedLogService.OutputDir, "*.log"))
		if err != nil {
			slog.Warn("ファイルの取得に失敗しました", "error", err)
		}

		counts := make(map[string]map[model.Role]*Count)

		for _, filePath := range filePaths {
			file, err := os.Open(filePath)
			if err != nil {
				slog.Warn("ファイルの読み込みに失敗しました", "error", err)
			}
			defer file.Close()

			teamsRole := make(map[string]model.Role)
			errorTeams := []string{}
			winSide := model.T_NONE

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				values := strings.Split(line, ",")
				if values[0] == "0" && values[1] == "status" {
					team := strings.TrimRight(values[5], "1234567890")
					role := model.RoleFromString(values[3])
					teamsRole[team] = role
				}
				if values[1] == "status" {
					team := strings.TrimRight(values[5], "1234567890")
					status := values[4]
					if status == "" {
						errorTeams = append(errorTeams, team)
					}
				}
				if values[1] == "result" {
					winSide = model.TeamFromString(values[4])
				}
			}

			for team, role := range teamsRole {
				if _, exists := counts[team]; !exists {
					counts[team] = make(map[model.Role]*Count)
				}
				if _, exists := counts[team][role]; !exists {
					counts[team][role] = &Count{}
				}
				counts[team][role].GameCount++
				if winSide == model.T_VILLAGER && role != model.R_WEREWOLF && role != model.R_POSSESSED {
					counts[team][role].WinCount++
				}
				if winSide == model.T_WEREWOLF && (role == model.R_WEREWOLF || role == model.R_POSSESSED) {
					counts[team][role].WinCount++
				}
				if slices.Contains(errorTeams, team) {
					counts[team][role].ErrorCount++
				}
				if winSide == model.T_NONE {
					counts[team][role].NoneCount++
				}
			}
		}

		for team, roles := range counts {
			for role, count := range roles {
				slog.Info("Count", "team", team, "role", role, "count", count)
			}
		}
	}
}

type Count struct {
	GameCount  int
	WinCount   int
	ErrorCount int
	NoneCount  int
}
