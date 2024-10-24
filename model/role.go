package model

import "encoding/json"

type Role struct {
	Name    string  // 名前
	Team    Team    // 陣営
	Species Species // 種族
}

var (
	R_WEREWOLF  = Role{Name: "WEREWOLF", Team: T_WEREWOLF, Species: S_WEREWOLF} // 人狼
	R_POSSESSED = Role{Name: "POSSESSED", Team: T_WEREWOLF, Species: S_HUMAN}   // 狂人
	R_SEER      = Role{Name: "SEER", Team: T_VILLAGER, Species: S_HUMAN}        // 占い師
	R_BODYGUARD = Role{Name: "BODYGUARD", Team: T_VILLAGER, Species: S_HUMAN}   // 騎士
	R_VILLAGER  = Role{Name: "VILLAGER", Team: T_VILLAGER, Species: S_HUMAN}    // 市民
	R_MEDIUM    = Role{Name: "MEDIUM", Team: T_VILLAGER, Species: S_HUMAN}      // 霊媒師
)

type Team string

const (
	T_VILLAGER Team = "VILLAGER" // 市民陣営
	T_WEREWOLF Team = "WEREWOLF" // 人狼陣営
	T_NONE     Team = "NONE"     // なし
)

type Species string

const (
	S_HUMAN    Species = "HUMAN"    // 人間
	S_WEREWOLF Species = "WEREWOLF" // 人狼
)

func (r Role) String() string {
	return r.Name
}

func (r Role) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

func Roles(num int) map[Role]int {
	switch num {
	case 5:
		return map[Role]int{
			R_WEREWOLF:  1,
			R_POSSESSED: 0,
			R_SEER:      1,
			R_BODYGUARD: 0,
			R_VILLAGER:  3,
			R_MEDIUM:    0,
		}
	}
	return nil
}
