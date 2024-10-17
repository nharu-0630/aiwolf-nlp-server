package model

import (
	"encoding/json"
	"errors"

	"github.com/nharu-0630/aiwolf-nlp-server/config"
)

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

func NewSettings() (*Settings, error) {
	roleNumMap := Roles(config.GAME_AGENT_COUNT)
	if roleNumMap == nil {
		return nil, errors.New("対応する役職の人数がありません")
	}
	settings := &Settings{
		RoleNumMap:       *roleNumMap,
		MaxTalk:          config.MAX_TALK_COUNT_PER_AGENT,
		MaxTalkTurn:      config.MAX_TALK_COUNT,
		MaxWhisper:       config.MAX_WHISPER_COUNT_PER_AGENT,
		MaxWhisperTurn:   config.MAX_WHISPER_COUNT,
		MaxSkip:          config.MAX_SKIP_COUNT,
		IsEnableNoAttack: config.IS_ENABLE_NO_ATTACK,
		IsVoteVisible:    config.IS_VOTE_VISIBLE,
		IsTalkOnFirstDay: config.IS_TALK_ON_FIRST_DAY,
		ResponseTimeout:  config.RESPONSE_TIMEOUT,
		ActionTimeout:    config.ACTION_TIMEOUT,
		MaxRevote:        config.MAX_REVOTE_COUNT,
		MaxAttackRevote:  config.MAX_ATTACK_REVOTE_COUNT,
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
