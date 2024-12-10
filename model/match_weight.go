package model

import (
	"encoding/json"
)

type MatchWeight struct {
	RoleIdxs map[Role][]int `json:"role_idxs"`
	Weight   float64        `json:"weight"`
}

func (mw MatchWeight) Equal(other MatchWeight) bool {
	if len(mw.RoleIdxs) != len(other.RoleIdxs) {
		return false
	}
	for role, idxs := range mw.RoleIdxs {
		targetIdxs, exists := other.RoleIdxs[role]
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
		for key, count := range idxCount {
			if otherCount, exists := targetCount[key]; !exists || count != otherCount {
				return false
			}
		}
	}
	return true
}

func (mw MatchWeight) MarshalJSON() ([]byte, error) {
	roleIdxs := make(map[string][]int)
	for role, idxs := range mw.RoleIdxs {
		roleIdxs[role.String()] = idxs
	}
	return json.Marshal(&struct {
		RoleIdxs map[string][]int `json:"role_idxs"`
		Weight   float64          `json:"weight"`
	}{
		RoleIdxs: roleIdxs,
		Weight:   mw.Weight,
	})
}
