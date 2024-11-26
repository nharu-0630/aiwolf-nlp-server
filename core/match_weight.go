package core

import (
	"encoding/json"

	"github.com/kano-lab/aiwolf-nlp-server/model"
)

type MatchWeight struct {
	RoleIdxs map[model.Role][]int `json:"role_idxs"`
	Weight   float64              `json:"weight"`
}

func (mw MatchWeight) Less(other MatchWeight) bool {
	return mw.Weight < other.Weight
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

func (mw MatchWeight) UnmarshalJSON(data []byte) error {
	aux := &struct {
		RoleIdxs map[string][]int `json:"role_idxs"`
		Weight   float64          `json:"weight"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	mw.RoleIdxs = make(map[model.Role][]int)
	for role, idxs := range aux.RoleIdxs {
		mw.RoleIdxs[model.RoleFromString(role)] = idxs
	}
	mw.Weight = aux.Weight
	return nil
}
