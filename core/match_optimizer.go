package core

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/kano-lab/aiwolf-nlp-server/model"
	"golang.org/x/exp/rand"
)

type MatchOptimizer struct {
	TeamCount        int                    `json:"team_count"`
	GameCount        int                    `json:"game_count"`
	OutputPath       string                 `json:"output_path"`
	RoleNumMap       map[model.Role]int     `json:"role_num_map"`
	IdxTeamMap       map[int]string         `json:"idx_team_map"`
	ScheduledMatches []map[model.Role][]int `json:"scheduled_matches"`
	EndedMatches     []map[model.Role][]int `json:"ended_matches"`
}

func (mo *MatchOptimizer) MarshalJSON() ([]byte, error) {
	roleNumMap := make(map[string]int)
	for k, v := range mo.RoleNumMap {
		roleNumMap[k.String()] = v
	}
	scheduledMatches := make([]map[string][]int, len(mo.ScheduledMatches))
	for i, match := range mo.ScheduledMatches {
		scheduledMatches[i] = make(map[string][]int)
		for role, idx := range match {
			scheduledMatches[i][role.String()] = idx
		}
	}
	endedMatches := make([]map[string][]int, len(mo.EndedMatches))
	for i, match := range mo.EndedMatches {
		endedMatches[i] = make(map[string][]int)
		for role, idx := range match {
			endedMatches[i][role.String()] = idx
		}
	}
	type Alias MatchOptimizer
	return json.Marshal(&struct {
		*Alias
		RoleNumMap       map[string]int     `json:"role_num_map"`
		ScheduledMatches []map[string][]int `json:"scheduled_matches"`
		EndedMatches     []map[string][]int `json:"ended_matches"`
	}{
		Alias:            (*Alias)(mo),
		RoleNumMap:       roleNumMap,
		ScheduledMatches: scheduledMatches,
		EndedMatches:     endedMatches,
	})
}

