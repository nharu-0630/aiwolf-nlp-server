package model

import "encoding/json"

type Settings struct {
	RoleNumMap       map[Role]int `json:"roleNumMap"`       // 役職の人数
	MaxTalk          int          `json:"maxTalk"`          // 1エージェントが発言できる最大回数
	MaxTalkTurn      int          `json:"maxTalkTurn"`      // 1日の全体の発言回数
	MaxWhisper       int          `json:"maxWhisper"`       // 1エージェントが囁きできる最大回数
	MaxWhisperTurn   int          `json:"maxWhisperTurn"`   // 1日の全体の囁き回数
	MaxSkip          int          `json:"maxSkip"`          // 最大スキップ回数
	IsEnableNoAttack bool         `json:"isEnableNoAttack"` // 襲撃なしの日を許可するか
	IsVoteVisible    bool         `json:"isVoteVisible"`    // 投票の結果を見せるか
	IsTalkOnFirstDay bool         `json:"isTalkOnFirstDay"` // 1日目の発言を許可するか
	ResponseTimeout  int          `json:"responseTimeout"`  // タイムアウト時間
	ActionTimeout    int          `json:"actionTimeout"`    // タイムアウト時間
	MaxRevote        int          `json:"maxRevote"`        // 再投票回数
	MaxAttackRevote  int          `json:"maxAttackRevote"`  // 襲撃再投票回数
}

func (s Settings) MarshalJSON() ([]byte, error) {
	roleNumMap := make(map[string]int)
	for k, v := range s.RoleNumMap {
		roleNumMap[k.String()] = v
	}
	type Alias Settings
	return json.Marshal(&struct {
		RoleNumMap map[string]int `json:"roleNumMap"`
		*Alias
	}{
		RoleNumMap: roleNumMap,
		Alias:      (*Alias)(&s),
	})
}
