package model

type GameInfo struct {
	Day                  int              `json:"day"`
	Agent                Agent            `json:"agent"`
	MediumResult         Judge            `json:"medium_result"`
	DivineResult         Judge            `json:"divine_result"`
	ExecutedAgent        Agent            `json:"executed_agent"`
	LatestExecutedAgent  Agent            `json:"latest_executed_agent"`
	AttackedAgent        Agent            `json:"attacked_agent"`
	CursedFox            Agent            `json:"cursed_fox"`
	GuardedAgent         Agent            `json:"guarded_agent"`
	VoteList             []Vote           `json:"vote_list"`
	LatestVoteList       []Vote           `json:"latest_vote_list"`
	AttackVoteList       []Vote           `json:"attack_vote_list"`
	LatestAttackVoteList []Vote           `json:"latest_attack_vote_list"`
	TalkList             []Talk           `json:"talk_list"`
	WhisperList          []Talk           `json:"whisper_list"`
	StatusMap            map[Agent]Status `json:"status_map"`
	RoleMap              map[Agent]Role   `json:"role_map"`
	RemainTalkMap        map[Agent]int    `json:"remain_talk_map"`
	RemainWhisperMap     map[Agent]int    `json:"remain_whisper_map"`
	ExistingRoleList     []Role           `json:"existing_role_list"`
	LastDeadAgentList    []Agent          `json:"last_dead_agent_list"`
}