func (mo *MatchOptimizer) UnmarshalJSON(data []byte) error {
	type Alias MatchOptimizer
	aux := &struct {
		*Alias
		RoleNumMap       map[string]int     `json:"role_num_map"`
		ScheduledMatches []map[string][]int `json:"scheduled_matches"`
		EndedMatches     []map[string][]int `json:"ended_matches"`
	}{
		Alias: (*Alias)(mo),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	mo.RoleNumMap = make(map[model.Role]int)
	for role, num := range aux.RoleNumMap {
		mo.RoleNumMap[model.RoleFromString(role)] = num
	}
	mo.ScheduledMatches = make([]map[model.Role][]int, len(aux.ScheduledMatches))
	for i, match := range aux.ScheduledMatches {
		mo.ScheduledMatches[i] = make(map[model.Role][]int)
		for role, idx := range match {
			mo.ScheduledMatches[i][model.RoleFromString(role)] = idx
		}
	}
	mo.EndedMatches = make([]map[model.Role][]int, len(aux.EndedMatches))
	for i, match := range aux.EndedMatches {
		mo.EndedMatches[i] = make(map[model.Role][]int)
		for role, idx := range match {
			mo.EndedMatches[i][model.RoleFromString(role)] = idx
		}
	}
	return nil
}

func NewMatchOptimizer(config model.Config) (*MatchOptimizer, error) {
	data, err := os.ReadFile(config.MatchOptimizer.OutputPath)
	if err != nil {
		slog.Warn("マッチオプティマイザの読み込みに失敗しました", "error", err)
		return NewMatchOptimizerFromConfig(config)
	}
	var mo MatchOptimizer
	if err := json.Unmarshal(data, &mo); err != nil {
		slog.Error("マッチオプティマイザのパースに失敗しました", "error", err)
		return nil, err
	}
	mo.save()
	return &mo, nil
}

func NewMatchOptimizerFromConfig(config model.Config) (*MatchOptimizer, error) {
	slog.Info("マッチオプティマイザを作成します")
	roleNumMap := model.Roles(config.Game.AgentCount)
	if roleNumMap == nil {
		return nil, errors.New("対応する役職の人数がありません")
	}
	mo := &MatchOptimizer{
		TeamCount:  config.MatchOptimizer.TeamCount,
		GameCount:  config.MatchOptimizer.GameCount,
		OutputPath: config.MatchOptimizer.OutputPath,
		RoleNumMap: roleNumMap,
		IdxTeamMap: map[int]string{},
	}
	mo.Initialize()
	return mo, nil
}

func (mo *MatchOptimizer) GetScheduledMatchesWithTeam() []map[model.Role][]string {
	matches := []map[model.Role][]string{}
	for _, match := range mo.ScheduledMatches {
		idxMatch := make(map[model.Role][]string)
		for role, idxs := range match {
			idxMatch[role] = make([]string, len(idxs))
			for i, idx := range idxs {
				idxMatch[role][i] = mo.IdxTeamMap[idx]
			}
		}
		matches = append(matches, idxMatch)
	}
	return matches
}

func (mo *MatchOptimizer) UpdateTeam(team string) {
	for _, t := range mo.IdxTeamMap {
		if t == team {
			return
		}
	}
	idx := len(mo.IdxTeamMap)
	mo.IdxTeamMap[idx] = team
}

func (mo *MatchOptimizer) Initialize() error {
	slog.Info("マッチオプティマイザを初期化します")
	mo.EndedMatches = []map[model.Role][]int{}

	// 各ロールの理論的な出現回数を計算
	theoretical := make(map[model.Role]float64)
	var roles []model.Role
	for role, num := range mo.RoleNumMap {
		if num > 0 {
			theoretical[role] = float64(num*mo.GameCount) / float64(mo.TeamCount)
			for i := 0; i < num; i++ {
				roles = append(roles, role)
			}
		}
	}
	slog.Info("各役職の理論値を計算しました", "theoretical", theoretical)
	maxAttempts := mo.GameCount * mo.TeamCount * 5
	var bestScheduledMatches []map[model.Role][]int
	bestDeviation := -1.0

	slog.Info("マッチング最適化を開始します", "attempts", maxAttempts)

	for attempt := 0; attempt < maxAttempts; attempt++ {

		teamRoleCounts := make(map[int]map[model.Role]int)
		for i := 0; i < mo.TeamCount; i++ {
			teamRoleCounts[i] = make(map[model.Role]int)
			for role := range theoretical {
				teamRoleCounts[i][role] = 0
			}
		}

		scheduledMatches := []map[model.Role][]int{}
		success := true

		for i := 0; i < mo.GameCount; i++ {
			availableTeams := make([]int, mo.TeamCount)
			for j := 0; j < mo.TeamCount; j++ {
				availableTeams[j] = j
			}
			rand.Shuffle(len(availableTeams), func(i, j int) {
				availableTeams[i], availableTeams[j] = availableTeams[j], availableTeams[i]
			})

			match := make(map[model.Role][]int)
			usedTeams := make(map[int]bool)

			shuffledRoles := append([]model.Role{}, roles...)
			rand.Shuffle(len(shuffledRoles), func(i, j int) {
				shuffledRoles[i], shuffledRoles[j] = shuffledRoles[j], shuffledRoles[i]
			})

			for _, role := range shuffledRoles {
				bestTeam := mo.findBestTeam(availableTeams, usedTeams, teamRoleCounts, role, theoretical)
				if bestTeam == -1 {
					success = false
					break
				}

				match[role] = append(match[role], bestTeam)
				usedTeams[bestTeam] = true
				teamRoleCounts[bestTeam][role]++
			}

			if !success {
				break
			}

			scheduledMatches = append(scheduledMatches, match)
		}

		currentDeviation := 0.0
		if len(scheduledMatches) > 0 {
			for team := 0; team < mo.TeamCount; team++ {
				for role := range theoretical {
					currentCount := float64(teamRoleCounts[team][role])
					currentDeviation += (currentCount - theoretical[role]) * (currentCount - theoretical[role])
				}
			}

			if bestScheduledMatches == nil || currentDeviation < bestDeviation {
				slog.Info("より良い解が見つかりました", "deviation", currentDeviation)
				bestScheduledMatches = scheduledMatches
				bestDeviation = currentDeviation
			}
		}

		if success {
			mo.ScheduledMatches = scheduledMatches
			mo.save()
			slog.Info("マッチング最適化が成功しました", "attempts", attempt+1)
			return nil
		}
	}

	if bestScheduledMatches != nil {
		slog.Info("最良の解を採用します", "bestDeviation", bestDeviation)
		mo.ScheduledMatches = bestScheduledMatches
		mo.save()
		return nil
	}
	return errors.New("最適なマッチングが見つかりませんでした")
}

func (mo *MatchOptimizer) findBestTeam(
	availableTeams []int,
	usedTeams map[int]bool,
	teamRoleCounts map[int]map[model.Role]int,
	role model.Role,
	theoretical map[model.Role]float64,
) int {
	bestTeam := -1
	minDeviation := -1.0

	for _, team := range availableTeams {
		if usedTeams[team] {
			continue
		}

		// 理論値を超える場合はスキップ
		if float64(teamRoleCounts[team][role]+1) > theoretical[role] {
			continue
		}

		// 全ロールの理論値からの偏差を計算
		deviation := 0.0
		for r, theoreticalValue := range theoretical {
			currentCount := float64(teamRoleCounts[team][r])
			if r == role {
				currentCount++
			}
			deviation += (currentCount - theoreticalValue) * (currentCount - theoreticalValue)
		}

		if bestTeam == -1 || deviation < minDeviation {
			bestTeam = team
			minDeviation = deviation
		}
	}

	return bestTeam
}

func (mo *MatchOptimizer) teamToIdx(team string) int {
	for idx, agent := range mo.IdxTeamMap {
		if agent == team {
			return idx
		}
	}
	return -1
}

func (mo *MatchOptimizer) addEndedMatch(match map[model.Role][]string) {
	idxMatch := make(map[model.Role][]int)
	for role, teams := range match {
		idxMatch[role] = make([]int, len(teams))
		for i, team := range teams {
			idxMatch[role][i] = mo.teamToIdx(team)
		}
	}
	mo.EndedMatches = append(mo.EndedMatches, idxMatch)
	slog.Info("マッチ履歴を追加しました", "length", len(mo.EndedMatches))
	mo.save()
}

func (mo *MatchOptimizer) save() error {
	jsonData, err := json.Marshal(mo)
	if err != nil {
		return err
	}
	dir := filepath.Dir(mo.OutputPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0755)
	}
	file, err := os.Create(mo.OutputPath)
	if err != nil {
		return err
	}
	defer file.Close()
	file.Write(jsonData)
	return nil
}
