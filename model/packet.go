package model

type Packet struct {
	Request        Request     `json:"request"`
	GameInfo       GameInfo    `json:"game_info"`
	GameSetting    GameSetting `json:"game_setting"`
	TalkHistory    []Talk      `json:"talk_history"`
	WhisperHistory []Talk      `json:"whisper_history"`
}
