package model

type GameStatus struct {
	Day              int              // 日付
	MediumResult     *Judge           // 霊媒師の結果
	DivineResult     *Judge           // 占い師の結果
	ExecutedAgent    *Agent           // 処刑されたエージェント
	AttackedAgent    *Agent           // 襲撃されたエージェント
	Guard            *Guard           // 護衛されたエージェント
	VoteList         []Vote           // 投票リスト
	AttackVoteList   []Vote           // 襲撃投票リスト
	TalkList         []Talk           // 会話リスト
	WhisperList      []Talk           // 囁きリスト
	StatusMap        map[Agent]Status // エージェントと生存の対応
	RemainTalkMap    map[Agent]int    // エージェントと残り発言回数の対応
	RemainWhisperMap map[Agent]int    // エージェントと残り囁き回数の対応
}

func NewInitializeGameStatus(agents []*Agent) GameStatus {
	status := GameStatus{
		Day:              0,
		MediumResult:     nil,
		DivineResult:     nil,
		ExecutedAgent:    nil,
		AttackedAgent:    nil,
		Guard:            nil,
		VoteList:         []Vote{},
		AttackVoteList:   []Vote{},
		TalkList:         []Talk{},
		WhisperList:      []Talk{},
		StatusMap:        make(map[Agent]Status),
		RemainTalkMap:    make(map[Agent]int),
		RemainWhisperMap: make(map[Agent]int),
	}
	for _, agent := range agents {
		status.StatusMap[*agent] = S_ALIVE
	}
	return status
}

func (g GameStatus) NextDay() GameStatus {
	status := GameStatus{
		Day:              g.Day + 1,
		MediumResult:     nil,
		DivineResult:     nil,
		ExecutedAgent:    nil,
		AttackedAgent:    nil,
		Guard:            nil,
		VoteList:         []Vote{},
		AttackVoteList:   []Vote{},
		TalkList:         []Talk{},
		WhisperList:      []Talk{},
		StatusMap:        make(map[Agent]Status),
		RemainTalkMap:    make(map[Agent]int),
		RemainWhisperMap: make(map[Agent]int),
	}
	for agent, s := range g.StatusMap {
		status.StatusMap[agent] = s
	}
	return status
}

func (g GameStatus) ResetRemainTalkMap(count int) {
	for agent, status := range g.StatusMap {
		if status == S_ALIVE {
			g.RemainTalkMap[agent] = count
		} else {
			g.RemainTalkMap[agent] = 0
		}
	}
}

func (g GameStatus) ResetRemainWhisperMap(count int) {
	for agent, status := range g.StatusMap {
		if status == S_ALIVE {
			if agent.Role == R_WEREWOLF {
				g.RemainWhisperMap[agent] = count
			} else {
				g.RemainWhisperMap[agent] = 0
			}
		} else {
			g.RemainWhisperMap[agent] = 0
		}
	}
}
