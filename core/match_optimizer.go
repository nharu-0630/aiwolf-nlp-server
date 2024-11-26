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
	outputPath       string                 `json:"-"`
	TeamCount        int                    `json:"team_count"`
	GameCount        int                    `json:"game_count"`
	RoleNumMap       map[model.Role]int     `json:"role_num_map"`
	IdxTeamMap       map[int]string         `json:"idx_team_map"`
	ScheduledMatches []MatchWeight          `json:"scheduled_matches"`
	EndedMatches     []map[model.Role][]int `json:"ended_matches"`
}

func (mo *MatchOptimizer) MarshalJSON() ([]byte, error) {
	roleNumMap := make(map[string]int)
	for k, v := range mo.RoleNumMap {
		roleNumMap[k.String()] = v
	}
	endedMatches := make([]map[string][]int, len(mo.EndedMatches))
	for i, match := range mo.EndedMatches {
		endedMatches[i] = make(map[string][]int)
		for role, idxs := range match {
			endedMatches[i][role.String()] = idxs
		}
	}
	type Alias MatchOptimizer
	return json.Marshal(&struct {
		*Alias
		RoleNumMap   map[string]int     `json:"role_num_map"`
		EndedMatches []map[string][]int `json:"ended_matches"`
	}{
		Alias:        (*Alias)(mo),
		RoleNumMap:   roleNumMap,
		EndedMatches: endedMatches,
	})
}

func (mo *MatchOptimizer) UnmarshalJSON(data []byte) error {
	type Alias MatchOptimizer
	aux := &struct {
		*Alias
		RoleNumMap       map[string]int     `json:"role_num_map"`
		EndedMatches     []map[string][]int `json:"ended_matches"`
		ScheduledMatches []struct {
			RoleIdxs map[string][]int `json:"role_idxs"`
			Weight   float64          `json:"weight"`
		} `json:"scheduled_matches"`
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
	mo.EndedMatches = make([]map[model.Role][]int, len(aux.EndedMatches))
	for i, match := range aux.EndedMatches {
		mo.EndedMatches[i] = make(map[model.Role][]int)
		for role, idxs := range match {
			mo.EndedMatches[i][model.RoleFromString(role)] = idxs
		}
	}
	mo.ScheduledMatches = make([]MatchWeight, len(aux.ScheduledMatches))
	for i, match := range aux.ScheduledMatches {
		mw := MatchWeight{
			RoleIdxs: make(map[model.Role][]int),
			Weight:   match.Weight,
		}
		for role, idxs := range match.RoleIdxs {
			mw.RoleIdxs[model.RoleFromString(role)] = idxs
		}
		mo.ScheduledMatches[i] = mw
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
	mo.outputPath = config.MatchOptimizer.OutputPath
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
		outputPath: config.MatchOptimizer.OutputPath,
		RoleNumMap: roleNumMap,
		IdxTeamMap: map[int]string{},
	}
	mo.initialize()
	return mo, nil
}

func (mo *MatchOptimizer) getScheduledMatchesWithTeam() []map[model.Role][]string {
	matches := []map[model.Role][]string{}
	for _, match := range mo.ScheduledMatches {
		idxMatch := make(map[model.Role][]string)
		for role, idxs := range match.RoleIdxs {
			idxMatch[role] = make([]string, len(idxs))
			for i, idx := range idxs {
				idxMatch[role][i] = mo.IdxTeamMap[idx]
			}
		}
		matches = append(matches, idxMatch)
	}
	return matches
}

func (mo *MatchOptimizer) updateTeam(team string) {
	for _, t := range mo.IdxTeamMap {
		if t == team {
			slog.Info("チームが既に登録されています", "team", team)
			return
		}
	}
	idx := len(mo.IdxTeamMap)
	if idx >= mo.TeamCount {
		slog.Warn("チーム数が上限に達しているため追加できません", "team", team)
		return
	}
	mo.IdxTeamMap[idx] = team
	slog.Info("チームを追加しました", "team", team, "idx", idx)
	mo.save()
}

func (mo *MatchOptimizer) initialize() error {
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
			for _, match := range scheduledMatches {
				mw := MatchWeight{
					RoleIdxs: match,
					Weight:   1.0,
				}
				mo.ScheduledMatches = append(mo.ScheduledMatches, mw)
			}
			mo.save()
			slog.Info("マッチング最適化が成功しました", "attempts", attempt+1)
			return nil
		}
	}

	if bestScheduledMatches != nil {
		for _, match := range bestScheduledMatches {
			mw := MatchWeight{
				RoleIdxs: match,
				Weight:   1.0,
			}
			mo.ScheduledMatches = append(mo.ScheduledMatches, mw)
		}
		mo.save()
		slog.Info("最良の解を採用します", "bestDeviation", bestDeviation)
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

	for i, scheduledMatch := range mo.ScheduledMatches {
		if scheduledMatch.Equal(MatchWeight{RoleIdxs: idxMatch}) {
			mo.ScheduledMatches = append(mo.ScheduledMatches[:i], mo.ScheduledMatches[i+1:]...)
			slog.Info("スケジュールされたマッチから削除しました", "length", len(mo.ScheduledMatches))
			break
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
	dir := filepath.Dir(mo.outputPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0755)
	}
	file, err := os.Create(mo.outputPath)
	if err != nil {
		return err
	}
	defer file.Close()
	file.Write(jsonData)
	return nil
}
