package model

type Role struct {
	Team    Team
	Species Species
}

var (
	BODYGUARD = Role{Team: T_VILLAGER, Species: S_HUMAN}
	FREEMASON = Role{Team: T_VILLAGER, Species: S_HUMAN}
	MEDIUM    = Role{Team: T_VILLAGER, Species: S_HUMAN}
	POSSESSED = Role{Team: T_WEREWOLF, Species: S_HUMAN}
	SEER      = Role{Team: T_VILLAGER, Species: S_HUMAN}
	VILLAGER  = Role{Team: T_VILLAGER, Species: S_HUMAN}
	WEREWOLF  = Role{Team: T_WEREWOLF, Species: S_WEREWOLF}
	FOX       = Role{Team: T_OTHERS, Species: S_HUMAN}
	ANY       = Role{Team: T_ANY, Species: S_ANY}
)
