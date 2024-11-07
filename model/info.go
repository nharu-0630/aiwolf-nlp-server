package model

import "encoding/json"

type Info struct {
	Day              int              `json:"day"`                      // 日付
	Agent            *Agent           `json:"agent,omitempty"`          // 自身のエージェント
	MediumResult     *Judge           `json:"mediumResult,omitempty"`   // 霊媒師の結果
	DivineResult     *Judge           `json:"divineResult,omitempty"`   // 占い師の結果
	ExecutedAgent    *Agent           `json:"executedAgent,omitempty"`  // 処刑されたエージェント
	AttackedAgent    *Agent           `json:"attackedAgent,omitempty"`  // 襲撃されたエージェント
	VoteList         []Vote           `json:"voteList,omitempty"`       // 投票リスト
	AttackVoteList   []Vote           `json:"attackVoteList,omitempty"` // 襲撃投票リスト
	TalkList         []Talk           `json:"-"`                        // 会話リスト
	WhisperList      []Talk           `json:"-"`                        // 囁きリスト
	StatusMap        map[Agent]Status `json:"statusMap"`                // エージェントと生死の対応
	RoleMap          map[Agent]Role   `json:"roleMap"`                  // エージェントと役職の対応
	RemainTalkMap    map[Agent]int    `json:"remainTalkMap"`            // エージェントと残り発言回数の対応
	RemainWhisperMap map[Agent]int    `json:"remainWhisperMap"`         // エージェントと残り囁き回数の対応
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

func NewInfo(agent *Agent, gameStatus *GameStatus, lastGameStatus *GameStatus, settings *Settings) Info {
	info := Info{
		Day:   gameStatus.Day,
		Agent: agent,
	}
	if lastGameStatus != nil {
		if lastGameStatus.MediumResult != nil && agent.Role == R_MEDIUM {
			info.MediumResult = lastGameStatus.MediumResult
		}
		if lastGameStatus.DivineResult != nil && agent.Role == R_SEER {
			info.DivineResult = lastGameStatus.DivineResult
		}
		if lastGameStatus.ExecutedAgent != nil {
			info.ExecutedAgent = lastGameStatus.ExecutedAgent
		}
		if lastGameStatus.AttackedAgent != nil {
			info.AttackedAgent = lastGameStatus.AttackedAgent
		}
		if settings.IsVoteVisible {
			info.VoteList = lastGameStatus.VoteList
		}
		if settings.IsVoteVisible && agent.Role == R_WEREWOLF {
			info.AttackVoteList = lastGameStatus.AttackVoteList
		}
	}
	info.TalkList = gameStatus.TalkList
	if agent.Role == R_WEREWOLF {
		info.WhisperList = gameStatus.WhisperList
	}
	info.StatusMap = gameStatus.StatusMap
	roleMap := make(map[Agent]Role)
	roleMap[*agent] = agent.Role
	if agent.Role == R_WEREWOLF {
		for a := range gameStatus.StatusMap {
			if a.Role == R_WEREWOLF {
				roleMap[a] = a.Role
			}
		}
	}
	info.RoleMap = roleMap
	info.RemainTalkMap = gameStatus.RemainTalkMap
	if agent.Role == R_WEREWOLF {
		info.RemainWhisperMap = gameStatus.RemainWhisperMap
	}
	return info
}
