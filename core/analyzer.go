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

	slog.Info("マッチオプティマイザの統計データを分析します")
	for idx, team := range mo.IdxTeamMap {
		slog.Info("登録済みチームを取得しました", "idx", idx, "team", team)

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
		slog.Info("スケジュールされた役職を取得しました", "idx", idx, "roles", scheduledRoles)

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
		slog.Info("終了した役職を取得しました", "idx", idx, "roles", endedRoles)
	}

	if config.DeprecatedLogService.Enable {
		slog.Info("ログサービスの統計データを分析します")

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
			var winSide *model.Team

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				values := strings.Split(line, ",")
				if len(values) == 6 && values[1] == "status" {
					if values[0] == "0" {
						team := strings.TrimRight(values[5], "1234567890")
						role := model.RoleFromString(values[3])
						teamsRole[team] = role
					} else {
						team := strings.TrimRight(values[5], "1234567890")
						status := values[4]
						if status == "" {
							errorTeams = append(errorTeams, team)
						}
					}
				}
				if len(values) == 5 && values[1] == "result" {
					side := model.TeamFromString(values[4])
					winSide = &side
				}
			}

			if len(teamsRole) == 0 {
				slog.Warn("役職が取得できませんでした", "file", filePath)
				continue
			}

			if winSide == nil {
				slog.Warn("結果が取得できませんでした", "file", filePath)
				continue
			}

			for team, role := range teamsRole {
				if _, exists := counts[team]; !exists {
					counts[team] = make(map[model.Role]*Count)
				}
				if _, exists := counts[team][role]; !exists {
					counts[team][role] = &Count{}
				}

				if slices.Contains(errorTeams, team) {
					counts[team][role].Error++
				}

				if *winSide == model.T_NONE {
					counts[team][role].None++
				} else {
					counts[team][role].Succeed++

					if *winSide == model.T_VILLAGER && role != model.R_WEREWOLF && role != model.R_POSSESSED {
						counts[team][role].Win++
					} else if *winSide == model.T_WEREWOLF && (role == model.R_WEREWOLF || role == model.R_POSSESSED) {
						counts[team][role].Win++
					} else {
						counts[team][role].Lose++
					}
				}
			}
		}

		for team, roles := range counts {
			count := &Count{}
			for role, count := range roles {
				slog.Info("統計データを取得しました", "team", team, "role", role, "win", count.Win, "lose", count.Lose, "error", count.Error, "none", count.None, "succeed", count.Succeed)
				count.Win += count.Win
				count.Lose += count.Lose
				count.Error += count.Error
				count.None += count.None
				count.Succeed += count.Succeed
			}
			slog.Info("統計データを取得しました", "team", team, "win", count.Win, "lose", count.Lose, "error", count.Error, "none", count.None, "succeed", count.Succeed)
		}
	}
}

type Count struct {
	Succeed int
	None    int
	Win     int
	Lose    int
	Error   int
}
