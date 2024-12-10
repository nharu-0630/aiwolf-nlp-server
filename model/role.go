package model

import (
	"encoding/json"
)

type Role struct {
	Name    string
	Team    Team
	Species Species
}

var (
	R_WEREWOLF  = Role{Name: "WEREWOLF", Team: T_WEREWOLF, Species: S_WEREWOLF}
	R_POSSESSED = Role{Name: "POSSESSED", Team: T_WEREWOLF, Species: S_HUMAN}
	R_SEER      = Role{Name: "SEER", Team: T_VILLAGER, Species: S_HUMAN}
	R_BODYGUARD = Role{Name: "BODYGUARD", Team: T_VILLAGER, Species: S_HUMAN}
	R_VILLAGER  = Role{Name: "VILLAGER", Team: T_VILLAGER, Species: S_HUMAN}
	R_MEDIUM    = Role{Name: "MEDIUM", Team: T_VILLAGER, Species: S_HUMAN}
)

type Team string

const (
	T_VILLAGER Team = "VILLAGER"
	T_WEREWOLF Team = "WEREWOLF"
	T_NONE     Team = "NONE"
)

func TeamFromString(s string) Team {
	switch s {
	case "VILLAGER":
		return T_VILLAGER
	case "WEREWOLF":
		return T_WEREWOLF
	}
	return T_NONE
}

type Species string

const (
	S_HUMAN    Species = "HUMAN"
	S_WEREWOLF Species = "WEREWOLF"
)

func SpeciesFromString(s string) Species {
	switch s {
	case "HUMAN":
		return S_HUMAN
	case "WEREWOLF":
		return S_WEREWOLF
	}
	return ""
}

func (r Role) String() string {
	return r.Name
}

func (r Role) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

func RoleFromString(s string) Role {
	switch s {
	case "WEREWOLF":
		return R_WEREWOLF
	case "POSSESSED":
		return R_POSSESSED
	case "SEER":
		return R_SEER
	case "BODYGUARD":
		return R_BODYGUARD
	case "VILLAGER":
		return R_VILLAGER
	case "MEDIUM":
		return R_MEDIUM
	}
	return Role{}
}

func Roles(num int) map[Role]int {
	switch num {
	case 5:
		return map[Role]int{
			R_WEREWOLF:  1,
			R_POSSESSED: 1,
			R_SEER:      1,
			R_BODYGUARD: 0,
			R_VILLAGER:  2,
			R_MEDIUM:    0,
		}
	}
	return nil
}
