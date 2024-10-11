package model

import "github.com/nharu-0630/aiwolf-nlp-server/config"

type GameStatus struct {
	Day              int              // 日付
	MediumResult     Judge            // 霊媒師の結果
	DivineResult     Judge            // 占い師の結果
	ExecutedAgent    Agent            // 処刑されたエージェント
	AttackedAgent    Agent            // 襲撃されたエージェント
	Guard            Guard            // 護衛されたエージェント
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
		MediumResult:     Judge{},
		DivineResult:     Judge{},
		ExecutedAgent:    Agent{},
		AttackedAgent:    Agent{},
		Guard:            Guard{},
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
		MediumResult:     Judge{},
		DivineResult:     Judge{},
		ExecutedAgent:    Agent{},
		AttackedAgent:    Agent{},
		Guard:            Guard{},
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

func (g GameStatus) ResetRemainTalkMap() {
	for agent, status := range g.StatusMap {
		if status == S_ALIVE {
			g.RemainTalkMap[agent] = config.MAX_TALK_COUNT_PER_AGENT
		} else {
			g.RemainTalkMap[agent] = 0
		}
	}
}

func (g GameStatus) ResetRemainWhisperMap() {
	for agent, status := range g.StatusMap {
		if status == S_ALIVE {
			if agent.Role == R_WEREWOLF {
				g.RemainWhisperMap[agent] = config.MAX_WHISPER_COUNT_PER_AGENT
			} else {
				g.RemainWhisperMap[agent] = 0
			}
		} else {
			g.RemainWhisperMap[agent] = 0
		}
	}
}

func (g GameStatus) ConvertToInfo(agent *Agent, settings Settings) Info {
	info := Info{
		Day:   g.Day,
		Agent: *agent,
	}
	if g.MediumResult != (Judge{}) && agent.Role == R_MEDIUM {
		info.MediumResult = g.MediumResult
	}
	if g.DivineResult != (Judge{}) && agent.Role == R_SEER {
		info.DivineResult = g.DivineResult
	}
	if g.ExecutedAgent != (Agent{}) {
		info.ExecutedAgent = g.ExecutedAgent
	}
	if g.AttackedAgent != (Agent{}) {
		info.AttackedAgent = g.AttackedAgent
	}
	if settings.IsVoteVisible {
		info.VoteList = g.VoteList
	}
	if settings.IsVoteVisible && agent.Role == R_WEREWOLF {
		info.AttackVoteList = g.AttackVoteList
	}
	info.TalkList = g.TalkList
	if agent.Role == R_WEREWOLF {
		info.WhisperList = g.WhisperList
	}
	info.StatusMap = g.StatusMap
	roleMap := make(map[Agent]Role)
	roleMap[*agent] = agent.Role
	if agent.Role == R_WEREWOLF {
		for a := range g.StatusMap {
			if a.Role == R_WEREWOLF {
				roleMap[a] = a.Role
			}
		}
	}
	info.RoleMap = roleMap
	info.RemainTalkMap = g.RemainTalkMap
	if agent.Role == R_WEREWOLF {
		info.RemainWhisperMap = g.RemainWhisperMap
	}
	return info
}
