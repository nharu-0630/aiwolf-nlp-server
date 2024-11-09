package model

type Packet struct {
	Request        *Request  `json:"request"`
	Info           *Info     `json:"info,omitempty"`
	Settings       *Settings `json:"setting,omitempty"`
	TalkHistory    *[]Talk   `json:"talkHistory,omitempty"`
	WhisperHistory *[]Talk   `json:"whisperHistory,omitempty"`
}
