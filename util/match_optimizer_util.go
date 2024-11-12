package util

import "github.com/kano-lab/aiwolf-nlp-server/model"

func EqualMatch(a map[model.Role][]int, b map[model.Role][]int) bool {
	if len(a) != len(b) {
		return false
	}
	for role, idxs := range a {
		targetIdxs, exists := b[role]
		if !exists || len(idxs) != len(targetIdxs) {
			return false
		}

		idxCount := make(map[int]int)
		for _, val := range idxs {
			idxCount[val]++
		}

		targetCount := make(map[int]int)
		for _, val := range targetIdxs {
			targetCount[val]++
		}

		if len(idxCount) != len(targetCount) {
			return false
		}
		for key, countA := range idxCount {
			if countB, exists := targetCount[key]; !exists || countA != countB {
				return false
			}
		}
	}
	return true
}
