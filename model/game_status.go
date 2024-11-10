package model

type GameStatus struct {
	Day              int
	MediumResult     *Judge
	DivineResult     *Judge
	ExecutedAgent    *Agent
	AttackedAgent    *Agent
	Guard            *Guard
	Votes            []Vote
	AttackVotes      []Vote
	Talks            []Talk
	Whispers         []Talk
	StatusMap        map[Agent]Status
	RemainTalkMap    map[Agent]int
	RemainWhisperMap map[Agent]int
}

func NewInitializeGameStatus(agents []*Agent) GameStatus {
	status := GameStatus{
		Day:              0,
		MediumResult:     nil,
		DivineResult:     nil,
		ExecutedAgent:    nil,
		AttackedAgent:    nil,
		Guard:            nil,
		Votes:            []Vote{},
		AttackVotes:      []Vote{},
		Talks:            []Talk{},
		Whispers:         []Talk{},
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
		Votes:            []Vote{},
		AttackVotes:      []Vote{},
		Talks:            []Talk{},
		Whispers:         []Talk{},
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

func (g GameStatus) ClearRemainTalkMap() {
	for agent := range g.RemainTalkMap {
		delete(g.RemainTalkMap, agent)
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

func (g GameStatus) ClearRemainWhisperMap() {
	for agent := range g.RemainWhisperMap {
		delete(g.RemainWhisperMap, agent)
	}
}
