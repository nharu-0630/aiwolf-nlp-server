package model

import (
	"encoding/json"
	"errors"
)

type Settings struct {
	RoleNumMap       map[Role]int `json:"roleNumMap"`       // 役職の人数
	MaxTalk          int          `json:"maxTalk"`          // 1日あたりの1エージェントの最大発言回数
	MaxTalkTurn      int          `json:"maxTalkTurn"`      // 1日あたりの全体の発言回数
	MaxWhisper       int          `json:"maxWhisper"`       // 1日あたりの1エージェントの最大囁き回数
	MaxWhisperTurn   int          `json:"maxWhisperTurn"`   // 1日あたりの全体の囁き回数
	MaxSkip          int          `json:"maxSkip"`          // 1日あたりの最大スキップ回数
	IsEnableNoAttack bool         `json:"isEnableNoAttack"` // 襲撃なしの日を許可するか
	IsVoteVisible    bool         `json:"isVoteVisible"`    // 投票の結果を公開するか
	IsTalkOnFirstDay bool         `json:"isTalkOnFirstDay"` // 1日目の発言を許可するか
	ResponseTimeout  int          `json:"responseTimeout"`  // タイムアウト時間
	ActionTimeout    int          `json:"actionTimeout"`    // タイムアウト時間
	MaxRevote        int          `json:"maxRevote"`        // 最大投票回数
	MaxAttackRevote  int          `json:"maxAttackRevote"`  // 最大襲撃再投票回数
}

func NewSettings(config Config) (Settings, error) {
	roleNumMap := Roles(config.Game.AgentCount)
	if roleNumMap == nil {
		return Settings{}, errors.New("対応する役職の人数がありません")
	}
	settings := Settings{
		RoleNumMap:       roleNumMap,
		MaxTalk:          config.Game.Talk.MaxCount.PerAgent,
		MaxTalkTurn:      config.Game.Talk.MaxCount.PerDay,
		MaxWhisper:       config.Game.Whisper.MaxCount.PerAgent,
		MaxWhisperTurn:   config.Game.Whisper.MaxCount.PerDay,
		MaxSkip:          config.Game.Skip.MaxCount,
		IsEnableNoAttack: config.Game.AllowNoAttack,
		IsVoteVisible:    config.Game.VoteVisibility,
		IsTalkOnFirstDay: config.Game.TalkOnFirstDay,
		ResponseTimeout:  int(config.Game.Timeout.Response.Milliseconds()),
		ActionTimeout:    int(config.Game.Timeout.Action.Milliseconds()),
		MaxRevote:        config.Game.Vote.MaxCount,
		MaxAttackRevote:  config.Game.Attack.MaxCount,
	}
	return settings, nil
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
