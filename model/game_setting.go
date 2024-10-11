package model

type GameSetting struct {
	RoleNumMap          map[Role]int `json:"role_num_map"`
	MaxTalk             int          `json:"max_talk"`
	MaxTalkTurn         int          `json:"max_talk_turn"`
	MaxWhisper          int          `json:"max_whisper"`
	MaxWhisperTurn      int          `json:"max_whisper_turn"`
	MaxSkip             int          `json:"max_skip"`
	IsEnableNoAttack    bool         `json:"is_enable_no_attack"`
	IsVoteVisible       bool         `json:"is_vote_visible"`
	IsTalkOnFirstDay    bool         `json:"is_talk_on_first_day"`
	ResponseTimeout     int          `json:"response_timeout"`
	ActionTimeout       int          `json:"action_timeout"`
	MaxRevote           int          `json:"max_revote"`
	MaxAttackRevote     int          `json:"max_attack_revote"`
	IsEnableRoleRequest bool         `json:"is_enable_role_request"`
}
