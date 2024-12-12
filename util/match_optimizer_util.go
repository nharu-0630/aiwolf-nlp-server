package util

import (
	"math"
	"math/rand"

	"github.com/kano-lab/aiwolf-nlp-server/model"
)

func teamToIdx(idxTeamMap map[int]string, team string) int {
	for idx, agent := range idxTeamMap {
		if agent == team {
			return idx
		}
	}
	return -1
}

func TeamNameMatchToIdxMatch(idxTeamMap map[int]string, match map[model.Role][]string) map[model.Role][]int {
	idxMatch := make(map[model.Role][]int)
	for role, teams := range match {
		idxMatch[role] = make([]int, len(teams))
		for i, team := range teams {
			idxMatch[role][i] = teamToIdx(idxTeamMap, team)
		}
	}
	return idxMatch
}

func IdxMatchToTeamNameMatch(idxTeamMap map[int]string, match map[model.Role][]int) map[model.Role][]string {
	teamMatch := make(map[model.Role][]string)
	for role, idxs := range match {
		teamMatch[role] = make([]string, len(idxs))
		for i, idx := range idxs {
			teamMatch[role][i] = idxTeamMap[idx]
		}
	}
	return teamMatch
}

func findBestIdx(
	idxs map[int]bool,
	idxRoleCounts map[int]map[model.Role]int,
	targetRole model.Role,
	theoretical map[model.Role]float64,
) int {
	bestIdx := -1
	minDeviation := math.MaxFloat64
	minSubDeviation := 0
	for idx, state := range idxs {
		if !state {
			continue
		}
		// 全ロールの理論値からの偏差を計算
		deviationByTheoretical := 0.0
		for role, value := range theoretical {
			count := float64(idxRoleCounts[idx][role])
			if role == targetRole {
				count++
			}
			deviationByTheoretical += (count - value) * (count - value)
		}
		// 最も偏差が小さいチームを選択
		deviationByTeams := 0
		for _, count := range idxRoleCounts[idx] {
			deviationByTeams += count
		}
		if bestIdx == -1 || (deviationByTheoretical < minDeviation && deviationByTeams <= minSubDeviation) {
			bestIdx = idx
			minDeviation = deviationByTheoretical
			minSubDeviation = deviationByTeams
		}
	}
	return bestIdx
}

func GenerateMatches(gameCount int, teamCount int, roles []model.Role, theoretical map[model.Role]float64) ([]map[model.Role][]int, float64) {
	matches := []map[model.Role][]int{}
	failed := false

	idxRoleCounts := make(map[int]map[model.Role]int)
	for i := 0; i < teamCount; i++ {
		idxRoleCounts[i] = make(map[model.Role]int)
		for role := range theoretical {
			idxRoleCounts[i][role] = 0
		}
	}

	for i := 0; i < gameCount; i++ {
		match := make(map[model.Role][]int)
		idxs := make(map[int]bool)
		for j := 0; j < teamCount; j++ {
			idxs[j] = true
		}

		shuffledRoles := append([]model.Role{}, roles...)
		rand.Shuffle(len(shuffledRoles), func(i, j int) {
			shuffledRoles[i], shuffledRoles[j] = shuffledRoles[j], shuffledRoles[i]
		})

		for _, role := range shuffledRoles {
			bestIdx := findBestIdx(idxs, idxRoleCounts, role, theoretical)
			if bestIdx == -1 {
				failed = true
				break
			}
			match[role] = append(match[role], bestIdx)
			idxs[bestIdx] = true
			idxRoleCounts[bestIdx][role]++
		}
		if failed {
			break
		}
		matches = append(matches, match)
	}
	deviation := CalcDeviation(idxRoleCounts, theoretical)
	return matches, deviation
}

func CalcDeviation(counts map[int]map[model.Role]int, theoretical map[model.Role]float64) float64 {
	// 偏差を計算
	if len(counts) == 0 {
		return -1.0
	}
	deviation := 0.0
	for _, roleCounts := range counts {
		for role, theoreticalValue := range theoretical {
			count := float64(roleCounts[role])
			deviation += (count - theoreticalValue) * (count - theoreticalValue)
		}
	}
	return deviation
}

func CalcTheoretical(roleNumMap map[model.Role]int, gameCount int, teamCount int) (map[model.Role]float64, []model.Role) {
	// 各ロールの理論的な出現回数を計算
	theoretical := make(map[model.Role]float64)
	var roles []model.Role
	for role, num := range roleNumMap {
		if num > 0 {
			theoretical[role] = float64(num*gameCount) / float64(teamCount)
			for i := 0; i < num; i++ {
				roles = append(roles, role)
			}
		}
	}
	return theoretical, roles
}
