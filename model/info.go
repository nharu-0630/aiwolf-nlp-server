package model

import "encoding/json"

type Info struct {
	Day              int              `json:"day"`                     // 日付
	Agent            Agent            `json:"agent,omitempty"`         // 自身のエージェント
	MediumResult     Judge            `json:"mediumResult,omitempty"`  // 霊媒師の結果
	DivineResult     Judge            `json:"divineResult,omitempty"`  // 占い師の結果
	ExecutedAgent    Agent            `json:"executedAgent,omitempty"` // 処刑されたエージェント
	AttackedAgent    Agent            `json:"attackedAgent,omitempty"` // 襲撃されたエージェント
	VoteList         []Vote           `json:"voteList"`                // 投票リスト
	AttackVoteList   []Vote           `json:"attackVoteList"`          // 襲撃投票リスト
	TalkList         []Talk           `json:"talkList"`                // 会話リスト
	WhisperList      []Talk           `json:"whisperList"`             // 囁きリスト
	StatusMap        map[Agent]Status `json:"statusMap"`               // エージェントと生死の対応
	RoleMap          map[Agent]Role   `json:"roleMap"`                 // エージェントと役職の対応
	RemainTalkMap    map[Agent]int    `json:"remainTalkMap"`           // エージェントと残り発言回数の対応
	RemainWhisperMap map[Agent]int    `json:"remainWhisperMap"`        // エージェントと残り囁き回数の対応
}

func (i Info) MarshalJSON() ([]byte, error) {
	statusMap := make(map[string]Status)
	for k, v := range i.StatusMap {
		statusMap[k.String()] = v
	}
	roleMap := make(map[string]Role)
	for k, v := range i.RoleMap {
		roleMap[k.String()] = v
	}
	remainTalkMap := make(map[string]int)
	for k, v := range i.RemainTalkMap {
		remainTalkMap[k.String()] = v
	}
	remainWhisperMap := make(map[string]int)
	for k, v := range i.RemainWhisperMap {
		remainWhisperMap[k.String()] = v
	}
	type Alias Info
	return json.Marshal(&struct {
		StatusMap        map[string]Status `json:"statusMap"`
		RoleMap          map[string]Role   `json:"roleMap"`
		RemainTalkMap    map[string]int    `json:"remainTalkMap"`
		RemainWhisperMap map[string]int    `json:"remainWhisperMap"`
		*Alias
	}{
		StatusMap:        statusMap,
		RoleMap:          roleMap,
		RemainTalkMap:    remainTalkMap,
		RemainWhisperMap: remainWhisperMap,
		Alias:            (*Alias)(&i),
	})
}

func (i Info) IsEmpty() bool {
	return i.Day == 0 &&
		i.Agent == Agent{} &&
		i.MediumResult == Judge{} &&
		i.DivineResult == Judge{} &&
		i.ExecutedAgent == Agent{} &&
		i.AttackedAgent == Agent{} &&
		len(i.VoteList) == 0 &&
		len(i.AttackVoteList) == 0 &&
		len(i.TalkList) == 0 &&
		len(i.WhisperList) == 0 &&
		len(i.StatusMap) == 0 &&
		len(i.RoleMap) == 0 &&
		len(i.RemainTalkMap) == 0 &&
		len(i.RemainWhisperMap) == 0
}
