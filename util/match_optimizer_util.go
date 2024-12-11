package util

import (
	"log/slog"
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

func findBestTeam(
	availableTeams []int,
	usedTeams map[int]bool,
	teamRoleCounts map[int]map[model.Role]int,
	role model.Role,
	theoretical map[model.Role]float64,
) int {
	bestTeam := -1
	minDeviation := -1.0
	minSubDeviation := -1
	for _, team := range availableTeams {
		slog.Info("usedTeams", "usedTeams", usedTeams)
		if usedTeams[team] {
			slog.Info("continue team", "team", team)
			continue
		}

		// 理論値を超える場合はスキップ
		if float64(teamRoleCounts[team][role]) >= theoretical[role] {
			slog.Info("continue role", "role", role)
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

		subDeviation := 0
		for _, count := range teamRoleCounts[team] {
			subDeviation += count
		}

		if bestTeam == -1 || (deviation < minDeviation && subDeviation <= minSubDeviation) {
			bestTeam = team
			minDeviation = deviation
			minSubDeviation = subDeviation
		}

		slog.Info("usedTeams", "usedTeams", usedTeams)
	}
	return bestTeam
}

func GenerateMatches(gameCount int, teamCount int, roles []model.Role, theoretical map[model.Role]float64) ([]map[model.Role][]int, map[int]map[model.Role]int, bool) {
	counts := make(map[int]map[model.Role]int)
	for i := 0; i < teamCount; i++ {
		counts[i] = make(map[model.Role]int)
		for role := range theoretical {
			counts[i][role] = 0
		}
	}

	matches := []map[model.Role][]int{}
	success := true

	for i := 0; i < gameCount; i++ {
		availableTeams := make([]int, teamCount)
		for j := 0; j < teamCount; j++ {
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
			bestTeam := findBestTeam(availableTeams, usedTeams, counts, role, theoretical)
			if bestTeam == -1 {
				success = false
				break
			}
			match[role] = append(match[role], bestTeam)
			usedTeams[bestTeam] = true
			counts[bestTeam][role]++
		}

		if !success {
			break
		}

		matches = append(matches, match)
	}
	return matches, counts, true
}

func CalcDeviation(counts map[int]map[model.Role]int, theoretical map[model.Role]float64) float64 {
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
